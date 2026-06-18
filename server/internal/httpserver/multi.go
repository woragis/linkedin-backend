package httpserver

import (
	"net/http"

	"github.com/unipe/linkedin/backend/server/internal/middleware"
	"github.com/unipe/linkedin/backend/server/internal/platform/realm"
)

// MultiApp holds isolated service stacks per realm (volume vs live).
type MultiApp struct {
	Volume *App
	Live   *App
}

// AppFor returns the service stack for the request realm.
func (m *MultiApp) AppFor(r *http.Request) *App {
	if m == nil {
		return nil
	}
	switch realm.FromContext(r.Context()) {
	case realm.Volume:
		if m.Volume != nil {
			return m.Volume
		}
	default:
		if m.Live != nil {
			return m.Live
		}
	}
	if m.Live != nil {
		return m.Live
	}
	return m.Volume
}

func (m *MultiApp) require(h func(*App, http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app := m.AppFor(r)
		if app == nil || app.Auth == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		middleware.RequireAuth(app.Auth, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h(app, w, r)
		})).ServeHTTP(w, r)
	})
}
