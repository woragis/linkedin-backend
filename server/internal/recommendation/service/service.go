package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	experimentrepo "github.com/unipe/linkedin/backend/server/internal/experiment/repository"
	recorepo "github.com/unipe/linkedin/backend/server/internal/recommendation/repository"
	"gorm.io/gorm"
)

const ScoringMethodRuleBased = "rule_based_affinity"
const ConnectionAcceptanceModel = "connection_acceptance"

type Service struct {
	repo           *recorepo.Repository
	experimentRepo *experimentrepo.Repository
}

func New(repo *recorepo.Repository, experimentRepo *experimentrepo.Repository) *Service {
	return &Service{repo: repo, experimentRepo: experimentRepo}
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

type PeopleMetaResponse struct {
	ScoringMethod string                    `json:"scoring_method"`
	MLModel       *experimentrepo.MLModel     `json:"ml_model,omitempty"`
	Suggestions   []PersonSuggestion        `json:"suggestions"`
}

func (s *Service) PeopleWithMeta(ctx context.Context, viewerID uuid.UUID) (*PeopleMetaResponse, error) {
	suggestions, err := s.People(ctx, viewerID)
	if err != nil {
		return nil, err
	}
	model, err := s.experimentRepo.ActiveMLModel(ctx, ConnectionAcceptanceModel)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return &PeopleMetaResponse{
		ScoringMethod: ScoringMethodRuleBased,
		MLModel:       model,
		Suggestions:   suggestions,
	}, nil
}
