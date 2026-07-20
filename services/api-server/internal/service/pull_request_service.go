package service

import (
	"context"
	"fmt"

	"github.com/localrepo/api-server/internal/models"
	"github.com/localrepo/api-server/internal/repository"
	pb "github.com/localrepo/api-server/proto/generated"
)

type PullRequestService interface {
	CreatePullRequest(ctx context.Context, pr *models.PullRequest) error
	ListPullRequests(ctx context.Context, repoID uint) ([]models.PullRequest, error)
	GetPullRequestByID(ctx context.Context, prID uint) (*models.PullRequest, error)
	MergePullRequest(ctx context.Context, prID uint, repoName string, authorName, authorEmail string) (string, error)
}

type pullRequestService struct {
	repo       repository.PullRequestRepository
	gitService GitService
}

func NewPullRequestService(repo repository.PullRequestRepository, gitService GitService) PullRequestService {
	return &pullRequestService{
		repo:       repo,
		gitService: gitService,
	}
}

func (s *pullRequestService) CreatePullRequest(ctx context.Context, pr *models.PullRequest) error {
	pr.Status = "open"
	return s.repo.Create(ctx, pr)
}

func (s *pullRequestService) ListPullRequests(ctx context.Context, repoID uint) ([]models.PullRequest, error) {
	return s.repo.ListByRepo(ctx, repoID)
}

func (s *pullRequestService) GetPullRequestByID(ctx context.Context, prID uint) (*models.PullRequest, error) {
	return s.repo.GetByID(ctx, prID)
}

func (s *pullRequestService) MergePullRequest(ctx context.Context, prID uint, repoName string, authorName, authorEmail string) (string, error) {
	pr, err := s.repo.GetByID(ctx, prID)
	if err != nil {
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
	if err := s.repo.Update(ctx, pr); err != nil {
		return "", fmt.Errorf("failed to update pull request status: %w", err)
	}

	return commitHash, nil
}
