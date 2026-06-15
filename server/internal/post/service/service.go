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

type PostView struct {
	models.Post
	ReactionCount int64 `json:"reaction_count"`
	CommentCount  int64 `json:"comment_count"`
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
	return s.view(ctx, p.ID)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*PostView, error) {
	return s.view(ctx, id)
}

func (s *Service) view(ctx context.Context, id uuid.UUID) (*PostView, error) {
	p, err := s.posts.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound(apperrors.CodePostNotFound, apperrors.MsgPostNotFound)
		}
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	rc, _ := s.posts.ReactionCount(ctx, id)
	cc, _ := s.posts.CommentCount(ctx, id)
	return &PostView{Post: *p, ReactionCount: rc, CommentCount: cc}, nil
}

func (s *Service) React(ctx context.Context, userID, postID uuid.UUID, kind string) error {
	if _, err := s.view(ctx, postID); err != nil {
		return err
	}
	if kind == "" {
		kind = "like"
	}
	if err := s.posts.AddReaction(ctx, postID, userID, kind); err != nil {
		return apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return nil
}

type CreateCommentRequest struct {
	Body string `json:"body"`
}

func (s *Service) Comment(ctx context.Context, userID, postID uuid.UUID, req CreateCommentRequest) (*models.Comment, error) {
	body := strings.TrimSpace(req.Body)
	if body == "" {
		return nil, apperrors.Invalid(apperrors.CodeCommentInvalid, "body is required")
	}
	if _, err := s.view(ctx, postID); err != nil {
		return nil, err
	}
	c := &models.Comment{ID: uuid.New(), PostID: postID, AuthorID: userID, Body: body}
	if err := s.posts.AddComment(ctx, c); err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return c, nil
}
