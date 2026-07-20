package repository

import (
	"context"

	"github.com/localrepo/api-server/internal/models"
	"gorm.io/gorm"
)

type PullRequestRepository interface {
	Create(ctx context.Context, pr *models.PullRequest) error
	ListByRepo(ctx context.Context, repoID uint) ([]models.PullRequest, error)
	GetByID(ctx context.Context, prID uint) (*models.PullRequest, error)
	Update(ctx context.Context, pr *models.PullRequest) error
}

type pullRequestRepository struct {
	db *gorm.DB
}

func NewPullRequestRepository(db *gorm.DB) PullRequestRepository {
	return &pullRequestRepository{db: db}
}

func (r *pullRequestRepository) Create(ctx context.Context, pr *models.PullRequest) error {
	return r.db.WithContext(ctx).Create(pr).Error
}

func (r *pullRequestRepository) ListByRepo(ctx context.Context, repoID uint) ([]models.PullRequest, error) {
	var prs []models.PullRequest
	err := r.db.WithContext(ctx).Where("repository_id = ?", repoID).Preload("Author").Find(&prs).Error
	return prs, err
}

func (r *pullRequestRepository) GetByID(ctx context.Context, prID uint) (*models.PullRequest, error) {
	var pr models.PullRequest
	err := r.db.WithContext(ctx).Preload("Author").First(&pr, prID).Error
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *pullRequestRepository) Update(ctx context.Context, pr *models.PullRequest) error {
	return r.db.WithContext(ctx).Save(pr).Error
}
