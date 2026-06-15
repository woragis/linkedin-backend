package service

import (
	"context"
	"fmt"

	authsvc "github.com/unipe/linkedin/backend/server/internal/auth/service"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	catalogrepo "github.com/unipe/linkedin/backend/server/internal/catalog/repository"
	profilerepo "github.com/unipe/linkedin/backend/server/internal/profile/repository"
	profilesvc "github.com/unipe/linkedin/backend/server/internal/profile/service"
)

type Service struct {
	auth     *authsvc.Service
	profiles *profilesvc.Service
	catalog  *catalogrepo.Repository
	repo     *profilerepo.Repository
}

func New(auth *authsvc.Service, profiles *profilesvc.Service, catalog *catalogrepo.Repository, repo *profilerepo.Repository) *Service {
	return &Service{auth: auth, profiles: profiles, catalog: catalog, repo: repo}
}

type SeedResult struct {
	UsersCreated int      `json:"users_created"`
	Slugs        []string `json:"slugs"`
}

type demoUser struct {
	Email    string
	Password string
	FullName string
	Headline string
	Location string
	Skills   []string
	School   string
	Field    string
	Company  string
	Title    string
}

var demoUsers = []demoUser{
	{Email: "ana@demo.linkedin", Password: "demo12345", FullName: "Ana Silva", Headline: "Backend Engineer · Go", Location: "Recife", Skills: []string{"Go", "PostgreSQL", "Redis"}, School: "UNIPe", Field: "Ciência da Computação", Company: "Lokra", Title: "Backend Developer"},
	{Email: "bruno@demo.linkedin", Password: "demo12345", FullName: "Bruno Costa", Headline: "Data Scientist", Location: "Recife", Skills: []string{"Python", "Statistics", "Machine Learning"}, School: "UNIPe", Field: "Estatística", Company: "Lokra", Title: "Data Analyst"},
	{Email: "carla@demo.linkedin", Password: "demo12345", FullName: "Carla Mendes", Headline: "Frontend Developer", Location: "Olinda", Skills: []string{"TypeScript", "React", "Next.js"}, School: "UFPE", Field: "Ciência da Computação", Company: "Freelance", Title: "Frontend Engineer"},
	{Email: "diego@demo.linkedin", Password: "demo12345", FullName: "Diego Alves", Headline: "DevOps Engineer", Location: "Recife", Skills: []string{"Docker", "Kubernetes", "Go"}, School: "UNIPe", Field: "Ciência da Computação", Company: "Lokra", Title: "DevOps"},
	{Email: "elisa@demo.linkedin", Password: "demo12345", FullName: "Elisa Rocha", Headline: "Product Manager", Location: "São Paulo", Skills: []string{"Product", "Analytics"}, School: "USP", Field: "Administração", Company: "Startup XYZ", Title: "PM"},
	{Email: "felipe@demo.linkedin", Password: "demo12345", FullName: "Felipe Nunes", Headline: "ML Engineer", Location: "Recife", Skills: []string{"Python", "PyTorch", "Go"}, School: "UNIPe", Field: "Ciência da Computação", Company: "Lokra", Title: "ML Engineer"},
	{Email: "gabi@demo.linkedin", Password: "demo12345", FullName: "Gabriela Lima", Headline: "UX Designer", Location: "Recife", Skills: []string{"Figma", "UX Research"}, School: "CESAR School", Field: "Design", Company: "Agência Digital", Title: "UX Designer"},
	{Email: "henrique@demo.linkedin", Password: "demo12345", FullName: "Henrique Dias", Headline: "Full Stack Developer", Location: "João Pessoa", Skills: []string{"Go", "React", "PostgreSQL"}, School: "UFPB", Field: "Ciência da Computação", Company: "Consultoria", Title: "Full Stack"},
}

func (s *Service) SeedDemo(ctx context.Context) (*SeedResult, error) {
	result := &SeedResult{Slugs: []string{}}
	for _, du := range demoUsers {
		authOut, err := s.auth.Register(ctx, authsvc.RegisterRequest{
			Email:    du.Email,
			Password: du.Password,
			FullName: du.FullName,
		})
		if err != nil {
			if ae, ok := apperrors.As(err); ok && ae.Code == apperrors.CodeAuthEmailTaken {
				authOut, err = s.auth.Login(ctx, authsvc.LoginRequest{Email: du.Email, Password: du.Password})
				if err != nil {
					return nil, fmt.Errorf("seed login %s: %w", du.Email, err)
				}
			} else {
				return nil, fmt.Errorf("seed user %s: %w", du.Email, err)
			}
		} else {
			result.UsersCreated++
		}
		result.Slugs = append(result.Slugs, authOut.Slug)

		_, _ = s.profiles.PatchProfile(ctx, authOut.UserID, profilesvc.PatchProfileRequest{
			Headline: &du.Headline,
			Location: &du.Location,
		})
		_, _ = s.profiles.ReplaceSkills(ctx, authOut.UserID, profilesvc.ReplaceSkillsRequest{Skills: du.Skills})

		endYear := 2026
		startYear := 2022
		_, _ = s.profiles.CreateEducation(ctx, authOut.UserID, profilesvc.CreateEducationRequest{
			InstitutionName: du.School,
			FieldOfStudy:    du.Field,
			Degree:          "Bacharelado",
			StartYear:       &startYear,
			EndYear:         &endYear,
		})

		expStart := 2024
		_, _ = s.profiles.CreateExperience(ctx, authOut.UserID, profilesvc.CreateExperienceRequest{
			CompanyName: du.Company,
			Title:       du.Title,
			StartYear:   &expStart,
			IsCurrent:   true,
		})
	}
	return result, nil
}
