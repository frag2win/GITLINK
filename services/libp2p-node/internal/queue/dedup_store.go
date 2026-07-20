package queue

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ProcessedSyncRecord struct {
	TaskUUID    string    `json:"task_uuid"`
	RepoName    string    `json:"repo_name"`
	PeerID      string    `json:"peer_id"`
	CompletedAt time.Time `json:"completed_at"`
}

type DedupStore struct {
	mu       sync.RWMutex
	filePath string
	records  map[string]ProcessedSyncRecord
}

func NewDedupStore(dirPath string) (*DedupStore, error) {
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create dedup store directory: %w", err)
	}

	filePath := filepath.Join(dirPath, "processed_syncs.json")
	store := &DedupStore{
		filePath: filePath,
		records:  make(map[string]ProcessedSyncRecord),
	}

	if err := store.load(); err != nil {
		slog.Warn("dedup store: failed to load existing records, starting fresh", "error", err)
	}

	return store, nil
}

func (s *DedupStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var list []ProcessedSyncRecord
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	for _, rec := range list {
		s.records[rec.TaskUUID] = rec
	}
	slog.Info("dedup store: loaded existing records", "count", len(s.records))
	return nil
}

func (s *DedupStore) saveLocked() error {
	list := make([]ProcessedSyncRecord, 0, len(s.records))
	for _, rec := range s.records {
		list = append(list, rec)
	}

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}

	tmpFile := s.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpFile, s.filePath)
}

func (s *DedupStore) HasProcessed(taskUUID string) bool {
	if taskUUID == "" {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.records[taskUUID]
	return exists
}

func (s *DedupStore) RecordProcessed(taskUUID, repoName, peerID string) error {
	if taskUUID == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records[taskUUID] = ProcessedSyncRecord{
		TaskUUID:    taskUUID,
		RepoName:    repoName,
		PeerID:      peerID,
		CompletedAt: time.Now(),
	}

	return s.saveLocked()
}
