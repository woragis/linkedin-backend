package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func parseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	v := r.PathValue(name)
	id, err := uuid.Parse(v)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, apperrors.Invalid(apperrors.CodeInvalidID, apperrors.MsgInvalidID)
	}
	return id, nil
}

func decodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return apperrors.Invalid(apperrors.CodeAuthInvalidBody, "request body required")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return apperrors.Invalid(apperrors.CodeAuthInvalidBody, "invalid JSON body")
	}
	return nil
}
