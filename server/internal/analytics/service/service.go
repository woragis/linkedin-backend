package service

import (
	"context"

	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	analyticsrepo "github.com/unipe/linkedin/backend/server/internal/analytics/repository"
)

type Service struct {
	repo *analyticsrepo.Repository
}

func New(repo *analyticsrepo.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Overview(ctx context.Context) (*analyticsrepo.Overview, error) {
	o, err := s.repo.Overview(ctx)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return o, nil
}

func (s *Service) TopPosts(ctx context.Context, limit int) ([]analyticsrepo.TopPost, error) {
	rows, err := s.repo.TopPosts(ctx, limit)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return rows, nil
}

func (s *Service) Cohorts(ctx context.Context) ([]analyticsrepo.CohortRow, error) {
	rows, err := s.repo.Cohorts(ctx)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return rows, nil
}

func (s *Service) Churn(ctx context.Context, limit int) ([]analyticsrepo.ChurnUser, error) {
	rows, err := s.repo.ChurnUsers(ctx, limit)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return rows, nil
}

func (s *Service) DailyActive(ctx context.Context, days int) (any, error) {
	rows, err := s.repo.DailyActive(ctx, days)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return rows, nil
}
