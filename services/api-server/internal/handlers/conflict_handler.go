package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/localrepo/api-server/internal/middleware"
	"github.com/localrepo/api-server/internal/models"
	"github.com/localrepo/api-server/internal/service"
	"gorm.io/gorm"
)

type ConflictHandler struct {
	conflictService service.ConflictService
	repoService     service.RepoService
}

func NewConflictHandler(conflictService service.ConflictService, repoService service.RepoService) *ConflictHandler {
	return &ConflictHandler{
		conflictService: conflictService,
		repoService:     repoService,
	}
}

func (h *ConflictHandler) AnalyzeConflicts(c *fiber.Ctx) error {
	userID := middleware.UserIDFromContext(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	repoIDStr := c.Params("id")
	var repo *models.Repository
	var err error
	if id, errParse := strconv.ParseUint(repoIDStr, 10, 32); errParse == nil {
		repo, err = h.repoService.GetRepoByID(c.UserContext(), uint(id))
	} else {
		repo, err = h.repoService.GetRepoByName(c.UserContext(), repoIDStr)
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Repository not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	baseBranch := c.Query("base", "main")
	headBranch := c.Query("head", "feature")

	report, err := h.conflictService.AnalyzeConflicts(c.UserContext(), repo.ID, repo.Name, baseBranch, headBranch)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(report)
}
