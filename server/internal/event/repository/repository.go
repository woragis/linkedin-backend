package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/models"
	"github.com/unipe/linkedin/backend/server/internal/platform/outbox"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

type IncomingEvent struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
	At      *time.Time     `json:"at"`
}

func (r *Repository) InsertBatch(ctx context.Context, userID *uuid.UUID, events []IncomingEvent) (int, error) {
	if len(events) == 0 {
		return 0, nil
	}
	inserted := 0
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, ev := range events {
			if ev.Type == "" {
				continue
			}
			payload := ev.Payload
			if payload == nil {
				payload = map[string]any{}
			}
			raw, err := json.Marshal(payload)
			if err != nil {
				return err
			}
			createdAt := time.Now().UTC()
			if ev.At != nil {
				createdAt = ev.At.UTC()
			}
			e := models.Event{
				ID:        uuid.New(),
				UserID:    userID,
				EventType: ev.Type,
				Payload:   datatypes.JSON(raw),
				CreatedAt: createdAt,
			}
			if err := tx.Create(&e).Error; err != nil {
				return err
			}
			jobs := []outbox.Job{
				{JobType: "analytics.process_event", Payload: map[string]any{"event_id": e.ID.String()}},
			}
			if err := outbox.EnqueueTx(tx, jobs...); err != nil {
				return err
			}
			inserted++
		}
		return nil
	})
	return inserted, err
}
