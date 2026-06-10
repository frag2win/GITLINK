package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/localrepo/api-server/internal/database"
)

// ProtectBranch sets branch protection rules for a branch in a repository.
//
//	POST /api/v1/repos/:id/branches/:branch/protect
//	Body: { "requirePR": true }
func ProtectBranch(c *fiber.Ctx) error {
	repoID := c.Params("id")
	branch := c.Params("branch")

	var req struct {
		RequirePR bool `json:"requirePR"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	// Fetch repository
	var repo database.Repository
	if err := db.Conn.First(&repo, repoID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "repository not found"})
	}

	// Upsert BranchProtection
	protection := database.BranchProtection{
		RepositoryID: repo.ID,
		BranchName:   branch,
		RequirePR:    req.RequirePR,
	}

	// Wait, we need to check if it already exists to update it, or use GORM's Clauses(clause.OnConflict{})
	// For simplicity, find and update or create
	var existing database.BranchProtection
	err := db.Conn.Where("repository_id = ? AND branch_name = ?", repo.ID, branch).First(&existing).Error
	if err == nil {
		// Update
		existing.RequirePR = req.RequirePR
		db.Conn.Save(&existing)
	} else {
		// Create
		db.Conn.Create(&protection)
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"branch": branch,
		"protected": true,
		"requirePR": req.RequirePR,
	})
}
