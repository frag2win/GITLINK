package models

import "time"

type ReviewState string

const (
	ReviewStateApproved         ReviewState = "APPROVED"
	ReviewStateChangesRequested ReviewState = "CHANGES_REQUESTED"
	ReviewStateCommented        ReviewState = "COMMENTED"
)

type PullRequestReview struct {
	ID                uint                       `gorm:"primaryKey" json:"id"`
	ReviewUUID        string                     `gorm:"type:uuid;uniqueIndex;not null" json:"review_uuid"`
	PullRequestID     uint                       `gorm:"index;not null" json:"pull_request_id"`
	ReviewerID        uint                       `gorm:"index;not null" json:"reviewer_id"`
	Reviewer          User                       `gorm:"foreignKey:ReviewerID" json:"reviewer"`
	State             ReviewState                `gorm:"type:varchar(32);not null" json:"state"`
	Body              string                     `gorm:"type:text" json:"body"`
	ReviewedCommitSHA string                     `gorm:"type:varchar(64);not null" json:"reviewed_commit_sha"`
	LogicalClock      uint64                     `gorm:"default:1" json:"logical_clock"`
	OriginPeerID      string                     `gorm:"type:varchar(255)" json:"origin_peer_id"`
	Comments          []PullRequestReviewComment `gorm:"foreignKey:ReviewID" json:"comments"`
	CreatedAt         time.Time                  `json:"created_at"`
}
