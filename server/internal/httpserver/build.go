package httpserver

import (
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
	jwtmgr "github.com/unipe/linkedin/backend/server/internal/platform/jwt"
	"github.com/unipe/linkedin/backend/server/internal/platform/cache"
	"github.com/unipe/linkedin/backend/server/internal/platform/realm"
	"github.com/unipe/linkedin/backend/server/internal/search/elasticsearch"
	postrepo "github.com/unipe/linkedin/backend/server/internal/post/repository"
	postsvc "github.com/unipe/linkedin/backend/server/internal/post/service"
	profilerepo "github.com/unipe/linkedin/backend/server/internal/profile/repository"
	profilesvc "github.com/unipe/linkedin/backend/server/internal/profile/service"
	recorepo "github.com/unipe/linkedin/backend/server/internal/recommendation/repository"
	recosvc "github.com/unipe/linkedin/backend/server/internal/recommendation/service"
	seedsvc "github.com/unipe/linkedin/backend/server/internal/seed/service"
	searchrepo "github.com/unipe/linkedin/backend/server/internal/search/repository"
	searchsvc "github.com/unipe/linkedin/backend/server/internal/search/service"
	"gorm.io/gorm"
)

type BuildConfig struct {
	DB                *gorm.DB
	RedisURL          string
	ElasticsearchURL  string
	InternalJobSecret string
	JWTSecret         string
	CachePrefix       string
}

func BuildApp(cfg BuildConfig) *App {
	db := cfg.DB
	authRepository := authrepo.New(db)
	profileRepository := profilerepo.New(db)
	catalogRepository := catalogrepo.New(db)
	connectionRepository := connrepo.New(db)
	postRepository := postrepo.New(db)
	eventRepository := eventrepo.New(db)

	jwt, err := jwtmgr.NewManager(cfg.JWTSecret, 7*24*time.Hour)
	if err != nil {
		panic(err)
	}
	authService := authsvc.New(authRepository, db, jwt)
	profileService := profilesvc.New(profileRepository, catalogRepository, db)
	connectionService := connsvc.New(connectionRepository, db)
	experimentRepository := experimentrepo.New(db)
	experimentService := experimentsvc.New(experimentRepository)

	feedCache, _ := cache.NewFeedCache(cfg.RedisURL, 60*time.Second, cfg.CachePrefix)

	postService := postsvc.New(postRepository, connectionRepository, experimentService, feedCache, db)
	eventService := eventsvc.New(eventRepository)
	searchRepository := searchrepo.New(db)
	esClient := elasticsearch.New(cfg.ElasticsearchURL)
	searchService := searchsvc.New(searchRepository, esClient)
	recommendationRepository := recorepo.New(db)
	recommendationService := recosvc.New(recommendationRepository, experimentRepository)
	graphRepository := graphrepo.New(db)
	graphService := graphsvc.New(graphRepository)
	analyticsRepository := analyticsrepo.New(db)
	analyticsService := analyticsvc.New(analyticsRepository, experimentRepository)
	seedService := seedsvc.New(authService, profileService, catalogRepository, profileRepository, connectionService, postService)

	return &App{
		DB:                db,
		InternalJobSecret: cfg.InternalJobSecret,
		JWTSecret:         cfg.JWTSecret,
		Auth:              authService,
		Profiles:          profileService,
		Connections:       connectionService,
		Posts:             postService,
		Events:            eventService,
		Search:            searchService,
		Recommendations:   recommendationService,
		Graph:             graphService,
		Analytics:         analyticsService,
		Experiments:       experimentService,
		Seed:              seedService,
	}
}

// NewMultiApp wires volume (primary DATABASE_URL) and live (DATABASE_URL_LIVE) stacks.
func NewMultiApp(volume, live *App) *MultiApp {
	if live == nil {
		live = volume
	}
	if volume == nil {
		volume = live
	}
	return &MultiApp{Volume: volume, Live: live}
}

func BuildConfigForRealm(
	db *gorm.DB,
	redisURL, esURL, internalToken, jwtSecret string,
	id realm.ID,
) BuildConfig {
	return BuildConfig{
		DB:                db,
		RedisURL:          redisURL,
		ElasticsearchURL:  esURL,
		InternalJobSecret: internalToken,
		JWTSecret:         jwtSecret,
		CachePrefix:       realm.CachePrefix(id),
	}
}
