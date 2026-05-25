package models

import "time"

// Contributor represents a peer that has access to one or more repositories.
type Contributor struct {
	// ID is the unique identifier for this contributor record.
	ID string `json:"id"`

	// Name is a human-friendly display name for the contributor.
	Name string `json:"name"`

	// PublicKey is the SSH public key in authorized_keys format
	// (e.g. "ssh-ed25519 AAAA... user@host").
	PublicKey string `json:"publicKey"`

	// PeerID is the libp2p peer identifier derived from the public key.
	PeerID string `json:"peerID"`

	// Repos is the list of repository IDs this contributor has access to.
	// In the database this is stored as a join table; here it is
	// populated on demand.
	Repos []string `json:"repos,omitempty"`

	// CreatedAt is the timestamp when the contributor was first registered.
	CreatedAt time.Time `json:"createdAt"`
}

// AddContributorRequest is the expected body for POST /repos/:id/contributors.
type AddContributorRequest struct {
	PeerID    string `json:"peerID"    validate:"required"`
	PublicKey string `json:"publicKey" validate:"required"`
	Role      string `json:"role"      validate:"required,oneof=read write admin"`
}
