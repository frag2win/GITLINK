// Package models defines the data structures persisted by the api-server.
package models

import "time"

// Repo represents a Git repository managed by the platform.
type Repo struct {
	// ID is the unique identifier (UUID or auto-increment).
	ID string `json:"id"`

	// Name is the human-readable repository name (e.g. "my-project").
	// Must be unique per owner.
	Name string `json:"name"`

	// Description is an optional short summary of the repository.
	Description string `json:"description"`

	// Owner is the PeerID of the repository creator / primary owner.
	Owner string `json:"owner"`

	// IsPrivate controls whether the repo is visible to all peers
	// or only to explicitly added contributors.
	IsPrivate bool `json:"isPrivate"`

	// CreatedAt is the timestamp when the repo was initialised.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt is the timestamp of the last metadata change.
	UpdatedAt time.Time `json:"updatedAt"`
}

// CreateRepoRequest is the expected request body for POST /repos.
type CreateRepoRequest struct {
	Name        string `json:"name"        validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
	IsPrivate   bool   `json:"isPrivate"`
}
