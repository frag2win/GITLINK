package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// ListContributors returns all contributors with access to a repository.
//
//	GET /api/v1/repos/:id/contributors
func ListContributors(c *fiber.Ctx) error {
	// TODO: Parse repo ID from URL param.
	// TODO: Verify the authenticated peer has read access.
	// TODO: Query the database for contributors linked to this repo.
	// TODO: Return JSON array of contributor objects.

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "ListContributors not implemented",
	})
}

// AddContributor grants a peer access to a repository.
//
//	POST /api/v1/repos/:id/contributors
//	Body: { "peerID": "Qm...", "publicKey": "ssh-ed25519 ...", "role": "write" }
func AddContributor(c *fiber.Ctx) error {
	// TODO: Parse and validate request body.
	// TODO: Verify the authenticated peer is the repo owner.
	// TODO: Create contributor record and permission entry in database.
	// TODO: Log audit event.
	// TODO: Return 201 with the new contributor.

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "AddContributor not implemented",
	})
}

// RemoveContributor revokes a peer's access to a repository.
//
//	DELETE /api/v1/repos/:id/contributors/:peerId
func RemoveContributor(c *fiber.Ctx) error {
	// TODO: Parse repo ID and peer ID from URL params.
	// TODO: Verify the authenticated peer is the repo owner.
	// TODO: Remove contributor and permission records from database.
	// TODO: Log audit event.
	// TODO: Return 204 No Content.

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "RemoveContributor not implemented",
	})
}
