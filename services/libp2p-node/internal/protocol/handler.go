// Package protocol defines the libp2p stream handler for the Git
// protocol. Remote peers open streams using the GitProtocolID and
// send/receive Git pack data through the api-server socket.
package protocol

import (
	"encoding/json"
	"io"
	"log"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/localrepo/libp2p-node/internal/socket"
)

// GitProtocolID is the libp2p protocol identifier for Git operations.
const GitProtocolID = protocol.ID("/localrepo/git/1.0.0")

// GitRequest is the message format a remote peer sends when initiating
// a Git operation over a libp2p stream.
type GitRequest struct {
	// Action is the Git operation: "clone", "push", "pull", "ls-refs".
	Action string `json:"action"`

	// Repo is the repository name.
	Repo string `json:"repo"`

	// Ref is an optional branch/tag reference.
	Ref string `json:"ref,omitempty"`

	// Payload carries base64-encoded pack data for push operations.
	Payload string `json:"payload,omitempty"`
}

// GitResponse is the message sent back to the remote peer.
type GitResponse struct {
	// Success indicates whether the operation completed without error.
	Success bool `json:"success"`

	// Error is a human-readable error message when Success is false.
	Error string `json:"error,omitempty"`

	// Payload carries base64-encoded pack data for clone/pull responses.
	Payload string `json:"payload,omitempty"`
}

// Handler processes incoming Git protocol streams.
type Handler struct {
	apiClient *socket.APIClient
}

// NewHandler creates a new Git protocol stream handler.
func NewHandler(apiClient *socket.APIClient) *Handler {
	return &Handler{apiClient: apiClient}
}

// HandleStream is the libp2p stream handler callback. It reads a
// GitRequest from the stream, forwards it to the api-server via
// the Unix socket, and writes the response back to the remote peer.
func (h *Handler) HandleStream(s network.Stream) {
	defer s.Close()

	remotePeer := s.Conn().RemotePeer().String()
	log.Printf("git protocol stream opened by peer %s", remotePeer)

	// Read the request from the stream
	var req GitRequest
	if err := json.NewDecoder(s).Decode(&req); err != nil {
		if err != io.EOF {
			log.Printf("git protocol: decode error from %s: %v", remotePeer, err)
		}
		writeError(s, "invalid request format")
		return
	}

	log.Printf("git protocol: peer=%s action=%s repo=%s", remotePeer, req.Action, req.Repo)

	// TODO: Forward the request to api-server via Unix socket.
	// TODO: Map GitRequest fields to socket.Request.
	// TODO: Include remotePeer as the PeerID for access control.
	// TODO: Read the socket response and convert to GitResponse.
	// TODO: Write the GitResponse to the libp2p stream.

	resp := &GitResponse{
		Success: false,
		Error:   "git protocol handler not fully implemented",
	}

	if err := json.NewEncoder(s).Encode(resp); err != nil {
		log.Printf("git protocol: encode error to %s: %v", remotePeer, err)
	}
}

// writeError sends a simple error response on the stream.
func writeError(s network.Stream, msg string) {
	resp := &GitResponse{Success: false, Error: msg}
	_ = json.NewEncoder(s).Encode(resp)
}
