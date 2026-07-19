package ipc

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"google.golang.org/protobuf/proto"

	pb "github.com/localrepo/api-server/proto/generated"
)

// GitError represents an error returned from the git-server.
type GitError struct {
	Code    string
	Message string
}

func (e *GitError) Error() string {
	return fmt.Sprintf("git_server_error [%s]: %s", e.Code, e.Message)
}

// ProtocolVersion matches the version expected by the Git Engine.
const ProtocolVersion int32 = 1

// GitClient sends commands to the Rust git-server over an abstracted Transport.
// Each method dials a fresh connection, sends one command, reads one response,
// and closes. Connection pooling is deferred to Phase 3.
type GitClient struct {
	transport Transport
	timeout   time.Duration
}

// NewGitClient creates a GitClient using the given transport.
// timeout is applied to every individual operation.
func NewGitClient(transport Transport, timeout time.Duration) *GitClient {
	return &GitClient{
		transport: transport,
		timeout:   timeout,
	}
}

// send dials the transport, writes a length-prefixed protobuf request,
// reads a length-prefixed protobuf response, and returns it.
// This is the single chokepoint all public methods flow through.
func (c *GitClient) send(
	ctx context.Context,
	req *pb.GitCommandRequest,
) (*pb.GitCommandResponse, error) {

	// Set protocol version
	req.ProtocolVersion = ProtocolVersion

	// Dial transport with context
	conn, err := c.transport.DialContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("git_client: dial transport: %w", err)
	}
	defer conn.Close()

	// Apply deadline to the full read+write cycle
	deadline := time.Now().Add(c.timeout)
	if err := conn.SetDeadline(deadline); err != nil {
		return nil, fmt.Errorf("git_client: set deadline: %w", err)
	}

	// Encode request to protobuf bytes
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("git_client: marshal request: %w", err)
	}

	// Write 4-byte big-endian length prefix
	length := uint32(len(reqBytes))
	if err := binary.Write(conn, binary.BigEndian, length); err != nil {
		return nil, fmt.Errorf("git_client: write length prefix: %w", err)
	}

	// Write protobuf bytes
	if _, err := conn.Write(reqBytes); err != nil {
		return nil, fmt.Errorf("git_client: write request body: %w", err)
	}

	// Read 4-byte big-endian length prefix from response
	var respLength uint32
	if err := binary.Read(conn, binary.BigEndian, &respLength); err != nil {
		return nil, fmt.Errorf("git_client: read response length: %w", err)
	}

	// Guard against absurd response sizes (prevent memory exhaustion)
	const maxResponseSize = 256 * 1024 * 1024 // 256 MB
	if respLength > maxResponseSize {
		return nil, fmt.Errorf("git_client: response size %d exceeds max %d", respLength, maxResponseSize)
	}

	// Read exactly respLength bytes
	respBytes := make([]byte, respLength)
	if _, err := io.ReadFull(conn, respBytes); err != nil {
		return nil, fmt.Errorf("git_client: read response body: %w", err)
	}

	// Decode protobuf response
	var resp pb.GitCommandResponse
	if err := proto.Unmarshal(respBytes, &resp); err != nil {
		return nil, fmt.Errorf("git_client: unmarshal response: %w", err)
	}

	return &resp, nil
}

// ── Public API — one method per command type ─────────────────────────────────

// CreateRepo creates a new bare repository with the given name.
func (c *GitClient) CreateRepo(ctx context.Context, name string) error {
	req := &pb.GitCommandRequest{
		Command: &pb.GitCommandRequest_CreateRepo{
			CreateRepo: &pb.CreateRepoRequest{Name: name},
		},
	}
	resp, err := c.send(ctx, req)
	if err != nil {
		return err
	}
	if errObj := resp.GetError(); errObj != nil {
		return &GitError{Code: errObj.GetCode(), Message: errObj.GetMessage()}
	}
	return nil
}

// DeleteRepo permanently removes a repository and all its data.
func (c *GitClient) DeleteRepo(ctx context.Context, name string) error {
	req := &pb.GitCommandRequest{
		Command: &pb.GitCommandRequest_DeleteRepo{
			DeleteRepo: &pb.DeleteRepoRequest{Name: name},
		},
	}
	resp, err := c.send(ctx, req)
	if err != nil {
		return err
	}
	if errObj := resp.GetError(); errObj != nil {
		return &GitError{Code: errObj.GetCode(), Message: errObj.GetMessage()}
	}
	return nil
}

