package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	catalogrepo "github.com/unipe/linkedin/backend/server/internal/catalog/repository"
	"github.com/unipe/linkedin/backend/server/internal/models"
	"github.com/unipe/linkedin/backend/server/internal/platform/outbox"
	profilerepo "github.com/unipe/linkedin/backend/server/internal/profile/repository"
	"gorm.io/gorm"
)

var slugFormat = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Service struct {
	profiles *profilerepo.Repository
	catalog  *catalogrepo.Repository
	db       *gorm.DB
}

func New(profiles *profilerepo.Repository, catalog *catalogrepo.Repository, db *gorm.DB) *Service {
	return &Service{profiles: profiles, catalog: catalog, db: db}
}

type ProfileView struct {
	UserID    uuid.UUID        `json:"user_id"`
	Slug      string           `json:"slug"`
	FullName  string           `json:"full_name"`
	Headline  string           `json:"headline"`
	Bio       string           `json:"bio"`
	Location  string           `json:"location"`
	BirthYear *int             `json:"birth_year,omitempty"`
	AvatarURL *string          `json:"avatar_url,omitempty"`
	Skills    []models.Skill   `json:"skills,omitempty"`
	Experiences []models.Experience `json:"experiences,omitempty"`
	Educations []models.Education   `json:"educations,omitempty"`
}

func (s *Service) GetMe(ctx context.Context, userID uuid.UUID) (*ProfileView, error) {
	return s.buildView(ctx, userID, true)
}

func (s *Service) GetBySlug(ctx context.Context, slugStr string) (*ProfileView, error) {
	p, err := s.profiles.GetProfileBySlug(ctx, slugStr)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound(apperrors.CodeProfileNotFound, apperrors.MsgProfileNotFound)
		}
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return s.buildView(ctx, p.UserID, true)
}

func (s *Service) buildView(ctx context.Context, userID uuid.UUID, withDetails bool) (*ProfileView, error) {
	p, err := s.profiles.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound(apperrors.CodeProfileNotFound, apperrors.MsgProfileNotFound)
		}
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	v := &ProfileView{
		UserID:    p.UserID,
		Slug:      p.Slug,
		FullName:  p.FullName,
		Headline:  p.Headline,
		Bio:       p.Bio,
		Location:  p.Location,
		BirthYear: p.BirthYear,
		AvatarURL: p.AvatarURL,
	}
	if !withDetails {
		return v, nil
	}
	skills, err := s.profiles.ListSkills(ctx, userID)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	v.Skills = skills
	exps, err := s.profiles.ListExperiences(ctx, userID)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	v.Experiences = exps
	edus, err := s.profiles.ListEducations(ctx, userID)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	v.Educations = edus
	return v, nil
}

type PatchProfileRequest struct {
	FullName  *string `json:"full_name"`
	Slug      *string `json:"slug"`
	Headline  *string `json:"headline"`
	Bio       *string `json:"bio"`
	Location  *string `json:"location"`
	BirthYear *int    `json:"birth_year"`
	AvatarURL *string `json:"avatar_url"`
}

