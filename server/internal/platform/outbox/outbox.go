package outbox

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Job struct {
	JobType string
	Payload map[string]any
}

func EnqueueTx(tx *gorm.DB, jobs ...Job) error {
	for _, j := range jobs {
		raw, err := json.Marshal(j.Payload)
		if err != nil {
			return fmt.Errorf("marshal outbox payload: %w", err)
		}
		row := models.OutboxJob{
			ID:      uuid.New(),
			JobType: j.JobType,
			Payload: datatypes.JSON(raw),
		}
		if err := tx.Create(&row).Error; err != nil {
			return fmt.Errorf("insert outbox job: %w", err)
		}
	}
	return nil
}

func Enqueue(ctx context.Context, db *gorm.DB, jobs ...Job) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return EnqueueTx(tx, jobs...)
	})
}
