package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	"github.com/unipe/linkedin/backend/server/internal/models"
)

type FeedResponse struct {
	Variant string     `json:"variant"`
	Posts   []PostView `json:"posts"`
}

func (s *Service) Feed(ctx context.Context, userID uuid.UUID, limit int) (*FeedResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	variant := "chronological"
	if s.experiments != nil {
		v, err := s.experiments.FeedVariant(ctx, userID)
		if err == nil {
			variant = v
		}
	}

	var cached FeedResponse
	if s.feedCache != nil && s.feedCache.Get(ctx, userID, variant, &cached) {
		return &cached, nil
	}

	var rows []models.Post
	var err error
	if s.experiments != nil && s.experiments.UseRankedFeed(variant) {
		rows, err = s.posts.RankedFeed(ctx, userID, limit)
	} else {
		peers, perr := s.connections.ListAcceptedPeerIDs(ctx, userID)
		if perr != nil {
			return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, perr)
		}
		rows, err = s.posts.Feed(ctx, userID, peers, limit)
	}
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}

	out := s.enrichPostViews(ctx, rows, userID)
	resp := &FeedResponse{Variant: variant, Posts: out}
	if s.feedCache != nil {
		s.feedCache.Set(ctx, userID, variant, resp)
	}
	return resp, nil
}
