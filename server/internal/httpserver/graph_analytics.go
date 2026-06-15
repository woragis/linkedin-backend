package httpserver

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	analyticsvc "github.com/unipe/linkedin/backend/server/internal/analytics/service"
	graphsvc "github.com/unipe/linkedin/backend/server/internal/graph/service"
)

type graphHandler struct {
	graph *graphsvc.Service
}

func newGraphHandler(graph *graphsvc.Service) *graphHandler {
	return &graphHandler{graph: graph}
}

func (h *graphHandler) userGraph(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	out, err := h.graph.UserGraph(r.Context(), userID)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *graphHandler) influencers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	out, err := h.graph.TopInfluencers(r.Context(), limit)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

type analyticsHandler struct {
	analytics *analyticsvc.Service
}

func newAnalyticsHandler(analytics *analyticsvc.Service) *analyticsHandler {
	return &analyticsHandler{analytics: analytics}
}

func (h *analyticsHandler) overview(w http.ResponseWriter, r *http.Request) {
	out, err := h.analytics.Overview(r.Context())
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *analyticsHandler) topPosts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	out, err := h.analytics.TopPosts(r.Context(), limit)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *analyticsHandler) cohorts(w http.ResponseWriter, r *http.Request) {
	out, err := h.analytics.Cohorts(r.Context())
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *analyticsHandler) churn(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	out, err := h.analytics.Churn(r.Context(), limit)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *analyticsHandler) dau(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	out, err := h.analytics.DailyActive(r.Context(), days)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
