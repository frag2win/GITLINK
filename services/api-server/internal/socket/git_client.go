// Package socket provides Unix domain socket clients for inter-process
// communication between the api-server and its sidecar containers.
package socket

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// GitClient communicates with the git-server sidecar over a Unix domain socket.
type GitClient struct {
	socketPath string
	timeout    time.Duration
}

// NewGitClient creates a new GitClient targeting the given socket path.
func NewGitClient(socketPath string) *GitClient {
	return &GitClient{
		socketPath: socketPath,
		timeout:    10 * time.Second,
	}
}

// Send sends a request to the git-server and waits for a response.
func (gc *GitClient) Send(req *Request) (*Response, error) {
	// TODO: Dial the Unix domain socket.
	// TODO: Set read/write deadlines based on gc.timeout.
	// TODO: Marshal the Request to JSON and write to the socket.
	// TODO: Read the response from the socket.
	// TODO: Unmarshal into a Response struct.
	// TODO: Close the connection.

	conn, err := net.DialTimeout("unix", gc.socketPath, gc.timeout)
	if err != nil {
		return nil, fmt.Errorf("connect to git-server: %w", err)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(gc.timeout))

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}

	var resp Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return &resp, nil
}

// InitRepo asks the git-server to initialise a new bare repository.
func (gc *GitClient) InitRepo(name string) (*Response, error) {
	return gc.Send(&Request{
		Action: ActionInitRepo,
		Params: map[string]string{"name": name},
	})
}

// DeleteRepo asks the git-server to remove a repository.
func (gc *GitClient) DeleteRepo(name string) (*Response, error) {
	return gc.Send(&Request{
		Action: ActionDeleteRepo,
		Params: map[string]string{"name": name},
	})
}

// ListBranches asks the git-server for all branches in a repo.
func (gc *GitClient) ListBranches(repoName string) (*Response, error) {
	return gc.Send(&Request{
		Action: ActionListBranches,
		Params: map[string]string{"repo": repoName},
	})
}

// Log asks the git-server for commit history.
func (gc *GitClient) Log(repoName, branch string, limit int) (*Response, error) {
	return gc.Send(&Request{
		Action: ActionLog,
		Params: map[string]string{
			"repo":   repoName,
			"branch": branch,
			"limit":  fmt.Sprintf("%d", limit),
		},
	})
}

// LsTree asks the git-server for a directory listing at the given ref/path.
func (gc *GitClient) LsTree(repoName, ref, path string) (*Response, error) {
	return gc.Send(&Request{
		Action: ActionLsTree,
		Params: map[string]string{
			"repo": repoName,
			"ref":  ref,
			"path": path,
		},
	})
}

// CatFile asks the git-server for the content of a file at the given ref/path.
func (gc *GitClient) CatFile(repoName, ref, path string) (*Response, error) {
	return gc.Send(&Request{
		Action: ActionCatFile,
		Params: map[string]string{
			"repo": repoName,
			"ref":  ref,
			"path": path,
		},
	})
}
