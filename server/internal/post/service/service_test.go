package service

import (
	"testing"

	"github.com/unipe/linkedin/backend/server/internal/apperrors"
)

func TestNormalizeReactionKind_EmptyDefaultsToLike(t *testing.T) {
	got, err := normalizeReactionKind("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "like" {
		t.Fatalf("expected like, got %q", got)
	}
}

func TestNormalizeReactionKind_ValidKinds(t *testing.T) {
	kinds := []string{"like", "celebrate", "support", "insightful", "love", "funny"}
	for _, kind := range kinds {
		got, err := normalizeReactionKind(kind)
		if err != nil {
			t.Fatalf("kind %q: unexpected error: %v", kind, err)
		}
		if got != kind {
			t.Fatalf("kind %q: expected %q, got %q", kind, kind, got)
		}
	}
}

func TestNormalizeReactionKind_InvalidReturnsError(t *testing.T) {
	_, err := normalizeReactionKind("dislike")
	if err == nil {
		t.Fatal("expected error for invalid kind")
	}
	ae, ok := apperrors.As(err)
	if !ok {
		t.Fatalf("expected apperrors.Error, got %T", err)
	}
	if ae.Code != apperrors.CodePostInvalidBody {
		t.Fatalf("expected code %q, got %q", apperrors.CodePostInvalidBody, ae.Code)
	}
}
