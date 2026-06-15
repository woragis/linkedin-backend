package httpserver

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	"github.com/unipe/linkedin/backend/server/internal/middleware"
	recosvc "github.com/unipe/linkedin/backend/server/internal/recommendation/service"
	searchsvc "github.com/unipe/linkedin/backend/server/internal/search/service"
)

type searchHandler struct {
	search *searchsvc.Service
}

func newSearchHandler(search *searchsvc.Service) *searchHandler {
	return &searchHandler{search: search}
}

func (h *searchHandler) people(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	var viewerID *uuid.UUID
	if id, ok := middleware.UserIDFromContext(r.Context()); ok {
		viewerID = &id
	}
	out, err := h.search.SearchPeople(r.Context(), viewerID, q, limit)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *searchHandler) posts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	out, err := h.search.SearchPosts(r.Context(), q, limit)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

type recommendationHandler struct {
	reco *recosvc.Service
}

func newRecommendationHandler(reco *recosvc.Service) *recommendationHandler {
	return &recommendationHandler{reco: reco}
}

func (h *recommendationHandler) people(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	out, err := h.reco.People(r.Context(), userID)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *recommendationHandler) peopleMeta(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	out, err := h.reco.PeopleWithMeta(r.Context(), userID)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
