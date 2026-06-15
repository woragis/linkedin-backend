package service

import (
	"context"

	"github.com/google/uuid"
	experimentrepo "github.com/unipe/linkedin/backend/server/internal/experiment/repository"
)

const FeedExperimentName = "feed_ranking_v1"

type Service struct {
	repo *experimentrepo.Repository
}

func New(repo *experimentrepo.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) FeedVariant(ctx context.Context, userID uuid.UUID) (string, error) {
	return s.repo.FeedVariant(ctx, userID, FeedExperimentName)
}

func (s *Service) UseRankedFeed(variant string) bool {
	return variant == "ranked" || variant == "treatment"
}
