package middleware

import (
	"net/http"

	"github.com/unipe/linkedin/backend/server/internal/platform/realm"
)

func Chain(cfg Config, next http.Handler) http.Handler {
	h := http.Handler(next)
	h = Prometheus(h)
	h = CORS(cfg, h)
	h = AccessLog(h)
	h = realm.Middleware(h)
	h = RequestID(h)
	h = SecurityHeaders(h)
	return h
}
