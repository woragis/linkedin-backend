package repository

import (
	"context"
	"fmt"

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

func (r *Repository) GetProfileByUserID(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	var p models.Profile
	if err := r.db.WithContext(ctx).First(&p, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) GetProfileBySlug(ctx context.Context, slug string) (*models.Profile, error) {
	var p models.Profile
	if err := r.db.WithContext(ctx).First(&p, "slug = ?", slug).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) UpdateProfile(ctx context.Context, userID uuid.UUID, updates map[string]any) (*models.Profile, error) {
	if err := r.db.WithContext(ctx).Model(&models.Profile{}).Where("user_id = ?", userID).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return r.GetProfileByUserID(ctx, userID)
}

func (r *Repository) SlugTakenByOther(ctx context.Context, slug string, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Profile{}).
		Where("slug = ? AND user_id <> ?", slug, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) ListExperiences(ctx context.Context, userID uuid.UUID) ([]models.Experience, error) {
	var rows []models.Experience
	err := r.db.WithContext(ctx).Preload("Company").Where("user_id = ?", userID).
		Order("is_current DESC, start_year DESC NULLS LAST").Find(&rows).Error
	return rows, err
}

func (r *Repository) CreateExperience(ctx context.Context, e *models.Experience) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *Repository) GetExperience(ctx context.Context, userID, id uuid.UUID) (*models.Experience, error) {
	var e models.Experience
	err := r.db.WithContext(ctx).Preload("Company").Where("id = ? AND user_id = ?", id, userID).First(&e).Error
	return &e, err
}

func (r *Repository) UpdateExperience(ctx context.Context, userID, id uuid.UUID, updates map[string]any) (*models.Experience, error) {
	res := r.db.WithContext(ctx).Model(&models.Experience{}).Where("id = ? AND user_id = ?", id, userID).Updates(updates)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return r.GetExperience(ctx, userID, id)
}

func (r *Repository) DeleteExperience(ctx context.Context, userID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&models.Experience{}, "id = ? AND user_id = ?", id, userID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) ListEducations(ctx context.Context, userID uuid.UUID) ([]models.Education, error) {
	var rows []models.Education
	err := r.db.WithContext(ctx).Preload("Institution").Where("user_id = ?", userID).
		Order("end_year DESC NULLS LAST, start_year DESC NULLS LAST").Find(&rows).Error
	return rows, err
}

func (r *Repository) CreateEducation(ctx context.Context, e *models.Education) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *Repository) GetEducation(ctx context.Context, userID, id uuid.UUID) (*models.Education, error) {
	var e models.Education
	err := r.db.WithContext(ctx).Preload("Institution").Where("id = ? AND user_id = ?", id, userID).First(&e).Error
	return &e, err
}

func (r *Repository) UpdateEducation(ctx context.Context, userID, id uuid.UUID, updates map[string]any) (*models.Education, error) {
	res := r.db.WithContext(ctx).Model(&models.Education{}).Where("id = ? AND user_id = ?", id, userID).Updates(updates)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return r.GetEducation(ctx, userID, id)
}

func (r *Repository) DeleteEducation(ctx context.Context, userID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&models.Education{}, "id = ? AND user_id = ?", id, userID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) ListSkills(ctx context.Context, userID uuid.UUID) ([]models.Skill, error) {
	var rows []models.UserSkill
	if err := r.db.WithContext(ctx).Preload("Skill").Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	skills := make([]models.Skill, 0, len(rows))
	for _, us := range rows {
		skills = append(skills, us.Skill)
	}
	return skills, nil
}

func (r *Repository) ReplaceSkills(ctx context.Context, userID uuid.UUID, skillIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.UserSkill{}, "user_id = ?", userID).Error; err != nil {
			return err
		}
		for _, sid := range skillIDs {
			us := models.UserSkill{UserID: userID, SkillID: sid}
			if err := tx.Create(&us).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
