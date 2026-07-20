package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/localrepo/api-server/internal/service"
)

type NotificationHandler struct {
	notifService service.NotificationService
}

func NewNotificationHandler(notifService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifService: notifService}
}

func (h *NotificationHandler) GetNotifications(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	userID, ok := userIDVal.(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	unreadOnly := c.Query("unread") == "true"
	notifs, err := h.notifService.GetUserNotifications(c.UserContext(), userID, unreadOnly)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"notifications": notifs})
}

func (h *NotificationHandler) MarkRead(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	userID, ok := userIDVal.(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	idStr := c.Params("id")
	id, _ := strconv.ParseUint(idStr, 10, 64)

	if err := h.notifService.MarkRead(c.UserContext(), uint(id), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Notification marked as read"})
}
