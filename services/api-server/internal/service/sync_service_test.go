package service_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/localrepo/api-server/internal/ipc"
	"github.com/localrepo/api-server/internal/models"
	"github.com/localrepo/api-server/internal/repository"
	"github.com/localrepo/api-server/internal/service"
	"gorm.io/gorm"
)

func TestExponentialBackoffCalculation(t *testing.T) {
	svc := service.NewSyncService(nil, nil, nil)

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 30 * time.Second},
		{2, 2 * time.Minute},
		{3, 10 * time.Minute},
		{4, 1 * time.Hour},
		{5, 24 * time.Hour},
	}

	for _, tt := range tests {
		got := svc.CalculateNextBackoff(tt.attempt)
		if got != tt.expected {
			t.Errorf("CalculateNextBackoff(%d) = %v; want %v", tt.attempt, got, tt.expected)
		}
	}
}

// mockSyncRepository implements a lightweight in-memory SyncRepository for testing concurrency and worker dispatch
type mockSyncRepository struct {
	repository.SyncRepository
	tasks        map[uint]*models.SyncTask
	mu           sync.Mutex
	processedIDs chan uint // optional: receives task IDs that reach MarkCompleted
}

func (m *mockSyncRepository) Create(ctx context.Context, task *models.SyncTask) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	task.ID = uint(len(m.tasks) + 1)
	m.tasks[task.ID] = task
	return nil
}

func (m *mockSyncRepository) GetByID(ctx context.Context, id uint) (*models.SyncTask, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	task, ok := m.tasks[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return task, nil
}

func (m *mockSyncRepository) ClaimTask(ctx context.Context, id uint) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	task, ok := m.tasks[id]
	if !ok {
		return false, gorm.ErrRecordNotFound
	}
	if task.Status == models.SyncTaskPending || task.Status == models.SyncTaskRetryScheduled {
		task.Status = models.SyncTaskRunning
		return true, nil
	}
	return false, nil
}

func (m *mockSyncRepository) MarkCompleted(ctx context.Context, id uint, bytesTransferred int64, durationMs int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if task, ok := m.tasks[id]; ok {
		task.Status = models.SyncTaskCompleted
	}
	if m.processedIDs != nil {
		select {
		case m.processedIDs <- id:
		default:
		}
	}
	return nil
}

func (m *mockSyncRepository) MarkFailed(ctx context.Context, id uint, errStr string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if task, ok := m.tasks[id]; ok {
		task.Status = models.SyncTaskFailed
	}
	return nil
}

func (m *mockSyncRepository) MarkRetry(ctx context.Context, id uint, errStr string, nextRetry time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if task, ok := m.tasks[id]; ok {
		task.Status = models.SyncTaskRetryScheduled
	}
	return nil
}

func (m *mockSyncRepository) UpdateHeartbeat(ctx context.Context, id uint) error { return nil }

type mockPeerService struct {
	service.PeerService
}

// mockPeerServiceTracking is a PeerService mock that reports successful dispatches
// via processedIDs, allowing TestSyncWorkerDirectTaskLookup to verify which task
// IDs actually reached DispatchSync (i.e., were correctly identified by processTask).
type mockPeerServiceTracking struct {
	service.PeerService
	processedIDs chan uint
}

func (m *mockPeerServiceTracking) DispatchSync(ctx context.Context, task *models.SyncTask) (*ipc.P2PSyncResponse, error) {
	if m.processedIDs != nil {
		select {
		case m.processedIDs <- task.ID:
		default:
		}
	}
	return &ipc.P2PSyncResponse{Status: "COMPLETED", BytesTransferred: 0, DurationMs: 1}, nil
}

func TestSyncWorkerGracefulRestartSafety(t *testing.T) {
	mockRepo := &mockSyncRepository{
		tasks: make(map[uint]*models.SyncTask),
	}
	mockPeers := &mockPeerService{}
	svc := service.NewSyncService(mockRepo, mockPeers, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Start worker
	svc.StartWorker(ctx)

	// 2. Stop worker twice to ensure no double-close channel panics occur
	svc.StopWorker()
	
	// Calling StopWorker second time should return immediately and not panic
	stoppedChan := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("StopWorker panicked on second call: %v", r)
			}
			stoppedChan <- true
		}()
		svc.StopWorker()
	}()

	select {
	case <-stoppedChan:
		// Passed!
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for second StopWorker call to complete")
	}

	// 3. Restart worker concurrently to check race condition safety
	restartWG := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		restartWG.Add(1)
		go func() {
			defer restartWG.Done()
			svc.RestartWorker(ctx)
		}()
	}
	restartWG.Wait()
}

// TestSyncWorkerDirectTaskLookup verifies that when two tasks exist in the queue,
// the sync worker correctly dispatches the specific task it claimed (by ID) rather
// than always re-fetching the newest row.
//
// This is a regression test for the §7 bug where processTask used
// ListTasks(limit=1, offset=0) — which always returned the newest row —
// instead of GetByID(taskID). If that bug were reintroduced, TaskA would never
// be dispatched once TaskB exists (it would always be stuck in "running" after claim).
func TestSyncWorkerDirectTaskLookup(t *testing.T) {
	// processedIDs records which task IDs were actually dispatched to DispatchSync.
	processedIDs := make(chan uint, 10)

	mockRepo := &mockSyncRepository{
		tasks:        make(map[uint]*models.SyncTask),
		processedIDs: processedIDs,
	}
	mockPeers := &mockPeerServiceTracking{processedIDs: processedIDs}

	svc := service.NewSyncService(mockRepo, mockPeers, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc.StartWorker(ctx)
	defer svc.StopWorker()

	// Enqueue Task A first (older, lower ID)
	taskA, err := svc.EnqueueSync(ctx, 1, "repo-a", "peer-1", 5, "corr-a")
	if err != nil {
		t.Fatalf("EnqueueSync taskA: %v", err)
	}

	// Enqueue Task B (newer, higher ID — this is what the old ListTasks bug returned)
	_, err = svc.EnqueueSync(ctx, 2, "repo-b", "peer-2", 5, "corr-b")
	if err != nil {
		t.Fatalf("EnqueueSync taskB: %v", err)
	}

	// Wait for both tasks to be dispatched. With the old bug, TaskA would be claimed
	// then immediately abandoned (task == nil after list mismatch), so only TaskB
	// would ever appear in processedIDs. The correct fix dispatches both.
	seen := make(map[uint]bool)
	deadline := time.After(3 * time.Second)
	for len(seen) < 2 {
		select {
		case id := <-processedIDs:
			seen[id] = true
		case <-deadline:
			t.Fatalf("timeout: only %d of 2 tasks dispatched; seen=%v (regression: processTask may be fetching wrong task)", len(seen), seen)
		}
	}

	// The critical assertion: TaskA (the older, lower-ID task) MUST have been
	// dispatched. The old ListTasks bug would skip it once TaskB existed.
	if !seen[taskA.ID] {
		t.Errorf("TaskA (id=%d, repo-a) was never dispatched — regression: processTask fetched wrong task by ID", taskA.ID)
	}
}

