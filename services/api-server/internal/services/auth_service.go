package services

import (
	"fmt"
	"time"

	"github.com/localrepo/api-server/internal/database"
)

// AuthService handles authentication and authorization logic.
type AuthService struct {
	db *database.DB
}

// NewAuthService creates an AuthService with the given database.
func NewAuthService(db *database.DB) *AuthService {
	return &AuthService{db: db}
}

// TokenClaims holds the decoded contents of a session token.
type TokenClaims struct {
	PeerID    string
	PublicKey string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// VerifySignature checks that the given signature was produced by the
// private key corresponding to publicKey over the challenge string.
func (s *AuthService) VerifySignature(publicKey, challenge, signature string) error {
	// TODO: Decode the SSH public key from authorized_keys format.
	// TODO: Decode the base64 signature.
	// TODO: Verify using the appropriate crypto library (ed25519 / ecdsa).

	return fmt.Errorf("AuthService.VerifySignature not implemented")
}

// IssueToken creates a short-lived session token for the authenticated peer.
func (s *AuthService) IssueToken(peerID, publicKey string) (string, int64, error) {
	// TODO: Create a JWT or HMAC-signed token containing the peer ID.
	// TODO: Set expiry (e.g. 24 hours).
	// TODO: Return (token, expiresAtUnix, nil).

	return "", 0, fmt.Errorf("AuthService.IssueToken not implemented")
}

// ValidateToken parses and validates a session token, returning the claims.
func (s *AuthService) ValidateToken(token string) (*TokenClaims, error) {
	// TODO: Verify the token signature / HMAC.
	// TODO: Check expiry.
	// TODO: Return the decoded claims.

	return nil, fmt.Errorf("AuthService.ValidateToken not implemented")
}

// HasAccess checks whether the given peer has the required role on a repo.
func (s *AuthService) HasAccess(peerID, repoID, requiredRole string) (bool, error) {
	// TODO: Query the permissions table for (repo_id, peer_id).
	// TODO: Compare the stored role against requiredRole (read < write < admin).

	return false, fmt.Errorf("AuthService.HasAccess not implemented")
}
