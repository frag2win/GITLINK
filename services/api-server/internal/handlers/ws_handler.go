package handlers

import (
	"bufio"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/localrepo/api-server/internal/middleware"
	"github.com/localrepo/api-server/internal/models"
	"github.com/localrepo/api-server/internal/service"
)

type WSHandler struct {
	hub          service.WebSocketHub
	notifService service.NotificationService
}

func NewWSHandler(hub service.WebSocketHub, notifService service.NotificationService) *WSHandler {
	return &WSHandler{
		hub:          hub,
		notifService: notifService,
	}
}

func (h *WSHandler) StreamNotifications(c *fiber.Ctx) error {
	userID := middleware.UserIDFromContext(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	sendCh := make(chan service.WSMessage, 64)
	h.hub.Register(userID, sendCh)

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer h.hub.Unregister(userID, sendCh)

		// Replay missed notifications if last_event_id is provided in query parameter or header
		lastEventIDStr := c.Query("last_event_id")
		if lastEventIDStr == "" {
			lastEventIDStr = c.Get("Last-Event-ID")
		}
		if lastEventIDStr != "" {
			if lastID, err := strconv.ParseUint(lastEventIDStr, 10, 64); err == nil && lastID > 0 {
				if notifs, err := h.notifService.GetUserNotifications(c.UserContext(), userID, true); err == nil {
					for _, n := range notifs {
						data, _ := json.Marshal(service.WSMessage{
							EventID: uint64(n.ID),
							Event: models.DomainEvent{
								Type:    n.Type,
								UserID:  n.UserID,
								Title:   n.Title,
								Message: n.Message,
								Link:    n.Link,
							},
						})
						_, _ = w.WriteString("data: " + string(data) + "\n\n")
						_ = w.Flush()
					}
				}
			}
		}

		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case msg, ok := <-sendCh:
				if !ok {
					return
				}
				data, _ := json.Marshal(msg)
				_, _ = w.WriteString("data: " + string(data) + "\n\n")
				if err := w.Flush(); err != nil {
					return
				}
			case <-ticker.C:
				_, _ = w.WriteString(": keepalive\n\n")
				if err := w.Flush(); err != nil {
					return
				}
			}
		}
	})

	return nil
}
