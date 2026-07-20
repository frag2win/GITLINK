package queue

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cgi"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

var (
	queueDir   = "/app/queue"
	dedupStore *DedupStore
	mu         sync.Mutex
)

// Init initializes the queue directory and persistent deduplication store.
func Init(queuePath string) {
	if queuePath != "" {
		queueDir = queuePath
	}
	if err := os.MkdirAll(queueDir, 0755); err != nil {
		slog.Error("queue: failed to create queue directory", "error", err)
	}

	ds, err := NewDedupStore(queueDir)
	if err != nil {
		slog.Error("queue: failed to initialize dedup store", "error", err)
	} else {
		dedupStore = ds
	}
}

// SyncResult carries execution metrics back to caller
type SyncResult struct {
	TaskUUID         string `json:"task_uuid"`
	RepoName         string `json:"repo_name"`
	TargetPeerID     string `json:"target_peer_id"`
	Status           string `json:"status"` // COMPLETED, ALREADY_APPLIED, FAILED
	BytesTransferred int64  `json:"bytes_transferred"`
	DurationMs       int64  `json:"duration_ms"`
	Error            string `json:"error,omitempty"`
}

// ExecuteSync executes an atomic sync attempt using DedupStore guard.
func ExecuteSync(ctx context.Context, h host.Host, taskUUID, repoName, peerIDStr string, correlationID string) *SyncResult {
	startTime := time.Now()
	logger := slog.With("task_uuid", taskUUID, "correlation_id", correlationID, "repo", repoName, "peer", peerIDStr)

	if dedupStore != nil && dedupStore.HasProcessed(taskUUID) {
		logger.Info("queue: task already processed, returning ALREADY_APPLIED")
		return &SyncResult{
			TaskUUID:     taskUUID,
			RepoName:     repoName,
			TargetPeerID: peerIDStr,
			Status:       "ALREADY_APPLIED",
			DurationMs:   time.Since(startTime).Milliseconds(),
		}
	}

	targetPeer, err := peer.Decode(peerIDStr)
	if err != nil {
		logger.Error("queue: invalid peer ID", "error", err)
		return &SyncResult{
			TaskUUID:     taskUUID,
			RepoName:     repoName,
			TargetPeerID: peerIDStr,
			Status:       "FAILED",
			Error:        fmt.Sprintf("invalid peer ID: %v", err),
			DurationMs:   time.Since(startTime).Milliseconds(),
		}
	}

	// Try to connect to peer
	if err := h.Connect(ctx, peer.AddrInfo{ID: targetPeer}); err != nil {
		logger.Warn("queue: target peer offline or unreachable", "error", err)
		return &SyncResult{
			TaskUUID:     taskUUID,
			RepoName:     repoName,
			TargetPeerID: peerIDStr,
			Status:       "FAILED",
			Error:        fmt.Sprintf("peer connection failed: %v", err),
			DurationMs:   time.Since(startTime).Milliseconds(),
		}
	}

	localRepoPath := filepath.Join(queueDir, peerIDStr, repoName+".git")
	proxyPort := os.Getenv("PROXY_PORT")
	if proxyPort == "" {
		proxyPort = "4000"
	}

	remoteURL := fmt.Sprintf("http://127.0.0.1:%s/p2p/%s/%s", proxyPort, peerIDStr, repoName)
	cmd := exec.CommandContext(ctx, "git", "push", "--all", remoteURL)
	cmd.Dir = localRepoPath

	out, err := cmd.CombinedOutput()
	durationMs := time.Since(startTime).Milliseconds()

	if err != nil {
		logger.Error("queue: git push execution failed", "output", string(out), "error", err)
		return &SyncResult{
			TaskUUID:     taskUUID,
			RepoName:     repoName,
			TargetPeerID: peerIDStr,
			Status:       "FAILED",
			Error:        fmt.Sprintf("git push failed: %s", string(out)),
			DurationMs:   durationMs,
		}
	}

	logger.Info("queue: successfully synced repository", "duration_ms", durationMs)
	if dedupStore != nil {
		_ = dedupStore.RecordProcessed(taskUUID, repoName, peerIDStr)
	}

	return &SyncResult{
		TaskUUID:         taskUUID,
		RepoName:         repoName,
		TargetPeerID:     peerIDStr,
		Status:           "COMPLETED",
		BytesTransferred: int64(len(out)),
		DurationMs:       durationMs,
	}
}

func isValidName(s string) bool {
	if s == "" || len(s) > 100 {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.') {
			return false
		}
	}
	return !strings.Contains(s, "..")
}

// HandleOfflinePush processes a Git push request locally when the remote peer is offline.
func HandleOfflinePush(w http.ResponseWriter, r *http.Request, peerIDStr string, repoPath string, h host.Host, proxyPort string) {
	mu.Lock()
	defer mu.Unlock()

	service := r.URL.Query().Get("service")
	if r.Method == http.MethodGet && service != "git-receive-pack" {
		http.Error(w, "Offline queue only supports git push", http.StatusServiceUnavailable)
		return
	}
	if r.Method == http.MethodPost && !strings.HasSuffix(r.URL.Path, "git-receive-pack") {
		http.Error(w, "Offline queue only supports git push", http.StatusServiceUnavailable)
		return
	}

	repoName := strings.TrimPrefix(repoPath, "/")
	repoName = strings.Split(repoName, "/")[0]

	if !isValidName(peerIDStr) || !isValidName(repoName) {
		http.Error(w, "Invalid path parameters", http.StatusBadRequest)
		return
	}

	localRepoPath := filepath.Join(queueDir, peerIDStr, repoName+".git")

	if _, err := os.Stat(localRepoPath); os.IsNotExist(err) {
		if err := os.MkdirAll(localRepoPath, 0755); err != nil {
			http.Error(w, "Failed to create queue dir", http.StatusInternalServerError)
			return
		}
		cmd := exec.Command("git", "init", "--bare")
		cmd.Dir = localRepoPath
		if err := cmd.Run(); err != nil {
			http.Error(w, "Failed to init queue repo", http.StatusInternalServerError)
			return
		}
		slog.Info(fmt.Sprintf("queue: initialized offline repo at %s", localRepoPath))
	}

	gitPath, err := exec.LookPath("git")
	if err != nil {
		http.Error(w, "Git not installed", http.StatusInternalServerError)
		return
	}

	handler := &cgi.Handler{
		Path: gitPath,
		Args: []string{"http-backend"},
		Dir:  localRepoPath,
		Env: []string{
			"GIT_PROJECT_ROOT=" + filepath.Join(queueDir, peerIDStr),
			"GIT_HTTP_EXPORT_ALL=true",
			"REMOTE_USER=p2p",
		},
	}

	r.URL.Path = "/" + repoName + ".git" + strings.TrimPrefix(repoPath, "/"+repoName)
	slog.Info(fmt.Sprintf("queue: buffering push for peer %s, repo %s", peerIDStr, repoName))
	handler.ServeHTTP(w, r)
}
