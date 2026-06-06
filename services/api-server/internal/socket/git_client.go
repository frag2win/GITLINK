package socket

import (
    "context"
    "encoding/binary"
    "fmt"
    "io"
    "net"
    "time"

    "google.golang.org/protobuf/proto"

    pb "github.com/localrepo/api-server/proto/generated"
)

// GitClient sends commands to the Rust git-server over a Unix Domain Socket.
// Each method dials a fresh connection, sends one command, reads one response,
// and closes. Connection pooling is deferred to Phase 3.
type GitClient struct {
    socketPath string
    timeout    time.Duration
}

// NewGitClient creates a GitClient pointed at the given socket path.
// timeout is applied to every individual operation.
func NewGitClient(socketPath string, timeout time.Duration) *GitClient {
    return &GitClient{
        socketPath: socketPath,
        timeout:    timeout,
    }
}

// send dials the unix socket, writes a length-prefixed protobuf request,
// reads a length-prefixed protobuf response, and returns it.
// This is the single chokepoint all public methods flow through.
func (c *GitClient) send(
    ctx context.Context,
    req *pb.GitCommandRequest,
) (*pb.GitCommandResponse, error) {

    // Dial with context deadline
    dialer := net.Dialer{Timeout: c.timeout}
    conn, err := dialer.DialContext(ctx, "unix", c.socketPath)
    if err != nil {
        return nil, fmt.Errorf("git_client: dial unix socket: %w", err)
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

    // Write 4-byte big-endian length prefix (matches tokio LengthDelimitedCodec default)
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
    if !resp.GetSuccess() {
        return fmt.Errorf("git_client: create repo: %s", resp.GetErrorMessage())
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
    if !resp.GetSuccess() {
        return fmt.Errorf("git_client: delete repo: %s", resp.GetErrorMessage())
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
    if !resp.GetSuccess() {
        return nil, fmt.Errorf("git_client: list repos: %s", resp.GetErrorMessage())
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
    if !resp.GetSuccess() {
        return nil, fmt.Errorf("git_client: list commits: %s", resp.GetErrorMessage())
    }
    return resp.GetListCommits().GetCommits(), nil
}

// GetTree returns the file tree for a given repo, ref, and path.
func (c *GitClient) GetTree(
    ctx context.Context,
    repo, ref, path string,
) ([]*pb.TreeEntry, error) {
    req := &pb.GitCommandRequest{
        Command: &pb.GitCommandRequest_GetTree{
            GetTree: &pb.GetTreeRequest{
                RepoName: repo,
                RefOrHash: ref,
                Path:     path,
            },
        },
    }
    resp, err := c.send(ctx, req)
    if err != nil {
        return nil, err
    }
    if !resp.GetSuccess() {
        return nil, fmt.Errorf("git_client: get tree: %s", resp.GetErrorMessage())
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
                RepoName: repo,
                RefOrHash: ref,
                Path:     path,
            },
        },
    }
    resp, err := c.send(ctx, req)
    if err != nil {
        return nil, err
    }
    if !resp.GetSuccess() {
        return nil, fmt.Errorf("git_client: get file: %s", resp.GetErrorMessage())
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
    if !resp.GetSuccess() {
        return nil, fmt.Errorf("git_client: info refs: %s", resp.GetErrorMessage())
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
    if !resp.GetSuccess() {
        return nil, fmt.Errorf("git_client: upload pack: %s", resp.GetErrorMessage())
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
    if !resp.GetSuccess() {
        return nil, fmt.Errorf("git_client: receive pack: %s", resp.GetErrorMessage())
    }
    return resp.GetReceivePack().GetOutput(), nil
}
