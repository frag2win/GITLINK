package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/localrepo/api-server/internal/repository"
	"github.com/localrepo/api-server/internal/service"
)

var startTime = time.Now()

type HealthHandler struct {
	healthService service.HealthService
	syncRepo      repository.SyncRepository
	gitSvc        service.GitService
	peerSvc       service.PeerService
}

func NewHealthHandler(healthService service.HealthService, syncRepo repository.SyncRepository, gitSvc service.GitService, peerSvc service.PeerService) *HealthHandler {
	return &HealthHandler{
		healthService: healthService,
		syncRepo:      syncRepo,
		gitSvc:        gitSvc,
		peerSvc:       peerSvc,
	}
}

// Liveness endpoint: /health
func (h *HealthHandler) Liveness(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "alive",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// Readiness endpoint: /ready
func (h *HealthHandler) Readiness(c *fiber.Ctx) error {
	status := h.healthService.CheckHealth(c.Context())
	if status["status"] != "ok" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(status)
	}
	return c.JSON(fiber.Map{
		"status": "ready",
		"db":     "ok",
	})
}

// DeepDiagnostics endpoint: /api/v1/health/deep
func (h *HealthHandler) DeepDiagnostics(c *fiber.Ctx) error {
	startDB := time.Now()
	status := h.healthService.CheckHealth(c.UserContext())
	dbLatency := time.Since(startDB).Milliseconds()

	// Live Git Server check
	startGit := time.Now()
	gitStatus := "healthy"
	if _, err := h.gitSvc.ListRepositories(c.UserContext()); err != nil {
		gitStatus = "unhealthy"
	}
	gitLatency := time.Since(startGit).Milliseconds()

	// Live libp2p node check
	startP2P := time.Now()
	p2pStatus := "healthy"
	if _, err := h.peerSvc.GetConnectedPeers(c.UserContext()); err != nil {
		p2pStatus = "unhealthy"
	}
	p2pLatency := time.Since(startP2P).Milliseconds()

	pendingCount := int64(0)
	dlqCount := int64(0)
	if h.syncRepo != nil {
		if tasks, err := h.syncRepo.GetNextTasks(c.UserContext(), 100); err == nil {
			pendingCount = int64(len(tasks))
		}
		if dlqTasks, err := h.syncRepo.GetDLQTasks(c.UserContext()); err == nil {
			dlqCount = int64(len(dlqTasks))
		}
	}

	return c.JSON(fiber.Map{
		"status": "healthy",
		"database": fiber.Map{
			"status":     status["status"],
			"latency_ms": dbLatency,
		},
		"git_server": fiber.Map{
			"status":     gitStatus,
			"latency_ms": gitLatency,
		},
		"libp2p": fiber.Map{
			"status":     p2pStatus,
			"latency_ms": p2pLatency,
		},
		"sync_worker": fiber.Map{
			"status":      "running",
			"queue_depth": pendingCount,
			"pending_dlq": dlqCount,
		},
		"uptime_seconds": int64(time.Since(startTime).Seconds()),
	})
}
