package e2e_test

import (
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	pb "github.com/localrepo/api-server/proto/generated"
)

const tcpAddr = "127.0.0.1:9099" // Matches git-server's TCP fallback address

func sendRawRequest(t *testing.T, req *pb.GitCommandRequest) *pb.GitCommandResponse {
	conn, err := net.DialTimeout("tcp", tcpAddr, 2*time.Second)
	require.NoError(t, err)
	defer conn.Close()

	reqBytes, err := proto.Marshal(req)
	require.NoError(t, err)

	length := uint32(len(reqBytes))
	err = binary.Write(conn, binary.BigEndian, length)
	require.NoError(t, err)

	_, err = conn.Write(reqBytes)
	require.NoError(t, err)

	var respLength uint32
	err = binary.Read(conn, binary.BigEndian, &respLength)
	require.NoError(t, err, "failed to read response length")

	respBytes := make([]byte, respLength)
	_, err = io.ReadFull(conn, respBytes)
	require.NoError(t, err, "failed to read response body")

	var resp pb.GitCommandResponse
	err = proto.Unmarshal(respBytes, &resp)
	require.NoError(t, err)

	return &resp
}

func TestCompat_UnsupportedProtocolVersion(t *testing.T) {
	req := &pb.GitCommandRequest{
		ProtocolVersion: 999, // Unsupported
		Command: &pb.GitCommandRequest_ListRepos{
			ListRepos: &pb.ListReposRequest{},
		},
	}

	resp := sendRawRequest(t, req)
	assert.NotNil(t, resp.GetError())
	assert.Equal(t, "UnsupportedProtocol", resp.GetError().GetCode())
}
