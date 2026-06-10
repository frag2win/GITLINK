package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB wraps the gorm.DB connection.
type DB struct {
	Conn *gorm.DB
}

// User represents a user account
type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null"`
	Email    string `gorm:"uniqueIndex;not null"`
	PeerID   string `gorm:"uniqueIndex;not null"` // Link to libp2p identity
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

// Contributor represents a user who can access a repository
type Contributor struct {
	gorm.Model
	RepositoryID uint
	UserID       uint
	Role         string `gorm:"default:'read'"` // read, write, admin
	User         User
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

// New connects to PostgreSQL using the provided DSN (Data Source Name).
func New(dsn string) (*DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("dsn is empty")
	}

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres: %w", err)
	}

	return &DB{Conn: conn}, nil
}

// Close is a no-op for GORM (connection pooling is handled automatically).
func (db *DB) Close() error {
	sqlDB, err := db.Conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Migrate runs GORM auto-migrations.
func (db *DB) Migrate() error {
	log.Println("Running database migrations...")
	return db.Conn.AutoMigrate(
		&User{},
		&Repository{},
		&Contributor{},
		&BranchProtection{},
		&PullRequest{},
		&AuditLog{},
	)
}

// ---------- Repository CRUD ----------

// CreateRepo inserts a new repository record.
func (db *DB) CreateRepo(name, description, owner string, isPrivate bool) (string, error) {
	// Wait, we need to create or find the User first since we changed to relations
	var user User
	res := db.Conn.FirstOrCreate(&user, User{PeerID: owner, Username: owner, Email: owner + "@p2p.local"})
	if res.Error != nil {
		return "", res.Error
	}

	repo := Repository{
		Name:        name,
		Description: description,
		OwnerID:     user.ID,
		IsPrivate:   isPrivate,
	}

	if err := db.Conn.Create(&repo).Error; err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", repo.ID), nil
}

// GetRepo retrieves a single repo by ID.
func (db *DB) GetRepo(id string) (*Repository, error) {
	var repo Repository
	if err := db.Conn.First(&repo, id).Error; err != nil {
		return nil, err
	}
	return &repo, nil
}

// ListRepos returns all repos.
func (db *DB) ListRepos(peerID string, page, limit int) ([]Repository, error) {
	var repos []Repository
	offset := (page - 1) * limit
	if err := db.Conn.Limit(limit).Offset(offset).Find(&repos).Error; err != nil {
		return nil, err
	}
	return repos, nil
}

// DeleteRepo removes a repo record by ID.
func (db *DB) DeleteRepo(id string) error {
	return db.Conn.Delete(&Repository{}, id).Error
}

// ---------- Contributor CRUD ----------

// AddContributor links a peer to a repository with the given role.
func (db *DB) AddContributor(repoID string, peerID, publicKey, role string) error {
	return fmt.Errorf("AddContributor not implemented")
}

// RemoveContributor unlinks a peer from a repository.
func (db *DB) RemoveContributor(repoID string, peerID string) error {
	return fmt.Errorf("RemoveContributor not implemented")
}

// ---------- Audit Log ----------

// InsertAuditLog records an audit event.
func (db *DB) InsertAuditLog(peerID, operation, repoName, details string) error {
	logEntry := AuditLog{
		PeerID:    peerID,
		Operation: operation,
		RepoName:  repoName,
		Details:   details,
	}
	return db.Conn.Create(&logEntry).Error
}
