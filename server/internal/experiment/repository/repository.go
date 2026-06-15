package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FeedVariant(ctx context.Context, userID uuid.UUID, experimentName string) (string, error) {
	var row struct {
		ID       uuid.UUID
		Variants json.RawMessage
	}
	err := r.db.WithContext(ctx).Raw(`
SELECT id, variants FROM ab_experiments
WHERE name = ? AND status = 'active' LIMIT 1
`, experimentName).Scan(&row).Error
	if err != nil || row.ID == uuid.Nil {
		return "chronological", nil
	}

	var assigned string
	err = r.db.WithContext(ctx).Raw(`
SELECT variant FROM ab_assignments WHERE experiment_id = ? AND user_id = ?
`, row.ID, userID).Scan(&assigned).Error
	if err == nil && assigned != "" {
		return assigned, nil
	}

	var variantList []string
	if err := json.Unmarshal(row.Variants, &variantList); err != nil || len(variantList) == 0 {
		variantList = []string{"chronological", "ranked"}
	}
	idx := hashIndex(userID.String()) % len(variantList)
	assigned = variantList[idx]

	_ = r.db.WithContext(ctx).Exec(`
INSERT INTO ab_assignments (experiment_id, user_id, variant)
VALUES (?, ?, ?) ON CONFLICT DO NOTHING
`, row.ID, userID, assigned).Error

	return assigned, nil
}

func hashIndex(s string) int {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return int(h)
}
