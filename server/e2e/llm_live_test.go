//go:build e2e

package e2e_test

import (
	"net/http"
	"os"
	"strings"
	"testing"
)

var templateMarkers = []string{
	"Migrando nosso pipeline para Go",
	"Rodando experimento A/B no feed",
	"Next.js 16 no projeto novo",
	"Compose com worker-realtime",
	"DAU e retenção D7",
}

func isTemplateText(text string) bool {
	for _, m := range templateMarkers {
		if strings.Contains(text, m) {
			return true
		}
	}
	return false
}

func looksPortuguese(text string) bool {
	lower := strings.ToLower(text)
	for _, w := range []string{" de ", " em ", " que ", " para ", " com "} {
		if strings.Contains(lower, w) {
			return true
		}
	}
	for _, r := range text {
		if strings.ContainsRune("áàâãéêíóôõúç", r) {
			return true
		}
	}
	return len(text) >= 24
}

func TestE2E_LLM_LiveFiveUsersTwoPostsTenComments(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("OPENAI_API_KEY not set — export key and ensure API container has it in .env")
	}

	waitReady(t)
	seedDemoLive(t)

	code, run := httpJSON(t, http.MethodPost, "/v1/internal/live-llm-run", nil, liveInternalHeaders())
	if code != http.StatusCreated {
		if code == http.StatusInternalServerError {
			if msg, _ := run["message"].(string); strings.Contains(msg, "OPENAI_API_KEY") {
				t.Skipf("API missing OPENAI_API_KEY: %v", run)
			}
		}
		t.Fatalf("POST live-llm-run status=%d body=%v", code, run)
	}

	postsCreated, _ := run["posts_created"].(float64)
	commentsCreated, _ := run["comments_created"].(float64)
	if postsCreated != 2 {
		t.Fatalf("posts_created=%v want 2", run["posts_created"])
	}
	if commentsCreated != 10 {
		t.Fatalf("comments_created=%v want 10", run["comments_created"])
	}

	postIDs, ok := run["post_ids"].([]any)
	if !ok || len(postIDs) < 1 {
		t.Fatalf("missing post_ids in %v", run)
	}
	post1ID, _ := postIDs[0].(string)

	anaEmail, anaPwd := demoUser("ana")
	token := login(t, anaEmail, anaPwd)

	code, feedOut := httpJSON(t, http.MethodGet, "/v1/feed", nil, mergeHeaders(authHeader(token), liveRealmHeaders()))
	if code != http.StatusOK {
		t.Fatalf("GET feed status=%d body=%v", code, feedOut)
	}
	items, ok := feedOut["posts"].([]any)
	if !ok {
		t.Fatalf("feed posts missing: %v", feedOut)
	}

	var llmPostBody string
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		id, _ := item["id"].(string)
		body, _ := item["body"].(string)
		if id == post1ID {
			llmPostBody = body
			break
		}
	}
	if llmPostBody == "" {
		t.Fatalf("llm post %s not found in feed", post1ID)
	}
	if isTemplateText(llmPostBody) {
		t.Fatalf("llm post looks like seed template: %q", llmPostBody)
	}
	if !looksPortuguese(llmPostBody) {
		t.Fatalf("llm post does not look PT-BR: %q", llmPostBody)
	}

	comments := listCommentsLive(t, token, post1ID)
	if len(comments) < 9 {
		t.Fatalf("expected at least 9 root comments, got %d", len(comments))
	}

	llmComments := 0
	hasReply := false
	for _, raw := range comments {
		c, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		body, _ := c["body"].(string)
		if body != "" && !isTemplateText(body) && looksPortuguese(body) {
			llmComments++
		}
		replies, _ := c["replies"].([]any)
		if len(replies) > 0 {
			hasReply = true
			for _, rr := range replies {
				reply, ok := rr.(map[string]any)
				if !ok {
					continue
				}
				rb, _ := reply["body"].(string)
				if rb != "" && !isTemplateText(rb) && looksPortuguese(rb) {
					llmComments++
				}
			}
		}
	}
	if llmComments < 8 {
		t.Fatalf("expected mostly LLM comments, got %d non-template PT comments", llmComments)
	}
	if !hasReply {
		t.Fatal("expected at least one threaded reply from LLM scenario")
	}
}

func mergeHeaders(parts ...map[string]string) map[string]string {
	out := make(map[string]string)
	for _, p := range parts {
		for k, v := range p {
			out[k] = v
		}
	}
	return out
}
