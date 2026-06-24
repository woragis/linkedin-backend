package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/unipe/linkedin/backend/server/internal/models"
)

func TestNestCommentViews_OneRootOneReply(t *testing.T) {
	rootID := uuid.New()
	replyID := uuid.New()
	postID := uuid.New()
	authorID := uuid.New()

	rows := []models.Comment{
		{ID: rootID, PostID: postID, AuthorID: authorID, Body: "root"},
		{ID: replyID, PostID: postID, AuthorID: authorID, ParentCommentID: &rootID, Body: "reply"},
	}

	views := map[uuid.UUID]CommentView{
		rootID: {
			Comment:   rows[0],
			Replies:   []CommentView{{Comment: rows[1]}},
			ReactionSummary: ReactionSummary{},
		},
		replyID: {
			Comment:         rows[1],
			Replies:         []CommentView{},
			ReactionSummary: ReactionSummary{},
		},
	}

	roots := nestCommentViews(rows, views)
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	if roots[0].ID != rootID {
		t.Fatalf("expected root id %s, got %s", rootID, roots[0].ID)
	}
	if len(roots[0].Replies) != 1 {
		t.Fatalf("expected 1 reply, got %d", len(roots[0].Replies))
	}
	if roots[0].Replies[0].ID != replyID {
		t.Fatalf("expected reply id %s, got %s", replyID, roots[0].Replies[0].ID)
	}
}

func TestNestCommentViews_OrphanReplyIgnored(t *testing.T) {
	rootID := uuid.New()
	orphanID := uuid.New()
	missingParent := uuid.New()
	postID := uuid.New()
	authorID := uuid.New()

	rows := []models.Comment{
		{ID: rootID, PostID: postID, AuthorID: authorID, Body: "root"},
		{ID: orphanID, PostID: postID, AuthorID: authorID, ParentCommentID: &missingParent, Body: "orphan"},
	}

	views := map[uuid.UUID]CommentView{
		rootID: {
			Comment:         rows[0],
			Replies:         []CommentView{},
			ReactionSummary: ReactionSummary{},
		},
		orphanID: {
			Comment:         rows[1],
			Replies:         []CommentView{},
			ReactionSummary: ReactionSummary{},
		},
	}

	roots := nestCommentViews(rows, views)
	if len(roots) != 1 {
		t.Fatalf("expected 1 root (orphan reply excluded), got %d", len(roots))
	}
	if roots[0].ID != rootID {
		t.Fatalf("expected root id %s, got %s", rootID, roots[0].ID)
	}
}
