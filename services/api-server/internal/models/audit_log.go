package models

import "time"

// AuditLog records a single auditable event in the system.
// Every mutating operation (create, delete, push, access change) should
// produce an AuditLog entry.
type AuditLog struct {
	// ID is the unique identifier for this log entry.
	ID string `json:"id"`

	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`

	// PeerID identifies the peer that performed the operation.
	PeerID string `json:"peerID"`

	// Operation is a machine-readable label for the action performed
	// (e.g. "repo.create", "branch.delete", "contributor.add").
	Operation string `json:"operation"`

	// RepoName is the name of the repository involved, if any.
	RepoName string `json:"repoName,omitempty"`

	// Details contains free-form JSON with extra context about the event.
	Details string `json:"details,omitempty"`
}
