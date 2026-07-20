package handlers

import (
	"context"
	"strconv"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
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

func (h *WSHandler) UpgradeMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}

func (h *WSHandler) HandleConnection(c *websocket.Conn) {
	userIDVal := c.Locals("userID")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		c.Close()
		return
	}

	sendCh := make(chan service.WSMessage, 64)
	h.hub.Register(userID, sendCh)

	// writerCtx is cancelled when the reader exits (client disconnect or error).
	// This stops the writer goroutine before hub.Unregister closes sendCh.
	writerCtx, cancelWriter := context.WithCancel(context.Background())
	defer cancelWriter()
	defer h.hub.Unregister(userID, sendCh)

	// Replay missed notifications if last_event_id is provided in query parameter
	lastEventIDStr := c.Query("last_event_id")
	if lastEventIDStr != "" {
		if lastID, err := strconv.ParseUint(lastEventIDStr, 10, 64); err == nil && lastID > 0 {
			if notifs, err := h.notifService.GetUserNotifications(context.Background(), userID, true); err == nil {
				for _, n := range notifs {
					// Strict Filter: Only replay notifications that occurred AFTER last_event_id
					if uint64(n.ID) > lastID {
						_ = c.WriteJSON(service.WSMessage{
							EventID: uint64(n.ID),
							Event: models.DomainEvent{
								Type:    n.Type,
								UserID:  n.UserID,
								Title:   n.Title,
								Message: n.Message,
								Link:    n.Link,
							},
						})
					}
				}
			}
		}
	}

	// Writer loop: stops when writerCtx is cancelled or sendCh is closed.
	go func() {
		for {
			select {
			case <-writerCtx.Done():
				return
			case msg, ok := <-sendCh:
				if !ok {
					return
				}
				if err := c.WriteJSON(msg); err != nil {
					cancelWriter()
					return
				}
			}
		}
	}()

	// Reader loop (keepalive / close listener)
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
	// cancelWriter() fires via defer, stopping the writer goroutine cleanly
	// before the hub.Unregister defer closes sendCh.
}
