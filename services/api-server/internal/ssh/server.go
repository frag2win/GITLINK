package ssh

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/localrepo/api-server/internal/config"
	"github.com/localrepo/api-server/internal/service"
)

// Start begins the embedded SSH daemon
func Start(cfg *config.Config, sshService service.SSHService, repoService service.RepoService, authzService service.AuthorizationService) error {
	port := cfg.SSHPort
	if port == "" {
		port = "2222" // fallback
	}

	ssh.Handle(func(s ssh.Session) {
		cmd := s.Command()
		if len(cmd) == 0 {
			fmt.Fprintln(s, "Welcome to GitLink SSH!")
			s.Exit(0)
			return
		}

		gitCmd := cmd[0]
		if gitCmd != "git-upload-pack" && gitCmd != "git-receive-pack" {
			fmt.Fprintln(s, "Only git commands are allowed")
			s.Exit(1)
			return
		}

		if len(cmd) < 2 {
			fmt.Fprintln(s, "Missing repository path")
			s.Exit(1)
			return
		}

		// The repo path might have a leading slash and trailing .git
		// e.g. /myrepo.git or myrepo.git
		repoName := strings.Trim(cmd[1], "'\"/")
		repoName = strings.TrimSuffix(repoName, ".git")

		// Verify user has access (basic check)
		userIDVal := s.Context().Value("userID")
		if userIDVal == nil {
			fmt.Fprintln(s, "Authentication required")
			s.Exit(1)
			return
		}
		userID := userIDVal.(uint)

		// Verify repo exists in DB
		if _, err := repoService.GetRepoByName(context.Background(), repoName); err != nil {
			fmt.Fprintln(s, "Repository not found")
			s.Exit(1)
			return
		}

		// For write operations, perform a basic collaborator authorization check upfront.
		// Detailed branch protection checks will run inside the pre-receive hook itself.
		if gitCmd == "git-receive-pack" {
			res, err := authzService.AuthorizePush(context.Background(), userID, repoName, "refs/heads/temp-pre-check")
			if err != nil {
				slog.Error("SSH auth check failed", "error", err)
				fmt.Fprintln(s, "Internal authorization error")
				s.Exit(1)
				return
			}
			// Note: We bypass branch-protection check for this pre-check since branch is a dummy
			if !res.Allowed && !res.ProtectedBranch {
				fmt.Fprintf(s, "Permission denied: %s\n", res.Reason)
				s.Exit(1)
				return
			}
		}

		// Execute the git command locally
		repoPath := filepath.Join(cfg.ReposPath, repoName+".git")

		slog.Info("Executing SSH git command", "cmd", gitCmd, "repo", repoPath, "user_id", userID)

		execCmd := exec.Command(gitCmd, repoPath)
		execCmd.Env = append(os.Environ(), fmt.Sprintf("GITLINK_USER_ID=%d", userID))

		// Wire up stdin/stdout/stderr
		execCmd.Stdin = s
		execCmd.Stdout = s
		execCmd.Stderr = s.Stderr()

		if err := execCmd.Run(); err != nil {
			slog.Error("SSH command error", "error", err)
			slog.Debug("Execution failed. Was it pushed?", "error", err)
			s.Exit(1)
			return
		}

		s.Exit(0)
	})

	slog.Info("SSH server listening", "port", port)
	return ssh.ListenAndServe(":"+port, nil, ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		sshKey, err := sshService.AuthenticateByKey(key)
		if err != nil {
			slog.Warn("SSH Auth failed", "error", err)
			return false
		}

		// Attach user ID to the context so the session handler knows who logged in
		ctx.SetValue("userID", sshKey.UserID)
		slog.Info("SSH Auth succeeded", "user_id", sshKey.UserID, "key", sshKey.Name)
		return true
	}))
}
