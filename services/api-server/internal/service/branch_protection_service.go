package service

import (
	"context"
	"errors"

	"github.com/localrepo/api-server/internal/models"
	"github.com/localrepo/api-server/internal/repository"
	"gorm.io/gorm"
)

type BranchProtectionService interface {
	EnableProtection(ctx context.Context, repoID uint, branchName string, requirePR bool) error
	DisableProtection(ctx context.Context, repoID uint, branchName string) error
	GetProtection(ctx context.Context, repoID uint, branchName string) (*models.BranchProtection, error)
}

type branchProtectionService struct {
	repo repository.BranchProtectionRepository
}

func NewBranchProtectionService(repo repository.BranchProtectionRepository) BranchProtectionService {
	return &branchProtectionService{repo: repo}
}

func (s *branchProtectionService) EnableProtection(ctx context.Context, repoID uint, branchName string, requirePR bool) error {
	rule, err := s.repo.GetRule(ctx, repoID, branchName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			rule = &models.BranchProtection{
				RepositoryID: repoID,
				BranchName:   branchName,
				RequirePR:    requirePR,
			}
		} else {
			return err
		}
	} else {
		rule.RequirePR = requirePR
	}
	return s.repo.CreateRule(ctx, rule)
}

func (s *branchProtectionService) DisableProtection(ctx context.Context, repoID uint, branchName string) error {
	return s.repo.DeleteRule(ctx, repoID, branchName)
}

func (s *branchProtectionService) GetProtection(ctx context.Context, repoID uint, branchName string) (*models.BranchProtection, error) {
	return s.repo.GetRule(ctx, repoID, branchName)
}
