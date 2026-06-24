package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/models"
	postrepo "github.com/unipe/linkedin/backend/server/internal/post/repository"
)

type ReactionSummary map[string]int64

type PostView struct {
	models.Post
	ReactionCount   int64           `json:"reaction_count"`
	CommentCount    int64           `json:"comment_count"`
	ReactionSummary ReactionSummary `json:"reaction_summary"`
	MyReaction      string          `json:"my_reaction,omitempty"`
}

type CommentView struct {
	models.Comment
	ReactionCount   int64           `json:"reaction_count"`
	ReactionSummary ReactionSummary `json:"reaction_summary"`
	MyReaction      string          `json:"my_reaction,omitempty"`
	Replies         []CommentView   `json:"replies"`
}

func (s *Service) enrichPostView(ctx context.Context, p models.Post, viewerID uuid.UUID) PostView {
	rc, _ := s.posts.ReactionCount(ctx, p.ID)
	cc, _ := s.posts.CommentCount(ctx, p.ID)
	summary, _ := s.posts.ReactionSummary(ctx, postrepo.TargetPost, p.ID)
	my, _ := s.posts.MyContentReaction(ctx, postrepo.TargetPost, p.ID, viewerID)
	if summary == nil {
		summary = ReactionSummary{}
	}
	return PostView{
		Post:            p,
		ReactionCount:   rc,
		CommentCount:    cc,
		ReactionSummary: summary,
		MyReaction:      my,
	}
}

func (s *Service) enrichPostViews(ctx context.Context, rows []models.Post, viewerID uuid.UUID) []PostView {
	if len(rows) == 0 {
		return []PostView{}
	}
	ids := make([]uuid.UUID, len(rows))
	for i, p := range rows {
		ids[i] = p.ID
	}
	summaries, _ := s.posts.BatchReactionSummaries(ctx, postrepo.TargetPost, ids)
	myReactions, _ := s.posts.BatchMyReactions(ctx, postrepo.TargetPost, ids, viewerID)

	out := make([]PostView, 0, len(rows))
	for _, p := range rows {
		rc := int64(0)
		summary := ReactionSummary{}
		if summaries != nil {
			if m, ok := summaries[p.ID]; ok {
				summary = m
				for _, c := range m {
					rc += c
				}
			}
		}
		cc, _ := s.posts.CommentCount(ctx, p.ID)
		my := ""
		if myReactions != nil {
			my = myReactions[p.ID]
		}
		out = append(out, PostView{
			Post:            p,
			ReactionCount:   rc,
			CommentCount:    cc,
			ReactionSummary: summary,
			MyReaction:      my,
		})
	}
	return out
}

func (s *Service) buildCommentTree(
	ctx context.Context,
	rows []models.Comment,
	viewerID uuid.UUID,
) ([]CommentView, error) {
	if len(rows) == 0 {
		return []CommentView{}, nil
	}
	ids := make([]uuid.UUID, len(rows))
	for i, c := range rows {
		ids[i] = c.ID
	}
	summaries, err := s.posts.BatchReactionSummaries(ctx, postrepo.TargetComment, ids)
	if err != nil {
		return nil, err
	}
	myReactions, err := s.posts.BatchMyReactions(ctx, postrepo.TargetComment, ids, viewerID)
	if err != nil {
		return nil, err
	}

	views := make(map[uuid.UUID]CommentView, len(rows))
	for _, c := range rows {
		summary := ReactionSummary{}
		var rc int64
		if summaries != nil {
			if m, ok := summaries[c.ID]; ok {
				summary = m
				for _, n := range m {
					rc += n
				}
			}
		}
		my := ""
		if myReactions != nil {
			my = myReactions[c.ID]
		}
		views[c.ID] = CommentView{
			Comment:         c,
			ReactionCount:   rc,
			ReactionSummary: summary,
			MyReaction:      my,
			Replies:         []CommentView{},
		}
	}

	for _, c := range rows {
		if c.ParentCommentID == nil {
			continue
		}
		child := views[c.ID]
		parent, ok := views[*c.ParentCommentID]
		if !ok {
			continue
		}
		parent.Replies = append(parent.Replies, child)
		views[*c.ParentCommentID] = parent
	}

	roots := make([]CommentView, 0)
	for _, c := range rows {
		if c.ParentCommentID != nil {
			continue
		}
		roots = append(roots, views[c.ID])
	}
	return roots, nil
}

func (s *Service) enrichSingleComment(ctx context.Context, c models.Comment, viewerID uuid.UUID) (*CommentView, error) {
	rows, err := s.buildCommentTree(ctx, []models.Comment{c}, viewerID)
	if err != nil || len(rows) == 0 {
		view := CommentView{Comment: c, ReactionSummary: ReactionSummary{}, Replies: []CommentView{}}
		return &view, err
	}
	return &rows[0], nil
}
