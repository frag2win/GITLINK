package service

import (
	"context"

	"github.com/localrepo/api-server/internal/models"
	"github.com/localrepo/api-server/internal/repository"
)

// AuditService handles logging of critical system events.
type AuditService interface {
	LogAction(ctx context.Context, action, repoName, actor, details string) error
}

type auditService struct {
	repo repository.AuditRepository
}

func NewAuditService(repo repository.AuditRepository) AuditService {
	return &auditService{repo: repo}
}

func (s *auditService) LogAction(ctx context.Context, action, repoName, actor, details string) error {
	log := &models.AuditLog{
		Operation: action,
		RepoName:  repoName,
		PeerID:    actor,
		Details:   details,
	}
	return s.repo.Create(ctx, log)
}
