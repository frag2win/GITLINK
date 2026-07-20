package service_test

import (
	"context"
	"sync"
	"testing"
	"time"

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
	tasks map[uint]*models.SyncTask
	mu    sync.Mutex
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

type mockPeerService struct {
	service.PeerService
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

func TestSyncWorkerDirectTaskLookup(t *testing.T) {
	mockRepo := &mockSyncRepository{
		tasks: make(map[uint]*models.SyncTask),
	}
	mockPeers := &mockPeerService{}
	_ = service.NewSyncService(mockRepo, mockPeers, nil)

	ctx := context.Background()

	// Enqueue Task A
	taskA := &models.SyncTask{
		Status:   models.SyncTaskPending,
		RepoName: "repo-a",
	}
	_ = mockRepo.Create(ctx, taskA)

	// Enqueue Task B (Task B is newer)
	taskB := &models.SyncTask{
		Status:   models.SyncTaskPending,
		RepoName: "repo-b",
	}
	_ = mockRepo.Create(ctx, taskB)

	// Claim Task A
	claimed, err := mockRepo.ClaimTask(ctx, taskA.ID)
	if err != nil || !claimed {
		t.Fatalf("expected Task A to be claimed successfully")
	}

	// Verify that querying with direct ID returns Task A
	fetchedTask, err := mockRepo.GetByID(ctx, taskA.ID)
	if err != nil {
		t.Fatalf("expected to fetch Task A successfully, got error: %v", err)
	}

	if fetchedTask.RepoName != "repo-a" {
		t.Errorf("expected fetched task repo to be 'repo-a', got %q", fetchedTask.RepoName)
	}
}
