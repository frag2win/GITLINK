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

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Auth routes (public)
	api.Post("/auth/register", handlers.Register)
	api.Post("/auth/login", handlers.Login)

	// Protected routes — require valid JWT
	protected := api.Group("", middleware.Auth())

	// Authenticated User info & keys
	protected.Get("/auth/me", handlers.GetMe)
	protected.Post("/user/keys", handlers.AddSSHKey)
	protected.Get("/user/keys", handlers.ListSSHKeys)
	protected.Delete("/user/keys/:id", handlers.DeleteSSHKey)

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
	protected.Post("/repos/:id/branches/:branch/protect", handlers.ProtectBranch)

	// Pull Request routes
	protected.Get("/repos/:id/pulls", handlers.ListPullRequests)
	protected.Post("/repos/:id/pulls", handlers.CreatePullRequest)
	protected.Post("/repos/:id/pulls/:pr_id/merge", handlers.MergePullRequest)

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

	// ---- Internal hooks (called by git-server) ----
	app.Post("/internal/hooks/pre-receive", handlers.PreReceiveHook)

	// ---- Static file serving for the web UI ----
	app.Static("/", "./ui/dist", fiber.Static{
		Index:    "index.html",
		Compress: true,
	})

	// Git Smart HTTP fallback routes
	app.Get("/:repo/info/refs", handlers.InfoRefs)
	app.Post("/:repo/git-upload-pack", handlers.UploadPack)
	app.Post("/:repo/git-receive-pack", handlers.ReceivePack)

	// SPA fallback — serve index.html for any unmatched routes
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("./ui/dist/index.html")
	})
}
