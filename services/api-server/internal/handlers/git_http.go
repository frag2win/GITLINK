package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// InfoRefs handles: GET /:repo/info/refs?service=git-upload-pack
func InfoRefs(c *fiber.Ctx) error {
	repo := c.Params("repo")
	service := c.Query("service")

	if service != "git-upload-pack" && service != "git-receive-pack" {
		return c.Status(fiber.StatusBadRequest).SendString("invalid service")
	}

	c.Set("Content-Type", "application/x-"+service+"-advertisement")
	c.Set("Cache-Control", "no-cache")

	respBytes, err := gitClient.InfoRefs(c.Context(), repo, service)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	header := pktLine("# service=" + service + "\n")
	flush := "0000"

	c.Write([]byte(header))
	c.Write([]byte(flush))
	c.Write(respBytes)
	return nil
}

// UploadPack handles: POST /:repo/git-upload-pack
func UploadPack(c *fiber.Ctx) error {
	repo := c.Params("repo")

	c.Set("Content-Type", "application/x-git-upload-pack-result")
	c.Set("Cache-Control", "no-cache")

	respBytes, err := gitClient.UploadPack(c.Context(), repo, c.Body())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.Send(respBytes)
}

// ReceivePack handles: POST /:repo/git-receive-pack
func ReceivePack(c *fiber.Ctx) error {
	repo := c.Params("repo")

	c.Set("Content-Type", "application/x-git-receive-pack-result")
	c.Set("Cache-Control", "no-cache")

	respBytes, err := gitClient.ReceivePack(c.Context(), repo, c.Body())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.Send(respBytes)
}

// pktLine formats a string as a Git pkt-line
func pktLine(s string) string {
	length := len(s) + 4
	return fmt.Sprintf("%04x%s", length, s)
}
