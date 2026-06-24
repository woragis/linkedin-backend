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

func (r *Repository) RankedFeed(ctx context.Context, userID uuid.UUID, limit int) ([]models.Post, error) {
	if limit <= 0 {
		limit = 50
	}
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Table("user_feed_scores").
		Select("post_id").
		Where("user_id = ?", userID).
		Order("score DESC").
		Limit(limit).
		Scan(&ids).Error
	if err != nil || len(ids) == 0 {
		var rows []models.Post
		err = r.db.WithContext(ctx).Preload("Author").
			Where(`deleted_at IS NULL AND (
				author_id = ? OR author_id IN (
					SELECT CASE WHEN requester_id = ? THEN addressee_id ELSE requester_id END
					FROM connections WHERE status = 'accepted' AND ? IN (requester_id, addressee_id)
				)
			)`, userID, userID, userID).
			Order("created_at DESC").
			Limit(limit).
			Find(&rows).Error
		return rows, err
	}
	var rows []models.Post
	err = r.db.WithContext(ctx).Preload("Author").
		Where("deleted_at IS NULL AND id IN ?", ids).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	// preserve rank order
	byID := make(map[uuid.UUID]models.Post, len(rows))
	for _, p := range rows {
		byID[p.ID] = p
	}
	ordered := make([]models.Post, 0, len(ids))
	for _, id := range ids {
		if p, ok := byID[id]; ok {
			ordered = append(ordered, p)
		}
	}
	return ordered, nil
}

func (r *Repository) AddReaction(ctx context.Context, postID, userID uuid.UUID, kind string) error {
	return r.UpsertContentReaction(ctx, TargetPost, postID, userID, kind)
}

func (r *Repository) ReactionCount(ctx context.Context, postID uuid.UUID) (int64, error) {
	return r.ContentReactionCount(ctx, TargetPost, postID)
}

func (r *Repository) GetCommentByID(ctx context.Context, id uuid.UUID) (*models.Comment, error) {
	var c models.Comment
	err := r.db.WithContext(ctx).
		Preload("Author").
		Where("deleted_at IS NULL").
		First(&c, "id = ?", id).Error
	return &c, err
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