// ListRepos returns all repository names known to the git-server.
func (c *GitClient) ListRepos(ctx context.Context) ([]string, error) {
	req := &pb.GitCommandRequest{
		Command: &pb.GitCommandRequest_ListRepos{
			ListRepos: &pb.ListReposRequest{},
		},
	}
	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, err
	}
	if errObj := resp.GetError(); errObj != nil {
		return nil, &GitError{Code: errObj.GetCode(), Message: errObj.GetMessage()}
	}

	var names []string
	if listResp := resp.GetListRepos(); listResp != nil {
		for _, r := range listResp.GetRepos() {
			names = append(names, r.Name)
		}
	}
	return names, nil
}

// ListCommits returns commits on the given branch of the given repo.
func (c *GitClient) ListCommits(
	ctx context.Context,
	repo, branch string,
	limit uint32,
) ([]*pb.CommitInfo, error) {
	req := &pb.GitCommandRequest{
		Command: &pb.GitCommandRequest_ListCommits{
			ListCommits: &pb.ListCommitsRequest{
				RepoName: repo,
				Branch:   branch,
				Limit:    int32(limit),
			},
		},
	}
	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, err
	}
	if errObj := resp.GetError(); errObj != nil {
		return nil, &GitError{Code: errObj.GetCode(), Message: errObj.GetMessage()}
	}
	return resp.GetListCommits().GetCommits(), nil
}

// GetCommit returns details of a single commit by its hash.
func (c *GitClient) GetCommit(ctx context.Context, repo, hash string) (*pb.CommitInfo, error) {
	req := &pb.GitCommandRequest{
		Command: &pb.GitCommandRequest_GetCommit{
			GetCommit: &pb.GetCommitRequest{
				RepoName: repo,
				Hash:     hash,
			},
		},
	}
	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, err
	}
	if errObj := resp.GetError(); errObj != nil {
		return nil, &GitError{Code: errObj.GetCode(), Message: errObj.GetMessage()}
	}
	res := resp.GetGetCommit()
	if res == nil || res.Commit == nil {
		return nil, fmt.Errorf("git_client: get commit returned empty commit info")
	}
	return res.Commit, nil
}

// GetDiff returns a unified diff string between base and target hashes/refs.
func (c *GitClient) GetDiff(ctx context.Context, repo, base, target string) (string, error) {
	req := &pb.GitCommandRequest{
		Command: &pb.GitCommandRequest_GetDiff{
			GetDiff: &pb.GetDiffRequest{
				RepoName:   repo,
				BaseHash:   base,
				TargetHash: target,
			},
		},
	}
	resp, err := c.send(ctx, req)
	if err != nil {
		return "", err
	}
	if errObj := resp.GetError(); errObj != nil {
		return "", &GitError{Code: errObj.GetCode(), Message: errObj.GetMessage()}
	}
	res := resp.GetGetDiff()
	if res == nil {
		return "", fmt.Errorf("git_client: get diff returned empty payload")
	}
	return res.DiffText, nil
}

// GetTree returns the file tree for a given repo, ref, and path.
func (c *GitClient) GetTree(
	ctx context.Context,
	repo, ref, path string,
) ([]*pb.TreeEntry, error) {
	req := &pb.GitCommandRequest{
		Command: &pb.GitCommandRequest_GetTree{
			GetTree: &pb.GetTreeRequest{
				RepoName:  repo,
				RefOrHash: ref,
				Path:      path,
			},
		},
	}
	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, err
	}
	if errObj := resp.GetError(); errObj != nil {
		return nil, &GitError{Code: errObj.GetCode(), Message: errObj.GetMessage()}
	}
	return resp.GetGetTree().GetEntries(), nil
}

