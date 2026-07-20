package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/localrepo/api-server/internal/repository"
	"github.com/localrepo/api-server/internal/service"
)

type SyncHandler struct {
	syncService service.SyncService
	peerService service.PeerService
	syncRepo    repository.SyncRepository
}

func NewSyncHandler(syncService service.SyncService, peerService service.PeerService, syncRepo repository.SyncRepository) *SyncHandler {
	return &SyncHandler{
		syncService: syncService,
		peerService: peerService,
		syncRepo:    syncRepo,
	}
}

func (h *SyncHandler) GetPeers(c *fiber.Ctx) error {
	peers, err := h.peerService.GetConnectedPeers(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch connected peers: " + err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"peers": peers,
		"total": len(peers),
	})
}

func (h *SyncHandler) GetQueue(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	status := c.Query("status", "")

	tasks, total, err := h.syncRepo.ListTasks(c.UserContext(), limit, offset, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list sync queue tasks: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"tasks":  tasks,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *SyncHandler) GetMetrics(c *fiber.Ctx) error {
	metrics, err := h.syncService.GetMetrics(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch sync metrics: " + err.Error(),
		})
	}
	return c.JSON(metrics)
}

func (h *SyncHandler) RetryTask(c *fiber.Ctx) error {
	taskIDStr := c.Params("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid task ID",
		})
	}

	if err := h.syncService.RetryTask(c.UserContext(), uint(taskID)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retry task: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Task queued for immediate retry",
		"task_id": taskID,
	})
}

type TriggerSyncRequest struct {
	RepositoryID uint   `json:"repository_id"`
	RepoName     string `json:"repo_name"`
	TargetPeerID string `json:"target_peer_id"`
	Priority     int    `json:"priority"`
}

func (h *SyncHandler) TriggerSync(c *fiber.Ctx) error {
	var req TriggerSyncRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	if req.RepoName == "" || req.TargetPeerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "repo_name and target_peer_id are required",
		})
	}

	task, err := h.syncService.EnqueueSync(c.UserContext(), req.RepositoryID, req.RepoName, req.TargetPeerID, req.Priority, "")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to trigger sync: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Sync task enqueued",
		"task":    task,
	})
}
