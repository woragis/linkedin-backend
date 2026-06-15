package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	eventrepo "github.com/unipe/linkedin/backend/server/internal/event/repository"
)

type Service struct {
	repo *eventrepo.Repository
}

func New(repo *eventrepo.Repository) *Service {
	return &Service{repo: repo}
}

type IngestRequest struct {
	Events []eventrepo.IncomingEvent `json:"events"`
}

type IngestResult struct {
	Accepted int `json:"accepted"`
}

func (s *Service) Ingest(ctx context.Context, userID *uuid.UUID, req IngestRequest) (*IngestResult, error) {
	if len(req.Events) == 0 {
		return nil, apperrors.Invalid(apperrors.CodeEventsInvalid, "events array is required")
	}
	if len(req.Events) > 100 {
		return nil, apperrors.Invalid(apperrors.CodeEventsInvalid, "max 100 events per batch")
	}
	n, err := s.repo.InsertBatch(ctx, userID, req.Events)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return &IngestResult{Accepted: n}, nil
}
