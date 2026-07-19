package repository

import (
	"context"

	"github.com/localrepo/api-server/internal/models"
	"gorm.io/gorm"
)

type AuditRepository interface {
	Create(ctx context.Context, log *models.AuditLog) error
}

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(ctx context.Context, log *models.AuditLog) error {
	return getDB(ctx, r.db).Create(log).Error
}
