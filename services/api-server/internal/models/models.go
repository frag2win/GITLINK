package models

import (
	"gorm.io/gorm"
)

// User represents a user account
type User struct {
	gorm.Model
	Username     string `gorm:"uniqueIndex;not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null;default:''"`
	PeerID       string `gorm:"uniqueIndex;not null"` // Link to libp2p identity
	IsAdmin      bool   `gorm:"default:false" json:"is_admin"`
	SSHKeys      []SSHKey
}

// SSHKey represents a user's public SSH key for Git authentication.
type SSHKey struct {
	gorm.Model
	UserID      uint   `gorm:"not null;index"`
	Name        string `gorm:"not null"`
	PublicKey   string `gorm:"type:text;not null"`
	Fingerprint string `gorm:"uniqueIndex;not null"`
}

// RepositoryCollaborator links a User to a Repository with a specific role.
type RepositoryCollaborator struct {
	gorm.Model
	UserID       uint   `gorm:"not null;index"`
	RepositoryID uint   `gorm:"not null;index"`
	Role         string `gorm:"not null"` // "owner", "admin", "write", "read"
}

// Repository represents a git repository managed by the platform
type Repository struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	OwnerID     uint
	Owner       User `gorm:"foreignKey:OwnerID"`
	IsPrivate   bool `gorm:"default:false"`
}

// BranchProtection represents protection rules for a specific branch
type BranchProtection struct {
	gorm.Model
	RepositoryID uint   `gorm:"uniqueIndex:idx_repo_branch"`
	BranchName   string `gorm:"uniqueIndex:idx_repo_branch"`
	RequirePR    bool   `gorm:"default:false"`
}

// PullRequest represents a PR
type PullRequest struct {
	gorm.Model
	RepositoryID uint
	Repository   Repository
	AuthorID     uint
	Author       User
	Title        string `gorm:"not null"`
	Description  string
	BaseBranch   string `gorm:"not null"`
	HeadBranch   string `gorm:"not null"`
	Status       string `gorm:"default:'open'"` // open, merged, closed
}

// AuditLog records an audit event.
type AuditLog struct {
	gorm.Model
	PeerID    string
	Operation string
	RepoName  string
	Details   string
}
