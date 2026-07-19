package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/localrepo/api-server/internal/middleware"
	"github.com/localrepo/api-server/internal/service"
)

type RepoHandler struct {
	repoService service.RepoService
}

func NewRepoHandler(repoService service.RepoService) *RepoHandler {
	return &RepoHandler{
		repoService: repoService,
	}
}

func (h *RepoHandler) ListRepos(c *fiber.Ctx) error {
	userID := middleware.UserIDFromContext(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	repos, err := h.repoService.ListRepositories(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(repos)
}

func (h *RepoHandler) GetRepo(c *fiber.Ctx) error {
	repoName := c.Params("id")
	repo, err := h.repoService.GetRepoByName(c.Context(), repoName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(repo)
}

func (h *RepoHandler) CreateRepo(c *fiber.Ctx) error {
	userID := middleware.UserIDFromContext(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	if err := h.repoService.CreateRepository(c.Context(), userID, req.Name, req.Description); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status": "created",
		"name":   req.Name,
	})
}

func (h *RepoHandler) DeleteRepo(c *fiber.Ctx) error {
	repoName := c.Params("id")
	if err := h.repoService.DeleteRepository(c.Context(), repoName); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "deleted",
	})
}
