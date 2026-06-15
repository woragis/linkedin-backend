package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Suggestion struct {
	UserID   uuid.UUID `json:"user_id"`
	Slug     string    `json:"slug"`
	FullName string    `json:"full_name"`
	Headline string    `json:"headline"`
	Score    float64   `json:"score"`
	Rank     int       `json:"rank"`
	Reasons  string    `json:"reasons"`
}

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListForViewer(ctx context.Context, viewerID uuid.UUID, limit int) ([]Suggestion, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	var rows []Suggestion
	err := r.db.WithContext(ctx).Raw(`
SELECT p.user_id, p.slug, p.full_name, p.headline, s.score, s.rank, s.reasons::text
FROM user_connection_suggestions s
JOIN profiles p ON p.user_id = s.suggested_user_id
WHERE s.viewer_id = ?
ORDER BY s.rank ASC
LIMIT ?
`, viewerID, limit).Scan(&rows).Error
	return rows, err
}
