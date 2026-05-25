// Package router wires up all Fiber routes and handler groups.
package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/localrepo/api-server/internal/config"
	"github.com/localrepo/api-server/internal/database"
	"github.com/localrepo/api-server/internal/handlers"
	"github.com/localrepo/api-server/internal/middleware"
)

// Setup registers all API route groups and static file serving on the
// given Fiber app.
func Setup(app *fiber.App, db *database.DB, cfg *config.Config) {
	// ---- API v1 route group ----
	api := app.Group("/api/v1")

	// Auth routes (public)
	api.Post("/auth", handlers.Authenticate)

	// Protected routes — require valid peer identity
	protected := api.Group("", middleware.Auth())

	// Repository routes
	protected.Get("/repos", handlers.ListRepos)
	protected.Get("/repos/:id", handlers.GetRepo)
	protected.Post("/repos", handlers.CreateRepo)
	protected.Delete("/repos/:id", handlers.DeleteRepo)

	// Branch routes
	protected.Get("/repos/:id/branches", handlers.ListBranches)
	protected.Get("/repos/:id/branches/:branch", handlers.GetBranch)
	protected.Post("/repos/:id/branches", handlers.CreateBranch)
	protected.Delete("/repos/:id/branches/:branch", handlers.DeleteBranch)

	// Commit routes
	protected.Get("/repos/:id/commits", handlers.ListCommits)
	protected.Get("/repos/:id/commits/:sha", handlers.GetCommit)

	// Contributor routes
	protected.Get("/repos/:id/contributors", handlers.ListContributors)
	protected.Post("/repos/:id/contributors", handlers.AddContributor)
	protected.Delete("/repos/:id/contributors/:peerId", handlers.RemoveContributor)

	// File browser routes
	protected.Get("/repos/:id/files/*", handlers.BrowseFiles)
	protected.Get("/repos/:id/blob/*", handlers.GetFileContent)

	// ---- Static file serving for the web UI ----
	app.Static("/", "./ui/dist", fiber.Static{
		Index:    "index.html",
		Compress: true,
	})

	// SPA fallback — serve index.html for any unmatched routes
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("./ui/dist/index.html")
	})
}
