package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/models"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindBetween(ctx context.Context, a, b uuid.UUID) (*models.Connection, error) {
	var c models.Connection
	err := r.db.WithContext(ctx).Where(
		"(requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)",
		a, b, b, a,
	).First(&c).Error
	return &c, err
}

func (r *Repository) Create(ctx context.Context, c *models.Connection) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.Connection, error) {
	var c models.Connection
	err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error
	return &c, err
}

func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&models.Connection{}).Where("id = ?", id).
		Update("status", status).Error
}

func (r *Repository) ListAcceptedPeerIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var rows []models.Connection
	err := r.db.WithContext(ctx).Where(
		"status = ? AND (requester_id = ? OR addressee_id = ?)",
		"accepted", userID, userID,
	).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	peers := make([]uuid.UUID, 0, len(rows))
	for _, c := range rows {
		if c.RequesterID == userID {
			peers = append(peers, c.AddresseeID)
		} else {
			peers = append(peers, c.RequesterID)
		}
	}
	return peers, nil
}

func (r *Repository) ListAccepted(ctx context.Context, userID uuid.UUID) ([]models.Connection, error) {
	var rows []models.Connection
	err := r.db.WithContext(ctx).Where(
		"status = ? AND (requester_id = ? OR addressee_id = ?)",
		"accepted", userID, userID,
	).Order("updated_at DESC").Find(&rows).Error
	return rows, err
}

func (r *Repository) ListPendingIncoming(ctx context.Context, userID uuid.UUID) ([]models.Connection, error) {
	var rows []models.Connection
	err := r.db.WithContext(ctx).Where("status = ? AND addressee_id = ?", "pending", userID).
		Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *Repository) ListAcceptedWithPeers(ctx context.Context, userID uuid.UUID) ([]ConnectionPeer, error) {
	var rows []ConnectionPeer
	err := r.db.WithContext(ctx).Raw(`
SELECT c.id,
       p.user_id, p.slug, p.full_name, p.headline, p.avatar_url,
       c.updated_at AS connected_at
FROM connections c
JOIN profiles p ON p.user_id = CASE
  WHEN c.requester_id = ? THEN c.addressee_id ELSE c.requester_id END
WHERE c.status = 'accepted' AND ? IN (c.requester_id, c.addressee_id)
ORDER BY c.updated_at DESC
`, userID, userID).Scan(&rows).Error
	return rows, err
}

type ConnectionPeer struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id" gorm:"column:user_id"`
	Slug        string    `json:"slug"`
	FullName    string    `json:"full_name"`
	Headline    string    `json:"headline"`
	AvatarURL   *string   `json:"avatar_url"`
	ConnectedAt time.Time `json:"connected_at" gorm:"column:connected_at"`
}

func (r *Repository) UserExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// IsSimulatorUser reports synthetic agents created by worker-simulator.
func (r *Repository) IsSimulatorUser(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("simulator_agents").Where("user_id = ?", userID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) CountAccepted(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Connection{}).Where(
		"status = ? AND (requester_id = ? OR addressee_id = ?)",
		"accepted", userID, userID,
	).Count(&count).Error
	return count, err
}

func (r *Repository) MutualCount(ctx context.Context, a, b uuid.UUID) (int64, error) {
	query := `
SELECT COUNT(*)::bigint FROM connections c1
JOIN connections c2 ON c2.status = 'accepted'
  AND c2.requester_id <> c2.addressee_id
WHERE c1.status = 'accepted'
  AND ((c1.requester_id = ? AND c1.addressee_id = c2.requester_id)
    OR (c1.addressee_id = ? AND c1.requester_id = c2.requester_id))
  AND ((c2.requester_id = ? AND c2.addressee_id = c1.requester_id)
    OR (c2.addressee_id = ? AND c2.requester_id = c1.addressee_id))
  AND c1.requester_id <> c1.addressee_id
`
	var count int64
	err := r.db.WithContext(ctx).Raw(query, a, a, b, b).Scan(&count).Error
	if err != nil {
		return 0, fmt.Errorf("mutual count: %w", err)
	}
	return count, nil
}
