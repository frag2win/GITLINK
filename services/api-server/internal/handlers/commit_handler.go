package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// ListCommits returns a paginated list of commits for a repository.
//
//	GET /api/v1/repos/:id/commits?branch=main&page=1&limit=30
func ListCommits(c *fiber.Ctx) error {
	// TODO: Parse repo ID, branch, page, and limit from query/params.
	// TODO: Verify read access for the authenticated peer.
	// TODO: Send "log" command to git-server via socket.
	// TODO: Return JSON array of commit summaries (SHA, message, author, date).

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "ListCommits not implemented",
	})
}

// GetCommit returns the full details of a single commit.
//
//	GET /api/v1/repos/:id/commits/:sha
func GetCommit(c *fiber.Ctx) error {
	// TODO: Parse repo ID and commit SHA from URL params.
	// TODO: Verify read access.
	// TODO: Send "show" command to git-server via socket.
	// TODO: Return JSON with commit details (SHA, tree, parent(s), author, committer, message, diff stats).

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "GetCommit not implemented",
	})
}
