package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/unipe/linkedin/backend/server/internal/observability/metrics"
)

func Prometheus(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		path := r.URL.Path
		if path == "/metrics" || path == "/health" || path == "/ready" {
			return
		}
		metrics.HTTPRequests.WithLabelValues(r.Method, path, strconv.Itoa(rec.status)).Inc()
		metrics.HTTPDuration.WithLabelValues(r.Method, path).Observe(time.Since(start).Seconds())
	})
}
