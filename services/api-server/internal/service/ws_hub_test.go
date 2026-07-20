package service_test

import (
	"testing"
	"time"

	"github.com/localrepo/api-server/internal/models"
	"github.com/localrepo/api-server/internal/service"
)

func TestWebSocketHubMonotonicEventID(t *testing.T) {
	bus := service.NewEventBus()
	hub := service.NewWebSocketHub(bus)

	sendCh := make(chan service.WSMessage, 10)
	hub.Register(1, sendCh)

	event1 := models.DomainEvent{
		Type:      models.NotificationTypePROpened,
		UserID:    1,
		Title:     "Test PR 1",
		Timestamp: time.Now(),
	}

	event2 := models.DomainEvent{
		Type:      models.NotificationTypePRApproved,
		UserID:    1,
		Title:     "Test PR 2",
		Timestamp: time.Now(),
	}

	id1 := hub.Broadcast(event1)
	id2 := hub.Broadcast(event2)

	if id2 <= id1 {
		t.Fatalf("expected monotonic event IDs (id2 > id1), got id1=%d, id2=%d", id1, id2)
	}

	msg1 := <-sendCh
	if msg1.EventID != id1 {
		t.Errorf("expected msg1 EventID %d, got %d", id1, msg1.EventID)
	}

	msg2 := <-sendCh
	if msg2.EventID != id2 {
		t.Errorf("expected msg2 EventID %d, got %d", id2, msg2.EventID)
	}
}
