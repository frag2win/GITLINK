package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/localrepo/api-server/internal/repository"
	"gorm.io/gorm"
)

type AuthorizationResult struct {
	Allowed         bool
	Reason          string
	Role            string
	ProtectedBranch bool
	Policy          string
}

type AuthorizationService interface {
	AuthorizePush(ctx context.Context, userID uint, repoName string, branchName string) (*AuthorizationResult, error)
}

type authorizationService struct {
	repoRepo         repository.RepoRepository
	contributorRepo  repository.ContributorRepository
	branchProtection repository.BranchProtectionRepository
}

func NewAuthorizationService(
	repoRepo repository.RepoRepository,
	contributorRepo repository.ContributorRepository,
	branchProtection repository.BranchProtectionRepository,
) AuthorizationService {
	return &authorizationService{
		repoRepo:         repoRepo,
		contributorRepo:  contributorRepo,
		branchProtection: branchProtection,
	}
}

func (s *authorizationService) AuthorizePush(ctx context.Context, userID uint, repoName string, branchName string) (*AuthorizationResult, error) {
	if userID == 0 {
		return &AuthorizationResult{
			Allowed: false,
			Reason:  "Authentication required",
			Role:    "none",
		}, nil
	}

	// 1. Fetch Repository
	repo, err := s.repoRepo.FindByName(ctx, repoName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AuthorizationResult{
				Allowed: false,
				Reason:  "Repository not found",
				Role:    "none",
			}, nil
		}
		return nil, fmt.Errorf("authz: fetch repo: %w", err)
	}

	// 2. Resolve User Role
	role := "none"
	if repo.OwnerID == userID {
		role = "owner"
	} else {
		dbRole, err := s.contributorRepo.FindRole(ctx, repo.ID, userID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("authz: fetch collaborator: %w", err)
			}
		} else {
			role = dbRole
		}
	}

	// Verify write permission
	if role != "owner" && role != "admin" && role != "write" {
		return &AuthorizationResult{
			Allowed: false,
			Reason:  "Write access denied",
			Role:    role,
		}, nil
	}

	// 3. Check Branch Protection Rules
	rule, err := s.branchProtection.GetRule(ctx, repo.ID, branchName)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("authz: fetch branch protection: %w", err)
		}
	} else if rule.RequirePR {
		return &AuthorizationResult{
			Allowed:         false,
			Reason:          fmt.Sprintf("Branch '%s' is protected. Direct pushes are disabled. Please create a Pull Request.", branchName),
			Role:            role,
			ProtectedBranch: true,
			Policy:          "RequirePR",
		}, nil
	}

	return &AuthorizationResult{
		Allowed: true,
		Reason:  "Authorized successfully",
		Role:    role,
	}, nil
}
