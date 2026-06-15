package apperrors

import (
	"encoding/json"
	"net/http"
)

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func HTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if ae, ok := As(err); ok {
		switch ae.Kind {
		case KindInvalid:
			return http.StatusBadRequest
		case KindNotFound:
			return http.StatusNotFound
		case KindForbidden:
			return http.StatusForbidden
		case KindUnauthorized:
			return http.StatusUnauthorized
		case KindConflict:
			return http.StatusConflict
		case KindInternal:
			return http.StatusInternalServerError
		case KindUnavailable:
			return http.StatusServiceUnavailable
		default:
			return http.StatusInternalServerError
		}
	}
	return http.StatusInternalServerError
}

func PublicPayload(err error) (code string, message string) {
	if err == nil {
		return "", ""
	}
	if ae, ok := As(err); ok {
		return ae.Code, ae.Message
	}
	return CodeInternal, MsgInternal
}

func WriteError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	status := HTTPStatus(err)
	code, msg := PublicPayload(err)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorBody{Code: code, Message: msg})
}
