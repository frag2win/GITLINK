package models

import "time"

type NotificationType string

const (
	NotificationTypePROpened          NotificationType = "PR_OPENED"
	NotificationTypePRReviewed        NotificationType = "PR_REVIEWED"
	NotificationTypePRApproved        NotificationType = "PR_APPROVED"
	NotificationTypePRMerged          NotificationType = "PR_MERGED"
	NotificationTypeCommentAdded      NotificationType = "COMMENT_ADDED"
	NotificationTypeSyncCompleted     NotificationType = "SYNC_COMPLETED"
	NotificationTypeRepositoryInvited NotificationType = "REPOSITORY_INVITED"
)

type Notification struct {
	ID        uint             `gorm:"primaryKey" json:"id"`
	UserID    uint             `gorm:"index;not null" json:"user_id"`
	Type      NotificationType `gorm:"type:varchar(64);not null" json:"type"`
	Title     string           `gorm:"type:varchar(255);not null" json:"title"`
	Message   string           `gorm:"type:text" json:"message"`
	Link      string           `gorm:"type:varchar(512)" json:"link"`
	IsRead    bool             `gorm:"default:false;index" json:"is_read"`
	CreatedAt time.Time        `json:"created_at"`
}

type DomainEvent struct {
	Type      NotificationType       `json:"type"`
	UserID    uint                   `json:"user_id"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Link      string                 `json:"link"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}
