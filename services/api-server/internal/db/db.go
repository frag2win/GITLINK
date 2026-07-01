package db

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// User represents a user account
type User struct {
	gorm.Model
	Username     string `gorm:"uniqueIndex;not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null;default:''"`
	PeerID       string `gorm:"uniqueIndex;not null"` // Link to libp2p identity
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

// Init initializes the database connection and auto-migrates schemas
func Init() error {
	dsn := os.Getenv("API_DB_URL")
	if dsn == "" {
		log.Println("API_DB_URL not set, skipping database initialization")
		return nil
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	log.Println("Connected to PostgreSQL successfully")

	// Auto Migrate the schemas
	err = db.AutoMigrate(&User{}, &SSHKey{}, &RepositoryCollaborator{}, &Repository{}, &BranchProtection{}, &PullRequest{})
	if err != nil {
		return err
	}

	log.Println("Database schemas migrated successfully")
	DB = db
	return nil
}
