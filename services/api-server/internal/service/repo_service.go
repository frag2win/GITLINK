package service

import (
	"context"
	"fmt"

	"github.com/localrepo/api-server/internal/models"
	"github.com/localrepo/api-server/internal/repository"
)

type RepoService interface {
	GetRepoByName(ctx context.Context, name string) (*models.Repository, error)
	GetRepoByID(ctx context.Context, id uint) (*models.Repository, error)
	CreateRepository(ctx context.Context, ownerID uint, name, description string) error
	DeleteRepository(ctx context.Context, name string) error
	ListRepositories(ctx context.Context, userID uint) ([]models.Repository, error)
}

type repoService struct {
	repoRepo     repository.RepoRepository
	gitService   GitService
	auditService AuditService
	txManager    repository.TransactionManager
}

// NewRepoService creates a new RepoService containing all business logic for repositories.
func NewRepoService(
	repoRepo repository.RepoRepository,
	gitService GitService,
	auditService AuditService,
	txManager repository.TransactionManager,
) RepoService {
	return &repoService{
		repoRepo:     repoRepo,
		gitService:   gitService,
		auditService: auditService,
		txManager:    txManager,
	}
}

func (s *repoService) GetRepoByName(ctx context.Context, name string) (*models.Repository, error) {
	repo, err := s.repoRepo.FindByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("repository not found: %w", err)
	}
	return repo, nil
}

func (s *repoService) GetRepoByID(ctx context.Context, id uint) (*models.Repository, error) {
	repo, err := s.repoRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("repository not found: %w", err)
	}
	return repo, nil
}

func (s *repoService) CreateRepository(ctx context.Context, ownerID uint, name, description string) error {
	// 1. Git Create (Strong Consistency: do external system first)
	err := s.gitService.CreateRepository(ctx, name)
	if err != nil {
		return fmt.Errorf("git engine repo create failed: %w", err)
	}

	// 2. Database transaction (Insert Repo, Insert Ownership, Audit Log)
	err = s.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		repo := &models.Repository{
			Name:        name,
			Description: description,
			OwnerID:     ownerID,
		}
		if err := s.repoRepo.Create(txCtx, repo); err != nil {
			return fmt.Errorf("db create repo: %w", err)
		}

		// Insert Ownership (Deferred to future auth phase)
		
		// Audit Log
		if err := s.auditService.LogAction(txCtx, "CREATE_REPO", name, "system", "Created new repository"); err != nil {
			return fmt.Errorf("audit log: %w", err)
		}

		return nil
	})

	if err != nil {
		// Compensation: Cleanup Git repository since DB transaction rolled back.
		s.compensateFailedCreate(name)
		return fmt.Errorf("transaction failed, rolled back DB and compensated Git: %w", err)
	}

	return nil
}

func (s *repoService) DeleteRepository(ctx context.Context, name string) error {
	// Check if repository exists
	repo, err := s.repoRepo.FindByName(ctx, name)
	if err != nil {
		return fmt.Errorf("repository not found: %w", err)
	}

	// 1. Database Transaction (Audit + Remove)
	err = s.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := s.auditService.LogAction(txCtx, "DELETE_REPO", name, "system", "Deleted repository"); err != nil {
			return fmt.Errorf("audit log: %w", err)
		}
		if err := s.repoRepo.Delete(txCtx, repo.ID); err != nil {
			return fmt.Errorf("db delete repo: %w", err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	// 2. Git Delete
	if err := s.gitService.DeleteRepository(ctx, name); err != nil {
		// NOTE: DB is deleted but Git is still there (orphaned). 
		// We could add an async cleanup job or dead-letter queue for this.
		return fmt.Errorf("db deleted but git delete failed: %w", err)
	}

	return nil
}

// compensateFailedCreate is an internal helper that handles cleanup
// if the database transaction fails after a Git repository is created.
// This decouples the compensation logic from a standard user-initiated delete.
func (s *repoService) compensateFailedCreate(name string) {
	// For now, we perform a hard delete via the GitService.
	// In the future, this could mark the directory for cleanup or move it to a trash folder.
	_ = s.gitService.DeleteRepository(context.Background(), name)
}

func (s *repoService) ListRepositories(ctx context.Context, userID uint) ([]models.Repository, error) {
	return s.repoRepo.List(ctx, userID)
}
