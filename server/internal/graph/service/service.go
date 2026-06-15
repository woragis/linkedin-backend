package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	graphrepo "github.com/unipe/linkedin/backend/server/internal/graph/repository"
)

type Service struct {
	repo *graphrepo.Repository
}

func New(repo *graphrepo.Repository) *Service {
	return &Service{repo: repo}
}

type GraphView struct {
	Nodes []graphrepo.GraphNode `json:"nodes"`
	Edges []graphrepo.GraphEdge `json:"edges"`
}

func (s *Service) UserGraph(ctx context.Context, userID uuid.UUID) (*GraphView, error) {
	nodes, edges, err := s.repo.Subgraph(ctx, userID)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return &GraphView{Nodes: nodes, Edges: edges}, nil
}

func (s *Service) TopInfluencers(ctx context.Context, limit int) ([]graphrepo.GraphNode, error) {
	rows, err := s.repo.TopInfluencers(ctx, limit)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return rows, nil
}

type LinkPredictionView struct {
	UserID   uuid.UUID `json:"user_id"`
	Slug     string    `json:"slug"`
	FullName string    `json:"full_name"`
	Headline string    `json:"headline"`
	Score    float64   `json:"score"`
	Reasons  []string  `json:"reasons"`
}

func (s *Service) LinkPredictions(ctx context.Context, viewerID uuid.UUID, limit int) ([]LinkPredictionView, error) {
	rows, err := s.repo.LinkPredictions(ctx, viewerID, limit)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	out := make([]LinkPredictionView, 0, len(rows))
	for _, r := range rows {
		var reasons []string
		if r.Reasons != "" {
			_ = json.Unmarshal([]byte(r.Reasons), &reasons)
		}
		out = append(out, LinkPredictionView{
			UserID:   r.UserID,
			Slug:     r.Slug,
			FullName: r.FullName,
			Headline: r.Headline,
			Score:    r.Score,
			Reasons:  reasons,
		})
	}
	return out, nil
}
