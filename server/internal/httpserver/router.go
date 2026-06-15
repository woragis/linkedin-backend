package httpserver

import (
	"net/http"

	"github.com/unipe/linkedin/backend/server/internal/middleware"
)

func Mount(mux *http.ServeMux, app *App) {
	mux.HandleFunc("GET /health", handleHealth)
	if app != nil && app.DB != nil {
		mux.HandleFunc("GET /ready", handleReady(app.DB))
	}

	if app == nil || app.Auth == nil {
		return
	}

	ah := newAuthHandler(app.Auth)
	mux.HandleFunc("POST /v1/auth/register", ah.register)
	mux.HandleFunc("POST /v1/auth/login", ah.login)

	if app.Profiles == nil {
		return
	}

	ph := newProfileHandler(app.Profiles)
	require := func(h http.HandlerFunc) http.Handler {
		return middleware.RequireAuth(app.Auth, h)
	}

	mux.Handle("GET /v1/me", require(http.HandlerFunc(ph.me)))
	mux.Handle("PATCH /v1/me/profile", require(http.HandlerFunc(ph.patchMe)))

	mux.HandleFunc("GET /v1/users/{slug}", ph.getBySlug)

	mux.Handle("GET /v1/me/experiences", require(http.HandlerFunc(ph.listExperiences)))
	mux.Handle("POST /v1/me/experiences", require(http.HandlerFunc(ph.createExperience)))
	mux.Handle("PATCH /v1/me/experiences/{id}", require(http.HandlerFunc(ph.patchExperience)))
	mux.Handle("DELETE /v1/me/experiences/{id}", require(http.HandlerFunc(ph.deleteExperience)))

	mux.Handle("GET /v1/me/educations", require(http.HandlerFunc(ph.listEducations)))
	mux.Handle("POST /v1/me/educations", require(http.HandlerFunc(ph.createEducation)))
	mux.Handle("PATCH /v1/me/educations/{id}", require(http.HandlerFunc(ph.patchEducation)))
	mux.Handle("DELETE /v1/me/educations/{id}", require(http.HandlerFunc(ph.deleteEducation)))

	mux.Handle("GET /v1/me/skills", require(http.HandlerFunc(ph.listSkills)))
	mux.Handle("PUT /v1/me/skills", require(http.HandlerFunc(ph.replaceSkills)))

	if app.Seed != nil {
		ih := newInternalHandler(app.InternalJobSecret, app.Seed)
		mux.HandleFunc("POST /v1/internal/seed-demo", ih.seedDemo)
	}
}
