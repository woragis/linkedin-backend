package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PersonHit struct {
	UserID   uuid.UUID `json:"user_id"`
	Slug     string    `json:"slug"`
	FullName string    `json:"full_name"`
	Headline string    `json:"headline"`
	Location string    `json:"location"`
	Score    float64   `json:"score"`
}

type PostHit struct {
	PostID     uuid.UUID `json:"post_id"`
	AuthorID   uuid.UUID `json:"author_id"`
	AuthorName string    `json:"author_name"`
	Body       string    `json:"body"`
	Score      float64   `json:"score"`
}

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SearchPeople(ctx context.Context, q string, limit int) ([]PersonHit, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	pattern := "%" + strings.TrimSpace(q) + "%"
	rows, err := r.db.WithContext(ctx).Raw(`
SELECT p.user_id, p.slug, p.full_name, p.headline, p.location, 1.0::float8 AS score
FROM profiles p
WHERE p.full_name ILIKE ? OR p.headline ILIKE ? OR p.bio ILIKE ? OR p.location ILIKE ?
ORDER BY p.full_name
LIMIT ?
`, pattern, pattern, pattern, pattern, limit).Rows()
	if err != nil {
		return nil, fmt.Errorf("search people: %w", err)
	}
	defer rows.Close()
	var out []PersonHit
	for rows.Next() {
		var h PersonHit
		if err := rows.Scan(&h.UserID, &h.Slug, &h.FullName, &h.Headline, &h.Location, &h.Score); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, nil
}

func (r *Repository) SearchPosts(ctx context.Context, q string, limit int) ([]PostHit, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	pattern := "%" + strings.TrimSpace(q) + "%"
	rows, err := r.db.WithContext(ctx).Raw(`
SELECT po.id, po.author_id, pr.full_name, po.body, 1.0::float8 AS score
FROM posts po
JOIN profiles pr ON pr.user_id = po.author_id
WHERE po.deleted_at IS NULL AND po.body ILIKE ?
ORDER BY po.created_at DESC
LIMIT ?
`, pattern, limit).Rows()
	if err != nil {
		return nil, fmt.Errorf("search posts: %w", err)
	}
	defer rows.Close()
	var out []PostHit
	for rows.Next() {
		var h PostHit
		if err := rows.Scan(&h.PostID, &h.AuthorID, &h.AuthorName, &h.Body, &h.Score); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, nil
}

func (r *Repository) AffinityBoost(ctx context.Context, viewerID uuid.UUID, targets []uuid.UUID) (map[uuid.UUID]float64, error) {
	out := make(map[uuid.UUID]float64, len(targets))
	if viewerID == uuid.Nil || len(targets) == 0 {
		return out, nil
	}
	type row struct {
		TargetID uuid.UUID
		Score    float64
	}
	var rows []row
	err := r.db.WithContext(ctx).Table("user_pair_affinity").
		Select("target_id, score").
		Where("viewer_id = ? AND target_id IN ?", viewerID, targets).
		Scan(&rows).Error
	if err != nil {
		return out, nil
	}
	for _, rw := range rows {
		out[rw.TargetID] = rw.Score
	}
	return out, nil
}
