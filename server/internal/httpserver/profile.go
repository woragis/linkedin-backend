package httpserver

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	"github.com/unipe/linkedin/backend/server/internal/middleware"
	profilesvc "github.com/unipe/linkedin/backend/server/internal/profile/service"
)

type profileHandler struct {
	profiles *profilesvc.Service
}

func newProfileHandler(profiles *profilesvc.Service) *profileHandler {
	return &profileHandler{profiles: profiles}
}

func (h *profileHandler) me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.Unauthorized(apperrors.CodeAuthUnauthorized, apperrors.MsgAuthUnauthorized))
		return
	}
	out, err := h.profiles.GetMe(r.Context(), userID)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *profileHandler) patchMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.Unauthorized(apperrors.CodeAuthUnauthorized, apperrors.MsgAuthUnauthorized))
		return
	}
	var req profilesvc.PatchProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		apperrors.WriteError(w, err)
		return
	}
	out, err := h.profiles.PatchProfile(r.Context(), userID, req)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *profileHandler) getBySlug(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	out, err := h.profiles.GetBySlug(r.Context(), slug)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *profileHandler) listExperiences(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	out, err := h.profiles.ListExperiences(r.Context(), userID)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *profileHandler) createExperience(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	var req profilesvc.CreateExperienceRequest
	if err := decodeJSON(r, &req); err != nil {
		apperrors.WriteError(w, err)
		return
	}
	out, err := h.profiles.CreateExperience(r.Context(), userID, req)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (h *profileHandler) patchExperience(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	var req profilesvc.PatchExperienceRequest
	if err := decodeJSON(r, &req); err != nil {
		apperrors.WriteError(w, err)
		return
	}
	out, err := h.profiles.PatchExperience(r.Context(), userID, id, req)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *profileHandler) deleteExperience(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	if err := h.profiles.DeleteExperience(r.Context(), userID, id); err != nil {
		apperrors.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *profileHandler) listEducations(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	out, err := h.profiles.ListEducations(r.Context(), userID)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *profileHandler) createEducation(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	var req profilesvc.CreateEducationRequest
	if err := decodeJSON(r, &req); err != nil {
		apperrors.WriteError(w, err)
		return
	}
	out, err := h.profiles.CreateEducation(r.Context(), userID, req)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (h *profileHandler) patchEducation(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	var req profilesvc.PatchEducationRequest
	if err := decodeJSON(r, &req); err != nil {
		apperrors.WriteError(w, err)
		return
	}
	out, err := h.profiles.PatchEducation(r.Context(), userID, id, req)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *profileHandler) deleteEducation(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	if err := h.profiles.DeleteEducation(r.Context(), userID, id); err != nil {
		apperrors.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *profileHandler) listSkills(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	out, err := h.profiles.ListSkills(r.Context(), userID)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *profileHandler) replaceSkills(w http.ResponseWriter, r *http.Request) {
	userID := mustUser(w, r)
	if userID == uuid.Nil {
		return
	}
	var req profilesvc.ReplaceSkillsRequest
	if err := decodeJSON(r, &req); err != nil {
		apperrors.WriteError(w, err)
		return
	}
	out, err := h.profiles.ReplaceSkills(r.Context(), userID, req)
	if err != nil {
		apperrors.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func mustUser(w http.ResponseWriter, r *http.Request) uuid.UUID {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apperrors.WriteError(w, apperrors.Unauthorized(apperrors.CodeAuthUnauthorized, apperrors.MsgAuthUnauthorized))
		return uuid.Nil
	}
	return userID
}