// GetFile returns the raw content of a file at the given ref and path.
func (c *GitClient) GetFile(
	ctx context.Context,
	repo, ref, path string,
) ([]byte, error) {
	req := &pb.GitCommandRequest{
		Command: &pb.GitCommandRequest_GetFile{
			GetFile: &pb.GetFileRequest{
				RepoName:  repo,
				RefOrHash: ref,
				Path:      path,
			},
		},
	}
	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, err
	}
	if errObj := resp.GetError(); errObj != nil {
		return nil, &GitError{Code: errObj.GetCode(), Message: errObj.GetMessage()}
	}
	return resp.GetGetFile().GetContent(), nil
}

// InfoRefs handles the Git Smart HTTP info/refs discovery phase
func (c *GitClient) InfoRefs(ctx context.Context, repo, service string) ([]byte, error) {
	req := &pb.GitCommandRequest{
		Command: &pb.GitCommandRequest_InfoRefs{
			InfoRefs: &pb.InfoRefsRequest{
				RepoName: repo,
				Service:  service,
			},
		},
	}
	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, err
	}
	if errObj := resp.GetError(); errObj != nil {
		return nil, &GitError{Code: errObj.GetCode(), Message: errObj.GetMessage()}
	}
	return resp.GetInfoRefs().GetOutput(), nil
}

// UploadPack handles the git-upload-pack command for fetching/cloning
func (c *GitClient) UploadPack(ctx context.Context, repo string, body []byte) ([]byte, error) {
	req := &pb.GitCommandRequest{
		Command: &pb.GitCommandRequest_UploadPack{
			UploadPack: &pb.UploadPackRequest{
				RepoName: repo,
				Body:     body,
			},
		},
	}
	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, err
	}
	if errObj := resp.GetError(); errObj != nil {
		return nil, &GitError{Code: errObj.GetCode(), Message: errObj.GetMessage()}
	}
	return resp.GetUploadPack().GetOutput(), nil
}

// ReceivePack handles the git-receive-pack command for pushing
func (c *GitClient) ReceivePack(ctx context.Context, repo string, body []byte) ([]byte, error) {
	req := &pb.GitCommandRequest{
		Command: &pb.GitCommandRequest_ReceivePack{
			ReceivePack: &pb.ReceivePackRequest{
				RepoName: repo,
				Body:     body,
			},
		},
	}
	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, err
	}
	if errObj := resp.GetError(); errObj != nil {
		return nil, &GitError{Code: errObj.GetCode(), Message: errObj.GetMessage()}
	}
	return resp.GetReceivePack().GetOutput(), nil
}

// MergePullRequest triggers the rust backend to perform a 3-way merge
func (c *GitClient) MergePullRequest(ctx context.Context, req *pb.MergePullRequest) (string, error) {
	cmdReq := &pb.GitCommandRequest{
		Command: &pb.GitCommandRequest_MergePullRequest{
			MergePullRequest: req,
		},
	}
	resp, err := c.send(ctx, cmdReq)
	if err != nil {
		return "", err
	}
	if errObj := resp.GetError(); errObj != nil {
		return "", &GitError{Code: errObj.GetCode(), Message: errObj.GetMessage()}
	}

	mergeResp := resp.GetMergePullRequest()
	if mergeResp == nil {
		return "", fmt.Errorf("git_client: merge PR returned empty payload")
	}

	switch outcome := mergeResp.Outcome.(type) {
	case *pb.MergePullRequestResponse_MergeCommitHash:
		return outcome.MergeCommitHash, nil
	case *pb.MergePullRequestResponse_Conflicts:
		return "", fmt.Errorf("git_client: merge PR conflicts: %d files conflicted", len(outcome.Conflicts.GetConflicts()))
	default:
		return "", fmt.Errorf("git_client: unknown merge outcome")
	}
}

func (c *GitClient) CreateBranch(ctx context.Context, repo, name, target string) (*pb.BranchInfo, error) {
	req := &pb.GitCommandRequest{
		Command: &pb.GitCommandRequest_CreateBranch{
			CreateBranch: &pb.CreateBranchRequest{
				RepoName:       repo,
				BranchName:     name,
				TargetCommitId: target,
			},
		},
	}
	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, err
	}
	if errObj := resp.GetError(); errObj != nil {
		return nil, &GitError{Code: errObj.GetCode(), Message: errObj.GetMessage()}
	}
	b := resp.GetCreateBranch()
	if b == nil || b.Branch == nil {
		return nil, fmt.Errorf("git_client: create branch returned empty branch info")
	}
	return b.Branch, nil
}

