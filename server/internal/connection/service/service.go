package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	connrepo "github.com/unipe/linkedin/backend/server/internal/connection/repository"
	"github.com/unipe/linkedin/backend/server/internal/models"
	"github.com/unipe/linkedin/backend/server/internal/platform/outbox"
	"gorm.io/gorm"
)

type Service struct {
	repo *connrepo.Repository
	db   *gorm.DB
}

func New(repo *connrepo.Repository, db *gorm.DB) *Service {
	return &Service{repo: repo, db: db}
}

type RequestInput struct {
	TargetUserID uuid.UUID `json:"target_user_id"`
}

func (s *Service) Request(ctx context.Context, userID uuid.UUID, in RequestInput) (*models.Connection, error) {
	if in.TargetUserID == userID {
		return nil, apperrors.Invalid(apperrors.CodeConnectionSelf, apperrors.MsgConnectionSelf)
	}
	exists, err := s.repo.UserExists(ctx, in.TargetUserID)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	if !exists {
		return nil, apperrors.NotFound(apperrors.CodeProfileNotFound, apperrors.MsgProfileNotFound)
	}
	if existing, err := s.repo.FindBetween(ctx, userID, in.TargetUserID); err == nil {
		switch existing.Status {
		case "accepted":
			return nil, apperrors.Conflict(apperrors.CodeConnectionExists, apperrors.MsgConnectionExists)
		case "pending":
			if accepted, acceptErr := s.tryAutoAcceptSimulator(ctx, existing); acceptErr != nil {
				return nil, acceptErr
			} else if accepted != nil {
				return accepted, nil
			}
			return nil, apperrors.Conflict(apperrors.CodeConnectionExists, apperrors.MsgConnectionExists)
		default:
			// rejected — allow new request (update initiator)
			updates := map[string]any{"status": "pending", "requester_id": userID, "addressee_id": in.TargetUserID}
			if err := s.db.WithContext(ctx).Model(&models.Connection{}).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
				return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
			}
			existing.Status = "pending"
			existing.RequesterID = userID
			existing.AddresseeID = in.TargetUserID
			if accepted, acceptErr := s.tryAutoAcceptSimulator(ctx, existing); acceptErr != nil {
				return nil, acceptErr
			} else if accepted != nil {
				return accepted, nil
			}
			return s.repo.GetByID(ctx, existing.ID)
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}

	c := &models.Connection{
		ID:          uuid.New(),
		RequesterID: userID,
		AddresseeID: in.TargetUserID,
		Status:      "pending",
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	if accepted, acceptErr := s.tryAutoAcceptSimulator(ctx, c); acceptErr != nil {
		return nil, acceptErr
	} else if accepted != nil {
		return accepted, nil
	}
	_ = outbox.Enqueue(ctx, s.db,
		outbox.Job{JobType: "graph.recompute_user", Payload: map[string]any{"user_id": userID.String()}},
		outbox.Job{JobType: "graph.recompute_user", Payload: map[string]any{"user_id": in.TargetUserID.String()}},
	)
	return c, nil
}

// tryAutoAcceptSimulator accepts immediately when either party is a synthetic agent.
func (s *Service) tryAutoAcceptSimulator(ctx context.Context, c *models.Connection) (*models.Connection, error) {
	if c.Status == "accepted" {
		return c, nil
	}
	sim, err := s.repo.IsSimulatorUser(ctx, c.AddresseeID)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	if !sim {
		sim, err = s.repo.IsSimulatorUser(ctx, c.RequesterID)
		if err != nil {
			return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
		}
	}
	if !sim {
		return nil, nil
	}
	if err := s.repo.UpdateStatus(ctx, c.ID, "accepted"); err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	c.Status = "accepted"
	_ = outbox.Enqueue(ctx, s.db,
		outbox.Job{JobType: "graph.recompute_user", Payload: map[string]any{"user_id": c.RequesterID.String()}},
		outbox.Job{JobType: "graph.recompute_user", Payload: map[string]any{"user_id": c.AddresseeID.String()}},
	)
	return c, nil
}

func (s *Service) Accept(ctx context.Context, userID, connectionID uuid.UUID) (*models.Connection, error) {
	c, err := s.loadAndAuthorize(ctx, connectionID, userID, true)
	if err != nil {
		return nil, err
	}
	if c.Status != "pending" {
		return nil, apperrors.Invalid(apperrors.CodeConnectionInvalid, "connection is not pending")
	}
	if err := s.repo.UpdateStatus(ctx, c.ID, "accepted"); err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	c.Status = "accepted"
	_ = outbox.Enqueue(ctx, s.db,
		outbox.Job{JobType: "graph.recompute_user", Payload: map[string]any{"user_id": c.RequesterID.String()}},
		outbox.Job{JobType: "graph.recompute_user", Payload: map[string]any{"user_id": c.AddresseeID.String()}},
	)
	return c, nil
}

func (s *Service) Reject(ctx context.Context, userID, connectionID uuid.UUID) (*models.Connection, error) {
	c, err := s.loadAndAuthorize(ctx, connectionID, userID, true)
	if err != nil {
		return nil, err
	}
	if c.Status != "pending" {
		return nil, apperrors.Invalid(apperrors.CodeConnectionInvalid, "connection is not pending")
	}
	if err := s.repo.UpdateStatus(ctx, c.ID, "rejected"); err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	c.Status = "rejected"
	return c, nil
}

type ConnectionView struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Slug        string    `json:"slug"`
	FullName    string    `json:"full_name"`
	Headline    string    `json:"headline"`
	AvatarURL   *string   `json:"avatar_url"`
	ConnectedAt time.Time `json:"connected_at"`
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]ConnectionView, error) {
	rows, err := s.repo.ListAcceptedWithPeers(ctx, userID)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	out := make([]ConnectionView, 0, len(rows))
	for _, r := range rows {
		out = append(out, ConnectionView{
			ID:          r.ID,
			UserID:      r.UserID,
			Slug:        r.Slug,
			FullName:    r.FullName,
			Headline:    r.Headline,
			AvatarURL:   r.AvatarURL,
			ConnectedAt: r.ConnectedAt,
		})
	}
	return out, nil
}

func (s *Service) ListPending(ctx context.Context, userID uuid.UUID) ([]models.Connection, error) {
	rows, err := s.repo.ListPendingIncoming(ctx, userID)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return rows, nil
}

func (s *Service) loadAndAuthorize(ctx context.Context, id, userID uuid.UUID, addresseeOnly bool) (*models.Connection, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound(apperrors.CodeConnectionNotFound, apperrors.MsgConnectionNotFound)
		}
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	if addresseeOnly && c.AddresseeID != userID {
		return nil, apperrors.Forbidden(apperrors.CodeConnectionForbidden, apperrors.MsgConnectionForbidden)
	}
	if !addresseeOnly && c.RequesterID != userID && c.AddresseeID != userID {
		return nil, apperrors.Forbidden(apperrors.CodeConnectionForbidden, apperrors.MsgConnectionForbidden)
	}
	return c, nil
}
