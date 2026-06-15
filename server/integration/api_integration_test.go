//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	analyticsrepo "github.com/unipe/linkedin/backend/server/internal/analytics/repository"
	analyticsvc "github.com/unipe/linkedin/backend/server/internal/analytics/service"
	authrepo "github.com/unipe/linkedin/backend/server/internal/auth/repository"
	authsvc "github.com/unipe/linkedin/backend/server/internal/auth/service"
	catalogrepo "github.com/unipe/linkedin/backend/server/internal/catalog/repository"
	connrepo "github.com/unipe/linkedin/backend/server/internal/connection/repository"
	connsvc "github.com/unipe/linkedin/backend/server/internal/connection/service"
	eventrepo "github.com/unipe/linkedin/backend/server/internal/event/repository"
	eventsvc "github.com/unipe/linkedin/backend/server/internal/event/service"
	experimentrepo "github.com/unipe/linkedin/backend/server/internal/experiment/repository"
	experimentsvc "github.com/unipe/linkedin/backend/server/internal/experiment/service"
	graphrepo "github.com/unipe/linkedin/backend/server/internal/graph/repository"
	graphsvc "github.com/unipe/linkedin/backend/server/internal/graph/service"
	"github.com/unipe/linkedin/backend/server/internal/httpserver"
	"github.com/unipe/linkedin/backend/server/internal/middleware"
	"github.com/unipe/linkedin/backend/server/internal/migrate"
	"github.com/unipe/linkedin/backend/server/internal/platform/cache"
	jwtmgr "github.com/unipe/linkedin/backend/server/internal/platform/jwt"
	"github.com/unipe/linkedin/backend/server/internal/platform/postgres"
	postrepo "github.com/unipe/linkedin/backend/server/internal/post/repository"
	postsvc "github.com/unipe/linkedin/backend/server/internal/post/service"
	profilerepo "github.com/unipe/linkedin/backend/server/internal/profile/repository"
	profilesvc "github.com/unipe/linkedin/backend/server/internal/profile/service"
	recorepo "github.com/unipe/linkedin/backend/server/internal/recommendation/repository"
	recosvc "github.com/unipe/linkedin/backend/server/internal/recommendation/service"
	seedsvc "github.com/unipe/linkedin/backend/server/internal/seed/service"
	searchrepo "github.com/unipe/linkedin/backend/server/internal/search/repository"
	"github.com/unipe/linkedin/backend/server/internal/search/elasticsearch"
	searchsvc "github.com/unipe/linkedin/backend/server/internal/search/service"
)

func TestRegisterLoginProfilePatch(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}

	db, err := postgres.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	dir := migrate.ResolveDir()
	if dir != "" {
		mctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		if err := migrate.Up(mctx, sqlDB, dir); err != nil {
			cancel()
			t.Fatal(err)
		}
		cancel()
	}

	jwtSecret := "integration-test-secret"
	jwt, err := jwtmgr.NewManager(jwtSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	authRepository := authrepo.New(db)
	profileRepository := profilerepo.New(db)
	catalogRepository := catalogrepo.New(db)
	connectionRepository := connrepo.New(db)
	postRepository := postrepo.New(db)
	eventRepository := eventrepo.New(db)

	authService := authsvc.New(authRepository, db, jwt)
	profileService := profilesvc.New(profileRepository, catalogRepository, db)
	connectionService := connsvc.New(connectionRepository, db)
	experimentRepository := experimentrepo.New(db)
	experimentService := experimentsvc.New(experimentRepository)
	feedCache, _ := cache.NewFeedCache("", 60*time.Second)
	postService := postsvc.New(postRepository, connectionRepository, experimentService, feedCache, db)
	eventService := eventsvc.New(eventRepository)
	searchRepository := searchrepo.New(db)
	searchService := searchsvc.New(searchRepository, elasticsearch.New(""))
	recommendationRepository := recorepo.New(db)
	recommendationService := recosvc.New(recommendationRepository, experimentRepository)
	graphRepository := graphrepo.New(db)
	graphService := graphsvc.New(graphRepository)
	analyticsRepository := analyticsrepo.New(db)
	analyticsService := analyticsvc.New(analyticsRepository, experimentRepository)
	seedService := seedsvc.New(authService, profileService, catalogRepository, profileRepository, connectionService, postService)

	app := &httpserver.App{
		DB:              db,
		InternalJobSecret: "test-internal",
		JWTSecret:       jwtSecret,
		Auth:            authService,
		Profiles:        profileService,
		Connections:     connectionService,
		Posts:           postService,
		Events:          eventService,
		Search:          searchService,
		Recommendations: recommendationService,
		Graph:           graphService,
		Analytics:       analyticsService,
		Experiments:     experimentService,
		Seed:            seedService,
	}

	srv := httptest.NewServer(httpserver.NewHandler(app, middleware.Config{}))
	defer srv.Close()

	email := fmt.Sprintf("integration-%d@test.local", time.Now().UnixNano())
	password := "integration123"

	code, reg := httpJSON(t, srv.URL, http.MethodPost, "/v1/auth/register", map[string]any{
		"email":     email,
		"password":  password,
		"full_name": "Integration User",
	}, nil)
	if code != http.StatusCreated {
		t.Fatalf("register status=%d body=%v", code, reg)
	}
	token, _ := reg["token"].(string)
	if token == "" {
		t.Fatal("missing token")
	}
	auth := map[string]string{"Authorization": "Bearer " + token}

	code, login := httpJSON(t, srv.URL, http.MethodPost, "/v1/auth/login", map[string]any{
		"email":    email,
		"password": password,
	}, nil)
	if code != http.StatusOK {
		t.Fatalf("login status=%d body=%v", code, login)
	}

	headline := "Integration headline"
	code, profile := httpJSON(t, srv.URL, http.MethodPatch, "/v1/me/profile", map[string]any{
		"headline": headline,
	}, auth)
	if code != http.StatusOK {
		t.Fatalf("patch profile status=%d body=%v", code, profile)
	}
	if profile["headline"] != headline {
		t.Fatalf("expected headline %q, got %v", headline, profile["headline"])
	}
}

func httpJSON(t *testing.T, base, method, path string, body any, headers map[string]string) (int, map[string]any) {
	t.Helper()
	var r io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		r = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, base+path, r)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if len(data) > 0 {
		if err := json.Unmarshal(data, &out); err != nil {
			t.Fatalf("%s %s: decode json: %v body=%s", method, path, err, string(data))
		}
	}
	return res.StatusCode, out
}
