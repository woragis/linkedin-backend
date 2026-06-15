package httpserver

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	postsvc "github.com/unipe/linkedin/backend/server/internal/post/service"
)

type postHandler struct {
	posts *postsvc.Service
}

func newPostHandler(posts *postsvc.Service) *postHandler {
	return &postHandler{posts: posts}
}

func (h *postHandler) create(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	var req postsvc.CreatePostRequest
	if err := decodeJSON(r, &req); err != nil {
		apperrors.WriteError(w, err)
		return
	}
	out, err := h.posts.Create(r.Context(), userID, req)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (h *postHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	out, err := h.posts.Get(r.Context(), id)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *postHandler) react(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	postID, err := parseUUIDParam(r, "id")
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	var body struct {
		Kind string `json:"kind"`
	}
	_ = decodeJSON(r, &body)
	if err := h.posts.React(r.Context(), userID, postID, body.Kind); err != nil {
		apperrors.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *postHandler) comment(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	postID, err := parseUUIDParam(r, "id")
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	var req postsvc.CreateCommentRequest
	if err := decodeJSON(r, &req); err != nil {
		apperrors.WriteError(w, err)
		return
	}
	out, err := h.posts.Comment(r.Context(), userID, postID, req)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (h *postHandler) feed(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	out, err := h.posts.Feed(r.Context(), userID, limit)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
