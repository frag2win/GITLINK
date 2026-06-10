package protocol

import (
	"bufio"
	"log"
	"net/http"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/localrepo/libp2p-node/internal/socket"
)

// GitProtocolID is the libp2p protocol identifier for Git operations.
const GitProtocolID = protocol.ID("/localrepo/git/1.0.0")

// Handler processes incoming Git protocol streams.
type Handler struct {
	apiClient *socket.APIClient
}

// NewHandler creates a new Git protocol stream handler.
func NewHandler(apiClient *socket.APIClient) *Handler {
	return &Handler{apiClient: apiClient}
}

// HandleStream is the libp2p stream handler callback. It reads an HTTP 
// request from the libp2p stream, forwards it to the local api-server 
// running on port 3000, and streams the HTTP response back over libp2p.
func (h *Handler) HandleStream(s network.Stream) {
	defer s.Close()

	remotePeer := s.Conn().RemotePeer().String()
	log.Printf("git protocol stream opened by peer %s", remotePeer)

	// Read the raw HTTP request from the libp2p stream
	req, err := http.ReadRequest(bufio.NewReader(s))
	if err != nil {
		log.Printf("git protocol: decode HTTP request error from %s: %v", remotePeer, err)
		return
	}

	log.Printf("git protocol: proxying %s %s from peer %s", req.Method, req.URL.Path, remotePeer)

	// Forward to local api-server
	req.URL.Scheme = "http"
	req.URL.Host = "127.0.0.1:3000"
	req.RequestURI = "" // Must be cleared for client requests

	// Optional: Inject the Peer ID into a header for auth
	req.Header.Set("X-Peer-Id", remotePeer)

	// Send request to the api-server
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("git protocol: failed to proxy request to api-server: %v", err)
		return
	}
	defer resp.Body.Close()

	// Write the raw HTTP response back to the libp2p stream
	if err := resp.Write(s); err != nil {
		log.Printf("git protocol: write response error to %s: %v", remotePeer, err)
	}
}
