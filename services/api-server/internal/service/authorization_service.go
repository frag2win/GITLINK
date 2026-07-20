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
	AuthorizeRead(ctx context.Context, userID uint, repoID uint) (bool, error)
	AuthorizeWrite(ctx context.Context, userID uint, repoID uint) (bool, error)
}

type authorizationService struct {
	repoRepo         repository.RepoRepository
	contributorRepo  repository.ContributorRepository
	branchProtection repository.BranchProtectionRepository
	teamRepo         repository.TeamRepository
}

func NewAuthorizationService(
	repoRepo repository.RepoRepository,
	contributorRepo repository.ContributorRepository,
	branchProtection repository.BranchProtectionRepository,
	teamRepo repository.TeamRepository,
) AuthorizationService {
	return &authorizationService{
		repoRepo:         repoRepo,
		contributorRepo:  contributorRepo,
		branchProtection: branchProtection,
		teamRepo:         teamRepo,
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

	// 2. Resolve User Role (Direct Collaborator OR Team Permission)
	role := "none"
	if repo.OwnerID == userID {
		role = "owner"
	} else {
		dbRole, err := s.contributorRepo.FindRole(ctx, repo.ID, userID)
		if err == nil {
			role = dbRole
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("authz: fetch collaborator: %w", err)
		}

		// Fallback to Team RBAC if direct collaborator role is not sufficient
		if role == "none" && s.teamRepo != nil {
			teamRole, err := s.teamRepo.GetUserRepoRole(ctx, userID, repo.ID)
			if err == nil && teamRole != "" {
				role = teamRole
			}
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

func (s *authorizationService) AuthorizeRead(ctx context.Context, userID uint, repoID uint) (bool, error) {
	if userID == 0 {
		return false, nil
	}
	repo, err := s.repoRepo.FindByID(ctx, repoID)
	if err != nil {
		return false, err
	}

	// Owners can always read
	if repo.OwnerID == userID {
		return true, nil
	}

	// Public repositories can be read by anyone
	if !repo.IsPrivate {
		return true, nil
	}

	// For private repositories, check collaborator role or team role
	role, err := s.contributorRepo.FindRole(ctx, repoID, userID)
	if err == nil && (role == "owner" || role == "admin" || role == "write" || role == "read") {
		return true, nil
	}

	if s.teamRepo != nil {
		teamRole, err := s.teamRepo.GetUserRepoRole(ctx, userID, repoID)
		if err == nil && (teamRole == "owner" || teamRole == "admin" || teamRole == "write" || teamRole == "read") {
			return true, nil
		}
	}

	return false, nil
}

func (s *authorizationService) AuthorizeWrite(ctx context.Context, userID uint, repoID uint) (bool, error) {
	if userID == 0 {
		return false, nil
	}
	repo, err := s.repoRepo.FindByID(ctx, repoID)
	if err != nil {
		return false, err
	}

	// Owners can always write
	if repo.OwnerID == userID {
		return true, nil
	}

	// Check collaborator role or team role for write permission
	role, err := s.contributorRepo.FindRole(ctx, repoID, userID)
	if err == nil && (role == "owner" || role == "admin" || role == "write") {
		return true, nil
	}

	if s.teamRepo != nil {
		teamRole, err := s.teamRepo.GetUserRepoRole(ctx, userID, repoID)
		if err == nil && (teamRole == "owner" || teamRole == "admin" || teamRole == "write") {
			return true, nil
		}
	}

	return false, nil
}
