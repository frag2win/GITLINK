package service

import (
	"context"
	"fmt"

	"github.com/localrepo/api-server/internal/models"
	pb "github.com/localrepo/api-server/proto/generated"
	"gorm.io/gorm"
)

type PullRequestService interface {
	CreatePullRequest(ctx context.Context, pr *models.PullRequest) error
	ListPullRequests(ctx context.Context, repoID uint) ([]models.PullRequest, error)
	GetPullRequestByID(ctx context.Context, prID uint) (*models.PullRequest, error)
	MergePullRequest(ctx context.Context, prID uint, repoName string, authorName, authorEmail string) (string, error)
}

type pullRequestService struct {
	db         *gorm.DB
	gitService GitService
}

func NewPullRequestService(db *gorm.DB, gitService GitService) PullRequestService {
	return &pullRequestService{
		db:         db,
		gitService: gitService,
	}
}

func (s *pullRequestService) CreatePullRequest(ctx context.Context, pr *models.PullRequest) error {
	pr.Status = "open"
	return s.db.WithContext(ctx).Create(pr).Error
}

func (s *pullRequestService) ListPullRequests(ctx context.Context, repoID uint) ([]models.PullRequest, error) {
	var prs []models.PullRequest
	// Preload Author for frontend to show who created it
	err := s.db.WithContext(ctx).Where("repository_id = ?", repoID).Preload("Author").Find(&prs).Error
	return prs, err
}

func (s *pullRequestService) GetPullRequestByID(ctx context.Context, prID uint) (*models.PullRequest, error) {
	var pr models.PullRequest
	err := s.db.WithContext(ctx).Preload("Author").First(&pr, prID).Error
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (s *pullRequestService) MergePullRequest(ctx context.Context, prID uint, repoName string, authorName, authorEmail string) (string, error) {
	var pr models.PullRequest
	if err := s.db.WithContext(ctx).First(&pr, prID).Error; err != nil {
		return "", err
	}

	if pr.Status == "merged" {
		return "", fmt.Errorf("pull request is already merged")
	}

	// 1. Call GitService Merge
	req := &pb.MergePullRequest{
		RepoName:       repoName,
		BaseBranch:     pr.BaseBranch,
		HeadBranch:     pr.HeadBranch,
		AuthorName:     authorName,
		AuthorEmail:    authorEmail,
		CommitMessage:  fmt.Sprintf("Merge pull request #%d from %s", pr.ID, pr.HeadBranch),
	}

	commitHash, err := s.gitService.MergePullRequest(ctx, req)
	if err != nil {
		return "", fmt.Errorf("git merge failed: %w", err)
	}

	// 2. Update DB status
	pr.Status = "merged"
	if err := s.db.WithContext(ctx).Save(&pr).Error; err != nil {
		return "", fmt.Errorf("failed to update pull request status: %w", err)
	}

	return commitHash, nil
}
