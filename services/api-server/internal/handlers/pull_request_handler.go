package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/localrepo/api-server/internal/database"
	pb "github.com/localrepo/api-server/proto/generated"
)

// CreatePullRequest creates a new PR.
//
//	POST /api/v1/repos/:id/pulls
//	Body: { "title": "...", "description": "...", "baseBranch": "main", "headBranch": "feature" }
func CreatePullRequest(c *fiber.Ctx) error {
	repoID := c.Params("id")

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		BaseBranch  string `json:"baseBranch"`
		HeadBranch  string `json:"headBranch"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	var repo database.Repository
	if err := db.Conn.First(&repo, repoID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "repository not found"})
	}

	pr := database.PullRequest{
		RepositoryID: repo.ID,
		Title:        req.Title,
		Description:  req.Description,
		BaseBranch:   req.BaseBranch,
		HeadBranch:   req.HeadBranch,
		Status:       "open",
	}

	if err := db.Conn.Create(&pr).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(pr)
}

// ListPullRequests lists PRs for a repo.
//
//	GET /api/v1/repos/:id/pulls
func ListPullRequests(c *fiber.Ctx) error {
	repoID := c.Params("id")

	var prs []database.PullRequest
	if err := db.Conn.Where("repository_id = ?", repoID).Find(&prs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(prs)
}

// MergePullRequest merges an open PR.
//
//	POST /api/v1/repos/:id/pulls/:pr_id/merge
func MergePullRequest(c *fiber.Ctx) error {
	repoID := c.Params("id")
	prID := c.Params("pr_id")

	var repo database.Repository
	if err := db.Conn.First(&repo, repoID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "repository not found"})
	}

	var pr database.PullRequest
	if err := db.Conn.First(&pr, prID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "PR not found"})
	}

	if pr.Status != "open" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "PR is not open"})
	}

	// Trigger Rust backend
	req := &pb.MergePullRequest{
		RepoName:      repo.Name,
		BaseBranch:    pr.BaseBranch,
		HeadBranch:    pr.HeadBranch,
		AuthorName:    "System",
		AuthorEmail:   "system@p2p.local",
		CommitMessage: "Merge pull request #" + prID + " from " + pr.HeadBranch,
	}

	hash, err := gitClient.MergePullRequest(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Update PR status
	pr.Status = "merged"
	db.Conn.Save(&pr)

	return c.JSON(fiber.Map{
		"status": "merged",
		"hash":   hash,
	})
}
