//go:build e2e

package e2e_test

import (
	"net/http"
	"testing"
)

func TestE2E_R2_FiveUsersTwoPostsTenInteractions(t *testing.T) {
	waitReady(t)
	seedDemo(t)

	ana := login(t, demoUser("ana"))
	bruno := login(t, demoUser("bruno"))
	carla := login(t, demoUser("carla"))
	diego := login(t, demoUser("diego"))
	elisa := login(t, demoUser("elisa"))

	post1 := createPost(t, ana, "E2E R2 post 1")
	post2 := createPost(t, bruno, "E2E R2 post 2")

	// 1. ana top comment on post1
	c1 := createComment(t, ana, post1, "ana root comment")

	// 2. bruno reply to ana comment
	c1Reply := replyComment(t, bruno, post1, c1, "bruno reply")

	// 3. carla like on root
	reactComment(t, carla, c1, "like")

	// 4. diego celebrate on reply
	reactComment(t, diego, c1Reply, "celebrate")

	// 5. elisa top comment on post1
	c2 := createComment(t, elisa, post1, "elisa root comment")

	// 6. ana insightful on elisa comment
	reactComment(t, ana, c2, "insightful")

	// 7. bruno top comment on post2
	c3 := createComment(t, bruno, post2, "bruno root on post2")

	// 8. carla reply to bruno comment
	c3Reply := replyComment(t, carla, post2, c3, "carla reply on post2")

	// 9. diego love on bruno comment
	reactComment(t, diego, c3, "love")

	// 10. elisa funny on reply
	reactComment(t, elisa, c3Reply, "funny")

	comments := listComments(t, ana, post1)
	if len(comments) != 2 {
		t.Fatalf("expected 2 root comments on post1, got %d", len(comments))
	}

	var rootAna, rootElisa map[string]any
	for _, c := range comments {
		m, ok := c.(map[string]any)
		if !ok {
			t.Fatal("comment is not an object")
		}
		switch m["id"] {
		case c1:
			rootAna = m
		case c2:
			rootElisa = m
		}
	}
	if rootAna == nil || rootElisa == nil {
		t.Fatalf("missing expected root comments")
	}

	replies, ok := rootAna["replies"].([]any)
	if !ok || len(replies) != 1 {
		t.Fatalf("expected 1 reply on ana root, got %v", rootAna["replies"])
	}
	reply, ok := replies[0].(map[string]any)
	if !ok || reply["id"] != c1Reply {
		t.Fatalf("unexpected reply: %v", replies[0])
	}

	assertReactionCountE2E(t, rootAna, "like", 1)
	assertReactionCountE2E(t, reply, "celebrate", 1)
	assertReactionCountE2E(t, rootElisa, "insightful", 1)

	code, errBody := httpJSON(t, http.MethodPost, "/v1/posts/"+post1+"/comments", map[string]any{
		"body":              "nested too deep",
		"parent_comment_id": c1Reply,
	}, authHeader(ana))
	if code != http.StatusBadRequest {
		t.Fatalf("reply-to-reply expected 400, got %d body=%v", code, errBody)
	}
}
