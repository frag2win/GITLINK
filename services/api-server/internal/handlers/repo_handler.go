// Package handlers contains HTTP handler functions for the api-server.
package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// ListRepos returns all repositories visible to the authenticated peer.
//
//	GET /api/v1/repos?page=1&limit=20
func ListRepos(c *fiber.Ctx) error {
	// TODO: Parse pagination query params (page, limit).
	// TODO: Retrieve peer ID from context (set by auth middleware).
	// TODO: Call repo service to list repos for this peer.
	// TODO: Return JSON array of repo summaries.

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "ListRepos not implemented",
	})
}

// GetRepo returns the details of a single repository by ID.
//
//	GET /api/v1/repos/:id
func GetRepo(c *fiber.Ctx) error {
	// TODO: Parse repo ID from URL param.
	// TODO: Verify the peer has read access.
	// TODO: Call repo service to fetch repo metadata.
	// TODO: Return JSON with repo details.

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "GetRepo not implemented",
	})
}

// CreateRepo initialises a new bare Git repository.
//
//	POST /api/v1/repos
//	Body: { "name": "...", "description": "...", "isPrivate": false }
func CreateRepo(c *fiber.Ctx) error {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	if err := gitClient.CreateRepo(c.Context(), req.Name); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status": "created",
		"name":   req.Name,
	})
}

// DeleteRepo removes a repository and its data.
//
//	DELETE /api/v1/repos/:id
func DeleteRepo(c *fiber.Ctx) error {
	// TODO: Parse repo ID from URL param.
	// TODO: Verify the peer is the owner.
	// TODO: Send "delete" command to git-server via socket.
	// TODO: Remove repo record from database.
	// TODO: Log audit event.
	// TODO: Return 204 No Content.

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "DeleteRepo not implemented",
	})
}
