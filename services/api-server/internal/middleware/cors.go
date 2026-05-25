package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// SetupCORS registers CORS middleware on the Fiber app.
// During local development the UI dev server runs on a different port,
// so we allow all origins. In production this should be locked down.
func SetupCORS(app *fiber.App) {
	app.Use(cors.New(cors.Config{
		// TODO: Restrict AllowOrigins in production.
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))
}
