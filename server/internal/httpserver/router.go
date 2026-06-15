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

	require := func(h http.HandlerFunc) http.Handler {
		return middleware.RequireAuth(app.Auth, h)
	}

	if app.Events != nil {
		eh := newEventHandler(app.Events)
		mux.Handle("POST /v1/events", require(http.HandlerFunc(eh.ingest)))
	}

	if app.Profiles != nil {
		ph := newProfileHandler(app.Profiles)
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
	}

	if app.Connections != nil {
		ch := newConnectionHandler(app.Connections)
		mux.Handle("POST /v1/connections/request", require(http.HandlerFunc(ch.request)))
		mux.Handle("PATCH /v1/connections/{id}/accept", require(http.HandlerFunc(ch.accept)))
		mux.Handle("PATCH /v1/connections/{id}/reject", require(http.HandlerFunc(ch.reject)))
		mux.Handle("GET /v1/connections", require(http.HandlerFunc(ch.list)))
		mux.Handle("GET /v1/connections/pending", require(http.HandlerFunc(ch.pending)))
	}

	if app.Posts != nil {
		po := newPostHandler(app.Posts)
		mux.Handle("POST /v1/posts", require(http.HandlerFunc(po.create)))
		mux.HandleFunc("GET /v1/posts/{id}", po.get)
		mux.Handle("POST /v1/posts/{id}/reactions", require(http.HandlerFunc(po.react)))
		mux.Handle("POST /v1/posts/{id}/comments", require(http.HandlerFunc(po.comment)))
		mux.Handle("GET /v1/feed", require(http.HandlerFunc(po.feed)))
	}

	if app.Seed != nil {
		ih := newInternalHandler(app.InternalJobSecret, app.Seed)
		mux.HandleFunc("POST /v1/internal/seed-demo", ih.seedDemo)
	}
}
