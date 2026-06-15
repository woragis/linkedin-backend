package service

import (
	"context"
	"errors"

	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	analyticsrepo "github.com/unipe/linkedin/backend/server/internal/analytics/repository"
	experimentrepo "github.com/unipe/linkedin/backend/server/internal/experiment/repository"
	"gorm.io/gorm"
)

type Service struct {
	repo         *analyticsrepo.Repository
	experimentRepo *experimentrepo.Repository
}

func New(repo *analyticsrepo.Repository, experimentRepo *experimentrepo.Repository) *Service {
	return &Service{repo: repo, experimentRepo: experimentRepo}
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

func (s *Service) Experiments(ctx context.Context) ([]experimentrepo.ABExperimentResult, error) {
	rows, err := s.experimentRepo.ABExperimentResults(ctx)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return rows, nil
}

func (s *Service) MLModels(ctx context.Context) ([]experimentrepo.MLModel, error) {
	rows, err := s.experimentRepo.ListMLModels(ctx)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return rows, nil
}

func (s *Service) ActiveMLModel(ctx context.Context, modelName string) (*experimentrepo.MLModel, error) {
	m, err := s.experimentRepo.ActiveMLModel(ctx, modelName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return m, nil
}
