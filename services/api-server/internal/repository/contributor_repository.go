package repository

import (
	"context"

	"github.com/localrepo/api-server/internal/models"
	"gorm.io/gorm"
)

type ContributorRepository interface {
	FindRole(ctx context.Context, repoID, userID uint) (string, error)
	AddCollaborator(ctx context.Context, col *models.RepositoryCollaborator) error
	RemoveCollaborator(ctx context.Context, repoID, userID uint) error
	ListCollaborators(ctx context.Context, repoID uint) ([]models.RepositoryCollaborator, error)
}

type contributorRepository struct {
	db *gorm.DB
}

func NewContributorRepository(db *gorm.DB) ContributorRepository {
	return &contributorRepository{db: db}
}

func (r *contributorRepository) FindRole(ctx context.Context, repoID, userID uint) (string, error) {
	var col models.RepositoryCollaborator
	if err := getDB(ctx, r.db).Where("repository_id = ? AND user_id = ?", repoID, userID).First(&col).Error; err != nil {
		return "", err
	}
	return col.Role, nil
}

func (r *contributorRepository) AddCollaborator(ctx context.Context, col *models.RepositoryCollaborator) error {
	return getDB(ctx, r.db).Create(col).Error
}

func (r *contributorRepository) RemoveCollaborator(ctx context.Context, repoID, userID uint) error {
	return getDB(ctx, r.db).Where("repository_id = ? AND user_id = ?", repoID, userID).Delete(&models.RepositoryCollaborator{}).Error
}

func (r *contributorRepository) ListCollaborators(ctx context.Context, repoID uint) ([]models.RepositoryCollaborator, error) {
	var list []models.RepositoryCollaborator
	if err := getDB(ctx, r.db).Where("repository_id = ?", repoID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
