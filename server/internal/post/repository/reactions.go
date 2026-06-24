package repository

import (
	"context"

	"github.com/google/uuid"
)

const (
	TargetPost    = "post"
	TargetComment = "comment"
)

type ReactionRow struct {
	Kind  string
	Count int64
}

func (r *Repository) UpsertContentReaction(
	ctx context.Context,
	targetType string,
	targetID, userID uuid.UUID,
	kind string,
) error {
	err := r.db.WithContext(ctx).Exec(`
INSERT INTO content_reactions (target_type, target_id, user_id, kind)
VALUES (?, ?, ?, ?)
ON CONFLICT (target_type, target_id, user_id)
DO UPDATE SET kind = EXCLUDED.kind, created_at = now()
`, targetType, targetID, userID, kind).Error
	if err != nil {
		return err
	}
	if targetType == TargetPost {
		return r.db.WithContext(ctx).Exec(`
INSERT INTO reactions (post_id, user_id, kind)
VALUES (?, ?, ?)
ON CONFLICT (post_id, user_id)
DO UPDATE SET kind = EXCLUDED.kind, created_at = now()
`, targetID, userID, kind).Error
	}
	return nil
}

func (r *Repository) ContentReactionCount(ctx context.Context, targetType string, targetID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Raw(`
SELECT COUNT(*)::bigint FROM content_reactions
WHERE target_type = ? AND target_id = ?
`, targetType, targetID).Scan(&n).Error
	return n, err
}

func (r *Repository) ReactionSummary(
	ctx context.Context,
	targetType string,
	targetID uuid.UUID,
) (map[string]int64, error) {
	var rows []ReactionRow
	err := r.db.WithContext(ctx).Raw(`
SELECT kind, COUNT(*)::bigint AS count
FROM content_reactions
WHERE target_type = ? AND target_id = ?
GROUP BY kind
`, targetType, targetID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, row := range rows {
		out[row.Kind] = row.Count
	}
	return out, nil
}

func (r *Repository) MyContentReaction(
	ctx context.Context,
	targetType string,
	targetID, userID uuid.UUID,
) (string, error) {
	if userID == uuid.Nil {
		return "", nil
	}
	var kind string
	err := r.db.WithContext(ctx).Raw(`
SELECT kind FROM content_reactions
WHERE target_type = ? AND target_id = ? AND user_id = ?
`, targetType, targetID, userID).Scan(&kind).Error
	if err != nil {
		return "", nil
	}
	return kind, nil
}

type targetSummaryRow struct {
	TargetID uuid.UUID `gorm:"column:target_id"`
	Kind     string    `gorm:"column:kind"`
	Count    int64     `gorm:"column:count"`
}

func (r *Repository) BatchReactionSummaries(
	ctx context.Context,
	targetType string,
	targetIDs []uuid.UUID,
) (map[uuid.UUID]map[string]int64, error) {
	out := make(map[uuid.UUID]map[string]int64)
	if len(targetIDs) == 0 {
		return out, nil
	}
	var rows []targetSummaryRow
	err := r.db.WithContext(ctx).Raw(`
SELECT target_id, kind, COUNT(*)::bigint AS count
FROM content_reactions
WHERE target_type = ? AND target_id IN ?
GROUP BY target_id, kind
`, targetType, targetIDs).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if out[row.TargetID] == nil {
			out[row.TargetID] = make(map[string]int64)
		}
		out[row.TargetID][row.Kind] = row.Count
	}
	return out, nil
}

type myReactionRow struct {
	TargetID uuid.UUID `gorm:"column:target_id"`
	Kind     string    `gorm:"column:kind"`
}

func (r *Repository) BatchMyReactions(
	ctx context.Context,
	targetType string,
	targetIDs []uuid.UUID,
	userID uuid.UUID,
) (map[uuid.UUID]string, error) {
	out := make(map[uuid.UUID]string)
	if userID == uuid.Nil || len(targetIDs) == 0 {
		return out, nil
	}
	var rows []myReactionRow
	err := r.db.WithContext(ctx).Raw(`
SELECT target_id, kind
FROM content_reactions
WHERE target_type = ? AND target_id IN ? AND user_id = ?
`, targetType, targetIDs, userID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.TargetID] = row.Kind
	}
	return out, nil
}
