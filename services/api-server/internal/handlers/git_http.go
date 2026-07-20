package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/localrepo/api-server/internal/middleware"
	"github.com/localrepo/api-server/internal/service"
)

type GitHTTPHandler struct {
	gitService service.GitService
	authzSvc   service.AuthorizationService
}

func NewGitHTTPHandler(gitService service.GitService, authzSvc service.AuthorizationService) *GitHTTPHandler {
	return &GitHTTPHandler{
		gitService: gitService,
		authzSvc:   authzSvc,
	}
}

// InfoRefs handles: GET /:repo/info/refs?service=git-upload-pack
func (h *GitHTTPHandler) InfoRefs(c *fiber.Ctx) error {
	repo := c.Params("repo")
	svcName := c.Query("service")

	if svcName != "git-upload-pack" && svcName != "git-receive-pack" {
		return c.Status(fiber.StatusBadRequest).SendString("invalid service")
	}

	userID := middleware.UserIDFromContext(c)
	if svcName == "git-receive-pack" {
		if userID == 0 {
			return c.Status(fiber.StatusUnauthorized).SendString("Authentication required")
		}
		res, err := h.authzSvc.AuthorizePush(c.UserContext(), userID, repo, "main")
		if err != nil || !res.Allowed {
			return c.Status(fiber.StatusForbidden).SendString("Push access denied")
		}
	}

	c.Set("Content-Type", "application/x-"+svcName+"-advertisement")
	c.Set("Cache-Control", "no-cache")

	respBytes, err := h.gitService.InfoRefs(c.Context(), repo, svcName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	header := pktLine("# service=" + svcName + "\n")
	flush := "0000"

	c.Write([]byte(header))
	c.Write([]byte(flush))
	c.Write(respBytes)
	return nil
}

// UploadPack handles: POST /:repo/git-upload-pack
func (h *GitHTTPHandler) UploadPack(c *fiber.Ctx) error {
	repo := c.Params("repo")

	c.Set("Content-Type", "application/x-git-upload-pack-result")
	c.Set("Cache-Control", "no-cache")

	respBytes, err := h.gitService.UploadPack(c.Context(), repo, c.Body())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.Send(respBytes)
}

// ReceivePack handles: POST /:repo/git-receive-pack
func (h *GitHTTPHandler) ReceivePack(c *fiber.Ctx) error {
	repo := c.Params("repo")

	userID := middleware.UserIDFromContext(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).SendString("Authentication required")
	}

	res, err := h.authzSvc.AuthorizePush(c.UserContext(), userID, repo, "main")
	if err != nil || !res.Allowed {
		return c.Status(fiber.StatusForbidden).SendString("Push access denied")
	}

	c.Set("Content-Type", "application/x-git-receive-pack-result")
	c.Set("Cache-Control", "no-cache")

	respBytes, err := h.gitService.ReceivePack(c.Context(), repo, c.Body())
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
