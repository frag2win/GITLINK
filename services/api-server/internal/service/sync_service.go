package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/localrepo/api-server/internal/models"
	"github.com/localrepo/api-server/internal/repository"
	"github.com/google/uuid"
)

type SyncService interface {
	EnqueueSync(ctx context.Context, repoID uint, repoName, targetPeerID string, priority int, correlationID string) (*models.SyncTask, error)
	CalculateNextBackoff(attemptCount int) time.Duration
	StartWorker(ctx context.Context)
	StopWorker()
	RestartWorker(ctx context.Context)
	RetryTask(ctx context.Context, taskID uint) error
	GetMetrics(ctx context.Context) (map[string]interface{}, error)
}

type syncService struct {
	syncRepo     repository.SyncRepository
	peerService  PeerService
	dispatchChan chan uint
	stopChan     chan struct{}
	wg           sync.WaitGroup
	logger       *slog.Logger
	mu           sync.Mutex
	isStopped    bool
}

func NewSyncService(syncRepo repository.SyncRepository, peerService PeerService, logger *slog.Logger) SyncService {
	if logger == nil {
		logger = slog.Default()
	}
	return &syncService{
		syncRepo:     syncRepo,
		peerService:  peerService,
		dispatchChan: make(chan uint, 100),
		stopChan:     make(chan struct{}),
		logger:       logger,
	}
}

func (s *syncService) CalculateNextBackoff(attemptCount int) time.Duration {
	switch attemptCount {
	case 1:
		return 30 * time.Second
	case 2:
		return 2 * time.Minute
	case 3:
		return 10 * time.Minute
	case 4:
		return 1 * time.Hour
	default:
		return 24 * time.Hour
	}
}

func (s *syncService) EnqueueSync(ctx context.Context, repoID uint, repoName, targetPeerID string, priority int, correlationID string) (*models.SyncTask, error) {
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	task := &models.SyncTask{
		TaskUUID:      uuid.New().String(),
		CorrelationID: correlationID,
		RepositoryID:  repoID,
		RepoName:      repoName,
		TargetPeerID:  targetPeerID,
		Status:        models.SyncTaskPending,
		Priority:      priority,
		AttemptCount:  0,
		MaxAttempts:   5,
	}

	if err := s.syncRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to enqueue sync task: %w", err)
	}

	s.logger.Info("enqueued sync task", "task_uuid", task.TaskUUID, "correlation_id", correlationID, "repo", repoName, "peer", targetPeerID)

	// Non-blocking event-driven dispatch notification
	select {
	case s.dispatchChan <- task.ID:
	default:
		s.logger.Warn("dispatch channel full, recovery poller will pick up task", "task_id", task.ID)
	}

	return task, nil
}

func (s *syncService) StartWorker(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isStopped = false
	s.wg.Add(2)
	go s.eventLoop(ctx)
	go s.recoveryLoop(ctx)
	s.logger.Info("sync service worker started (event-driven + recovery poller active)")
}

func (s *syncService) StopWorker() {
	s.mu.Lock()
	if s.isStopped {
		s.mu.Unlock()
		return
	}
	s.isStopped = true
	close(s.stopChan)
	s.mu.Unlock()

	s.wg.Wait()
	s.logger.Info("sync service worker stopped")
}

func (s *syncService) eventLoop(ctx context.Context) {
	defer s.wg.Done()
	for {
		select {
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		case taskID := <-s.dispatchChan:
			s.processTask(ctx, taskID)
		}
	}
}

func (s *syncService) recoveryLoop(ctx context.Context) {
	defer s.wg.Done()
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Recover stale heartbeat tasks (>120s old)
			recovered, err := s.syncRepo.RecoverStaleTasks(ctx, 120*time.Second)
			if err != nil {
				s.logger.Error("failed to recover stale tasks", "error", err)
			} else if recovered > 0 {
				s.logger.Info("recovered stale in-progress tasks", "count", recovered)
			}

			// Fetch pending/retryable tasks ordered by priority
			tasks, err := s.syncRepo.GetNextTasks(ctx, 10)
			if err != nil {
				s.logger.Error("failed to fetch pending tasks in recovery poller", "error", err)
				continue
			}

			for _, t := range tasks {
				s.processTask(ctx, t.ID)
			}
		}
	}
}

func (s *syncService) processTask(ctx context.Context, taskID uint) {
	// Atomic Claim Task
	claimed, err := s.syncRepo.ClaimTask(ctx, taskID)
	if err != nil || !claimed {
		return // Another worker claimed it or task is no longer pending
	}

	// Fetch fresh task record
	task, err := s.syncRepo.GetByID(ctx, taskID)
	if err != nil || task == nil {
		return
	}

	logger := s.logger.With("task_uuid", task.TaskUUID, "correlation_id", task.CorrelationID, "repo", task.RepoName)
	logger.Info("executing sync task")

	// Call PeerService
	resp, err := s.peerService.DispatchSync(ctx, task)
	if err != nil {
		s.handleFailure(ctx, task, fmt.Sprintf("ipc call error: %v", err))
		return
	}

	if resp.Status == "COMPLETED" || resp.Status == "ALREADY_APPLIED" {
		logger.Info("sync task completed successfully", "status", resp.Status, "duration_ms", resp.DurationMs, "bytes", resp.BytesTransferred)
		_ = s.syncRepo.MarkCompleted(ctx, task.ID, resp.BytesTransferred, resp.DurationMs)
	} else {
		s.handleFailure(ctx, task, resp.Error)
	}
}

func (s *syncService) handleFailure(ctx context.Context, task *models.SyncTask, errStr string) {
	nextAttempt := task.AttemptCount + 1
	if nextAttempt >= task.MaxAttempts {
		s.logger.Error("sync task exceeded max attempts, marking FAILED", "task_uuid", task.TaskUUID, "attempts", nextAttempt, "error", errStr)
		_ = s.syncRepo.MarkFailed(ctx, task.ID, errStr)
		return
	}

	backoff := s.CalculateNextBackoff(nextAttempt)
	nextRetryAt := time.Now().Add(backoff)
	s.logger.Warn("sync task failed, scheduling retry", "task_uuid", task.TaskUUID, "attempt", nextAttempt, "next_retry_at", nextRetryAt, "error", errStr)
	_ = s.syncRepo.MarkRetry(ctx, task.ID, errStr, nextRetryAt)
}

func (s *syncService) RestartWorker(ctx context.Context) {
	s.logger.Info("Graceful worker restart initiated: pausing and draining sync worker")
	s.StopWorker()

	// Reset channels and waitgroup to initial state
	s.muReset()
	s.StartWorker(ctx)
	s.logger.Info("Sync worker successfully restarted and resumed processing")
}

func (s *syncService) muReset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopChan = make(chan struct{})
	s.dispatchChan = make(chan uint, 100)
	s.isStopped = false
}

func (s *syncService) RetryTask(ctx context.Context, taskID uint) error {
	return s.syncRepo.MarkRetry(ctx, taskID, "Manual retry requested via Dashboard", time.Now())
}

func (s *syncService) GetMetrics(ctx context.Context) (map[string]interface{}, error) {
	return s.syncRepo.GetSyncMetrics(ctx)
}
