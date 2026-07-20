package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/localrepo/api-server/internal/models"
	"github.com/localrepo/api-server/internal/service"
)

type TeamHandler struct {
	teamService service.TeamService
}

func NewTeamHandler(teamService service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

type CreateOrgRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *TeamHandler) CreateOrganization(c *fiber.Ctx) error {
	var req CreateOrgRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	userIDVal := c.Locals("userID")
	userID, ok := userIDVal.(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	org, err := h.teamService.CreateOrganization(c.UserContext(), req.Name, req.Description, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(org)
}

type CreateTeamRequest struct {
	OrganizationID uint   `json:"organization_id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
}

func (h *TeamHandler) CreateTeam(c *fiber.Ctx) error {
	var req CreateTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	team, err := h.teamService.CreateTeam(c.UserContext(), req.OrganizationID, req.Name, req.Description)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(team)
}

type AddMemberRequest struct {
	UserID uint            `json:"user_id"`
	Role   models.TeamRole `json:"role"`
}

func (h *TeamHandler) AddMember(c *fiber.Ctx) error {
	teamID, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	var req AddMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if err := h.teamService.AddMember(c.UserContext(), uint(teamID), req.UserID, req.Role); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Member added to team"})
}
