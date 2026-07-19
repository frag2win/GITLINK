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
	queueDir = "/app/queue"
	mu       sync.Mutex
)

// Init initializes the queue directory.
func Init(queuePath string) {
	if queuePath != "" {
		queueDir = queuePath
	}
	if err := os.MkdirAll(queueDir, 0755); err != nil {
		slog.Error("queue: failed to create queue directory", "error", err)
	}
}

// HandleOfflinePush processes a Git push request locally when the remote peer is offline.
func HandleOfflinePush(w http.ResponseWriter, r *http.Request, peerIDStr string, repoPath string, h host.Host, proxyPort string) {
	mu.Lock()
	defer mu.Unlock()

	// Ensure the request is related to pushing (receive-pack)
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
	repoName = strings.Split(repoName, "/")[0] // e.g. "myrepo"

	localRepoPath := filepath.Join(queueDir, peerIDStr, repoName+".git")

	// Create bare repo if it doesn't exist
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

	// Route to git http-backend
	gitPath, err := exec.LookPath("git")
	if err != nil {
		http.Error(w, "Git not installed", http.StatusInternalServerError)
		return
	}

	// Set up CGI handler for git-http-backend
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

	// Rewrite request so git http-backend understands it
	r.URL.Path = "/" + repoName + ".git" + strings.TrimPrefix(repoPath, "/"+repoName)

	slog.Info(fmt.Sprintf("queue: buffering push for peer %s, repo %s", peerIDStr, repoName))
	handler.ServeHTTP(w, r)

	// If this was the POST request (the actual push), schedule a sync
	if r.Method == http.MethodPost {
		go scheduleSync(peerIDStr, repoName, h)
	}
}

func scheduleSync(peerIDStr, repoName string, h host.Host) {
	targetPeer, err := peer.Decode(peerIDStr)
	if err != nil {
		slog.Info(fmt.Sprintf("queue: invalid peer ID %s: %v", peerIDStr, err))
		return
	}

	localRepoPath := filepath.Join(queueDir, peerIDStr, repoName+".git")

	for {
		time.Sleep(10 * time.Second)

		slog.Info(fmt.Sprintf("queue: attempting to sync %s to peer %s...", repoName, peerIDStr))

		// Try to connect to peer
		if err := h.Connect(context.Background(), peer.AddrInfo{ID: targetPeer}); err != nil {
			slog.Info(fmt.Sprintf("queue: peer %s still offline, retrying later...", peerIDStr))
			continue
		}

		// Peer is online!
		// We use standard git push, but we need to push over the proxy!
		// The proxy is running on localhost:4000 (or PROXY_PORT).
		proxyPort := os.Getenv("PROXY_PORT")
		if proxyPort == "" {
			proxyPort = "4000"
		}
		
		remoteURL := fmt.Sprintf("http://127.0.0.1:%s/p2p/%s/%s", proxyPort, peerIDStr, repoName)
		
		cmd := exec.Command("git", "push", "--all", remoteURL)
		cmd.Dir = localRepoPath
		out, err := cmd.CombinedOutput()
		if err != nil {
			slog.Info(fmt.Sprintf("queue: sync failed: %s", string(out)))
			// Might be a non-fast-forward or remote issue. We keep trying or log error.
			time.Sleep(30 * time.Second)
			continue
		}

		slog.Info(fmt.Sprintf("queue: successfully synced %s to peer %s!", repoName, peerIDStr))
		
		// Optionally, clear the queue by removing the repo or leaving it as a cache.
		// For now, leaving it is safer, git push will just say "Everything up-to-date" next time.
		break
	}
}
