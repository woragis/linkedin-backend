package httpserver

import "net/http"

func Mount(mux *http.ServeMux, app *App) {
	mux.HandleFunc("GET /health", handleHealth)
	if app != nil && app.DB != nil {
		mux.HandleFunc("GET /ready", handleReady(app.DB))
	}
}
