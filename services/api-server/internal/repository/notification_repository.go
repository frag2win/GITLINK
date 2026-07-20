package repository

import (
	"context"

	"github.com/localrepo/api-server/internal/models"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(ctx context.Context, notif *models.Notification) error
	GetUserNotifications(ctx context.Context, userID uint, unreadOnly bool) ([]models.Notification, error)
	MarkAsRead(ctx context.Context, id, userID uint) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notif *models.Notification) error {
	return r.db.WithContext(ctx).Create(notif).Error
}

func (r *notificationRepository) GetUserNotifications(ctx context.Context, userID uint, unreadOnly bool) ([]models.Notification, error) {
	var notifs []models.Notification
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}
	err := query.Order("id DESC").Limit(50).Find(&notifs).Error
	return notifs, err
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id, userID uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}
