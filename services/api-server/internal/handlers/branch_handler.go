package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/localrepo/api-server/internal/middleware"
	"github.com/localrepo/api-server/internal/service"
	"gorm.io/gorm"
)

type BranchHandler struct {
	gitService        service.GitService
	repoService       service.RepoService
	branchProtectSvc  service.BranchProtectionService
	authorizationSvc  service.AuthorizationService
}

func NewBranchHandler(
	gitService service.GitService,
	repoService service.RepoService,
	branchProtectSvc service.BranchProtectionService,
	authorizationSvc service.AuthorizationService,
) *BranchHandler {
	return &BranchHandler{
		gitService:       gitService,
		repoService:      repoService,
		branchProtectSvc: branchProtectSvc,
		authorizationSvc: authorizationSvc,
	}
}

func (h *BranchHandler) getAuthorizedRepo(c *fiber.Ctx) (string, uint, error) {
	userID := middleware.UserIDFromContext(c)
	if userID == 0 {
		return "", 0, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	repoIDStr := c.Params("id")
	repoID, err := strconv.ParseUint(repoIDStr, 10, 32)
	if err != nil {
		return "", 0, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid repository ID"})
	}

	repo, err := h.repoService.GetRepoByID(c.Context(), uint(repoID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", 0, c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Repository not found"})
		}
		return "", 0, c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	isAllowed := false
	var authErr error
	if c.Method() == fiber.MethodGet {
		isAllowed, authErr = h.authorizationSvc.AuthorizeRead(c.Context(), userID, repo.ID)
	} else {
		isAllowed, authErr = h.authorizationSvc.AuthorizeWrite(c.Context(), userID, repo.ID)
	}

	if authErr != nil || !isAllowed {
		return "", 0, c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden: Access denied"})
	}

	return repo.Name, uint(repoID), nil
}

func (h *BranchHandler) ListBranches(c *fiber.Ctx) error {
	repoName, _, err := h.getAuthorizedRepo(c)
	if err != nil {
		return err
	}

	branches, err := h.gitService.ListBranches(c.Context(), repoName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(branches)
}

func (h *BranchHandler) GetBranch(c *fiber.Ctx) error {
	repoName, _, err := h.getAuthorizedRepo(c)
	if err != nil {
		return err
	}

	branchName := c.Params("branch")
	branch, err := h.gitService.GetBranch(c.Context(), repoName, branchName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(branch)
}

func (h *BranchHandler) CreateBranch(c *fiber.Ctx) error {
	repoName, _, err := h.getAuthorizedRepo(c)
	if err != nil {
		return err
	}

	var req struct {
		Name   string `json:"name"`
		Target string `json:"target"` // Commit SHA or base branch name
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	branch, err := h.gitService.CreateBranch(c.Context(), repoName, req.Name, req.Target)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(branch)
}

func (h *BranchHandler) DeleteBranch(c *fiber.Ctx) error {
	repoName, repoID, err := h.getAuthorizedRepo(c)
	if err != nil {
		return err
	}

	branchName := c.Params("branch")

	// Verify they are allowed to delete (cannot delete if branch protection is enabled)
	rule, err := h.branchProtectSvc.GetProtection(c.Context(), repoID, branchName)
	if err == nil && rule.RequirePR {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Branch is protected and cannot be deleted"})
	}

	if err := h.gitService.DeleteBranch(c.Context(), repoName, branchName); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}

func (h *BranchHandler) ProtectBranch(c *fiber.Ctx) error {
	_, repoID, err := h.getAuthorizedRepo(c)
	if err != nil {
		return err
	}

	branchName := c.Params("branch")
	var req struct {
		RequirePR bool `json:"require_pr"`
	}
	// Default to true if body is empty
	req.RequirePR = true
	_ = c.BodyParser(&req)

	if err := h.branchProtectSvc.EnableProtection(c.Context(), repoID, branchName, req.RequirePR); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "protected", "require_pr": req.RequirePR})
}

