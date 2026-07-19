package service

import (
	"context"

	"github.com/localrepo/api-server/internal/repository"
)

type HealthService interface {
	CheckHealth(ctx context.Context) map[string]string
}

type healthService struct {
	repo repository.HealthRepository
}

func NewHealthService(repo repository.HealthRepository) HealthService {
	return &healthService{repo: repo}
}

func (s *healthService) CheckHealth(ctx context.Context) map[string]string {
	status := map[string]string{
		"database": "ok",
		"status":   "ok",
	}

	if err := s.repo.Ping(ctx); err != nil {
		status["database"] = "error"
		status["status"] = "degraded"
	}

	return status
}
