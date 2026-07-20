package models

import "time"

type SyncTaskStatus string

const (
	SyncTaskPending        SyncTaskStatus = "PENDING"
	SyncTaskRunning        SyncTaskStatus = "RUNNING"
	SyncTaskRetryScheduled SyncTaskStatus = "RETRY_SCHEDULED"
	SyncTaskCompleted      SyncTaskStatus = "COMPLETED"
	SyncTaskFailed         SyncTaskStatus = "FAILED"
	SyncTaskCancelled      SyncTaskStatus = "CANCELLED"
	SyncTaskExpired        SyncTaskStatus = "EXPIRED"
)

type SyncTask struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	TaskUUID         string         `gorm:"type:uuid;uniqueIndex;not null" json:"task_uuid"`
	CorrelationID    string         `gorm:"type:varchar(64);index;not null" json:"correlation_id"`
	RepositoryID     uint           `gorm:"index;not null" json:"repository_id"`
	RepoName         string         `gorm:"type:varchar(255);not null" json:"repo_name"`
	TargetPeerID     string         `gorm:"type:varchar(255);index;not null" json:"target_peer_id"`
	Status           SyncTaskStatus `gorm:"type:varchar(32);index;default:'PENDING'" json:"status"`
	Priority         int            `gorm:"default:0;index" json:"priority"`
	AttemptCount     int            `gorm:"default:0" json:"attempt_count"`
	MaxAttempts      int            `gorm:"default:5" json:"max_attempts"`
	LastError        string         `gorm:"type:text" json:"last_error,omitempty"`
	BytesTransferred int64          `gorm:"default:0" json:"bytes_transferred"`
	NextRetryAt      *time.Time     `gorm:"index" json:"next_retry_at,omitempty"`
	StartedAt        *time.Time     `json:"started_at,omitempty"`
	CompletedAt      *time.Time     `json:"completed_at,omitempty"`
	LastHeartbeat    *time.Time     `gorm:"index" json:"last_heartbeat,omitempty"`
	DurationMs       int64          `json:"duration_ms"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}
