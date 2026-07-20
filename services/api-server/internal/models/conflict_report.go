package models

import "time"

type ConflictHunk struct {
	StartLine       int    `json:"start_line"`
	EndLine         int    `json:"end_line"`
	Reason          string `json:"reason"`
	GitConflictType string `json:"git_conflict_type"` // e.g., "content", "modify_delete", "binary"
}

type ConflictingFile struct {
	FilePath string         `json:"file_path"`
	Hunks    []ConflictHunk `json:"hunks"`
}

type ConflictReport struct {
	ID              uint              `gorm:"primaryKey" json:"id"`
	RepositoryID    uint              `gorm:"index;not null" json:"repository_id"`
	BaseBranch      string            `gorm:"type:varchar(255);not null" json:"base_branch"`
	HeadBranch      string            `gorm:"type:varchar(255);not null" json:"head_branch"`
	MergeBaseSHA    string            `gorm:"type:varchar(64);not null" json:"merge_base_sha"`
	BaseCommit      string            `gorm:"type:varchar(64);not null" json:"base_commit"`
	HeadCommit      string            `gorm:"type:varchar(64);not null" json:"head_commit"`
	AnalysisVersion string            `gorm:"type:varchar(32);not null;default:'v1.0'" json:"analysis_version"`
	ConflictingFiles []ConflictingFile `gorm:"-" json:"conflicting_files"`
	CreatedAt       time.Time         `json:"created_at"`
}
