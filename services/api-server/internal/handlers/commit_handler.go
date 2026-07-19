package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/localrepo/api-server/internal/middleware"
	"github.com/localrepo/api-server/internal/service"
	"gorm.io/gorm"
)

type CommitHandler struct {
	gitService  service.GitService
	repoService service.RepoService
}

func NewCommitHandler(gitService service.GitService, repoService service.RepoService) *CommitHandler {
	return &CommitHandler{
		gitService:  gitService,
		repoService: repoService,
	}
}

func (h *CommitHandler) getAuthorizedRepo(c *fiber.Ctx) (string, error) {
	userID := middleware.UserIDFromContext(c)
	if userID == 0 {
		return "", c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	repoIDStr := c.Params("id")
	repoID, err := strconv.ParseUint(repoIDStr, 10, 32)
	if err != nil {
		return "", c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid repository ID"})
	}

	repo, err := h.repoService.GetRepoByID(c.Context(), uint(repoID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Repository not found"})
		}
		return "", c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return repo.Name, nil
}

func (h *CommitHandler) ListCommits(c *fiber.Ctx) error {
	repoName, err := h.getAuthorizedRepo(c)
	if err != nil {
		return err
	}

	branch := c.Query("branch", "main")
	limitStr := c.Query("limit", "20")
	limit, err := strconv.ParseUint(limitStr, 10, 32)
	if err != nil {
		limit = 20
	}

	commits, err := h.gitService.ListCommits(c.Context(), repoName, branch, uint32(limit))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(commits)
}

func (h *CommitHandler) GetCommit(c *fiber.Ctx) error {
	repoName, err := h.getAuthorizedRepo(c)
	if err != nil {
		return err
	}

	sha := c.Params("sha")
	commit, err := h.gitService.GetCommit(c.Context(), repoName, sha)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(commit)
}

