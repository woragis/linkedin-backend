package httpserver

import (
	authsvc "github.com/unipe/linkedin/backend/server/internal/auth/service"
	connsvc "github.com/unipe/linkedin/backend/server/internal/connection/service"
	eventsvc "github.com/unipe/linkedin/backend/server/internal/event/service"
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
	Seed            *seedsvc.Service
}
