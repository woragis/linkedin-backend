package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/models"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, p *models.Post) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	var p models.Post
	err := r.db.WithContext(ctx).
		Preload("Author").
		Where("deleted_at IS NULL").
		First(&p, "id = ?", id).Error
	return &p, err
}

func (r *Repository) Feed(ctx context.Context, userID uuid.UUID, peerIDs []uuid.UUID, limit int) ([]models.Post, error) {
	authorIDs := append(peerIDs, userID)
	var rows []models.Post
	q := r.db.WithContext(ctx).
		Preload("Author").
		Where("deleted_at IS NULL AND author_id IN ?", authorIDs).
		Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&rows).Error
	return rows, err
}

func (r *Repository) AddReaction(ctx context.Context, postID, userID uuid.UUID, kind string) error {
	rxn := models.Reaction{PostID: postID, UserID: userID, Kind: kind}
	return r.db.WithContext(ctx).Save(&rxn).Error
}

func (r *Repository) ReactionCount(ctx context.Context, postID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.Reaction{}).Where("post_id = ?", postID).Count(&n).Error
	return n, err
}

func (r *Repository) AddComment(ctx context.Context, c *models.Comment) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *Repository) ListComments(ctx context.Context, postID uuid.UUID) ([]models.Comment, error) {
	var rows []models.Comment
	err := r.db.WithContext(ctx).Preload("Author").
		Where("post_id = ? AND deleted_at IS NULL", postID).
		Order("created_at ASC").Find(&rows).Error
	return rows, err
}

func (r *Repository) CommentCount(ctx context.Context, postID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.Comment{}).
		Where("post_id = ? AND deleted_at IS NULL", postID).Count(&n).Error
	return n, err
}
