package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GraphNode struct {
	UserID      uuid.UUID `json:"user_id"`
	Slug        string    `json:"slug"`
	FullName    string    `json:"full_name"`
	Headline    string    `json:"headline"`
	PageRank    float64   `json:"pagerank"`
	Degree      int       `json:"degree"`
	CommunityID *int      `json:"community_id,omitempty"`
}

type GraphEdge struct {
	Source uuid.UUID `json:"source"`
	Target uuid.UUID `json:"target"`
}

type LinkPrediction struct {
	UserID   uuid.UUID `json:"user_id"`
	Slug     string    `json:"slug"`
	FullName string    `json:"full_name"`
	Headline string    `json:"headline"`
	Score    float64   `json:"score"`
	Reasons  string    `json:"reasons"`
}

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Subgraph(ctx context.Context, userID uuid.UUID) ([]GraphNode, []GraphEdge, error) {
	var nodes []GraphNode
	err := r.db.WithContext(ctx).Raw(`
SELECT p.user_id, p.slug, p.full_name, p.headline,
       COALESCE(gm.pagerank, 0), COALESCE(gm.degree, 0), gm.community_id
FROM profiles p
LEFT JOIN user_graph_metrics gm ON gm.user_id = p.user_id
WHERE p.user_id = ? OR p.user_id IN (
  SELECT CASE WHEN requester_id = ? THEN addressee_id ELSE requester_id END
  FROM connections WHERE status = 'accepted' AND ? IN (requester_id, addressee_id)
)
`, userID, userID, userID).Scan(&nodes).Error
	if err != nil {
		return nil, nil, err
	}

	var edges []GraphEdge
	err = r.db.WithContext(ctx).Raw(`
SELECT requester_id, addressee_id FROM connections
WHERE status = 'accepted' AND (
  requester_id IN (
    SELECT user_id FROM profiles p WHERE p.user_id = ? OR p.user_id IN (
      SELECT CASE WHEN requester_id = ? THEN addressee_id ELSE requester_id END
      FROM connections WHERE status = 'accepted' AND ? IN (requester_id, addressee_id)
    )
  )
  AND addressee_id IN (
    SELECT user_id FROM profiles p WHERE p.user_id = ? OR p.user_id IN (
      SELECT CASE WHEN requester_id = ? THEN addressee_id ELSE requester_id END
      FROM connections WHERE status = 'accepted' AND ? IN (requester_id, addressee_id)
    )
  )
)
`, userID, userID, userID, userID, userID, userID).Scan(&edges).Error
	return nodes, edges, err
}

func (r *Repository) TopInfluencers(ctx context.Context, limit int) ([]GraphNode, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	var nodes []GraphNode
	err := r.db.WithContext(ctx).Raw(`
SELECT p.user_id, p.slug, p.full_name, p.headline, gm.pagerank, gm.degree, gm.community_id
FROM user_graph_metrics gm
JOIN profiles p ON p.user_id = gm.user_id
ORDER BY gm.pagerank DESC
LIMIT ?
`, limit).Scan(&nodes).Error
	return nodes, err
}

func (r *Repository) LinkPredictions(ctx context.Context, viewerID uuid.UUID, limit int) ([]LinkPrediction, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	var rows []LinkPrediction
	err := r.db.WithContext(ctx).Raw(`
SELECT p.user_id, p.slug, p.full_name, p.headline, a.score, a.reasons::text
FROM user_pair_affinity a
JOIN profiles p ON p.user_id = a.target_id
WHERE a.viewer_id = ?
  AND NOT EXISTS (
    SELECT 1 FROM connections c
    WHERE c.status = 'accepted'
      AND ((c.requester_id = ? AND c.addressee_id = a.target_id)
        OR (c.addressee_id = ? AND c.requester_id = a.target_id))
  )
ORDER BY a.score DESC
LIMIT ?
`, viewerID, viewerID, viewerID, limit).Scan(&rows).Error
	return rows, err
}
