package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	authsvc "github.com/unipe/linkedin/backend/server/internal/auth/service"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
)

const userIDKey ctxKey = 2

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(userIDKey).(uuid.UUID)
	return v, ok && v != uuid.Nil
}

func RequireAuth(auth *authsvc.Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := extractUserID(r, auth)
		if err != nil {
			apperrors.WriteError(w, err)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractUserID(r *http.Request, auth *authsvc.Service) (uuid.UUID, error) {
	token := bearerToken(r)
	if token == "" {
		return uuid.Nil, apperrors.Unauthorized(apperrors.CodeAuthUnauthorized, apperrors.MsgAuthUnauthorized)
	}
	return auth.ParseToken(token)
}

func bearerToken(r *http.Request) string {
	h := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}
