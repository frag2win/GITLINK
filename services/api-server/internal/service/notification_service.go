package service

import (
	"context"

	"github.com/localrepo/api-server/internal/models"
	"github.com/localrepo/api-server/internal/repository"
)

type NotificationService interface {
	GetUserNotifications(ctx context.Context, userID uint, unreadOnly bool) ([]models.Notification, error)
	MarkRead(ctx context.Context, id, userID uint) error
}

type notificationService struct {
	notifRepo repository.NotificationRepository
	eventBus  EventBus
}

func NewNotificationService(notifRepo repository.NotificationRepository, eventBus EventBus) NotificationService {
	s := &notificationService{
		notifRepo: notifRepo,
		eventBus:  eventBus,
	}

	if eventBus != nil {
		eventBus.Subscribe(s.handleDomainEvent)
	}

	return s
}

func (s *notificationService) handleDomainEvent(event models.DomainEvent) {
	if event.UserID == 0 {
		return
	}
	ctx := context.Background()
	notif := &models.Notification{
		UserID:  event.UserID,
		Type:    event.Type,
		Title:   event.Title,
		Message: event.Message,
		Link:    event.Link,
		IsRead:  false,
	}
	_ = s.notifRepo.Create(ctx, notif)
}

func (s *notificationService) GetUserNotifications(ctx context.Context, userID uint, unreadOnly bool) ([]models.Notification, error) {
	return s.notifRepo.GetUserNotifications(ctx, userID, unreadOnly)
}

func (s *notificationService) MarkRead(ctx context.Context, id, userID uint) error {
	return s.notifRepo.MarkAsRead(ctx, id, userID)
}
