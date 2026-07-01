package ssh

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/localrepo/api-server/internal/config"
	"github.com/localrepo/api-server/internal/db"
)

// Start begins the embedded SSH daemon
func Start(cfg *config.Config) error {
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
		_ = userID // In a full implementation, we'd check DB to ensure userID can access repoName

		// Verify repo exists in DB
		var repo db.Repository
		if err := db.DB.Where("name = ?", repoName).First(&repo).Error; err != nil {
			fmt.Fprintln(s, "Repository not found")
			s.Exit(1)
			return
		}

		// Execute the git command locally
		repoPath := filepath.Join(cfg.ReposPath, repoName+".git")
		
		log.Printf("Executing %s on %s for user %d", gitCmd, repoPath, userID)

		execCmd := exec.Command(gitCmd, repoPath)
		
		// Wire up stdin/stdout/stderr
		execCmd.Stdin = s
		execCmd.Stdout = s
		execCmd.Stderr = s.Stderr()

		if err := execCmd.Run(); err != nil {
			log.Printf("SSH command error: %v", err)
			s.Exit(1)
			return
		}

		s.Exit(0)
	})

	log.Printf("SSH server listening on :%s", port)
	return ssh.ListenAndServe(":"+port, nil, ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		// Calculate the fingerprint of the incoming key
		hash := sha256.Sum256(key.Marshal())
		fingerprint := "SHA256:" + base64.RawStdEncoding.EncodeToString(hash[:])

		var sshKey db.SSHKey
		if err := db.DB.Where("fingerprint = ?", fingerprint).First(&sshKey).Error; err != nil {
			log.Printf("SSH Auth failed: key fingerprint %s not found", fingerprint)
			return false
		}

		// Attach user ID to the context so the session handler knows who logged in
		ctx.SetValue("userID", sshKey.UserID)
		log.Printf("SSH Auth succeeded for user %d (key: %s)", sshKey.UserID, sshKey.Name)
		return true
	}))
}
