package repository

import (
	"context"
	"errors"
	"time"

	"github.com/localrepo/api-server/internal/models"
	"gorm.io/gorm"
)

type SyncRepository interface {
	Create(ctx context.Context, task *models.SyncTask) error
	GetByUUID(ctx context.Context, uuid string) (*models.SyncTask, error)
	ClaimTask(ctx context.Context, id uint) (bool, error)
	GetNextTasks(ctx context.Context, limit int) ([]models.SyncTask, error)
	MarkCompleted(ctx context.Context, id uint, bytesTransferred int64, durationMs int64) error
	MarkRetry(ctx context.Context, id uint, errStr string, nextRetry time.Time) error
	MarkFailed(ctx context.Context, id uint, errStr string) error
	UpdateHeartbeat(ctx context.Context, id uint) error
	RecoverStaleTasks(ctx context.Context, timeoutThreshold time.Duration) (int64, error)
	ListTasks(ctx context.Context, limit, offset int, status string) ([]models.SyncTask, int64, error)
	GetSyncMetrics(ctx context.Context) (map[string]interface{}, error)
	GetDLQTasks(ctx context.Context) ([]models.SyncTask, error)
	ReplayDLQTask(ctx context.Context, id uint) error
}

type syncRepository struct {
	db *gorm.DB
}

func NewSyncRepository(db *gorm.DB) SyncRepository {
	return &syncRepository{db: db}
}

func (r *syncRepository) Create(ctx context.Context, task *models.SyncTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *syncRepository) GetByUUID(ctx context.Context, uuid string) (*models.SyncTask, error) {
	var task models.SyncTask
	err := r.db.WithContext(ctx).Where("task_uuid = ?", uuid).First(&task).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

func (r *syncRepository) ClaimTask(ctx context.Context, id uint) (bool, error) {
	now := time.Now()
	res := r.db.WithContext(ctx).
		Model(&models.SyncTask{}).
		Where("id = ? AND status IN (?, ?)", id, models.SyncTaskPending, models.SyncTaskRetryScheduled).
		Updates(map[string]interface{}{
			"status":         models.SyncTaskRunning,
			"started_at":     now,
			"last_heartbeat": now,
		})

	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *syncRepository) GetNextTasks(ctx context.Context, limit int) ([]models.SyncTask, error) {
	var tasks []models.SyncTask
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("(status = ?) OR (status = ? AND (next_retry_at IS NULL OR next_retry_at <= ?))",
			models.SyncTaskPending, models.SyncTaskRetryScheduled, now).
		Order("priority DESC, next_retry_at ASC, created_at ASC").
		Limit(limit).
		Find(&tasks).Error

	return tasks, err
}

func (r *syncRepository) MarkCompleted(ctx context.Context, id uint, bytesTransferred int64, durationMs int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.SyncTask{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":            models.SyncTaskCompleted,
			"completed_at":      now,
			"bytes_transferred": bytesTransferred,
			"duration_ms":       durationMs,
			"last_error":        "",
		}).Error
}

func (r *syncRepository) MarkRetry(ctx context.Context, id uint, errStr string, nextRetry time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.SyncTask{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        models.SyncTaskRetryScheduled,
			"attempt_count": gorm.Expr("attempt_count + 1"),
			"last_error":    errStr,
			"next_retry_at": nextRetry,
		}).Error
}

func (r *syncRepository) MarkFailed(ctx context.Context, id uint, errStr string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.SyncTask{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        models.SyncTaskFailed,
			"attempt_count": gorm.Expr("attempt_count + 1"),
			"last_error":    errStr,
			"completed_at":  now,
		}).Error
}

func (r *syncRepository) UpdateHeartbeat(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.SyncTask{}).
		Where("id = ? AND status = ?", id, models.SyncTaskRunning).
		Update("last_heartbeat", now).Error
}

func (r *syncRepository) RecoverStaleTasks(ctx context.Context, timeoutThreshold time.Duration) (int64, error) {
	threshold := time.Now().Add(-timeoutThreshold)
	res := r.db.WithContext(ctx).
		Model(&models.SyncTask{}).
		Where("status = ? AND last_heartbeat < ?", models.SyncTaskRunning, threshold).
		Updates(map[string]interface{}{
			"status":     models.SyncTaskPending,
			"last_error": "Task execution heartbeat timed out; returned to pending pool",
		})
	return res.RowsAffected, res.Error
}

func (r *syncRepository) ListTasks(ctx context.Context, limit, offset int, status string) ([]models.SyncTask, int64, error) {
	var tasks []models.SyncTask
	var total int64

	query := r.db.WithContext(ctx).Model(&models.SyncTask{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&tasks).Error
	return tasks, total, err
}

func (r *syncRepository) GetSyncMetrics(ctx context.Context) (map[string]interface{}, error) {
	var totalTasks, completedTasks, failedTasks, pendingTasks int64
	var totalBytes int64
	var avgDuration float64

	db := r.db.WithContext(ctx)
	db.Model(&models.SyncTask{}).Count(&totalTasks)
	db.Model(&models.SyncTask{}).Where("status = ?", models.SyncTaskCompleted).Count(&completedTasks)
	db.Model(&models.SyncTask{}).Where("status = ?", models.SyncTaskFailed).Count(&failedTasks)
	db.Model(&models.SyncTask{}).Where("status IN (?, ?, ?)", models.SyncTaskPending, models.SyncTaskRunning, models.SyncTaskRetryScheduled).Count(&pendingTasks)

	db.Model(&models.SyncTask{}).Where("status = ?", models.SyncTaskCompleted).Select("COALESCE(SUM(bytes_transferred), 0)").Scan(&totalBytes)
	db.Model(&models.SyncTask{}).Where("status = ?", models.SyncTaskCompleted).Select("COALESCE(AVG(duration_ms), 0)").Scan(&avgDuration)

	var successRate float64
	if totalTasks > 0 {
		successRate = (float64(completedTasks) / float64(totalTasks)) * 100.0
	}

	return map[string]interface{}{
		"total_tasks":       totalTasks,
		"completed_tasks":   completedTasks,
		"failed_tasks":      failedTasks,
		"pending_tasks":     pendingTasks,
		"total_bytes":       totalBytes,
		"avg_duration_ms":   avgDuration,
		"success_rate_pct":  successRate,
	}, nil
}

func (r *syncRepository) GetDLQTasks(ctx context.Context) ([]models.SyncTask, error) {
	var tasks []models.SyncTask
	err := r.db.WithContext(ctx).
		Where("status = ?", models.SyncTaskFailed).
		Order("id DESC").
		Limit(100).
		Find(&tasks).Error
	return tasks, err
}

func (r *syncRepository) ReplayDLQTask(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.SyncTask{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        models.SyncTaskPending,
			"attempt_count": 0,
			"last_error":    "Manually replayed from DLQ",
			"next_retry_at": nil,
		}).Error
}
