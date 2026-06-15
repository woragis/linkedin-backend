package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	recorepo "github.com/unipe/linkedin/backend/server/internal/recommendation/repository"
)

type Service struct {
	repo *recorepo.Repository
}

func New(repo *recorepo.Repository) *Service {
	return &Service{repo: repo}
}

type PersonSuggestion struct {
	UserID   uuid.UUID `json:"user_id"`
	Slug     string    `json:"slug"`
	FullName string    `json:"full_name"`
	Headline string    `json:"headline"`
	Score    float64   `json:"score"`
	Rank     int       `json:"rank"`
	Reasons  []string  `json:"reasons"`
}

func (s *Service) People(ctx context.Context, viewerID uuid.UUID) ([]PersonSuggestion, error) {
	rows, err := s.repo.ListForViewer(ctx, viewerID, 10)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	out := make([]PersonSuggestion, 0, len(rows))
	for _, r := range rows {
		var reasons []string
		if r.Reasons != "" {
			_ = json.Unmarshal([]byte(r.Reasons), &reasons)
		}
		out = append(out, PersonSuggestion{
			UserID:   r.UserID,
			Slug:     r.Slug,
			FullName: r.FullName,
			Headline: r.Headline,
			Score:    r.Score,
			Rank:     r.Rank,
			Reasons:  reasons,
		})
	}
	return out, nil
}
