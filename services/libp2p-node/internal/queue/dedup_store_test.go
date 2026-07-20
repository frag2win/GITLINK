package queue_test

import (
	"os"
	"testing"

	"github.com/localrepo/libp2p-node/internal/queue"
)

func TestDedupStorePersistenceAndIdempotency(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dedup_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	taskUUID := "123e4567-e89b-12d3-a456-426614174000"

	// 1. First instance
	store1, err := queue.NewDedupStore(tempDir)
	if err != nil {
		t.Fatalf("failed to create DedupStore: %v", err)
	}

	if store1.HasProcessed(taskUUID) {
		t.Errorf("expected HasProcessed to be false initially")
	}

	if err := store1.RecordProcessed(taskUUID, "testrepo", "peer123"); err != nil {
		t.Fatalf("failed to RecordProcessed: %v", err)
	}

	if !store1.HasProcessed(taskUUID) {
		t.Errorf("expected HasProcessed to be true after recording")
	}

	// 2. Simulate node restart: Create brand new instance reading from same tempDir
	store2, err := queue.NewDedupStore(tempDir)
	if err != nil {
		t.Fatalf("failed to re-create DedupStore: %v", err)
	}

	if !store2.HasProcessed(taskUUID) {
		t.Errorf("expected HasProcessed to be true after node restart (persistent disk check failed)")
	}
}
