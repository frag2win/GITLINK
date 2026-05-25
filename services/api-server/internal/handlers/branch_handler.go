package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// ListBranches returns all branches for a given repository.
//
//	GET /api/v1/repos/:id/branches
func ListBranches(c *fiber.Ctx) error {
	// TODO: Parse repo ID from URL param.
	// TODO: Verify read access for the authenticated peer.
	// TODO: Send "list-branches" command to git-server via socket.
	// TODO: Return JSON array of branch objects.

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "ListBranches not implemented",
	})
}

// GetBranch returns metadata for a specific branch.
//
//	GET /api/v1/repos/:id/branches/:branch
func GetBranch(c *fiber.Ctx) error {
	// TODO: Parse repo ID and branch name from URL params.
	// TODO: Verify read access.
	// TODO: Send "get-branch" command to git-server via socket.
	// TODO: Return JSON with branch details (head commit, etc.).

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "GetBranch not implemented",
	})
}

// CreateBranch creates a new branch in a repository.
//
//	POST /api/v1/repos/:id/branches
//	Body: { "name": "feature-x", "startPoint": "main" }
func CreateBranch(c *fiber.Ctx) error {
	// TODO: Parse and validate request body.
	// TODO: Verify write access for the authenticated peer.
	// TODO: Send "create-branch" command to git-server via socket.
	// TODO: Log audit event.
	// TODO: Return 201 with the new branch.

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "CreateBranch not implemented",
	})
}

// DeleteBranch removes a branch from a repository.
//
//	DELETE /api/v1/repos/:id/branches/:branch
func DeleteBranch(c *fiber.Ctx) error {
	// TODO: Parse repo ID and branch name from URL params.
	// TODO: Verify write access.
	// TODO: Send "delete-branch" command to git-server via socket.
	// TODO: Log audit event.
	// TODO: Return 204 No Content.

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "DeleteBranch not implemented",
	})
}
