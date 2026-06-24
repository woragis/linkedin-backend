//go:build e2e

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
)

func authHeader(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

func login(t *testing.T, email, password string) string {
	t.Helper()
	code, out := httpJSON(t, http.MethodPost, "/v1/auth/login", map[string]any{
		"email":    email,
		"password": password,
	}, nil)
	if code != http.StatusOK {
		t.Fatalf("login %s status=%d body=%v", email, code, out)
	}
	token, _ := out["token"].(string)
	if token == "" {
		t.Fatal("missing token")
	}
	return token
}

func createPost(t *testing.T, token, body string) string {
	t.Helper()
	code, out := httpJSON(t, http.MethodPost, "/v1/posts", map[string]any{
		"body": body,
	}, authHeader(token))
	if code != http.StatusCreated {
		t.Fatalf("create post status=%d body=%v", code, out)
	}
	id, _ := out["id"].(string)
	if id == "" {
		t.Fatal("missing post id")
	}
	return id
}

func createComment(t *testing.T, token, postID, body string) string {
	t.Helper()
	code, out := httpJSON(t, http.MethodPost, "/v1/posts/"+postID+"/comments", map[string]any{
		"body": body,
	}, authHeader(token))
	if code != http.StatusCreated {
		t.Fatalf("create comment status=%d body=%v", code, out)
	}
	id, _ := out["id"].(string)
	if id == "" {
		t.Fatal("missing comment id")
	}
	return id
}

func replyComment(t *testing.T, token, postID, parentID, body string) string {
	t.Helper()
	code, out := httpJSON(t, http.MethodPost, "/v1/posts/"+postID+"/comments", map[string]any{
		"body":              body,
		"parent_comment_id": parentID,
	}, authHeader(token))
	if code != http.StatusCreated {
		t.Fatalf("reply comment status=%d body=%v", code, out)
	}
	id, _ := out["id"].(string)
	if id == "" {
		t.Fatal("missing reply id")
	}
	return id
}

func reactComment(t *testing.T, token, commentID, kind string) {
	t.Helper()
	code, _ := httpJSON(t, http.MethodPost, "/v1/comments/"+commentID+"/reactions", map[string]any{
		"kind": kind,
	}, authHeader(token))
	if code != http.StatusNoContent {
		t.Fatalf("react comment %s kind=%s status=%d", commentID, kind, code)
	}
}

func seedDemo(t *testing.T) {
	t.Helper()
	code, body := httpJSON(t, http.MethodPost, "/v1/internal/seed-demo", nil, map[string]string{
		"X-Internal-Token": internalToken(),
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("POST seed-demo status=%d body=%v", code, body)
	}
}

func listComments(t *testing.T, token, postID string) []any {
	t.Helper()
	code, arr := httpJSONArray(t, http.MethodGet, "/v1/posts/"+postID+"/comments", nil, authHeader(token))
	if code != http.StatusOK {
		t.Fatalf("list comments status=%d", code)
	}
	return arr
}

func assertReactionCountE2E(t *testing.T, comment map[string]any, kind string, want int64) {
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

func demoUser(name string) (email, password string) {
	return fmt.Sprintf("%s@demo.linkedin", name), "demo12345"
}
