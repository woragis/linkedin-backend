package service

import (
	"context"
	"fmt"

	authsvc "github.com/unipe/linkedin/backend/server/internal/auth/service"
	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	catalogrepo "github.com/unipe/linkedin/backend/server/internal/catalog/repository"
	connsvc "github.com/unipe/linkedin/backend/server/internal/connection/service"
	postsvc "github.com/unipe/linkedin/backend/server/internal/post/service"
	profilerepo "github.com/unipe/linkedin/backend/server/internal/profile/repository"
	profilesvc "github.com/unipe/linkedin/backend/server/internal/profile/service"
)

type Service struct {
	auth        *authsvc.Service
	profiles    *profilesvc.Service
	catalog     *catalogrepo.Repository
	repo        *profilerepo.Repository
	connections *connsvc.Service
	posts       *postsvc.Service
}

func New(
	auth *authsvc.Service,
	profiles *profilesvc.Service,
	catalog *catalogrepo.Repository,
	repo *profilerepo.Repository,
	connections *connsvc.Service,
	posts *postsvc.Service,
) *Service {
	return &Service{
		auth:        auth,
		profiles:    profiles,
		catalog:     catalog,
		repo:        repo,
		connections: connections,
		posts:       posts,
	}
}

type SeedResult struct {
	UsersCreated int      `json:"users_created"`
	Slugs        []string `json:"slugs"`
	PostsCreated int      `json:"posts_created"`
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
	Post     string
}

var demoUsers = []demoUser{
	{Email: "ana@demo.linkedin", Password: "demo12345", FullName: "Ana Silva", Headline: "Backend Engineer · Go", Location: "Recife", Skills: []string{"Go", "PostgreSQL", "Redis"}, School: "UNIPe", Field: "Ciência da Computação", Company: "Lokra", Title: "Backend Developer", Post: "Migrando nosso pipeline para Go + Postgres. Performance subiu bastante!"},
	{Email: "bruno@demo.linkedin", Password: "demo12345", FullName: "Bruno Costa", Headline: "Data Scientist", Location: "Recife", Skills: []string{"Python", "Statistics", "Machine Learning"}, School: "UNIPe", Field: "Estatística", Company: "Lokra", Title: "Data Analyst", Post: "Rodando experimento A/B no feed — estatística aplicada em produção."},
	{Email: "carla@demo.linkedin", Password: "demo12345", FullName: "Carla Mendes", Headline: "Frontend Developer", Location: "Olinda", Skills: []string{"TypeScript", "React", "Next.js"}, School: "UFPE", Field: "Ciência da Computação", Company: "Freelance", Title: "Frontend Engineer", Post: "Next.js 16 no projeto novo. DX excelente."},
	{Email: "diego@demo.linkedin", Password: "demo12345", FullName: "Diego Alves", Headline: "DevOps Engineer", Location: "Recife", Skills: []string{"Docker", "Kubernetes", "Go"}, School: "UNIPe", Field: "Ciência da Computação", Company: "Lokra", Title: "DevOps", Post: "Compose com worker-realtime + worker-batch separados. Ficou limpo."},
	{Email: "elisa@demo.linkedin", Password: "demo12345", FullName: "Elisa Rocha", Headline: "Product Manager", Location: "São Paulo", Skills: []string{"Product", "Analytics"}, School: "USP", Field: "Administração", Company: "Startup XYZ", Title: "PM", Post: "DAU e retenção D7 são nossas métricas norte este sprint."},
	{Email: "felipe@demo.linkedin", Password: "demo12345", FullName: "Felipe Nunes", Headline: "ML Engineer", Location: "Recife", Skills: []string{"Python", "PyTorch", "Go"}, School: "UNIPe", Field: "Ciência da Computação", Company: "Lokra", Title: "ML Engineer", Post: "Treinando modelo de link prediction nas conexões da plataforma."},
	{Email: "gabi@demo.linkedin", Password: "demo12345", FullName: "Gabriela Lima", Headline: "UX Designer", Location: "Recife", Skills: []string{"Figma", "UX Research"}, School: "CESAR School", Field: "Design", Company: "Agência Digital", Title: "UX Designer", Post: "Prototipando a visualização do grafo social na página /network."},
	{Email: "henrique@demo.linkedin", Password: "demo12345", FullName: "Henrique Dias", Headline: "Full Stack Developer", Location: "João Pessoa", Skills: []string{"Go", "React", "PostgreSQL"}, School: "UFPB", Field: "Ciência da Computação", Company: "Consultoria", Title: "Full Stack", Post: "Graph analytics + rede social = projeto perfeito para o 5-grafo."},
}

func (s *Service) SeedDemo(ctx context.Context) (*SeedResult, error) {
	result := &SeedResult{Slugs: []string{}}
	userIDs := make([]authsvc.AuthResponse, 0, len(demoUsers))

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
		userIDs = append(userIDs, *authOut)

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

		if du.Post != "" {
			if _, err := s.posts.Create(ctx, authOut.UserID, postsvc.CreatePostRequest{Body: du.Post}); err == nil {
				result.PostsCreated++
			}
		}
	}

	// mesh connections between demo users
	for i, a := range userIDs {
		for j, b := range userIDs {
			if i >= j {
				continue
			}
			_, _ = s.connections.Request(ctx, a.UserID, connsvc.RequestInput{TargetUserID: b.UserID})
			// auto-accept for demo
			pending, _ := s.connections.ListPending(ctx, b.UserID)
			for _, c := range pending {
				if c.RequesterID == a.UserID {
					_, _ = s.connections.Accept(ctx, b.UserID, c.ID)
				}
			}
		}
	}

	return result, nil
}
