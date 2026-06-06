package socket_test

import (
    "context"
    "encoding/binary"
    "net"
    "os"
    "testing"
    "time"

    "google.golang.org/protobuf/proto"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/localrepo/api-server/internal/socket"
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
        Success: true,
        Result: &pb.GitCommandResponse_CreateRepo{
            CreateRepo: &pb.CreateRepoResponse{},
        },
    })

    client := socket.NewGitClient(sockPath, 5*time.Second)
    err := client.CreateRepo(context.Background(), "my-repo")
    assert.NoError(t, err)
}

func TestListRepos_ReturnsNames(t *testing.T) {
    sockPath := "/tmp/test-git-list.sock"

    startMockRustServer(t, sockPath, &pb.GitCommandResponse{
        Success: true,
        Result: &pb.GitCommandResponse_ListRepos{
            ListRepos: &pb.ListReposResponse{
                Repos: []*pb.RepoInfo{
                    {Name: "repo-a"},
                    {Name: "repo-b"},
                },
            },
        },
    })

    client := socket.NewGitClient(sockPath, 5*time.Second)
    names, err := client.ListRepos(context.Background())
    assert.NoError(t, err)
    assert.Equal(t, []string{"repo-a", "repo-b"}, names)
}

func TestSend_ErrorResponse_ReturnsError(t *testing.T) {
    sockPath := "/tmp/test-git-err.sock"

    startMockRustServer(t, sockPath, &pb.GitCommandResponse{
        Success:      false,
        ErrorMessage: "repo already exists",
    })

    client := socket.NewGitClient(sockPath, 5*time.Second)
    err := client.CreateRepo(context.Background(), "existing-repo")
    assert.ErrorContains(t, err, "repo already exists")
}

func TestSend_Timeout_ReturnsError(t *testing.T) {
    // Point at a socket that does not exist
    client := socket.NewGitClient("/tmp/nonexistent.sock", 100*time.Millisecond)
    err := client.CreateRepo(context.Background(), "any-repo")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "dial unix socket")
}
