package service_test

import (
	"context"
	"testing"

	"github.com/localrepo/api-server/internal/models"
	"github.com/localrepo/api-server/internal/service"
	pb "github.com/localrepo/api-server/proto/generated"
)

type mockPRRepo struct{}

func (m *mockPRRepo) Create(ctx context.Context, pr *models.PullRequest) error { return nil }
func (m *mockPRRepo) ListByRepo(ctx context.Context, repoID uint) ([]models.PullRequest, error) {
	return []models.PullRequest{}, nil
}
func (m *mockPRRepo) GetByID(ctx context.Context, prID uint) (*models.PullRequest, error) {
	return &models.PullRequest{
		RepositoryID: 1,
		AuthorID:     1,
		Title:        "Test PR",
		BaseBranch:   "main",
		HeadBranch:   "feature",
		Status:       "open",
	}, nil
}
func (m *mockPRRepo) Update(ctx context.Context, pr *models.PullRequest) error { return nil }

type mockReviewRepo struct {
	reviewedCommitSHA string
	reviewState       models.ReviewState
}

func (m *mockReviewRepo) CreateReview(ctx context.Context, review *models.PullRequestReview) error {
	return nil
}
func (m *mockReviewRepo) GetReviewsByPR(ctx context.Context, prID uint) ([]models.PullRequestReview, error) {
	return []models.PullRequestReview{}, nil
}
func (m *mockReviewRepo) GetLatestReviewByReviewer(ctx context.Context, prID, reviewerID uint) (*models.PullRequestReview, error) {
	return &models.PullRequestReview{
		State:             m.reviewState,
		ReviewedCommitSHA: m.reviewedCommitSHA,
	}, nil
}
func (m *mockReviewRepo) CreateThread(ctx context.Context, thread *models.ReviewThread) error {
	return nil
}
func (m *mockReviewRepo) GetThreadsByPR(ctx context.Context, prID uint) ([]models.ReviewThread, error) {
	return []models.ReviewThread{}, nil
}
func (m *mockReviewRepo) ResolveThread(ctx context.Context, threadID, userID uint) error {
	return nil
}
func (m *mockReviewRepo) AddThreadComment(ctx context.Context, comment *models.PullRequestReviewComment) error {
	return nil
}

type mockGitService struct{}

func (m *mockGitService) CreateRepository(ctx context.Context, name string) error { return nil }
func (m *mockGitService) DeleteRepository(ctx context.Context, name string) error { return nil }
func (m *mockGitService) ListRepositories(ctx context.Context) ([]string, error) {
	return []string{}, nil
}
func (m *mockGitService) ListCommits(ctx context.Context, repo, branch string, limit uint32) ([]*pb.CommitInfo, error) {
	return []*pb.CommitInfo{}, nil
}
func (m *mockGitService) GetCommit(ctx context.Context, repo, hash string) (*pb.CommitInfo, error) {
	return nil, nil
}
func (m *mockGitService) GetDiff(ctx context.Context, repo, base, target string) (string, error) {
	return "", nil
}
func (m *mockGitService) GetTree(ctx context.Context, repo, ref, path string) ([]*pb.TreeEntry, error) {
	return []*pb.TreeEntry{}, nil
}
func (m *mockGitService) GetFile(ctx context.Context, repo, ref, path string) ([]byte, error) {
	return nil, nil
}
func (m *mockGitService) InfoRefs(ctx context.Context, repo, gitService string) ([]byte, error) {
	return nil, nil
}
func (m *mockGitService) UploadPack(ctx context.Context, repo string, body []byte) ([]byte, error) {
	return nil, nil
}
func (m *mockGitService) ReceivePack(ctx context.Context, repo string, body []byte) ([]byte, error) {
	return nil, nil
}
func (m *mockGitService) MergePullRequest(ctx context.Context, req *pb.MergePullRequest) (string, error) {
	return "commit123", nil
}
func (m *mockGitService) CreateBranch(ctx context.Context, repo, name, target string) (*pb.BranchInfo, error) {
	return &pb.BranchInfo{Name: name}, nil
}
func (m *mockGitService) DeleteBranch(ctx context.Context, repo, name string) error { return nil }
func (m *mockGitService) ListBranches(ctx context.Context, repo string) ([]*pb.BranchInfo, error) {
	return []*pb.BranchInfo{}, nil
}
func (m *mockGitService) GetBranch(ctx context.Context, repo, name string) (*pb.BranchInfo, error) {
	return &pb.BranchInfo{Name: name}, nil
}
func (m *mockGitService) CreateTag(ctx context.Context, repo, name, target, message, taggerName, taggerEmail string) (*pb.TagInfo, error) {
	return &pb.TagInfo{Name: name}, nil
}
func (m *mockGitService) DeleteTag(ctx context.Context, repo, name string) error { return nil }
func (m *mockGitService) ListTags(ctx context.Context, repo string) ([]*pb.TagInfo, error) {
	return []*pb.TagInfo{}, nil
}

func TestStaleApprovalCommitSHAMismatch(t *testing.T) {
	prRepo := &mockPRRepo{}
	reviewRepo := &mockReviewRepo{
		reviewedCommitSHA: "sha_commit_v1",
		reviewState:       models.ReviewStateApproved,
	}
	gitSvc := &mockGitService{}

	svc := service.NewPullRequestService(prRepo, reviewRepo, gitSvc, nil)

	// Attempt merge when head commit moved to sha_commit_v2
	_, err := svc.MergePullRequest(context.Background(), 1, "testrepo", "user", "user@test.local", "sha_commit_v2")

	if err == nil {
		t.Fatalf("expected error due to stale approval, got nil")
	}

	expectedErr := "approval is stale: head commit moved from sha_commit_v1 to sha_commit_v2. Re-approval required"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestValidApprovalCommitSHAMatch(t *testing.T) {
	prRepo := &mockPRRepo{}
	reviewRepo := &mockReviewRepo{
		reviewedCommitSHA: "sha_commit_v1",
		reviewState:       models.ReviewStateApproved,
	}
	gitSvc := &mockGitService{}

	svc := service.NewPullRequestService(prRepo, reviewRepo, gitSvc, nil)

	commitHash, err := svc.MergePullRequest(context.Background(), 1, "testrepo", "user", "user@test.local", "sha_commit_v1")

	if err != nil {
		t.Fatalf("expected merge success for matching commit SHA, got error: %v", err)
	}

	if commitHash != "commit123" {
		t.Errorf("expected commit hash commit123, got %s", commitHash)
	}
}
