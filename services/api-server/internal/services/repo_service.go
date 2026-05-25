// Package services contains the business logic layer for the api-server.
// Each service orchestrates between handlers, database, and socket clients.
package services

import (
	"fmt"

	"github.com/localrepo/api-server/internal/database"
	"github.com/localrepo/api-server/internal/models"
	"github.com/localrepo/api-server/internal/socket"
)

// RepoService handles business logic for repository operations.
// It coordinates between the database (metadata) and the git-server
// socket (actual Git operations on disk).
type RepoService struct {
	db        *database.DB
	gitClient *socket.GitClient
	audit     *AuditService
}

// NewRepoService creates a RepoService with the given dependencies.
func NewRepoService(db *database.DB, gitClient *socket.GitClient, audit *AuditService) *RepoService {
	return &RepoService{
		db:        db,
		gitClient: gitClient,
		audit:     audit,
	}
}

// Create initialises a new repository both in the database and on disk.
func (s *RepoService) Create(peerID string, req models.CreateRepoRequest) (*models.Repo, error) {
	// TODO: Validate the request fields (name uniqueness, etc.).
	// TODO: Insert the repo record into the database.
	// TODO: Send InitRepo command to git-server via socket.
	// TODO: If git-server fails, rollback the database insert.
	// TODO: Log an audit event via s.audit.
	// TODO: Return the created Repo model.

	return nil, fmt.Errorf("RepoService.Create not implemented")
}

// Get retrieves a single repository by ID, verifying that the given
// peer has read access.
func (s *RepoService) Get(peerID, repoID string) (*models.Repo, error) {
	// TODO: Fetch repo from database.
	// TODO: Check access: repo must be public OR peer must be owner/contributor.
	// TODO: Return the Repo model.

	return nil, fmt.Errorf("RepoService.Get not implemented")
}

// List returns all repositories visible to the given peer.
func (s *RepoService) List(peerID string, page, limit int) ([]models.Repo, error) {
	// TODO: Query database for repos the peer can access.
	// TODO: Apply pagination.

	return nil, fmt.Errorf("RepoService.List not implemented")
}

// Delete removes a repository, both the metadata and the on-disk data.
func (s *RepoService) Delete(peerID, repoID string) error {
	// TODO: Verify the peer is the repository owner.
	// TODO: Send DeleteRepo command to git-server via socket.
	// TODO: Remove the repo record from the database.
	// TODO: Log an audit event.

	return fmt.Errorf("RepoService.Delete not implemented")
}
