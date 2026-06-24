package llm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	authrepo "github.com/unipe/linkedin/backend/server/internal/auth/repository"
	profilerepo "github.com/unipe/linkedin/backend/server/internal/profile/repository"
	postsvc "github.com/unipe/linkedin/backend/server/internal/post/service"
)

type Persona struct {
	UserID   uuid.UUID
	Email    string
	FullName string
	Headline string
}

type RunResult struct {
	PostsCreated    int      `json:"posts_created"`
	CommentsCreated int      `json:"comments_created"`
	PostIDs         []string `json:"post_ids"`
	LLMModel        string   `json:"llm_model"`
}

type Runner struct {
	llm      *Client
	auth     *authrepo.Repository
	profiles *profilerepo.Repository
	posts    *postsvc.Service
}

func NewRunner(llm *Client, auth *authrepo.Repository, profiles *profilerepo.Repository, posts *postsvc.Service) *Runner {
	return &Runner{llm: llm, auth: auth, profiles: profiles, posts: posts}
}

func (r *Runner) RunE2EScenario(ctx context.Context) (*RunResult, error) {
	if r.llm == nil || !r.llm.Enabled() {
		return nil, fmt.Errorf("OPENAI_API_KEY not configured")
	}

	emails := []string{
		"ana@demo.linkedin",
		"bruno@demo.linkedin",
		"carla@demo.linkedin",
		"diego@demo.linkedin",
		"elisa@demo.linkedin",
	}
	personas := make([]Persona, 0, len(emails))
	for _, email := range emails {
		p, err := r.loadPersona(ctx, email)
		if err != nil {
			return nil, err
		}
		personas = append(personas, p)
	}

	result := &RunResult{LLMModel: r.llm.Model()}

	// 2 posts (ana, bruno)
	postBodies := make([]string, 0, 2)
	postIDs := make([]uuid.UUID, 0, 2)
	for i := 0; i < 2; i++ {
		p := personas[i]
		body, err := r.llm.GeneratePost(ctx, p.FullName, p.Headline, "tecnologia e carreira")
		if err != nil {
			return nil, err
		}
		if IsTemplateText(body) || !LooksPortuguese(body) {
			return nil, fmt.Errorf("llm post looks invalid: %q", body)
		}
		created, err := r.posts.Create(ctx, p.UserID, postsvc.CreatePostRequest{Body: body})
		if err != nil {
			return nil, err
		}
		postBodies = append(postBodies, body)
		postIDs = append(postIDs, created.ID)
		result.PostIDs = append(result.PostIDs, created.ID.String())
		result.PostsCreated++
	}

	post1ID := postIDs[0]
	post1Body := postBodies[0]

	// 10 comment interactions on post1 (mix comments, replies, reactions via comments only)
	type step struct {
		personaIdx int
		replyTo    *uuid.UUID
		parentBody string
	}
	steps := []step{
		{1, nil, ""},          // bruno top comment
		{2, nil, ""},          // carla top comment
		{3, nil, ""},          // diego top comment
		{4, nil, ""},          // elisa top comment
		{0, nil, ""},          // ana top comment
		{2, nil, ""},          // carla another
		{3, nil, ""},          // diego another
		{1, nil, ""},          // bruno another
		{4, nil, ""},          // elisa another
		{0, nil, ""},          // ana another — will attach reply to first comment below
	}

	var firstCommentID uuid.UUID
	var firstCommentBody string
	for i, st := range steps {
		p := personas[st.personaIdx]
		var parentID *uuid.UUID
		parentText := ""
		if i == 9 && firstCommentID != uuid.Nil {
			parentID = &firstCommentID
			parentText = firstCommentBody
		}
		body, err := r.llm.GenerateComment(ctx, p.FullName, p.Headline, post1Body, parentText)
		if err != nil {
			return nil, err
		}
		if IsTemplateText(body) || !LooksPortuguese(body) {
			return nil, fmt.Errorf("llm comment looks invalid: %q", body)
		}
		req := postsvc.CreateCommentRequest{Body: body, ParentCommentID: parentID}
		created, err := r.posts.Comment(ctx, p.UserID, post1ID, req)
		if err != nil {
			return nil, err
		}
		if i == 0 {
			firstCommentID = created.ID
			firstCommentBody = body
		}
		result.CommentsCreated++
	}

	// 2nd post gets 0 comments in this scenario — total 10 comments on post1
	_ = postIDs[1]

	return result, nil
}

func (r *Runner) loadPersona(ctx context.Context, email string) (Persona, error) {
	user, err := r.auth.GetByEmail(ctx, email)
	if err != nil {
		return Persona{}, fmt.Errorf("user %s: %w", email, err)
	}
	profile, err := r.profiles.GetProfileByUserID(ctx, user.ID)
	if err != nil {
		return Persona{}, fmt.Errorf("profile %s: %w", email, err)
	}
	return Persona{
		UserID:   user.ID,
		Email:    email,
		FullName: profile.FullName,
		Headline: profile.Headline,
	}, nil
}
