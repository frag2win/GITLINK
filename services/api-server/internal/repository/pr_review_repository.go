package repository

import (
	"context"

	"github.com/localrepo/api-server/internal/models"
	"gorm.io/gorm"
)

type PRReviewRepository interface {
	CreateReview(ctx context.Context, review *models.PullRequestReview) error
	GetReviewsByPR(ctx context.Context, prID uint) ([]models.PullRequestReview, error)
	GetLatestReviewByReviewer(ctx context.Context, prID, reviewerID uint) (*models.PullRequestReview, error)
	CreateThread(ctx context.Context, thread *models.ReviewThread) error
	GetThreadsByPR(ctx context.Context, prID uint) ([]models.ReviewThread, error)
	ResolveThread(ctx context.Context, threadID, userID uint) error
	AddThreadComment(ctx context.Context, comment *models.PullRequestReviewComment) error
}

type prReviewRepository struct {
	db *gorm.DB
}

func NewPRReviewRepository(db *gorm.DB) PRReviewRepository {
	return &prReviewRepository{db: db}
}

func (r *prReviewRepository) CreateReview(ctx context.Context, review *models.PullRequestReview) error {
	return r.db.WithContext(ctx).Create(review).Error
}

func (r *prReviewRepository) GetReviewsByPR(ctx context.Context, prID uint) ([]models.PullRequestReview, error) {
	var reviews []models.PullRequestReview
	err := r.db.WithContext(ctx).
		Preload("Reviewer").
		Preload("Comments").
		Where("pull_request_id = ?", prID).
		Order("id ASC").
		Find(&reviews).Error
	return reviews, err
}

func (r *prReviewRepository) GetLatestReviewByReviewer(ctx context.Context, prID, reviewerID uint) (*models.PullRequestReview, error) {
	var review models.PullRequestReview
	err := r.db.WithContext(ctx).
		Where("pull_request_id = ? AND reviewer_id = ?", prID, reviewerID).
		Order("id DESC").
		First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *prReviewRepository) CreateThread(ctx context.Context, thread *models.ReviewThread) error {
	return r.db.WithContext(ctx).Create(thread).Error
}

func (r *prReviewRepository) GetThreadsByPR(ctx context.Context, prID uint) ([]models.ReviewThread, error) {
	var threads []models.ReviewThread
	err := r.db.WithContext(ctx).
		Preload("Comments.Author").
		Preload("ResolvedBy").
		Where("pull_request_id = ?", prID).
		Order("id ASC").
		Find(&threads).Error
	return threads, err
}

func (r *prReviewRepository) ResolveThread(ctx context.Context, threadID, userID uint) error {
	return r.db.WithContext(ctx).
		Model(&models.ReviewThread{}).
		Where("id = ?", threadID).
		Updates(map[string]interface{}{
			"is_resolved":    true,
			"resolved_by_id": userID,
		}).Error
}

func (r *prReviewRepository) AddThreadComment(ctx context.Context, comment *models.PullRequestReviewComment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}
