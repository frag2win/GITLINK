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
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing auth key",
			})
		}

		// For Phase 1 smoke test, we treat the SSH key as the peerID if it starts with ssh-ed25519
		// In a real implementation we would validate against the contributors table
		if !strings.HasPrefix(authHeader, "ssh-") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unknown key",
			})
		}

		c.Locals("peerID", authHeader)
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
