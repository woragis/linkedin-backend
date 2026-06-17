package service

import (
	"context"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	searchrepo "github.com/unipe/linkedin/backend/server/internal/search/repository"
	"github.com/unipe/linkedin/backend/server/internal/search/elasticsearch"
)

const (
	alphaES       = 0.55
	betaAffinity  = 0.35
	defaultLimit  = 20
)

type Service struct {
	repo *searchrepo.Repository
	es   *elasticsearch.Client
}

func New(repo *searchrepo.Repository, es *elasticsearch.Client) *Service {
	return &Service{repo: repo, es: es}
}

type PersonResult struct {
	searchrepo.PersonHit
	AffinityScore float64  `json:"affinity_score,omitempty"`
	FinalScore    float64  `json:"final_score"`
	Reasons       []string `json:"reasons,omitempty"`
}

func (s *Service) SearchPeople(ctx context.Context, viewerID *uuid.UUID, q string, limit int) ([]PersonResult, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, apperrors.Invalid(apperrors.CodeProfileInvalidBody, "query q is required")
	}
	if limit <= 0 || limit > 50 {
		limit = defaultLimit
	}

	var hits []searchrepo.PersonHit
	var err error
	if s.es != nil && s.es.Enabled() {
		hits, err = s.es.SearchPeople(ctx, q, limit*2)
	}
	if err != nil || len(hits) == 0 {
		hits, err = s.repo.SearchPeople(ctx, q, limit*2)
	}
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}

	var affinity map[uuid.UUID]float64
	if viewerID != nil {
		ids := make([]uuid.UUID, 0, len(hits))
		for _, h := range hits {
			ids = append(ids, h.UserID)
		}
		affinity, _ = s.repo.AffinityBoost(ctx, *viewerID, ids)
	}

	out := make([]PersonResult, 0, len(hits))
	for _, h := range hits {
		aff := affinity[h.UserID]
		final := alphaES*h.Score + betaAffinity*aff*10
		if aff == 0 {
			final = h.Score
		}
		out = append(out, PersonResult{
			PersonHit:     h,
			AffinityScore: aff,
			FinalScore:    final,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].FinalScore > out[j].FinalScore })
	if len(out) > limit {
		out = out[:limit]
	}
	if out == nil {
		out = []PersonResult{}
	}
	return out, nil
}

func (s *Service) SearchPosts(ctx context.Context, q string, limit int) ([]searchrepo.PostHit, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, apperrors.Invalid(apperrors.CodeProfileInvalidBody, "query q is required")
	}
	if limit <= 0 || limit > 50 {
		limit = defaultLimit
	}

	var hits []searchrepo.PostHit
	var err error
	if s.es != nil && s.es.Enabled() {
		hits, err = s.es.SearchPosts(ctx, q, limit)
	}
	if err != nil || len(hits) == 0 {
		hits, err = s.repo.SearchPosts(ctx, q, limit)
	}
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	if hits == nil {
		hits = []searchrepo.PostHit{}
	}
	return hits, nil
}
