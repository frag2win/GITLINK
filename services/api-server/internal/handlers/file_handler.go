package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// FileEntry represents a single entry in a directory listing.
type FileEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	Type  string `json:"type"` // "file", "dir", or "symlink"
	Size  int64  `json:"size,omitempty"`
	Mode  string `json:"mode"`
}

// BrowseFiles returns a directory listing for a path inside a repository.
//
//	GET /api/v1/repos/:id/files/*
//	Query: ?ref=main (optional, defaults to HEAD)
func BrowseFiles(c *fiber.Ctx) error {
	// TODO: Parse repo ID and file path from URL params.
	// TODO: Parse optional "ref" query param (branch / tag / SHA).
	// TODO: Verify read access for the authenticated peer.
	// TODO: Send "ls-tree" command to git-server via socket.
	// TODO: Return JSON array of FileEntry objects.

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "BrowseFiles not implemented",
	})
}

// GetFileContent returns the raw content of a file inside a repository.
//
//	GET /api/v1/repos/:id/blob/*
//	Query: ?ref=main (optional, defaults to HEAD)
func GetFileContent(c *fiber.Ctx) error {
	// TODO: Parse repo ID and file path from URL params.
	// TODO: Parse optional "ref" query param.
	// TODO: Verify read access.
	// TODO: Send "cat-file" command to git-server via socket.
	// TODO: Return the raw file content with appropriate Content-Type.

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "GetFileContent not implemented",
	})
}
