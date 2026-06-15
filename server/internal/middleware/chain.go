package middleware

import "net/http"

func Chain(cfg Config, next http.Handler) http.Handler {
	h := http.Handler(next)
	h = CORS(cfg, h)
	h = AccessLog(h)
	h = RequestID(h)
	h = SecurityHeaders(h)
	return h
}
