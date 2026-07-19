package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/localrepo/api-server/internal/config"
	"github.com/localrepo/api-server/internal/handlers"
	"github.com/localrepo/api-server/internal/middleware"
)

// Handlers holds all the dependencies required for routing
type Handlers struct {
	Auth        *handlers.AuthHandler
	SSH         *handlers.SSHHandler
	Repo        *handlers.RepoHandler
	Branch      *handlers.BranchHandler
	Pull        *handlers.PullRequestHandler
	Commit      *handlers.CommitHandler
	File        *handlers.FileHandler
	Contributor *handlers.ContributorHandler
	Health      *handlers.HealthHandler
	GitHTTP     *handlers.GitHTTPHandler
}

// Setup registers all API route groups and static file serving on the given Fiber app.
func Setup(app *fiber.App, h *Handlers, cfg *config.Config) {
	api := app.Group("/api/v1")

	// Health endpoints
	app.Get("/health", h.Health.CheckHealth)
	app.Get("/ready", h.Health.CheckHealth)
	app.Get("/live", h.Health.CheckHealth)

	// Auth routes (public)
	api.Post("/auth/register", h.Auth.Register)
	api.Post("/auth/login", h.Auth.Login)

	// Protected routes — require valid JWT
	protected := api.Group("", middleware.Auth(cfg.JWTSecret))

	// Authenticated User info & keys
	protected.Get("/auth/me", h.Auth.GetMe)
	protected.Post("/user/keys", h.SSH.AddSSHKey)
	protected.Get("/user/keys", h.SSH.ListSSHKeys)
	protected.Delete("/user/keys/:id", h.SSH.DeleteSSHKey)

	// Repository routes
	protected.Get("/repos", h.Repo.ListRepos)
	protected.Get("/repos/:id", h.Repo.GetRepo)
	protected.Post("/repos", h.Repo.CreateRepo)
	protected.Delete("/repos/:id", h.Repo.DeleteRepo)

	// Branch routes
	protected.Get("/repos/:id/branches", h.Branch.ListBranches)
	protected.Get("/repos/:id/branches/:branch", h.Branch.GetBranch)
	protected.Post("/repos/:id/branches", h.Branch.CreateBranch)
	protected.Delete("/repos/:id/branches/:branch", h.Branch.DeleteBranch)
	protected.Post("/repos/:id/branches/:branch/protect", h.Branch.ProtectBranch)

	// Pull Request routes
	protected.Get("/repos/:id/pulls", h.Pull.ListPullRequests)
	protected.Post("/repos/:id/pulls", h.Pull.CreatePullRequest)
	protected.Post("/repos/:id/pulls/:pr_id/merge", h.Pull.MergePullRequest)

	// Commit routes
	protected.Get("/repos/:id/commits", h.Commit.ListCommits)
	protected.Get("/repos/:id/commits/:sha", h.Commit.GetCommit)

	// Contributor routes
	protected.Get("/repos/:id/contributors", h.Contributor.ListContributors)
	protected.Post("/repos/:id/contributors", h.Contributor.AddContributor)
	protected.Delete("/repos/:id/contributors/:peerId", h.Contributor.RemoveContributor)

	// File browser routes
	protected.Get("/repos/:id/files", h.File.BrowseFiles)
	protected.Get("/repos/:id/files/*", h.File.BrowseFiles)
	protected.Get("/repos/:id/blob/*", h.File.GetFileContent)

	// ---- Static file serving for the web UI ----
	app.Static("/", "./ui/dist", fiber.Static{
		Index:    "index.html",
		Compress: true,
	})

	// Git Smart HTTP fallback routes
	app.Get("/:repo/info/refs", h.GitHTTP.InfoRefs)
	app.Post("/:repo/git-upload-pack", h.GitHTTP.UploadPack)
	app.Post("/:repo/git-receive-pack", h.GitHTTP.ReceivePack)

	// SPA fallback — serve index.html for any unmatched routes
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("./ui/dist/index.html")
	})
}
