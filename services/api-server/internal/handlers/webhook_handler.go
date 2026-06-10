package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/localrepo/api-server/internal/database"
)

// PreReceiveHook handles internal callbacks from the Git pre-receive hook.
// It checks if the push to the given branch is allowed.
//
//	POST /internal/hooks/pre-receive
//	Body: { "repo": "...", "branch": "...", "oldrev": "...", "newrev": "..." }
func PreReceiveHook(c *fiber.Ctx) error {
	var req struct {
		Repo   string `json:"repo"`
		Branch string `json:"branch"`
		OldRev string `json:"oldrev"`
		NewRev string `json:"newrev"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	log.Printf("pre-receive hook triggered: repo=%s branch=%s oldrev=%s newrev=%s", req.Repo, req.Branch, req.OldRev, req.NewRev)

	// Fetch the repository
	var repo database.Repository
	if err := db.Conn.Where("name = ?", req.Repo).First(&repo).Error; err != nil {
		// If repo not found in DB, just allow it for now (might be a system repo)
		return c.JSON(fiber.Map{"allowed": true})
	}

	// Check branch protection rules
	var protection database.BranchProtection
	if err := db.Conn.Where("repository_id = ? AND branch_name = ?", repo.ID, req.Branch).First(&protection).Error; err == nil {
		if protection.RequirePR {
			log.Printf("push rejected: branch %s in repo %s requires a PR", req.Branch, req.Repo)
			return c.JSON(fiber.Map{"allowed": false, "reason": "branch is protected, must use a pull request"})
		}
	}

	return c.JSON(fiber.Map{"allowed": true})
}
