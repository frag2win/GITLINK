package service

import (
	"sync"

	"github.com/localrepo/api-server/internal/models"
)

type EventListener func(event models.DomainEvent)

type EventBus interface {
	Publish(event models.DomainEvent)
	Subscribe(listener EventListener)
}

type eventBus struct {
	mu        sync.RWMutex
	listeners []EventListener
}

func NewEventBus() EventBus {
	return &eventBus{
		listeners: make([]EventListener, 0),
	}
}

func (b *eventBus) Subscribe(listener EventListener) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.listeners = append(b.listeners, listener)
}

func (b *eventBus) Publish(event models.DomainEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, l := range b.listeners {
		go func(listener EventListener) {
			defer func() {
				_ = recover() // Prevents unhandled listener panics from crashing api-server
			}()
			listener(event)
		}(l)
	}
}
