package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/localrepo/api-server/internal/middleware"
	"github.com/localrepo/api-server/internal/models"
	"github.com/localrepo/api-server/internal/repository"
	"github.com/localrepo/api-server/internal/service"
	"gorm.io/gorm"
)

type ContributorHandler struct {
	repoService      service.RepoService
	contributorRepo  repository.ContributorRepository
	userRepo         repository.UserRepository
	authorizationSvc service.AuthorizationService
}

func NewContributorHandler(
	repoService service.RepoService,
	contributorRepo repository.ContributorRepository,
	userRepo repository.UserRepository,
	authorizationSvc service.AuthorizationService,
) *ContributorHandler {
	return &ContributorHandler{
		repoService:      repoService,
		contributorRepo:  contributorRepo,
		userRepo:         userRepo,
		authorizationSvc: authorizationSvc,
	}
}

type ContributorDTO struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	AvatarURL    string `json:"avatarUrl"`
	SSHPublicKey string `json:"sshPublicKey"`
	JoinedAt     string `json:"joinedAt"`
}

func (h *ContributorHandler) getAuthorizedRepo(c *fiber.Ctx) (*models.Repository, error) {
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

	isAllowed := false
	var authErr error
	if c.Method() == fiber.MethodGet {
		isAllowed, authErr = h.authorizationSvc.AuthorizeRead(c.Context(), userID, repo.ID)
	} else {
		isAllowed, authErr = h.authorizationSvc.AuthorizeWrite(c.Context(), userID, repo.ID)
	}

	if authErr != nil || !isAllowed {
		return nil, c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden: Access denied"})
	}

	return repo, nil
}

func (h *ContributorHandler) ListContributors(c *fiber.Ctx) error {
	repo, err := h.getAuthorizedRepo(c)
	if err != nil {
		return err
	}

	collaborators, err := h.contributorRepo.ListCollaborators(c.Context(), repo.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	dtos := make([]ContributorDTO, 0, len(collaborators)+1)
	
	// Always include owner as admin/owner
	ownerUser, err := h.userRepo.FindByID(repo.OwnerID)
	if err == nil {
		dtos = append(dtos, ContributorDTO{
			ID:       strconv.Itoa(int(ownerUser.ID)),
			Name:     ownerUser.Username,
			Email:    ownerUser.Email,
			Role:     "admin", // Owner acts as admin
			JoinedAt: repo.CreatedAt.Format(time.RFC3339),
		})
	}

	for _, col := range collaborators {
		// Skip owner — already included above as admin
		if col.UserID == repo.OwnerID {
			continue
		}
		user, err := h.userRepo.FindByID(col.UserID)
		if err != nil {
			continue
		}
		dtos = append(dtos, ContributorDTO{
			ID:       strconv.Itoa(int(user.ID)),
			Name:     user.Username,
			Email:    user.Email,
			Role:     col.Role,
			JoinedAt: col.CreatedAt.Format(time.RFC3339),
		})
	}

	return c.JSON(dtos)
}

func (h *ContributorHandler) AddContributor(c *fiber.Ctx) error {
	repo, err := h.getAuthorizedRepo(c)
	if err != nil {
		return err
	}

	var req struct {
		Name         string `json:"name"`
		Email        string `json:"email"`
		Role         string `json:"role"`
		SSHPublicKey string `json:"sshPublicKey"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Find the user to add
	user, err := h.userRepo.FindByUsername(req.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": fmt.Sprintf("User '%s' not found", req.Name)})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Create collaborator link
	col := &models.RepositoryCollaborator{
		UserID:       user.ID,
		RepositoryID: repo.ID,
		Role:         req.Role,
	}

	if err := h.contributorRepo.AddCollaborator(c.Context(), col); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	dto := ContributorDTO{
		ID:       strconv.Itoa(int(user.ID)),
		Name:     user.Username,
		Email:    user.Email,
		Role:     col.Role,
		JoinedAt: time.Now().Format(time.RFC3339),
	}

	return c.Status(fiber.StatusCreated).JSON(dto)
}

func (h *ContributorHandler) RemoveContributor(c *fiber.Ctx) error {
	repo, err := h.getAuthorizedRepo(c)
	if err != nil {
		return err
	}

	peerIDStr := c.Params("peerId")
	contributorID, err := strconv.ParseUint(peerIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid contributor ID"})
	}

	if err := h.contributorRepo.RemoveCollaborator(c.Context(), repo.ID, uint(contributorID)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
