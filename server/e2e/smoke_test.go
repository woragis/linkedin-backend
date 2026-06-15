//go:build e2e

package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func apiBase() string {
	if v := os.Getenv("API_BASE"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "http://127.0.0.1:8080"
}

func internalToken() string {
	if v := os.Getenv("INTERNAL_JOB_TOKEN"); v != "" {
		return v
	}
	return "dev-internal-token"
}

func httpJSON(t *testing.T, method, path string, body any, headers map[string]string) (int, map[string]any) {
	t.Helper()
	var r io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		r = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, apiBase()+path, r)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if len(data) > 0 {
		if err := json.Unmarshal(data, &out); err != nil {
			var arr []any
			if err2 := json.Unmarshal(data, &arr); err2 == nil {
				return res.StatusCode, map[string]any{"_array": arr}
			}
			t.Fatalf("%s %s: decode json: %v body=%s", method, path, err, string(data))
		}
	}
	return res.StatusCode, out
}

func httpJSONArray(t *testing.T, method, path string, body any, headers map[string]string) (int, []any) {
	t.Helper()
	code, out := httpJSON(t, method, path, body, headers)
	if arr, ok := out["_array"].([]any); ok {
		return code, arr
	}
	return code, nil
}

func waitReady(t *testing.T) {
	t.Helper()
	res, err := http.Get(apiBase() + "/health")
	if err != nil {
		t.Skipf("API not available at %s: %v", apiBase(), err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Skipf("API health check failed at %s (status %d)", apiBase(), res.StatusCode)
	}

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		res, err := http.Get(apiBase() + "/ready")
		if err == nil && res.StatusCode == http.StatusOK {
			res.Body.Close()
			return
		}
		if res != nil {
			res.Body.Close()
		}
		time.Sleep(time.Second)
	}
	t.Skipf("API /ready not available at %s", apiBase())
}

func TestSmoke_DemoFlow(t *testing.T) {
	waitReady(t)

	code, _ := httpJSON(t, http.MethodGet, "/health", nil, nil)
	if code != http.StatusOK {
		t.Fatalf("GET /health status=%d", code)
	}

	code, seed := httpJSON(t, http.MethodPost, "/v1/internal/seed-demo", nil, map[string]string{
		"X-Internal-Token": internalToken(),
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("POST seed-demo status=%d body=%v", code, seed)
	}

	code, login := httpJSON(t, http.MethodPost, "/v1/auth/login", map[string]any{
		"email":    "ana@demo.linkedin",
		"password": "demo12345",
	}, nil)
	if code != http.StatusOK {
		t.Fatalf("login status=%d body=%v", code, login)
	}
	token, _ := login["token"].(string)
	if token == "" {
		t.Fatal("missing token")
	}
	auth := map[string]string{"Authorization": "Bearer " + token}

	code, post := httpJSON(t, http.MethodPost, "/v1/posts", map[string]any{
		"body": fmt.Sprintf("E2E smoke post at %s", time.Now().Format(time.RFC3339)),
	}, auth)
	if code != http.StatusCreated {
		t.Fatalf("create post status=%d body=%v", code, post)
	}
	postID, _ := post["id"].(string)
	if postID == "" {
		t.Fatal("missing post id")
	}

	code, comment := httpJSON(t, http.MethodPost, "/v1/posts/"+postID+"/comments", map[string]any{
		"body": "Smoke test comment",
	}, auth)
	if code != http.StatusCreated {
		t.Fatalf("create comment status=%d body=%v", code, comment)
	}

	code, comments := httpJSONArray(t, http.MethodGet, "/v1/posts/"+postID+"/comments", nil, nil)
	if code != http.StatusOK {
		t.Fatalf("list comments status=%d", code)
	}
	if len(comments) == 0 {
		t.Fatal("expected at least one comment")
	}

	code, connections := httpJSONArray(t, http.MethodGet, "/v1/connections", nil, auth)
	if code != http.StatusOK {
		t.Fatalf("list connections status=%d", code)
	}
	if len(connections) == 0 {
		t.Fatal("expected demo user to have connections")
	}
}
