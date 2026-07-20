package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type P2PClient struct {
	socketPath string
	timeout    time.Duration
}

func NewP2PClient(socketPath string, timeout time.Duration) *P2PClient {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &P2PClient{
		socketPath: socketPath,
		timeout:    timeout,
	}
}

type P2PSyncRequest struct {
	Command       string `json:"command"`
	TaskUUID      string `json:"task_uuid"`
	RepoName      string `json:"repo_name"`
	TargetPeerID  string `json:"target_peer_id"`
	Priority      int    `json:"priority"`
	CorrelationID string `json:"correlation_id"`
}

type P2PSyncResponse struct {
	TaskUUID         string `json:"task_uuid"`
	RepoName         string `json:"repo_name"`
	TargetPeerID     string `json:"target_peer_id"`
	Status           string `json:"status"` // COMPLETED, ALREADY_APPLIED, FAILED
	BytesTransferred int64  `json:"bytes_transferred"`
	DurationMs       int64  `json:"duration_ms"`
	Error            string `json:"error,omitempty"`
}

type PeerInfo struct {
	ID        string   `json:"id"`
	Addrs     []string `json:"addrs"`
	LatencyMs int64    `json:"latency_ms"`
}

func (c *P2PClient) TriggerSync(ctx context.Context, req *P2PSyncRequest) (*P2PSyncResponse, error) {
	conn, err := net.DialTimeout("unix", c.socketPath, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("dial p2p socket %s: %w", c.socketPath, err)
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Now().Add(c.timeout))
	}

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return nil, fmt.Errorf("write sync request to p2p node: %w", err)
	}

	var resp P2PSyncResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return nil, fmt.Errorf("read sync response from p2p node: %w", err)
	}

	return &resp, nil
}

func (c *P2PClient) GetPeers(ctx context.Context) ([]PeerInfo, error) {
	conn, err := net.DialTimeout("unix", c.socketPath, c.timeout)
	if err != nil {
		// If socket is unreachable, return empty list gracefully
		return []PeerInfo{}, nil
	}
	defer conn.Close()

	req := map[string]string{"command": "LIST_PEERS"}
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return nil, err
	}

	var peers []PeerInfo
	if err := json.NewDecoder(conn).Decode(&peers); err != nil {
		return []PeerInfo{}, nil
	}
	return peers, nil
}
