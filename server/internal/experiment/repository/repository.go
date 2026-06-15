package repository

import (
	"context"
	"encoding/json"
	"time"

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

type ABExperimentResult struct {
	ExperimentID   uuid.UUID `json:"experiment_id"`
	ExperimentName string    `json:"experiment_name"`
	PrimaryMetric  string    `json:"primary_metric"`
	Variant        string    `json:"variant"`
	SampleSize     int       `json:"sample_size"`
	MetricValue    float64   `json:"metric_value"`
	CILower        *float64  `json:"ci_lower,omitempty"`
	CIUpper        *float64  `json:"ci_upper,omitempty"`
	ComputedAt     time.Time `json:"computed_at"`
}

func (r *Repository) ABExperimentResults(ctx context.Context) ([]ABExperimentResult, error) {
	var rows []ABExperimentResult
	err := r.db.WithContext(ctx).Raw(`
SELECT DISTINCT ON (r.experiment_id, r.variant)
       r.experiment_id, e.name AS experiment_name, e.primary_metric, r.variant,
       r.sample_size, r.metric_value, r.ci_lower, r.ci_upper, r.computed_at
FROM ab_experiment_results r
JOIN ab_experiments e ON e.id = r.experiment_id
ORDER BY r.experiment_id, r.variant, r.computed_at DESC
`).Scan(&rows).Error
	return rows, err
}

type MLModel struct {
	ID           uuid.UUID       `json:"id"`
	ModelName    string          `json:"model_name"`
	Version      string          `json:"version"`
	Metrics      json.RawMessage `json:"metrics"`
	ArtifactPath *string         `json:"artifact_path,omitempty"`
	IsActive     bool            `json:"is_active"`
	TrainedAt    time.Time       `json:"trained_at"`
}

func (r *Repository) ActiveMLModel(ctx context.Context, modelName string) (*MLModel, error) {
	var m MLModel
	err := r.db.WithContext(ctx).Raw(`
SELECT id, model_name, version, metrics, artifact_path, is_active, trained_at
FROM model_versions
WHERE model_name = ? AND is_active = true
ORDER BY trained_at DESC
LIMIT 1
`, modelName).Scan(&m).Error
	if err != nil {
		return nil, err
	}
	if m.ID == uuid.Nil {
		return nil, gorm.ErrRecordNotFound
	}
	return &m, nil
}

func (r *Repository) ListMLModels(ctx context.Context) ([]MLModel, error) {
	var rows []MLModel
	err := r.db.WithContext(ctx).Raw(`
SELECT id, model_name, version, metrics, artifact_path, is_active, trained_at
FROM model_versions
ORDER BY model_name ASC, trained_at DESC
`).Scan(&rows).Error
	return rows, err
}
