package services

import (
	"log"

	"github.com/localrepo/api-server/internal/database"
)

// AuditService records auditable events in the database.
// It is used by other services to maintain a tamper-evident log of
// all significant operations.
type AuditService struct {
	db *database.DB
}

// NewAuditService creates an AuditService with the given database.
func NewAuditService(db *database.DB) *AuditService {
	return &AuditService{db: db}
}

// Log records an audit event. It is intentionally fire-and-forget:
// a failure to log an audit event should not block the primary operation,
// but it will be logged to stderr for operational alerting.
func (s *AuditService) Log(peerID, operation, repoName, details string) {
	if err := s.db.InsertAuditLog(peerID, operation, repoName, details); err != nil {
		log.Printf("AUDIT ERROR: failed to insert audit log: %v (peer=%s op=%s repo=%s)",
			err, peerID, operation, repoName)
	}
}

// Common audit operation constants.
const (
	OpRepoCreate        = "repo.create"
	OpRepoDelete        = "repo.delete"
	OpBranchCreate      = "branch.create"
	OpBranchDelete      = "branch.delete"
	OpContributorAdd    = "contributor.add"
	OpContributorRemove = "contributor.remove"
	OpAuthenticate      = "auth.authenticate"
	OpPeerClone         = "peer.clone"
	OpPeerPush          = "peer.push"
	OpPeerPull          = "peer.pull"
)