func (s *Service) PatchProfile(ctx context.Context, userID uuid.UUID, req PatchProfileRequest) (*ProfileView, error) {
	updates := map[string]any{}
	if req.FullName != nil {
		name := strings.TrimSpace(*req.FullName)
		if name == "" {
			return nil, apperrors.Invalid(apperrors.CodeProfileInvalidBody, "full_name cannot be empty")
		}
		updates["full_name"] = name
	}
	if req.Slug != nil {
		slugStr := strings.ToLower(strings.TrimSpace(*req.Slug))
		if !slugFormat.MatchString(slugStr) {
			return nil, apperrors.Invalid(apperrors.CodeProfileSlugInvalid, "invalid slug format")
		}
		taken, err := s.profiles.SlugTakenByOther(ctx, slugStr, userID)
		if err != nil {
			return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
		}
		if taken {
			return nil, apperrors.Conflict(apperrors.CodeProfileSlugTaken, apperrors.MsgProfileSlugTaken)
		}
		updates["slug"] = slugStr
	}
	if req.Headline != nil {
		updates["headline"] = strings.TrimSpace(*req.Headline)
	}
	if req.Bio != nil {
		updates["bio"] = strings.TrimSpace(*req.Bio)
	}
	if req.Location != nil {
		updates["location"] = strings.TrimSpace(*req.Location)
	}
	if req.BirthYear != nil {
		updates["birth_year"] = req.BirthYear
	}
	if req.AvatarURL != nil {
		updates["avatar_url"] = req.AvatarURL
	}
	if len(updates) == 0 {
		return s.GetMe(ctx, userID)
	}
	if _, err := s.profiles.UpdateProfile(ctx, userID, updates); err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	_ = outbox.Enqueue(ctx, s.db, outbox.Job{
		JobType: "search.index_profile",
		Payload: map[string]any{"user_id": userID.String()},
	})
	return s.GetMe(ctx, userID)
}

type CreateExperienceRequest struct {
	CompanyName string `json:"company_name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	StartYear   *int   `json:"start_year"`
	EndYear     *int   `json:"end_year"`
	IsCurrent   bool   `json:"is_current"`
}

func (s *Service) CreateExperience(ctx context.Context, userID uuid.UUID, req CreateExperienceRequest) (*models.Experience, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" || strings.TrimSpace(req.CompanyName) == "" {
		return nil, apperrors.Invalid(apperrors.CodeProfileInvalidBody, "title and company_name are required")
	}
	co, err := s.catalog.FindOrCreateCompany(ctx, req.CompanyName)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	e := &models.Experience{
		ID:          uuid.New(),
		UserID:      userID,
		CompanyID:   co.ID,
		Title:       title,
		Description: strings.TrimSpace(req.Description),
		StartYear:   req.StartYear,
		EndYear:     req.EndYear,
		IsCurrent:   req.IsCurrent,
	}
	if err := s.profiles.CreateExperience(ctx, e); err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return s.profiles.GetExperience(ctx, userID, e.ID)
}

type PatchExperienceRequest struct {
	CompanyName *string `json:"company_name"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	StartYear   *int    `json:"start_year"`
	EndYear     *int    `json:"end_year"`
	IsCurrent   *bool   `json:"is_current"`
}

func (s *Service) PatchExperience(ctx context.Context, userID, id uuid.UUID, req PatchExperienceRequest) (*models.Experience, error) {
	updates := map[string]any{}
	if req.CompanyName != nil {
		co, err := s.catalog.FindOrCreateCompany(ctx, *req.CompanyName)
		if err != nil {
			return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
		}
		updates["company_id"] = co.ID
	}
	if req.Title != nil {
		updates["title"] = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		updates["description"] = strings.TrimSpace(*req.Description)
	}
	if req.StartYear != nil {
		updates["start_year"] = req.StartYear
	}
	if req.EndYear != nil {
		updates["end_year"] = req.EndYear
	}
	if req.IsCurrent != nil {
		updates["is_current"] = *req.IsCurrent
	}
	exp, err := s.profiles.UpdateExperience(ctx, userID, id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound(apperrors.CodeExperienceNotFound, apperrors.MsgExperienceNotFound)
		}
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return exp, nil
}

func (s *Service) DeleteExperience(ctx context.Context, userID, id uuid.UUID) error {
	if err := s.profiles.DeleteExperience(ctx, userID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NotFound(apperrors.CodeExperienceNotFound, apperrors.MsgExperienceNotFound)
		}
		return apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return nil
}

func (s *Service) ListExperiences(ctx context.Context, userID uuid.UUID) ([]models.Experience, error) {
	rows, err := s.profiles.ListExperiences(ctx, userID)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return rows, nil
}

