package httpserver

import (
	"net/http"
	"strings"

	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	seedsvc "github.com/unipe/linkedin/backend/server/internal/seed/service"
)

type internalHandler struct {
	secret string
	seed   *seedsvc.Service
}

func newInternalHandler(secret string, seed *seedsvc.Service) *internalHandler {
	return &internalHandler{secret: secret, seed: seed}
}

func (h *internalHandler) seedDemo(w http.ResponseWriter, r *http.Request) {
	if h.secret != "" {
		token := strings.TrimSpace(r.Header.Get("X-Internal-Token"))
		if token != h.secret {
			apperrors.WriteError(w, apperrors.Forbidden(apperrors.CodeInternalUnauthorized, apperrors.MsgInternalUnauthorized))
			return
		}
	}
	out, err := h.seed.SeedDemo(r.Context())
	if err != nil {
		apperrors.WriteError(w, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err))
		return
	}
	writeJSON(w, http.StatusCreated, out)
}
