package httpserver

import (
	"net/http"

	authsvc "github.com/unipe/linkedin/backend/server/internal/auth/service"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
)

type authHandler struct {
	auth *authsvc.Service
}

func newAuthHandler(auth *authsvc.Service) *authHandler {
	return &authHandler{auth: auth}
}

func (h *authHandler) register(w http.ResponseWriter, r *http.Request) {
	var req authsvc.RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		apperrors.WriteError(w, err)
		return
	}
	out, err := h.auth.Register(r.Context(), req)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (h *authHandler) login(w http.ResponseWriter, r *http.Request) {
	var req authsvc.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		apperrors.WriteError(w, err)
		return
	}
	out, err := h.auth.Login(r.Context(), req)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
