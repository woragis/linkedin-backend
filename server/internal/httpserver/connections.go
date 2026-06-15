package httpserver

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	"github.com/unipe/linkedin/backend/server/internal/middleware"
	connsvc "github.com/unipe/linkedin/backend/server/internal/connection/service"
)

type connectionHandler struct {
	connections *connsvc.Service
}

func newConnectionHandler(connections *connsvc.Service) *connectionHandler {
	return &connectionHandler{connections: connections}
}

func (h *connectionHandler) request(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	var req connsvc.RequestInput
	if err := decodeJSON(r, &req); err != nil {
		apperrors.WriteError(w, err)
		return
	}
	out, err := h.connections.Request(r.Context(), userID, req)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (h *connectionHandler) accept(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	out, err := h.connections.Accept(r.Context(), userID, id)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *connectionHandler) reject(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	out, err := h.connections.Reject(r.Context(), userID, id)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *connectionHandler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.Unauthorized(apperrors.CodeAuthUnauthorized, apperrors.MsgAuthUnauthorized))
		return
	}
	out, err := h.connections.List(r.Context(), userID)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *connectionHandler) pending(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.Unauthorized(apperrors.CodeAuthUnauthorized, apperrors.MsgAuthUnauthorized))
		return
	}
	out, err := h.connections.ListPending(r.Context(), userID)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
