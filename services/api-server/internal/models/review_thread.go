package models

import "time"

type ReviewThread struct {
	ID            uint                       `gorm:"primaryKey" json:"id"`
	PullRequestID uint                       `gorm:"index;not null" json:"pull_request_id"`
	FilePath      string                     `gorm:"type:varchar(512);not null" json:"file_path"`
	LineNumber    int                        `gorm:"not null" json:"line_number"`
	IsResolved    bool                       `gorm:"default:false" json:"is_resolved"`
	ResolvedByID  *uint                      `json:"resolved_by_id,omitempty"`
	ResolvedBy    *User                      `gorm:"foreignKey:ResolvedByID" json:"resolved_by,omitempty"`
	Comments      []PullRequestReviewComment `gorm:"foreignKey:ThreadID" json:"comments"`
	CreatedAt     time.Time                  `json:"created_at"`
}

type PullRequestReviewComment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ThreadID  uint      `gorm:"index;not null" json:"thread_id"`
	ReviewID  uint      `gorm:"index;not null" json:"review_id"`
	AuthorID  uint      `gorm:"index;not null" json:"author_id"`
	Author    User      `gorm:"foreignKey:AuthorID" json:"author"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
