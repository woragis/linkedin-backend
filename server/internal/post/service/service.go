package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	connrepo "github.com/unipe/linkedin/backend/server/internal/connection/repository"
	experimentsvc "github.com/unipe/linkedin/backend/server/internal/experiment/service"
	"github.com/unipe/linkedin/backend/server/internal/models"
	"github.com/unipe/linkedin/backend/server/internal/platform/cache"
	"github.com/unipe/linkedin/backend/server/internal/platform/outbox"
	postrepo "github.com/unipe/linkedin/backend/server/internal/post/repository"
	"gorm.io/gorm"
)

type Service struct {
	posts       *postrepo.Repository
	connections *connrepo.Repository
	experiments *experimentsvc.Service
	feedCache   *cache.FeedCache
	db          *gorm.DB
}

func New(
	posts *postrepo.Repository,
	connections *connrepo.Repository,
	experiments *experimentsvc.Service,
	feedCache *cache.FeedCache,
	db *gorm.DB,
) *Service {
	return &Service{
		posts:       posts,
		connections: connections,
		experiments: experiments,
		feedCache:   feedCache,
		db:          db,
	}
}

type CreatePostRequest struct {
	Body string `json:"body"`
}

var validReactionKinds = map[string]struct{}{
	"like":        {},
	"celebrate":   {},
	"support":     {},
	"insightful":  {},
	"love":        {},
	"funny":       {},
}

func normalizeReactionKind(kind string) (string, error) {
	if kind == "" {
		return "like", nil
	}
	if _, ok := validReactionKinds[kind]; !ok {
		return "", apperrors.Invalid(apperrors.CodePostInvalidBody, "invalid reaction kind")
	}
	return kind, nil
}

func (s *Service) Create(ctx context.Context, authorID uuid.UUID, req CreatePostRequest) (*PostView, error) {
	body := strings.TrimSpace(req.Body)
	if body == "" {
		return nil, apperrors.Invalid(apperrors.CodePostInvalidBody, "body is required")
	}
	p := &models.Post{ID: uuid.New(), AuthorID: authorID, Body: body}
	if err := s.posts.Create(ctx, p); err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	_ = outbox.Enqueue(ctx, s.db, outbox.Job{
		JobType: "search.index_post",
		Payload: map[string]any{"post_id": p.ID.String()},
	})
	return s.view(ctx, p.ID, authorID)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*PostView, error) {
	return s.view(ctx, id, uuid.Nil)
}

func (s *Service) view(ctx context.Context, id uuid.UUID, viewerID uuid.UUID) (*PostView, error) {
	p, err := s.posts.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound(apperrors.CodePostNotFound, apperrors.MsgPostNotFound)
		}
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	out := s.enrichPostView(ctx, *p, viewerID)
	return &out, nil
}

func (s *Service) React(ctx context.Context, userID, postID uuid.UUID, kind string) error {
	normalized, err := normalizeReactionKind(kind)
	if err != nil {
		return err
	}
	if _, err := s.view(ctx, postID, userID); err != nil {
		return err
	}
	if err := s.posts.AddReaction(ctx, postID, userID, normalized); err != nil {
		return apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return nil
}

func (s *Service) ReactComment(ctx context.Context, userID, commentID uuid.UUID, kind string) error {
	normalized, err := normalizeReactionKind(kind)
	if err != nil {
		return err
	}
	c, err := s.posts.GetCommentByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NotFound(apperrors.CodeCommentNotFound, apperrors.MsgCommentNotFound)
		}
		return apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	if err := s.posts.UpsertContentReaction(ctx, postrepo.TargetComment, commentID, userID, normalized); err != nil {
		return apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	_ = c
	return nil
}

type CreateCommentRequest struct {
	Body            string     `json:"body"`
	ParentCommentID *uuid.UUID `json:"parent_comment_id,omitempty"`
}

func (s *Service) Comment(ctx context.Context, userID, postID uuid.UUID, req CreateCommentRequest) (*CommentView, error) {
	body := strings.TrimSpace(req.Body)
	if body == "" {
		return nil, apperrors.Invalid(apperrors.CodeCommentInvalid, "body is required")
	}
	if _, err := s.view(ctx, postID, userID); err != nil {
		return nil, err
	}
	var parentID *uuid.UUID
	if req.ParentCommentID != nil && *req.ParentCommentID != uuid.Nil {
		parent, err := s.posts.GetCommentByID(ctx, *req.ParentCommentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.NotFound(apperrors.CodeCommentNotFound, apperrors.MsgCommentNotFound)
			}
			return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
		}
		if parent.PostID != postID {
			return nil, apperrors.Invalid(apperrors.CodeCommentInvalid, "parent comment belongs to another post")
		}
		if parent.ParentCommentID != nil {
			return nil, apperrors.Invalid(apperrors.CodeCommentInvalid, "only one level of replies is allowed")
		}
		parentID = req.ParentCommentID
	}
	c := &models.Comment{
		ID:              uuid.New(),
		PostID:          postID,
		AuthorID:        userID,
		ParentCommentID: parentID,
		Body:            body,
	}
	if err := s.posts.AddComment(ctx, c); err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	created, err := s.posts.GetCommentByID(ctx, c.ID)
	if err != nil {
		return s.enrichSingleComment(ctx, *c, userID)
	}
	return s.enrichSingleComment(ctx, *created, userID)
}

func (s *Service) ListComments(ctx context.Context, postID, viewerID uuid.UUID) ([]CommentView, error) {
	if _, err := s.view(ctx, postID, viewerID); err != nil {
		return nil, err
	}
	rows, err := s.posts.ListComments(ctx, postID)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return s.buildCommentTree(ctx, rows, viewerID)
}
