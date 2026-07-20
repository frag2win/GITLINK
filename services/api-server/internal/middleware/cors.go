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
		AllowOrigins: "http://localhost:5173, http://localhost:3000, https://gitlink.local",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))
}
