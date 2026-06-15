package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	authrepo "github.com/unipe/linkedin/backend/server/internal/auth/repository"
	authsvc "github.com/unipe/linkedin/backend/server/internal/auth/service"
	catalogrepo "github.com/unipe/linkedin/backend/server/internal/catalog/repository"
	connrepo "github.com/unipe/linkedin/backend/server/internal/connection/repository"
	connsvc "github.com/unipe/linkedin/backend/server/internal/connection/service"
	eventrepo "github.com/unipe/linkedin/backend/server/internal/event/repository"
	eventsvc "github.com/unipe/linkedin/backend/server/internal/event/service"
	"github.com/unipe/linkedin/backend/server/internal/httpserver"
	"github.com/unipe/linkedin/backend/server/internal/middleware"
	"github.com/unipe/linkedin/backend/server/internal/migrate"
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

func main() {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	db, err := postgres.Open(dsn)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	if skip := strings.TrimSpace(os.Getenv("SKIP_SQL_MIGRATIONS")); skip != "1" && !strings.EqualFold(skip, "true") {
		dir := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR"))
		if dir == "" {
			dir = migrate.ResolveDir()
		}
		if dir != "" {
			sqlDB, err := db.DB()
			if err != nil {
				log.Fatalf("sql db: %v", err)
			}
			mctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			err = migrate.Up(mctx, sqlDB, dir)
			cancel()
			if err != nil {
				log.Fatalf("sql migrate: %v", err)
			}
			log.Printf("sql migrations applied from %s", dir)
		} else {
			log.Print("warning: SQL migrations skipped (set MIGRATIONS_DIR)")
		}
	}

	jwt, err := jwtmgr.NewManager(jwtSecret, 7*24*time.Hour)
	if err != nil {
		log.Fatalf("jwt: %v", err)
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
	postService := postsvc.New(postRepository, connectionRepository, db)
	eventService := eventsvc.New(eventRepository)
	searchRepository := searchrepo.New(db)
	esClient := elasticsearch.New(strings.TrimSpace(os.Getenv("ELASTICSEARCH_URL")))
	searchService := searchsvc.New(searchRepository, esClient)
	recommendationRepository := recorepo.New(db)
	recommendationService := recosvc.New(recommendationRepository)
	seedService := seedsvc.New(authService, profileService, catalogRepository, profileRepository, connectionService, postService)

	app := &httpserver.App{
		DB:                db,
		InternalJobSecret: os.Getenv("INTERNAL_JOB_TOKEN"),
		JWTSecret:         jwtSecret,
		Auth:              authService,
		Profiles:          profileService,
		Connections:       connectionService,
		Posts:             postService,
		Events:            eventService,
		Search:            searchService,
		Recommendations:   recommendationService,
		Seed:              seedService,
	}

	mwCfg := middleware.LoadConfigFromEnv()
	handler := httpserver.NewHandler(app, mwCfg)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("api listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
