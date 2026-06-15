package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	authrepo "github.com/unipe/linkedin/backend/server/internal/auth/repository"
	"github.com/unipe/linkedin/backend/server/internal/models"
	jwtmgr "github.com/unipe/linkedin/backend/server/internal/platform/jwt"
	"github.com/unipe/linkedin/backend/server/internal/platform/password"
	"github.com/unipe/linkedin/backend/server/internal/platform/slug"
	"gorm.io/gorm"
)

type Service struct {
	repo *authrepo.Repository
	db   *gorm.DB
	jwt  *jwtmgr.Manager
}

func New(repo *authrepo.Repository, db *gorm.DB, jwt *jwtmgr.Manager) *Service {
	return &Service{repo: repo, db: db, jwt: jwt}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    uuid.UUID `json:"user_id"`
	Slug      string    `json:"slug"`
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, apperrors.Invalid(apperrors.CodeAuthEmailInvalid, apperrors.MsgAuthEmailInvalid)
	}
	if len(req.Password) < 8 {
		return nil, apperrors.Invalid(apperrors.CodeAuthPasswordWeak, apperrors.MsgAuthPasswordWeak)
	}
	fullName := strings.TrimSpace(req.FullName)
	if fullName == "" {
		return nil, apperrors.Invalid(apperrors.CodeAuthInvalidBody, "full_name is required")
	}

	hash, err := password.Hash(req.Password)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}

	baseSlug := slug.FromName(fullName)
	uniqueSlug, err := slug.EnsureUnique(ctx, s.db, baseSlug)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hash,
	}
	profile := &models.Profile{
		Slug:     uniqueSlug,
		FullName: fullName,
	}

	if err := s.repo.CreateUserWithProfile(ctx, user, profile); err != nil {
		if authrepo.IsUniqueViolation(err) {
			return nil, apperrors.Conflict(apperrors.CodeAuthEmailTaken, apperrors.MsgAuthEmailTaken)
		}
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}

	token, exp, err := s.jwt.Issue(user.ID)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}

	return &AuthResponse{
		Token:     token,
		ExpiresAt: exp,
		UserID:    user.ID,
		Slug:      uniqueSlug,
	}, nil
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Unauthorized(apperrors.CodeAuthInvalidCredentials, apperrors.MsgAuthInvalidCredentials)
		}
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	if !password.Compare(user.PasswordHash, req.Password) {
		return nil, apperrors.Unauthorized(apperrors.CodeAuthInvalidCredentials, apperrors.MsgAuthInvalidCredentials)
	}

	var profile models.Profile
	if err := s.db.WithContext(ctx).First(&profile, "user_id = ?", user.ID).Error; err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}

	token, exp, err := s.jwt.Issue(user.ID)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}

	return &AuthResponse{
		Token:     token,
		ExpiresAt: exp,
		UserID:    user.ID,
		Slug:      profile.Slug,
	}, nil
}

func (s *Service) ParseToken(token string) (uuid.UUID, error) {
	id, err := s.jwt.Parse(token)
	if err != nil {
		return uuid.Nil, apperrors.Unauthorized(apperrors.CodeAuthTokenInvalid, apperrors.MsgAuthTokenInvalid)
	}
	return id, nil
}
