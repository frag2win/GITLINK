package ipc_test

import (
	"context"
	"encoding/binary"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/localrepo/api-server/internal/ipc"
	pb "github.com/localrepo/api-server/proto/generated"
)

// startMockRustServer starts a unix socket server that echoes
// back a fixed GitCommandResponse — simulates the Rust git-server.
func startMockRustServer(t *testing.T, sockPath string, resp *pb.GitCommandResponse) {
	t.Helper()
	os.Remove(sockPath)

	ln, err := net.Listen("unix", sockPath)
	require.NoError(t, err)

	t.Cleanup(func() {
		ln.Close()
		os.Remove(sockPath)
	})

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Read length-prefixed request (same framing as tokio LengthDelimitedCodec)
		var length uint32
		binary.Read(conn, binary.BigEndian, &length)
		buf := make([]byte, length)
		conn.Read(buf)

		// Write length-prefixed response
		respBytes, _ := proto.Marshal(resp)
		respLen := uint32(len(respBytes))
		binary.Write(conn, binary.BigEndian, respLen)
		conn.Write(respBytes)
	}()
}

func TestCreateRepo_SendsCorrectFrame(t *testing.T) {
	sockPath := "/tmp/test-git.sock"

	// Mock Rust returns success
	startMockRustServer(t, sockPath, &pb.GitCommandResponse{
		ProtocolVersion: 1,
		Error:           nil,
		Result: &pb.GitCommandResponse_CreateRepo{
			CreateRepo: &pb.CreateRepoResponse{},
		},
	})

	transport := ipc.NewUnixSocketTransport(sockPath, 5*time.Second)
	client := ipc.NewGitClient(transport, 5*time.Second)
	err := client.CreateRepo(context.Background(), "my-repo")
	assert.NoError(t, err)
}

func TestListRepos_ReturnsNames(t *testing.T) {
	sockPath := "/tmp/test-git-list.sock"

	startMockRustServer(t, sockPath, &pb.GitCommandResponse{
		ProtocolVersion: 1,
		Error:           nil,
		Result: &pb.GitCommandResponse_ListRepos{
			ListRepos: &pb.ListReposResponse{
				Repos: []*pb.RepoInfo{
					{Name: "repo-a"},
					{Name: "repo-b"},
				},
			},
		},
	})

	transport := ipc.NewUnixSocketTransport(sockPath, 5*time.Second)
	client := ipc.NewGitClient(transport, 5*time.Second)
	names, err := client.ListRepos(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []string{"repo-a", "repo-b"}, names)
}

func TestSend_ErrorResponse_ReturnsError(t *testing.T) {
	sockPath := "/tmp/test-git-err.sock"

	startMockRustServer(t, sockPath, &pb.GitCommandResponse{
		ProtocolVersion: 1,
		Error: &pb.GitError{
			Code:    "ErrRepositoryExists",
			Message: "repo already exists",
		},
	})

	transport := ipc.NewUnixSocketTransport(sockPath, 5*time.Second)
	client := ipc.NewGitClient(transport, 5*time.Second)
	err := client.CreateRepo(context.Background(), "existing-repo")
	assert.ErrorContains(t, err, "repo already exists")
}

func TestSend_Timeout_ReturnsError(t *testing.T) {
	// Point at a socket that does not exist
	transport := ipc.NewUnixSocketTransport("/tmp/nonexistent.sock", 100*time.Millisecond)
	client := ipc.NewGitClient(transport, 100*time.Millisecond)
	err := client.CreateRepo(context.Background(), "any-repo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dial transport")
}

func TestSend_MergeConflict_ReturnsStructuredError(t *testing.T) {
	sockPath := "/tmp/test-git-merge.sock"

	startMockRustServer(t, sockPath, &pb.GitCommandResponse{
		ProtocolVersion: 1,
		Error:           nil,
		Result: &pb.GitCommandResponse_MergePullRequest{
			MergePullRequest: &pb.MergePullRequestResponse{
				Outcome: &pb.MergePullRequestResponse_Conflicts{
					Conflicts: &pb.MergeConflictList{
						Conflicts: []*pb.MergeConflict{
							{
								Path:         "main.go",
								ConflictType: "BothModified",
							},
						},
					},
				},
			},
		},
	})

	transport := ipc.NewUnixSocketTransport(sockPath, 5*time.Second)
	client := ipc.NewGitClient(transport, 5*time.Second)
	
	_, err := client.MergePullRequest(context.Background(), &pb.MergePullRequest{
		RepoName: "my-repo",
	})
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git_client: merge PR conflicts: 1 files conflicted")
}