func (c *GitClient) DeleteBranch(ctx context.Context, repo, name string) error {
	req := &pb.GitCommandRequest{
		Command: &pb.GitCommandRequest_DeleteBranch{
			DeleteBranch: &pb.DeleteBranchRequest{
				RepoName:   repo,
				BranchName: name,
			},
		},
	}
	resp, err := c.send(ctx, req)
	if err != nil {
		return err
	}
	if errObj := resp.GetError(); errObj != nil {
		return &GitError{Code: errObj.GetCode(), Message: errObj.GetMessage()}
	}
	return nil
}

func (c *GitClient) ListBranches(ctx context.Context, repo string) ([]*pb.BranchInfo, error) {
	req := &pb.GitCommandRequest{
		Command: &pb.GitCommandRequest_ListBranches{
			ListBranches: &pb.ListBranchesRequest{
				RepoName: repo,
			},
		},
	}
	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, err
	}
	if errObj := resp.GetError(); errObj != nil {
		return nil, &GitError{Code: errObj.GetCode(), Message: errObj.GetMessage()}
	}
	res := resp.GetListBranches()
	if res == nil {
		return nil, fmt.Errorf("git_client: list branches returned empty payload")
	}
	return res.Branches, nil
}

func (c *GitClient) GetBranch(ctx context.Context, repo, name string) (*pb.BranchInfo, error) {
	req := &pb.GitCommandRequest{
		Command: &pb.GitCommandRequest_GetBranch{
			GetBranch: &pb.GetBranchRequest{
				RepoName:   repo,
				BranchName: name,
			},
		},
	}
	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, err
	}
	if errObj := resp.GetError(); errObj != nil {
		return nil, &GitError{Code: errObj.GetCode(), Message: errObj.GetMessage()}
	}
	res := resp.GetGetBranch()
	if res == nil || res.Branch == nil {
		return nil, fmt.Errorf("git_client: get branch returned empty branch info")
	}
	return res.Branch, nil
}

func (c *GitClient) CreateTag(ctx context.Context, repo, name, target, message, taggerName, taggerEmail string) (*pb.TagInfo, error) {
	req := &pb.GitCommandRequest{
		Command: &pb.GitCommandRequest_CreateTag{
			CreateTag: &pb.CreateTagRequest{
				RepoName:       repo,
				TagName:        name,
				TargetCommitId: target,
				Message:        message,
				TaggerName:     taggerName,
				TaggerEmail:    taggerEmail,
			},
		},
	}
	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, err
	}
	if errObj := resp.GetError(); errObj != nil {
		return nil, &GitError{Code: errObj.GetCode(), Message: errObj.GetMessage()}
	}
	res := resp.GetCreateTag()
	if res == nil || res.Tag == nil {
		return nil, fmt.Errorf("git_client: create tag returned empty tag info")
	}
	return res.Tag, nil
}

func (c *GitClient) DeleteTag(ctx context.Context, repo, name string) error {
	req := &pb.GitCommandRequest{
		Command: &pb.GitCommandRequest_DeleteTag{
			DeleteTag: &pb.DeleteTagRequest{
				RepoName: repo,
				TagName:  name,
			},
		},
	}
	resp, err := c.send(ctx, req)
	if err != nil {
		return err
	}
	if errObj := resp.GetError(); errObj != nil {
		return &GitError{Code: errObj.GetCode(), Message: errObj.GetMessage()}
	}
	return nil
}

func (c *GitClient) ListTags(ctx context.Context, repo string) ([]*pb.TagInfo, error) {
	req := &pb.GitCommandRequest{
		Command: &pb.GitCommandRequest_ListTags{
			ListTags: &pb.ListTagsRequest{
				RepoName: repo,
			},
		},
	}
	resp, err := c.send(ctx, req)
	if err != nil {
		return nil, err
	}
	if errObj := resp.GetError(); errObj != nil {
		return nil, &GitError{Code: errObj.GetCode(), Message: errObj.GetMessage()}
	}
	res := resp.GetListTags()
	if res == nil {
		return nil, fmt.Errorf("git_client: list tags returned empty payload")
	}
	return res.Tags, nil
}
