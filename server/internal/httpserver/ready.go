package httpserver

import (
	"log"
	"net/http"

	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	"gorm.io/gorm"
)

func handleReady(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("ready: sql db: %v", err)
			apperrors.WriteError(w, apperrors.UnavailableCause(apperrors.CodeReadySQLGetterFailed, apperrors.MsgReadySQLGetterFailed, err))
			return
		}
		if err := sqlDB.PingContext(r.Context()); err != nil {
			log.Printf("ready: db ping: %v", err)
			apperrors.WriteError(w, apperrors.UnavailableCause(apperrors.CodeReadyDatabasePingFailed, apperrors.MsgReadyDatabasePingFailed, err))
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	}
}
