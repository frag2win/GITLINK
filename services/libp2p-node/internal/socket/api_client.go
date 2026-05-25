// Package socket provides a Unix domain socket client for communicating
// with the api-server sidecar from the libp2p-node.
package socket

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// APIClient sends requests to the api-server over a Unix domain socket.
type APIClient struct {
	socketPath string
	timeout    time.Duration
}

// NewAPIClient creates an APIClient targeting the given socket path.
func NewAPIClient(socketPath string) *APIClient {
	return &APIClient{
		socketPath: socketPath,
		timeout:    15 * time.Second,
	}
}

// Request is the JSON message sent to the api-server socket.
type Request struct {
	// Action is the operation to perform (e.g. "peer-clone", "peer-push").
	Action string `json:"action"`

	// PeerID identifies the remote peer that originated the request.
	PeerID string `json:"peerID"`

	// Params carries action-specific key-value arguments.
	Params map[string]string `json:"params,omitempty"`

	// Payload is optional base64-encoded binary data.
	Payload string `json:"payload,omitempty"`
}

// Response is the JSON message received from the api-server socket.
type Response struct {
	// Success indicates whether the action completed without error.
	Success bool `json:"success"`

	// Error contains a human-readable error message when Success is false.
	Error string `json:"error,omitempty"`

	// Data carries the action-specific result.
	Data interface{} `json:"data,omitempty"`
}

// Send sends a request to the api-server and waits for a response.
func (c *APIClient) Send(req *Request) (*Response, error) {
	conn, err := net.DialTimeout("unix", c.socketPath, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("connect to api-server socket: %w", err)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(c.timeout))

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return nil, fmt.Errorf("write request to api-server: %w", err)
	}

	var resp Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return nil, fmt.Errorf("read response from api-server: %w", err)
	}

	return &resp, nil
}

// ForwardGitRequest is a convenience method that forwards a Git operation
// from a remote peer to the api-server.
func (c *APIClient) ForwardGitRequest(peerID, action, repo, ref, payload string) (*Response, error) {
	return c.Send(&Request{
		Action: action,
		PeerID: peerID,
		Params: map[string]string{
			"repo": repo,
			"ref":  ref,
		},
		Payload: payload,
	})
}
