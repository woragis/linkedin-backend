package httpserver

import (
	authsvc "github.com/unipe/linkedin/backend/server/internal/auth/service"
	analyticsvc "github.com/unipe/linkedin/backend/server/internal/analytics/service"
	connsvc "github.com/unipe/linkedin/backend/server/internal/connection/service"
	eventsvc "github.com/unipe/linkedin/backend/server/internal/event/service"
	experimentsvc "github.com/unipe/linkedin/backend/server/internal/experiment/service"
	graphsvc "github.com/unipe/linkedin/backend/server/internal/graph/service"
	postsvc "github.com/unipe/linkedin/backend/server/internal/post/service"
	profilesvc "github.com/unipe/linkedin/backend/server/internal/profile/service"
	recosvc "github.com/unipe/linkedin/backend/server/internal/recommendation/service"
	seedsvc "github.com/unipe/linkedin/backend/server/internal/seed/service"
	searchsvc "github.com/unipe/linkedin/backend/server/internal/search/service"
	"gorm.io/gorm"
)

type App struct {
	DB                *gorm.DB
	InternalJobSecret string
	JWTSecret         string

	Auth            *authsvc.Service
	Profiles        *profilesvc.Service
	Connections     *connsvc.Service
	Posts           *postsvc.Service
	Events          *eventsvc.Service
	Search          *searchsvc.Service
	Recommendations *recosvc.Service
	Graph           *graphsvc.Service
	Analytics       *analyticsvc.Service
	Experiments     *experimentsvc.Service
	Seed            *seedsvc.Service
}
