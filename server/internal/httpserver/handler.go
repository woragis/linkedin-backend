package httpserver

import (
	"net/http"

	"github.com/unipe/linkedin/backend/server/internal/middleware"
)

func NewHandler(multi *MultiApp, cfg middleware.Config) http.Handler {
	mux := http.NewServeMux()
	Mount(mux, multi)
	return middleware.Chain(cfg, mux)
}
