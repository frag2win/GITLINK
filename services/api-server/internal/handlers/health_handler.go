package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/localrepo/api-server/internal/service"
)

type HealthHandler struct {
	healthService service.HealthService
}

func NewHealthHandler(healthService service.HealthService) *HealthHandler {
	return &HealthHandler{healthService: healthService}
}

func (h *HealthHandler) CheckHealth(c *fiber.Ctx) error {
	status := h.healthService.CheckHealth(c.Context())

	if status["status"] != "ok" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(status)
	}
	return c.Status(fiber.StatusOK).JSON(status)
}
