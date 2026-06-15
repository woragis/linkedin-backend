package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/models"
	"github.com/unipe/linkedin/backend/server/internal/platform/slug"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindOrCreateInstitution(ctx context.Context, name string) (*models.Institution, error) {
	n := trim(name)
	if n == "" {
		return nil, fmt.Errorf("institution name required")
	}
	s := slug.FromName(n)
	var inst models.Institution
	err := r.db.WithContext(ctx).Where("slug = ?", s).First(&inst).Error
	if err == nil {
		return &inst, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	inst = models.Institution{ID: uuid.New(), Name: n, Slug: s}
	if err := r.db.WithContext(ctx).Create(&inst).Error; err != nil {
		// race: try fetch again
		if e2 := r.db.WithContext(ctx).Where("slug = ?", s).First(&inst).Error; e2 == nil {
			return &inst, nil
		}
		return nil, err
	}
	return &inst, nil
}

func (r *Repository) FindOrCreateCompany(ctx context.Context, name string) (*models.Company, error) {
	n := trim(name)
	if n == "" {
		return nil, fmt.Errorf("company name required")
	}
	s := slug.FromName(n)
	var co models.Company
	err := r.db.WithContext(ctx).Where("slug = ?", s).First(&co).Error
	if err == nil {
		return &co, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	co = models.Company{ID: uuid.New(), Name: n, Slug: s}
	if err := r.db.WithContext(ctx).Create(&co).Error; err != nil {
		if e2 := r.db.WithContext(ctx).Where("slug = ?", s).First(&co).Error; e2 == nil {
			return &co, nil
		}
		return nil, err
	}
	return &co, nil
}

func (r *Repository) FindOrCreateSkill(ctx context.Context, name string) (*models.Skill, error) {
	n := trim(name)
	if n == "" {
		return nil, fmt.Errorf("skill name required")
	}
	s := slug.FromName(n)
	var sk models.Skill
	err := r.db.WithContext(ctx).Where("slug = ?", s).First(&sk).Error
	if err == nil {
		return &sk, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	sk = models.Skill{ID: uuid.New(), Name: n, Slug: s}
	if err := r.db.WithContext(ctx).Create(&sk).Error; err != nil {
		if e2 := r.db.WithContext(ctx).Where("slug = ?", s).First(&sk).Error; e2 == nil {
			return &sk, nil
		}
		return nil, err
	}
	return &sk, nil
}

func trim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}
