package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// AuthRequest represents the body of an authentication request.
type AuthRequest struct {
	// PublicKey is the SSH public key in authorized_keys format
	// (e.g. "ssh-ed25519 AAAA... user@host").
	PublicKey string `json:"publicKey"`

	// Signature is a cryptographic signature proving possession of
	// the corresponding private key.
	Signature string `json:"signature"`

	// Challenge is the nonce that was signed.
	Challenge string `json:"challenge"`
}

// AuthResponse is returned on successful authentication.
type AuthResponse struct {
	// PeerID is the libp2p peer ID derived from the public key.
	PeerID string `json:"peerID"`

	// Token is a short-lived session token for subsequent API calls.
	Token string `json:"token"`

	// ExpiresAt is the UNIX timestamp when the token expires.
	ExpiresAt int64 `json:"expiresAt"`
}

// Authenticate verifies an SSH-key-based identity and returns a session token.
//
//	POST /api/v1/auth
//	Body: AuthRequest
func Authenticate(c *fiber.Ctx) error {
	// TODO: Parse and validate AuthRequest body.
	// TODO: Verify the signature against the public key and challenge.
	// TODO: Look up or register the contributor by public key.
	// TODO: Generate a short-lived session token.
	// TODO: Log audit event for authentication.
	// TODO: Return AuthResponse with token.

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "Authenticate not implemented",
	})
}
