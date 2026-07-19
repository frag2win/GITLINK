package ipc

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
)

// PushAuthorizer is implemented by services checking write permissions.
type PushAuthorizer interface {
	AuthorizePush(ctx context.Context, userID uint, repoName string, branchName string) (bool, string, error)
}

// PushAuthorizerFunc is an adapter to allow the use of ordinary functions as PushAuthorizers.
type PushAuthorizerFunc func(ctx context.Context, userID uint, repoName string, branchName string) (bool, string, error)

func (f PushAuthorizerFunc) AuthorizePush(ctx context.Context, userID uint, repoName string, branchName string) (bool, string, error) {
	return f(ctx, userID, repoName, branchName)
}

// AuthSocketServer listens on a Unix Domain Socket to provide an internal
// authorization endpoint for Git hooks (e.g., pre-receive).
type AuthSocketServer struct {
	socketPath string
	logger     *slog.Logger
	authorizer PushAuthorizer
}

func NewAuthSocketServer(socketPath string, authorizer PushAuthorizer, logger *slog.Logger) *AuthSocketServer {
	return &AuthSocketServer{
		socketPath: socketPath,
		logger:     logger,
		authorizer: authorizer,
	}
}

func (s *AuthSocketServer) Start() error {
	// Clean up old socket if it exists
	os.Remove(s.socketPath)

	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return err
	}

	// Ensure the socket is accessible
	if err := os.Chmod(s.socketPath, 0666); err != nil {
		s.logger.Warn("Failed to chmod auth socket", "error", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/hooks/pre-receive", s.handlePreReceive)

	s.logger.Info("Auth socket server listening", "path", s.socketPath)
	return http.Serve(listener, mux)
}

func (s *AuthSocketServer) handlePreReceive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Repo   string `json:"repo"`
		Branch string `json:"branch"`
		OldRev string `json:"oldrev"`
		NewRev string `json:"newrev"`
		UserID string `json:"user_id"` // Passed as env variable from SSH wrapper
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	s.logger.Info("Received pre-receive hook authorization request",
		"repo", req.Repo,
		"branch", req.Branch,
		"user_id", req.UserID,
	)

	// Parse user ID
	var userID uint
	if req.UserID != "" {
		if id, err := strconv.ParseUint(req.UserID, 10, 32); err == nil {
			userID = uint(id)
		}
	}

	// Authorize
	allowed, reason, err := s.authorizer.AuthorizePush(r.Context(), userID, req.Repo, req.Branch)
	if err != nil {
		s.logger.Error("Auth hook error", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if !allowed {
		s.logger.Warn("Push rejected", "repo", req.Repo, "branch", req.Branch, "reason", reason)
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{"allowed": false, "reason": reason})
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"allowed": true})
}
