package httpserver

import (
	authsvc "github.com/unipe/linkedin/backend/server/internal/auth/service"
	profilesvc "github.com/unipe/linkedin/backend/server/internal/profile/service"
	seedsvc "github.com/unipe/linkedin/backend/server/internal/seed/service"
	"gorm.io/gorm"
)

type App struct {
	DB                *gorm.DB
	InternalJobSecret string
	JWTSecret         string

	Auth     *authsvc.Service
	Profiles *profilesvc.Service
	Seed     *seedsvc.Service
}
