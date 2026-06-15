package service

import (
	"context"

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