type CreateEducationRequest struct {
	InstitutionName string `json:"institution_name"`
	FieldOfStudy    string `json:"field_of_study"`
	Degree          string `json:"degree"`
	StartYear       *int   `json:"start_year"`
	EndYear         *int   `json:"end_year"`
}

func (s *Service) CreateEducation(ctx context.Context, userID uuid.UUID, req CreateEducationRequest) (*models.Education, error) {
	if strings.TrimSpace(req.InstitutionName) == "" {
		return nil, apperrors.Invalid(apperrors.CodeProfileInvalidBody, "institution_name is required")
	}
	inst, err := s.catalog.FindOrCreateInstitution(ctx, req.InstitutionName)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	e := &models.Education{
		ID:            uuid.New(),
		UserID:        userID,
		InstitutionID: inst.ID,
		FieldOfStudy:  strings.TrimSpace(req.FieldOfStudy),
		Degree:        strings.TrimSpace(req.Degree),
		StartYear:     req.StartYear,
		EndYear:       req.EndYear,
	}
	if err := s.profiles.CreateEducation(ctx, e); err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return s.profiles.GetEducation(ctx, userID, e.ID)
}

type PatchEducationRequest struct {
	InstitutionName *string `json:"institution_name"`
	FieldOfStudy    *string `json:"field_of_study"`
	Degree          *string `json:"degree"`
	StartYear       *int    `json:"start_year"`
	EndYear         *int    `json:"end_year"`
}

func (s *Service) PatchEducation(ctx context.Context, userID, id uuid.UUID, req PatchEducationRequest) (*models.Education, error) {
	updates := map[string]any{}
	if req.InstitutionName != nil {
		inst, err := s.catalog.FindOrCreateInstitution(ctx, *req.InstitutionName)
		if err != nil {
			return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
		}
		updates["institution_id"] = inst.ID
	}
	if req.FieldOfStudy != nil {
		updates["field_of_study"] = strings.TrimSpace(*req.FieldOfStudy)
	}
	if req.Degree != nil {
		updates["degree"] = strings.TrimSpace(*req.Degree)
	}
	if req.StartYear != nil {
		updates["start_year"] = req.StartYear
	}
	if req.EndYear != nil {
		updates["end_year"] = req.EndYear
	}
	edu, err := s.profiles.UpdateEducation(ctx, userID, id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound(apperrors.CodeEducationNotFound, apperrors.MsgEducationNotFound)
		}
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return edu, nil
}

func (s *Service) DeleteEducation(ctx context.Context, userID, id uuid.UUID) error {
	if err := s.profiles.DeleteEducation(ctx, userID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NotFound(apperrors.CodeEducationNotFound, apperrors.MsgEducationNotFound)
		}
		return apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return nil
}

func (s *Service) ListEducations(ctx context.Context, userID uuid.UUID) ([]models.Education, error) {
	rows, err := s.profiles.ListEducations(ctx, userID)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return rows, nil
}

type ReplaceSkillsRequest struct {
	Skills []string `json:"skills"`
}

func (s *Service) ReplaceSkills(ctx context.Context, userID uuid.UUID, req ReplaceSkillsRequest) ([]models.Skill, error) {
	ids := make([]uuid.UUID, 0, len(req.Skills))
	seen := map[string]struct{}{}
	for _, name := range req.Skills {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		sk, err := s.catalog.FindOrCreateSkill(ctx, name)
		if err != nil {
			return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
		}
		ids = append(ids, sk.ID)
	}
	if err := s.profiles.ReplaceSkills(ctx, userID, ids); err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return s.profiles.ListSkills(ctx, userID)
}

func (s *Service) ListSkills(ctx context.Context, userID uuid.UUID) ([]models.Skill, error) {
	rows, err := s.profiles.ListSkills(ctx, userID)
	if err != nil {
		return nil, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err)
	}
	return rows, nil
}
