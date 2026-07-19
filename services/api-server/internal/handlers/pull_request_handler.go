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

type PullRequestHandler struct {
	pullSvc     service.PullRequestService
	repoService service.RepoService
}

func NewPullRequestHandler(pullSvc service.PullRequestService, repoService service.RepoService) *PullRequestHandler {
	return &PullRequestHandler{
		pullSvc:     pullSvc,
		repoService: repoService,
	}
}

func (h *PullRequestHandler) getAuthorizedRepo(c *fiber.Ctx) (*models.Repository, error) {
	userID := middleware.UserIDFromContext(c)
	if userID == 0 {
		return nil, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	repoIDStr := c.Params("id")
	var repo *models.Repository
	var err error
	if id, errParse := strconv.ParseUint(repoIDStr, 10, 32); errParse == nil {
		repo, err = h.repoService.GetRepoByID(c.Context(), uint(id))
	} else {
		repo, err = h.repoService.GetRepoByName(c.Context(), repoIDStr)
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Repository not found"})
		}
		return nil, c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return repo, nil
}

type PullRequestDTO struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	BaseBranch  string `json:"baseBranch"`
	HeadBranch  string `json:"headBranch"`
	Status      string `json:"status"` // open, merged, closed
	AuthorName  string `json:"authorName"`
	CreatedAt   string `json:"createdAt"`
}

func (h *PullRequestHandler) ListPullRequests(c *fiber.Ctx) error {
	repo, err := h.getAuthorizedRepo(c)
	if err != nil {
		return err
	}

	prs, err := h.pullSvc.ListPullRequests(c.Context(), repo.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	dtos := make([]PullRequestDTO, len(prs))
	for i, pr := range prs {
		dtos[i] = PullRequestDTO{
			ID:          pr.ID,
			Title:       pr.Title,
			Description: pr.Description,
			BaseBranch:  pr.BaseBranch,
			HeadBranch:  pr.HeadBranch,
			Status:      pr.Status,
			AuthorName:  pr.Author.Username,
			CreatedAt:   pr.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return c.JSON(dtos)
}

func (h *PullRequestHandler) CreatePullRequest(c *fiber.Ctx) error {
	repo, err := h.getAuthorizedRepo(c)
	if err != nil {
		return err
	}

	userID := middleware.UserIDFromContext(c)

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		BaseBranch  string `json:"baseBranch"`
		HeadBranch  string `json:"headBranch"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	pr := &models.PullRequest{
		RepositoryID: repo.ID,
		AuthorID:     userID,
		Title:        req.Title,
		Description:  req.Description,
		BaseBranch:   req.BaseBranch,
		HeadBranch:   req.HeadBranch,
	}

	if err := h.pullSvc.CreatePullRequest(c.Context(), pr); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	username, _ := c.Locals("username").(string)

	dto := PullRequestDTO{
		ID:          pr.ID,
		Title:       pr.Title,
		Description: pr.Description,
		BaseBranch:  pr.BaseBranch,
		HeadBranch:  pr.HeadBranch,
		Status:      pr.Status,
		AuthorName:  username,
		CreatedAt:   pr.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return c.Status(fiber.StatusCreated).JSON(dto)
}

func (h *PullRequestHandler) MergePullRequest(c *fiber.Ctx) error {
	repo, err := h.getAuthorizedRepo(c)
	if err != nil {
		return err
	}

	prIDStr := c.Params("pr_id")
	prID, err := strconv.ParseUint(prIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid PR ID"})
	}

	username, _ := c.Locals("username").(string)

	commitHash, err := h.pullSvc.MergePullRequest(c.Context(), uint(prID), repo.Name, username, username+"@gitlink.local")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "merged", "mergeCommitHash": commitHash})
}
