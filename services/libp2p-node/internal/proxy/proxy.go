package proxy

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/localrepo/libp2p-node/internal/protocol"
	"github.com/localrepo/libp2p-node/internal/queue"
)

// Server is a local HTTP server that proxies requests to remote peers over libp2p.
type Server struct {
	host host.Host
}

// NewServer creates a new local proxy server.
func NewServer(h host.Host) *Server {
	return &Server{host: h}
}

// Start runs the proxy server on the given address.
func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/p2p/", s.handleProxy)

	log.Printf("Starting local git proxy on http://%s", addr)
	return http.ListenAndServe(addr, mux)
}

// handleProxy intercepts requests like:
// GET /p2p/<peer-id>/<repo>/info/refs
// It extracts the peer ID, opens a libp2p stream to that peer, and proxies the HTTP request.
func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
	// Path should be /p2p/<peer-id>/<repo>/...
	parts := strings.SplitN(r.URL.Path, "/", 4)
	if len(parts) < 4 || parts[1] != "p2p" {
		http.Error(w, "invalid path format, expected /p2p/<peer-id>/<repo>/...", http.StatusBadRequest)
		return
	}

	peerIDStr := parts[2]
	repoPath := "/" + parts[3] // e.g. /myrepo/info/refs

	targetPeer, err := peer.Decode(peerIDStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid peer ID: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("git proxy: opening stream to peer %s for %s", peerIDStr, repoPath)

	// Open a libp2p stream to the target peer using our Git protocol ID
	stream, err := s.host.NewStream(r.Context(), targetPeer, protocol.GitProtocolID)
	if err != nil {
		log.Printf("git proxy: failed to open stream to peer %s: %v. Attempting offline queue.", peerIDStr, err)
		queue.HandleOfflinePush(w, r, peerIDStr, repoPath, s.host)
		return
	}
	defer stream.Close()

	// Rewrite the request URL so the remote api-server sees the correct path
	proxyReq := r.Clone(r.Context())
	proxyReq.URL.Path = repoPath
	proxyReq.RequestURI = "" // Must be empty for client requests

	// Write the HTTP request to the libp2p stream
	if err := proxyReq.Write(stream); err != nil {
		log.Printf("git proxy: failed to write request to stream: %v", err)
		http.Error(w, "failed to proxy request", http.StatusInternalServerError)
		return
	}

	// Read the HTTP response from the libp2p stream
	resp, err := http.ReadResponse(bufio.NewReader(stream), proxyReq)
	if err != nil {
		log.Printf("git proxy: failed to read response from stream: %v", err)
		http.Error(w, "failed to read response from peer", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy the response headers back to the local client
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// Stream the response body back to the local client
	if _, err := bufio.NewReader(resp.Body).WriteTo(w); err != nil {
		log.Printf("git proxy: failed to copy response body: %v", err)
	}
}
