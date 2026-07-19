package repository

import (
	"context"

	"github.com/localrepo/api-server/internal/models"
	"gorm.io/gorm"
)

type RepoRepository interface {
	FindByName(ctx context.Context, name string) (*models.Repository, error)
	FindByID(ctx context.Context, id uint) (*models.Repository, error)
	Create(ctx context.Context, repo *models.Repository) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, userID uint) ([]models.Repository, error)
}

type repoRepository struct {
	db *gorm.DB
}

func NewRepoRepository(db *gorm.DB) RepoRepository {
	return &repoRepository{db: db}
}

func (r *repoRepository) FindByName(ctx context.Context, name string) (*models.Repository, error) {
	var repo models.Repository
	if err := getDB(ctx, r.db).Where("name = ?", name).First(&repo).Error; err != nil {
		return nil, err
	}
	return &repo, nil
}

func (r *repoRepository) FindByID(ctx context.Context, id uint) (*models.Repository, error) {
	var repo models.Repository
	if err := getDB(ctx, r.db).First(&repo, id).Error; err != nil {
		return nil, err
	}
	return &repo, nil
}

func (r *repoRepository) Create(ctx context.Context, repo *models.Repository) error {
	return getDB(ctx, r.db).Create(repo).Error
}

func (r *repoRepository) Delete(ctx context.Context, id uint) error {
	return getDB(ctx, r.db).Delete(&models.Repository{}, id).Error
}

func (r *repoRepository) List(ctx context.Context, userID uint) ([]models.Repository, error) {
	var repos []models.Repository
	err := getDB(ctx, r.db).Where("owner_id = ? OR id IN (SELECT repository_id FROM repository_collaborators WHERE user_id = ?)", userID, userID).Find(&repos).Error
	return repos, err
}
