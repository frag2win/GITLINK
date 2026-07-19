package repository

import (
	"context"

	"github.com/localrepo/api-server/internal/models"
	"gorm.io/gorm"
)

type BranchProtectionRepository interface {
	GetRule(ctx context.Context, repoID uint, branchName string) (*models.BranchProtection, error)
	CreateRule(ctx context.Context, rule *models.BranchProtection) error
	DeleteRule(ctx context.Context, repoID uint, branchName string) error
}

type branchProtectionRepository struct {
	db *gorm.DB
}

func NewBranchProtectionRepository(db *gorm.DB) BranchProtectionRepository {
	return &branchProtectionRepository{db: db}
}

func (r *branchProtectionRepository) GetRule(ctx context.Context, repoID uint, branchName string) (*models.BranchProtection, error) {
	var rule models.BranchProtection
	if err := getDB(ctx, r.db).Where("repository_id = ? AND branch_name = ?", repoID, branchName).First(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *branchProtectionRepository) CreateRule(ctx context.Context, rule *models.BranchProtection) error {
	// Use Save to perform insert or update (upsert)
	return getDB(ctx, r.db).Save(rule).Error
}

func (r *branchProtectionRepository) DeleteRule(ctx context.Context, repoID uint, branchName string) error {
	return getDB(ctx, r.db).Where("repository_id = ? AND branch_name = ?", repoID, branchName).Delete(&models.BranchProtection{}).Error
}
