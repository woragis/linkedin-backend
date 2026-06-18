package middleware

import (
	"net/http"
	"strings"
)

func CORS(cfg Config, next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(cfg.CORSOrigins))
	for _, o := range cfg.CORSOrigins {
		allowed[o] = struct{}{}
	}
	allowHeaders := "Authorization, Content-Type, X-Request-ID, X-Internal-Token, X-App-Realm"
	allowMethods := "GET, POST, PATCH, PUT, DELETE, OPTIONS"
	exposeHeaders := "X-Request-ID"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if _, ok := allowed[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
				if cfg.CORSAllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
				w.Header().Set("Access-Control-Allow-Methods", allowMethods)
				w.Header().Set("Access-Control-Expose-Headers", exposeHeaders)
				w.Header().Set("Access-Control-Max-Age", "86400")
			}
		}
		if r.Method == http.MethodOptions {
			if origin != "" && w.Header().Get("Access-Control-Allow-Origin") != "" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if strings.HasPrefix(r.URL.Path, "/v1/") {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
