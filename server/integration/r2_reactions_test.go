//go:build integration

package integration_test

import (
	"net/http"
	"testing"
)

func TestR2_CommentThreadsAndReactions(t *testing.T) {
	srv, cleanup := setupIntegrationServer(t)
	defer cleanup()

	_, _, u1 := registerUser(t, srv, "u1")
	_, _, u2 := registerUser(t, srv, "u2")
	_, _, u3 := registerUser(t, srv, "u3")
	_, _, u4 := registerUser(t, srv, "u4")
	_, _, u5 := registerUser(t, srv, "u5")

	code, post1 := httpJSON(t, srv.URL, http.MethodPost, "/v1/posts", map[string]any{
		"body": "R2 post 1",
	}, authHeader(u1))
	if code != http.StatusCreated {
		t.Fatalf("create post1 status=%d body=%v", code, post1)
	}
	post1ID, _ := post1["id"].(string)

	code, post2 := httpJSON(t, srv.URL, http.MethodPost, "/v1/posts", map[string]any{
		"body": "R2 post 2",
	}, authHeader(u2))
	if code != http.StatusCreated {
		t.Fatalf("create post2 status=%d body=%v", code, post2)
	}
	post2ID, _ := post2["id"].(string)

	// 1. u1 top comment on post1
	code, c1 := httpJSON(t, srv.URL, http.MethodPost, "/v1/posts/"+post1ID+"/comments", map[string]any{
		"body": "u1 root comment",
	}, authHeader(u1))
	if code != http.StatusCreated {
		t.Fatalf("u1 comment status=%d body=%v", code, c1)
	}
	c1ID, _ := c1["id"].(string)

	// 2. u2 reply to u1 comment
	code, c1Reply := httpJSON(t, srv.URL, http.MethodPost, "/v1/posts/"+post1ID+"/comments", map[string]any{
		"body":              "u2 reply",
		"parent_comment_id": c1ID,
	}, authHeader(u2))
	if code != http.StatusCreated {
		t.Fatalf("u2 reply status=%d body=%v", code, c1Reply)
	}
	c1ReplyID, _ := c1Reply["id"].(string)

	// 3. u3 like on root
	if code := httpNoContent(t, srv.URL, http.MethodPost, "/v1/comments/"+c1ID+"/reactions", map[string]any{
		"kind": "like",
	}, authHeader(u3)); code != http.StatusNoContent {
		t.Fatalf("u3 like on root status=%d", code)
	}

	// 4. u4 celebrate on reply
	if code := httpNoContent(t, srv.URL, http.MethodPost, "/v1/comments/"+c1ReplyID+"/reactions", map[string]any{
		"kind": "celebrate",
	}, authHeader(u4)); code != http.StatusNoContent {
		t.Fatalf("u4 celebrate on reply status=%d", code)
	}

	// 5. u5 top comment on post1
	code, c2 := httpJSON(t, srv.URL, http.MethodPost, "/v1/posts/"+post1ID+"/comments", map[string]any{
		"body": "u5 root comment",
	}, authHeader(u5))
	if code != http.StatusCreated {
		t.Fatalf("u5 comment status=%d body=%v", code, c2)
	}
	c2ID, _ := c2["id"].(string)

	// 6. u1 insightful on u5 comment
	if code := httpNoContent(t, srv.URL, http.MethodPost, "/v1/comments/"+c2ID+"/reactions", map[string]any{
		"kind": "insightful",
	}, authHeader(u1)); code != http.StatusNoContent {
		t.Fatalf("u1 insightful status=%d", code)
	}

	// 7. u2 top comment on post2
	code, c3 := httpJSON(t, srv.URL, http.MethodPost, "/v1/posts/"+post2ID+"/comments", map[string]any{
		"body": "u2 root on post2",
	}, authHeader(u2))
	if code != http.StatusCreated {
		t.Fatalf("u2 post2 comment status=%d body=%v", code, c3)
	}
	c3ID, _ := c3["id"].(string)

	// 8. u3 reply to u2 comment
	code, c3Reply := httpJSON(t, srv.URL, http.MethodPost, "/v1/posts/"+post2ID+"/comments", map[string]any{
		"body":              "u3 reply on post2",
		"parent_comment_id": c3ID,
	}, authHeader(u3))
	if code != http.StatusCreated {
		t.Fatalf("u3 post2 reply status=%d body=%v", code, c3Reply)
	}
	c3ReplyID, _ := c3Reply["id"].(string)

	// 9. u4 love on u2 comment
	if code := httpNoContent(t, srv.URL, http.MethodPost, "/v1/comments/"+c3ID+"/reactions", map[string]any{
		"kind": "love",
	}, authHeader(u4)); code != http.StatusNoContent {
		t.Fatalf("u4 love status=%d", code)
	}

	// 10. u5 funny on reply
	if code := httpNoContent(t, srv.URL, http.MethodPost, "/v1/comments/"+c3ReplyID+"/reactions", map[string]any{
		"kind": "funny",
	}, authHeader(u5)); code != http.StatusNoContent {
		t.Fatalf("u5 funny status=%d", code)
	}

	// GET comments on post1 — verify tree and reaction summaries
	code, comments := httpJSONArray(t, srv.URL, http.MethodGet, "/v1/posts/"+post1ID+"/comments", nil, authHeader(u1))
	if code != http.StatusOK {
		t.Fatalf("list comments status=%d", code)
	}
	if len(comments) != 2 {
		t.Fatalf("expected 2 root comments on post1, got %d", len(comments))
	}

	var rootU1, rootU5 map[string]any
	for _, c := range comments {
		m, ok := c.(map[string]any)
		if !ok {
			t.Fatal("comment is not an object")
		}
		switch m["id"] {
		case c1ID:
			rootU1 = m
		case c2ID:
			rootU5 = m
		}
	}
	if rootU1 == nil || rootU5 == nil {
		t.Fatalf("missing expected root comments: u1=%v u5=%v", rootU1 != nil, rootU5 != nil)
	}

	replies, ok := rootU1["replies"].([]any)
	if !ok || len(replies) != 1 {
		t.Fatalf("expected 1 reply on u1 root, got %v", rootU1["replies"])
	}
	reply, ok := replies[0].(map[string]any)
	if !ok || reply["id"] != c1ReplyID {
		t.Fatalf("unexpected reply: %v", replies[0])
	}

	assertReactionCount(t, rootU1, "like", 1)
	assertReactionCount(t, reply, "celebrate", 1)
	assertReactionCount(t, rootU5, "insightful", 1)

	// reply-to-reply should be rejected
	code, errBody := httpJSON(t, srv.URL, http.MethodPost, "/v1/posts/"+post1ID+"/comments", map[string]any{
		"body":              "nested too deep",
		"parent_comment_id": c1ReplyID,
	}, authHeader(u1))
	if code != http.StatusBadRequest {
		t.Fatalf("reply-to-reply expected 400, got %d body=%v", code, errBody)
	}
}

func assertReactionCount(t *testing.T, comment map[string]any, kind string, want int64) {
	t.Helper()
	summary, ok := comment["reaction_summary"].(map[string]any)
	if !ok {
		t.Fatalf("missing reaction_summary on comment %v", comment["id"])
	}
	var got int64
	switch v := summary[kind].(type) {
	case float64:
		got = int64(v)
	case int64:
		got = v
	case int:
		got = int64(v)
	default:
		t.Fatalf("reaction_summary[%q] unexpected type %T", kind, summary[kind])
	}
	if got != want {
		t.Fatalf("comment %v reaction_summary[%q]=%d, want %d", comment["id"], kind, got, want)
	}
}
