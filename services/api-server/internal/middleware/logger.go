package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SetupLogger registers a request-logging middleware on the Fiber app.
// Every request is logged with method, path, status, latency, and the
// authenticated peer ID (when available). This forms the HTTP-level
// portion of the audit trail.
func SetupLogger(app *fiber.App) {
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		latency := time.Since(start)
		peerID := PeerIDFromContext(c)

		log.Printf(
			"[%s] %s %s — %d (%s) peer=%s",
			time.Now().Format(time.RFC3339),
			c.Method(),
			c.Path(),
			c.Response().StatusCode(),
			latency.Round(time.Microsecond),
			peerID,
		)

		return err
	})
}
