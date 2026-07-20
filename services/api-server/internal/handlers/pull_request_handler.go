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

type SubmitReviewRequest struct {
	State             models.ReviewState                `json:"state"`
	Body              string                            `json:"body"`
	ReviewedCommitSHA string                            `json:"reviewed_commit_sha"`
	Comments          []models.PullRequestReviewComment `json:"comments"`
}

func (h *PullRequestHandler) SubmitReview(c *fiber.Ctx) error {
	userID := middleware.UserIDFromContext(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	prIDStr := c.Params("pr_id")
	prID, err := strconv.ParseUint(prIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid PR ID"})
	}

	var req SubmitReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	review, err := h.pullSvc.SubmitReview(c.Context(), uint(prID), userID, req.State, req.Body, req.ReviewedCommitSHA, req.Comments)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(review)
}

func (h *PullRequestHandler) GetReviews(c *fiber.Ctx) error {
	prIDStr := c.Params("pr_id")
	prID, err := strconv.ParseUint(prIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid PR ID"})
	}

	reviews, err := h.pullSvc.GetReviews(c.Context(), uint(prID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"reviews": reviews})
}

func (h *PullRequestHandler) ResolveThread(c *fiber.Ctx) error {
	userID := middleware.UserIDFromContext(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	threadIDStr := c.Params("thread_id")
	threadID, err := strconv.ParseUint(threadIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Thread ID"})
	}

	if err := h.pullSvc.ResolveThread(c.Context(), uint(threadID), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Thread resolved"})
}

type MergePRRequest struct {
	HeadCommitSHA string `json:"head_commit_sha"`
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

	var req MergePRRequest
	_ = c.BodyParser(&req)

	username, _ := c.Locals("username").(string)

	commitHash, err := h.pullSvc.MergePullRequest(c.Context(), uint(prID), repo.Name, username, username+"@gitlink.local", req.HeadCommitSHA)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "merged", "mergeCommitHash": commitHash})
}
