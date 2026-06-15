package httpserver

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	"github.com/unipe/linkedin/backend/server/internal/middleware"
	eventsvc "github.com/unipe/linkedin/backend/server/internal/event/service"
)

type eventHandler struct {
	events *eventsvc.Service
}

func newEventHandler(events *eventsvc.Service) *eventHandler {
	return &eventHandler{events: events}
}

func (h *eventHandler) ingest(w http.ResponseWriter, r *http.Request) {
	var userID *uuid.UUID
	if id, ok := middleware.UserIDFromContext(r.Context()); ok {
		userID = &id
	}
	var req eventsvc.IngestRequest
	if err := decodeJSON(r, &req); err != nil {
		apperrors.WriteError(w, err)
		return
	}
	out, err := h.events.Ingest(r.Context(), userID, req)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, out)
}
