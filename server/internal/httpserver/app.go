package httpserver

import "gorm.io/gorm"

type App struct {
	DB                *gorm.DB
	InternalJobSecret string
}
