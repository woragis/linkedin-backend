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

	"github.com/unipe/linkedin/backend/server/internal/httpserver"
	"github.com/unipe/linkedin/backend/server/internal/middleware"
	"github.com/unipe/linkedin/backend/server/internal/migrate"
	"github.com/unipe/linkedin/backend/server/internal/platform/postgres"
	"github.com/unipe/linkedin/backend/server/internal/platform/realm"
	"gorm.io/gorm"
)

func main() {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	volumeDSN := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if volumeDSN == "" {
		log.Fatal("DATABASE_URL is required")
	}

	liveDSN := strings.TrimSpace(os.Getenv("DATABASE_URL_LIVE"))
	if liveDSN == "" {
		liveDSN = volumeDSN
	}

	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	redisURL := os.Getenv("REDIS_URL")
	esURL := strings.TrimSpace(os.Getenv("ELASTICSEARCH_URL"))
	internalToken := os.Getenv("INTERNAL_JOB_TOKEN")

	volumeDB, err := openAndMigrate(volumeDSN)
	if err != nil {
		log.Fatalf("volume database: %v", err)
	}

	var liveDB = volumeDB
	if liveDSN != volumeDSN {
		liveDB, err = openAndMigrate(liveDSN)
		if err != nil {
			log.Fatalf("live database: %v", err)
		}
	}

	volumeApp := httpserver.BuildApp(httpserver.BuildConfigForRealm(
		volumeDB, redisURL, esURL, internalToken, jwtSecret, realm.Volume,
	))
	liveApp := httpserver.BuildApp(httpserver.BuildConfigForRealm(
		liveDB, redisURL, esURL, internalToken, jwtSecret, realm.Live,
	))

	multi := httpserver.NewMultiApp(volumeApp, liveApp)

	mwCfg := middleware.LoadConfigFromEnv()
	handler := httpserver.NewHandler(multi, mwCfg)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("api listening on %s (realms: volume + live)", addr)
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

func openAndMigrate(dsn string) (*gorm.DB, error) {
	db, err := postgres.Open(dsn)
	if err != nil {
		return nil, err
	}

	if skip := strings.TrimSpace(os.Getenv("SKIP_SQL_MIGRATIONS")); skip != "1" && !strings.EqualFold(skip, "true") {
		dir := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR"))
		if dir == "" {
			dir = migrate.ResolveDir()
		}
		if dir != "" {
			sqlDB, err := db.DB()
			if err != nil {
				return nil, err
			}
			mctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			err = migrate.Up(mctx, sqlDB, dir)
			cancel()
			if err != nil {
				return nil, err
			}
			log.Printf("sql migrations applied from %s", dir)
		}
	}

	return db, nil
}
