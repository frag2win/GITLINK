package service

import (
	"context"

	"github.com/localrepo/api-server/internal/ipc"
	pb "github.com/localrepo/api-server/proto/generated"
)

// GitService encapsulates all Git-specific orchestration and IPC communication.
// It contains no HTTP, SQL, or Auth logic.
type GitService interface {
	CreateRepository(ctx context.Context, name string) error
	DeleteRepository(ctx context.Context, name string) error
	ListRepositories(ctx context.Context) ([]string, error)
	ListCommits(ctx context.Context, repo, branch string, limit uint32) ([]*pb.CommitInfo, error)
	GetCommit(ctx context.Context, repo, hash string) (*pb.CommitInfo, error)
	GetDiff(ctx context.Context, repo, base, target string) (string, error)
	GetTree(ctx context.Context, repo, ref, path string) ([]*pb.TreeEntry, error)
	GetFile(ctx context.Context, repo, ref, path string) ([]byte, error)
	InfoRefs(ctx context.Context, repo, gitService string) ([]byte, error)
	UploadPack(ctx context.Context, repo string, body []byte) ([]byte, error)
	ReceivePack(ctx context.Context, repo string, body []byte) ([]byte, error)
	MergePullRequest(ctx context.Context, req *pb.MergePullRequest) (string, error)
	CreateBranch(ctx context.Context, repo, name, target string) (*pb.BranchInfo, error)
	DeleteBranch(ctx context.Context, repo, name string) error
	ListBranches(ctx context.Context, repo string) ([]*pb.BranchInfo, error)
	GetBranch(ctx context.Context, repo, name string) (*pb.BranchInfo, error)
	CreateTag(ctx context.Context, repo, name, target, message, taggerName, taggerEmail string) (*pb.TagInfo, error)
	DeleteTag(ctx context.Context, repo, name string) error
	ListTags(ctx context.Context, repo string) ([]*pb.TagInfo, error)
}

type gitService struct {
	client *ipc.GitClient
}

// NewGitService creates a new GitService using the provided GitClient.
func NewGitService(client *ipc.GitClient) GitService {
	return &gitService{client: client}
}

func (s *gitService) CreateRepository(ctx context.Context, name string) error {
	return s.client.CreateRepo(ctx, name)
}

func (s *gitService) DeleteRepository(ctx context.Context, name string) error {
	return s.client.DeleteRepo(ctx, name)
}

func (s *gitService) ListRepositories(ctx context.Context) ([]string, error) {
	return s.client.ListRepos(ctx)
}

func (s *gitService) ListCommits(ctx context.Context, repo, branch string, limit uint32) ([]*pb.CommitInfo, error) {
	return s.client.ListCommits(ctx, repo, branch, limit)
}

func (s *gitService) GetCommit(ctx context.Context, repo, hash string) (*pb.CommitInfo, error) {
	return s.client.GetCommit(ctx, repo, hash)
}

func (s *gitService) GetDiff(ctx context.Context, repo, base, target string) (string, error) {
	return s.client.GetDiff(ctx, repo, base, target)
}

func (s *gitService) GetTree(ctx context.Context, repo, ref, path string) ([]*pb.TreeEntry, error) {
	return s.client.GetTree(ctx, repo, ref, path)
}

func (s *gitService) GetFile(ctx context.Context, repo, ref, path string) ([]byte, error) {
	return s.client.GetFile(ctx, repo, ref, path)
}

func (s *gitService) InfoRefs(ctx context.Context, repo, service string) ([]byte, error) {
	return s.client.InfoRefs(ctx, repo, service)
}

func (s *gitService) UploadPack(ctx context.Context, repo string, body []byte) ([]byte, error) {
	return s.client.UploadPack(ctx, repo, body)
}

func (s *gitService) ReceivePack(ctx context.Context, repo string, body []byte) ([]byte, error) {
	return s.client.ReceivePack(ctx, repo, body)
}

func (s *gitService) MergePullRequest(ctx context.Context, req *pb.MergePullRequest) (string, error) {
	return s.client.MergePullRequest(ctx, req)
}

func (s *gitService) CreateBranch(ctx context.Context, repo, name, target string) (*pb.BranchInfo, error) {
	return s.client.CreateBranch(ctx, repo, name, target)
}

func (s *gitService) DeleteBranch(ctx context.Context, repo, name string) error {
	return s.client.DeleteBranch(ctx, repo, name)
}

func (s *gitService) ListBranches(ctx context.Context, repo string) ([]*pb.BranchInfo, error) {
	return s.client.ListBranches(ctx, repo)
}

func (s *gitService) GetBranch(ctx context.Context, repo, name string) (*pb.BranchInfo, error) {
	return s.client.GetBranch(ctx, repo, name)
}

func (s *gitService) CreateTag(ctx context.Context, repo, name, target, message, taggerName, taggerEmail string) (*pb.TagInfo, error) {
	return s.client.CreateTag(ctx, repo, name, target, message, taggerName, taggerEmail)
}

func (s *gitService) DeleteTag(ctx context.Context, repo, name string) error {
	return s.client.DeleteTag(ctx, repo, name)
}

func (s *gitService) ListTags(ctx context.Context, repo string) ([]*pb.TagInfo, error) {
	return s.client.ListTags(ctx, repo)
}
