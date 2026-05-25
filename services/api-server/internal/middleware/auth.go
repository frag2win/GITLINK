// Package middleware provides Fiber middleware for the api-server.
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Auth returns a Fiber middleware handler that validates the peer's
// identity on each request. It checks for a session token in the
// Authorization header and injects the peer ID into the request context.
func Auth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// ---- Extract token from Authorization header ----
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing Authorization header",
			})
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid Authorization format, expected 'Bearer <token>'",
			})
		}

		// TODO: Validate the token (check signature, expiry).
		// TODO: Extract peer ID from validated token claims.
		// TODO: Optionally verify the peer exists in the contributors table.

		peerID := "" // placeholder — set from token claims
		_ = peerID

		// Store peer ID in Fiber locals for downstream handlers.
		// c.Locals("peerID", peerID)

		return c.Next()
	}
}

// PeerIDFromContext extracts the authenticated peer ID stored by the
// Auth middleware. Returns an empty string if not set.
func PeerIDFromContext(c *fiber.Ctx) string {
	if id, ok := c.Locals("peerID").(string); ok {
		return id
	}
	return ""
}
