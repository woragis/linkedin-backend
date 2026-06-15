package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Overview struct {
	DAU            int   `json:"dau"`
	MAU            int   `json:"mau"`
	TotalUsers     int64 `json:"total_users"`
	TotalPosts     int64 `json:"total_posts"`
	ChurnHighCount int64 `json:"churn_high_count"`
	ChurnMedium    int64 `json:"churn_medium_count"`
	ChurnLow       int64 `json:"churn_low_count"`
}

type TopPost struct {
	PostID    uuid.UUID `json:"post_id"`
	Body      string    `json:"body"`
	Author    string    `json:"author_name"`
	Views     int       `json:"views"`
	Reactions int       `json:"reactions"`
	Comments  int       `json:"comments"`
}

type CohortRow struct {
	CohortWeek   time.Time `json:"cohort_week"`
	WeekOffset   int       `json:"week_offset"`
	ActiveUsers  int       `json:"active_users"`
	CohortSize   int       `json:"cohort_size"`
	RetentionPct float64   `json:"retention_pct"`
}

type ChurnUser struct {
	UserID           uuid.UUID `json:"user_id"`
	Slug             string    `json:"slug"`
	FullName         string    `json:"full_name"`
	ChurnProbability float64   `json:"churn_probability"`
	RiskTier         string    `json:"risk_tier"`
}

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Overview(ctx context.Context) (*Overview, error) {
	var o Overview
	_ = r.db.WithContext(ctx).Raw(`
SELECT COALESCE(dau, 0) FROM analytics.daily_active_users
WHERE day = CURRENT_DATE
`).Scan(&o.DAU)

	_ = r.db.WithContext(ctx).Raw(`
SELECT COUNT(DISTINCT user_id)::int FROM events
WHERE user_id IS NOT NULL AND created_at >= now() - interval '30 days'
`).Scan(&o.MAU)

	r.db.WithContext(ctx).Table("users").Count(&o.TotalUsers)
	r.db.WithContext(ctx).Table("posts").Where("deleted_at IS NULL").Count(&o.TotalPosts)

	r.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM user_churn_scores WHERE risk_tier = 'high'`).Scan(&o.ChurnHighCount)
	r.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM user_churn_scores WHERE risk_tier = 'medium'`).Scan(&o.ChurnMedium)
	r.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM user_churn_scores WHERE risk_tier = 'low'`).Scan(&o.ChurnLow)

	return &o, nil
}

func (r *Repository) TopPosts(ctx context.Context, limit int) ([]TopPost, error) {
	if limit <= 0 {
		limit = 10
	}
	var rows []TopPost
	err := r.db.WithContext(ctx).Raw(`
SELECT pe.post_id, LEFT(po.body, 120), pr.full_name, pe.views, pe.reactions, pe.comments
FROM analytics.post_engagement_daily pe
JOIN posts po ON po.id = pe.post_id
JOIN profiles pr ON pr.user_id = po.author_id
WHERE pe.day = CURRENT_DATE
ORDER BY (pe.views + pe.reactions * 2 + pe.comments * 3) DESC
LIMIT ?
`, limit).Scan(&rows).Error
	return rows, err
}

func (r *Repository) Cohorts(ctx context.Context) ([]CohortRow, error) {
	var rows []CohortRow
	err := r.db.WithContext(ctx).Raw(`
SELECT cohort_week, week_offset, active_users, cohort_size,
       CASE WHEN cohort_size > 0 THEN (active_users::float / cohort_size) * 100 ELSE 0 END
FROM analytics.user_cohorts
ORDER BY cohort_week DESC, week_offset ASC
LIMIT 100
`).Scan(&rows).Error
	return rows, err
}

func (r *Repository) ChurnUsers(ctx context.Context, limit int) ([]ChurnUser, error) {
	if limit <= 0 {
		limit = 20
	}
	var rows []ChurnUser
	err := r.db.WithContext(ctx).Raw(`
SELECT cs.user_id, p.slug, p.full_name, cs.churn_probability, cs.risk_tier
FROM user_churn_scores cs
JOIN profiles p ON p.user_id = cs.user_id
ORDER BY cs.churn_probability DESC
LIMIT ?
`, limit).Scan(&rows).Error
	return rows, err
}

func (r *Repository) DailyActive(ctx context.Context, days int) ([]struct {
	Day time.Time `json:"day"`
	DAU int       `json:"dau"`
}, error) {
	if days <= 0 {
		days = 30
	}
	var rows []struct {
		Day time.Time `json:"day"`
		DAU int       `json:"dau"`
	}
	err := r.db.WithContext(ctx).Raw(`
SELECT day, dau FROM analytics.daily_active_users
WHERE day >= CURRENT_DATE - ?
ORDER BY day ASC
`, days).Scan(&rows).Error
	return rows, err
}
