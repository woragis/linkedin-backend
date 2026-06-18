package httpserver

import (
	"net/http"

	"github.com/unipe/linkedin/backend/server/internal/observability/metrics"
)

func Mount(mux *http.ServeMux, multi *MultiApp) {
	mux.HandleFunc("GET /health", handleHealth)
	mux.Handle("GET /metrics", metrics.Handler())

	if multi != nil {
		mux.HandleFunc("GET /ready", handleReadyMulti(multi))
	}

	if multi == nil || multi.Live == nil && multi.Volume == nil {
		return
	}

	mux.HandleFunc("POST /v1/auth/register", func(w http.ResponseWriter, r *http.Request) {
		app := multi.AppFor(r)
		if app == nil || app.Auth == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		newAuthHandler(app.Auth).register(w, r)
	})
	mux.HandleFunc("POST /v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		app := multi.AppFor(r)
		if app == nil || app.Auth == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		newAuthHandler(app.Auth).login(w, r)
	})

	mux.Handle("POST /v1/events", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newEventHandler(app.Events).ingest(w, r)
	}))

	mux.Handle("GET /v1/me", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newProfileHandler(app.Profiles).me(w, r)
	}))
	mux.Handle("PATCH /v1/me/profile", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newProfileHandler(app.Profiles).patchMe(w, r)
	}))
	mux.HandleFunc("GET /v1/users/{slug}", func(w http.ResponseWriter, r *http.Request) {
		app := multi.AppFor(r)
		if app == nil || app.Profiles == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		newProfileHandler(app.Profiles).getBySlug(w, r)
	})

	mux.Handle("GET /v1/me/experiences", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newProfileHandler(app.Profiles).listExperiences(w, r)
	}))
	mux.Handle("POST /v1/me/experiences", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newProfileHandler(app.Profiles).createExperience(w, r)
	}))
	mux.Handle("PATCH /v1/me/experiences/{id}", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newProfileHandler(app.Profiles).patchExperience(w, r)
	}))
	mux.Handle("DELETE /v1/me/experiences/{id}", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newProfileHandler(app.Profiles).deleteExperience(w, r)
	}))

	mux.Handle("GET /v1/me/educations", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newProfileHandler(app.Profiles).listEducations(w, r)
	}))
	mux.Handle("POST /v1/me/educations", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newProfileHandler(app.Profiles).createEducation(w, r)
	}))
	mux.Handle("PATCH /v1/me/educations/{id}", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newProfileHandler(app.Profiles).patchEducation(w, r)
	}))
	mux.Handle("DELETE /v1/me/educations/{id}", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newProfileHandler(app.Profiles).deleteEducation(w, r)
	}))

	mux.Handle("GET /v1/me/skills", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newProfileHandler(app.Profiles).listSkills(w, r)
	}))
	mux.Handle("PUT /v1/me/skills", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newProfileHandler(app.Profiles).replaceSkills(w, r)
	}))

	mux.Handle("POST /v1/connections/request", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newConnectionHandler(app.Connections).request(w, r)
	}))
	mux.Handle("PATCH /v1/connections/{id}/accept", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newConnectionHandler(app.Connections).accept(w, r)
	}))
	mux.Handle("PATCH /v1/connections/{id}/reject", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newConnectionHandler(app.Connections).reject(w, r)
	}))
	mux.Handle("GET /v1/connections", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newConnectionHandler(app.Connections).list(w, r)
	}))
	mux.Handle("GET /v1/connections/pending", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newConnectionHandler(app.Connections).pending(w, r)
	}))

	mux.Handle("POST /v1/posts", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newPostHandler(app.Posts).create(w, r)
	}))
	mux.HandleFunc("GET /v1/posts/{id}", func(w http.ResponseWriter, r *http.Request) {
		app := multi.AppFor(r)
		if app == nil || app.Posts == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		newPostHandler(app.Posts).get(w, r)
	})
	mux.Handle("POST /v1/posts/{id}/reactions", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newPostHandler(app.Posts).react(w, r)
	}))
	mux.Handle("POST /v1/posts/{id}/comments", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newPostHandler(app.Posts).comment(w, r)
	}))
	mux.HandleFunc("GET /v1/posts/{id}/comments", func(w http.ResponseWriter, r *http.Request) {
		app := multi.AppFor(r)
		if app == nil || app.Posts == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		newPostHandler(app.Posts).listComments(w, r)
	})
	mux.Handle("GET /v1/feed", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newPostHandler(app.Posts).feed(w, r)
	}))

	mux.HandleFunc("GET /v1/search/people", func(w http.ResponseWriter, r *http.Request) {
		app := multi.AppFor(r)
		if app == nil || app.Search == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		newSearchHandler(app.Search).people(w, r)
	})
	mux.HandleFunc("GET /v1/search/posts", func(w http.ResponseWriter, r *http.Request) {
		app := multi.AppFor(r)
		if app == nil || app.Search == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		newSearchHandler(app.Search).posts(w, r)
	})

	mux.Handle("GET /v1/recommendations/people", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newRecommendationHandler(app.Recommendations).people(w, r)
	}))
	mux.Handle("GET /v1/recommendations/meta", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newRecommendationHandler(app.Recommendations).peopleMeta(w, r)
	}))

	mux.Handle("GET /v1/network/graph", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newGraphHandler(app.Graph).userGraph(w, r)
	}))
	mux.HandleFunc("GET /v1/network/influencers", func(w http.ResponseWriter, r *http.Request) {
		app := multi.AppFor(r)
		if app == nil || app.Graph == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		newGraphHandler(app.Graph).influencers(w, r)
	})
	mux.Handle("GET /v1/network/link-predictions", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newGraphHandler(app.Graph).linkPredictions(w, r)
	}))

	mux.Handle("GET /v1/analytics/overview", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newAnalyticsHandler(app.Analytics).overview(w, r)
	}))
	mux.Handle("GET /v1/analytics/top-posts", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newAnalyticsHandler(app.Analytics).topPosts(w, r)
	}))
	mux.Handle("GET /v1/analytics/cohorts", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newAnalyticsHandler(app.Analytics).cohorts(w, r)
	}))
	mux.Handle("GET /v1/analytics/churn", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newAnalyticsHandler(app.Analytics).churn(w, r)
	}))
	mux.Handle("GET /v1/analytics/dau", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newAnalyticsHandler(app.Analytics).dau(w, r)
	}))
	mux.Handle("GET /v1/analytics/experiments", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newAnalyticsHandler(app.Analytics).experiments(w, r)
	}))
	mux.Handle("GET /v1/analytics/ml-models", multi.require(func(app *App, w http.ResponseWriter, r *http.Request) {
		newAnalyticsHandler(app.Analytics).mlModels(w, r)
	}))

	mux.HandleFunc("POST /v1/internal/seed-demo", func(w http.ResponseWriter, r *http.Request) {
		app := multi.AppFor(r)
		if app == nil || app.Seed == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		newInternalHandler(app.InternalJobSecret, app.Seed).seedDemo(w, r)
	})
}
