package service

import (
	"sync"
	"sync/atomic"

	"github.com/localrepo/api-server/internal/models"
)

type WSMessage struct {
	EventID uint64             `json:"event_id"`
	Event   models.DomainEvent `json:"event"`
}

type WebSocketHub interface {
	Register(userID uint, sendCh chan WSMessage)
	Unregister(userID uint, sendCh chan WSMessage)
	Broadcast(event models.DomainEvent) uint64
	GetNextEventID() uint64
}

type webSocketHub struct {
	mu           sync.RWMutex
	clients      map[uint][]chan WSMessage
	eventCounter uint64
	eventBus     EventBus
}

func NewWebSocketHub(eventBus EventBus) WebSocketHub {
	h := &webSocketHub{
		clients:  make(map[uint][]chan WSMessage),
		eventBus: eventBus,
	}

	if eventBus != nil {
		eventBus.Subscribe(h.handleDomainEvent)
	}

	return h
}

func (h *webSocketHub) GetNextEventID() uint64 {
	return atomic.AddUint64(&h.eventCounter, 1)
}

func (h *webSocketHub) Register(userID uint, sendCh chan WSMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[userID] = append(h.clients[userID], sendCh)
}

func (h *webSocketHub) Unregister(userID uint, sendCh chan WSMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()
	channels := h.clients[userID]
	for i, ch := range channels {
		if ch == sendCh {
			h.clients[userID] = append(channels[:i], channels[i+1:]...)
			close(ch)
			break
		}
	}
	if len(h.clients[userID]) == 0 {
		delete(h.clients, userID)
	}
}

func (h *webSocketHub) handleDomainEvent(event models.DomainEvent) {
	h.Broadcast(event)
}

func (h *webSocketHub) Broadcast(event models.DomainEvent) uint64 {
	eventID := h.GetNextEventID()
	msg := WSMessage{
		EventID: eventID,
		Event:   event,
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if event.UserID != 0 {
		if chans, ok := h.clients[event.UserID]; ok {
			for _, ch := range chans {
				select {
				case ch <- msg:
				default:
				}
			}
		}
	} else {
		// Broadcast to all connected clients
		for _, chans := range h.clients {
			for _, ch := range chans {
				select {
				case ch <- msg:
				default:
				}
			}
		}
	}

	return eventID
}
