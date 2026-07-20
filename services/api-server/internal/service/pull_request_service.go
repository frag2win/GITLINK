package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/localrepo/api-server/internal/models"
	"github.com/localrepo/api-server/internal/repository"
	pb "github.com/localrepo/api-server/proto/generated"
)

type PullRequestService interface {
	CreatePullRequest(ctx context.Context, pr *models.PullRequest) error
	ListPullRequests(ctx context.Context, repoID uint) ([]models.PullRequest, error)
	GetPullRequestByID(ctx context.Context, prID uint) (*models.PullRequest, error)
	SubmitReview(ctx context.Context, prID, reviewerID uint, state models.ReviewState, body string, headCommitSHA string, comments []models.PullRequestReviewComment) (*models.PullRequestReview, error)
	GetReviews(ctx context.Context, prID uint) ([]models.PullRequestReview, error)
	ResolveThread(ctx context.Context, threadID, userID uint) error
	MergePullRequest(ctx context.Context, prID uint, repoName string, authorName, authorEmail string, currentHeadSHA string) (string, error)
}

type pullRequestService struct {
	repo         repository.PullRequestRepository
	reviewRepo   repository.PRReviewRepository
	gitService   GitService
	eventBus     EventBus
}

func NewPullRequestService(repo repository.PullRequestRepository, reviewRepo repository.PRReviewRepository, gitService GitService, eventBus EventBus) PullRequestService {
	return &pullRequestService{
		repo:       repo,
		reviewRepo: reviewRepo,
		gitService: gitService,
		eventBus:   eventBus,
	}
}

func (s *pullRequestService) CreatePullRequest(ctx context.Context, pr *models.PullRequest) error {
	pr.Status = "open"
	if err := s.repo.Create(ctx, pr); err != nil {
		return err
	}

	if s.eventBus != nil {
		s.eventBus.Publish(models.DomainEvent{
			Type:      models.NotificationTypePROpened,
			UserID:    pr.AuthorID,
			Title:     "Pull Request Opened",
			Message:   fmt.Sprintf("Pull request '%s' opened", pr.Title),
			Link:      fmt.Sprintf("/repos/%d/pulls/%d", pr.RepositoryID, pr.ID),
			Timestamp: time.Now(),
		})
	}
	return nil
}

func (s *pullRequestService) ListPullRequests(ctx context.Context, repoID uint) ([]models.PullRequest, error) {
	return s.repo.ListByRepo(ctx, repoID)
}

func (s *pullRequestService) GetPullRequestByID(ctx context.Context, prID uint) (*models.PullRequest, error) {
	return s.repo.GetByID(ctx, prID)
}

func (s *pullRequestService) SubmitReview(ctx context.Context, prID, reviewerID uint, state models.ReviewState, body string, headCommitSHA string, comments []models.PullRequestReviewComment) (*models.PullRequestReview, error) {
	pr, err := s.repo.GetByID(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("pull request not found: %w", err)
	}

	review := &models.PullRequestReview{
		ReviewUUID:        uuid.New().String(),
		PullRequestID:     prID,
		ReviewerID:        reviewerID,
		State:             state,
		Body:              body,
		ReviewedCommitSHA: headCommitSHA,
		LogicalClock:      1,
		Comments:          comments,
	}

	if err := s.reviewRepo.CreateReview(ctx, review); err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	if s.eventBus != nil {
		notifType := models.NotificationTypePRReviewed
		if state == models.ReviewStateApproved {
			notifType = models.NotificationTypePRApproved
		}
		s.eventBus.Publish(models.DomainEvent{
			Type:      notifType,
			UserID:    pr.AuthorID,
			Title:     fmt.Sprintf("PR Review: %s", state),
			Message:   fmt.Sprintf("Reviewer submitted %s review on PR #%d", state, prID),
			Link:      fmt.Sprintf("/repos/%d/pulls/%d", pr.RepositoryID, prID),
			Timestamp: time.Now(),
		})
	}

	return review, nil
}

func (s *pullRequestService) GetReviews(ctx context.Context, prID uint) ([]models.PullRequestReview, error) {
	return s.reviewRepo.GetReviewsByPR(ctx, prID)
}

func (s *pullRequestService) ResolveThread(ctx context.Context, threadID, userID uint) error {
	return s.reviewRepo.ResolveThread(ctx, threadID, userID)
}

func (s *pullRequestService) MergePullRequest(ctx context.Context, prID uint, repoName string, authorName, authorEmail string, currentHeadSHA string) (string, error) {
	pr, err := s.repo.GetByID(ctx, prID)
	if err != nil {
		return "", err
	}

	if pr.Status == "merged" {
		return "", fmt.Errorf("pull request is already merged")
	}

	// Stale Approval Validation Check
	latestReview, err := s.reviewRepo.GetLatestReviewByReviewer(ctx, prID, pr.AuthorID)
	if err == nil && latestReview != nil {
		if latestReview.State == models.ReviewStateApproved && currentHeadSHA != "" && latestReview.ReviewedCommitSHA != currentHeadSHA {
			return "", fmt.Errorf("approval is stale: head commit moved from %s to %s. Re-approval required", latestReview.ReviewedCommitSHA, currentHeadSHA)
		}
	}

	req := &pb.MergePullRequest{
		RepoName:      repoName,
		BaseBranch:    pr.BaseBranch,
		HeadBranch:    pr.HeadBranch,
		AuthorName:    authorName,
		AuthorEmail:   authorEmail,
		CommitMessage: fmt.Sprintf("Merge pull request #%d from %s", pr.ID, pr.HeadBranch),
	}

	commitHash, err := s.gitService.MergePullRequest(ctx, req)
	if err != nil {
		return "", fmt.Errorf("git merge failed: %w", err)
	}

	pr.Status = "merged"
	if err := s.repo.Update(ctx, pr); err != nil {
		return "", fmt.Errorf("failed to update pull request status: %w", err)
	}

	if s.eventBus != nil {
		s.eventBus.Publish(models.DomainEvent{
			Type:      models.NotificationTypePRMerged,
			UserID:    pr.AuthorID,
			Title:     "Pull Request Merged",
			Message:   fmt.Sprintf("Pull request #%d was successfully merged", prID),
			Link:      fmt.Sprintf("/repos/%d/pulls/%d", pr.RepositoryID, prID),
			Timestamp: time.Now(),
		})
	}

	return commitHash, nil
}
