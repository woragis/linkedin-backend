package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	searchrepo "github.com/unipe/linkedin/backend/server/internal/search/repository"
)

type Client struct {
	base   string
	client *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		base:   strings.TrimRight(baseURL, "/"),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.base != ""
}

func (c *Client) SearchPeople(ctx context.Context, q string, limit int) ([]searchrepo.PersonHit, error) {
	body := map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":  q,
				"fields": []string{"full_name^3", "headline^2", "bio", "skills", "schools", "location"},
			},
		},
		"size": limit,
	}
	return c.searchPeople(ctx, "people", body)
}

func (c *Client) SearchPosts(ctx context.Context, q string, limit int) ([]searchrepo.PostHit, error) {
	body := map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":  q,
				"fields": []string{"body^2", "author_name"},
			},
		},
		"size": limit,
	}
	raw, err := c.post(ctx, "/posts/_search", body)
	if err != nil {
		return nil, err
	}
	var resp esResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	out := make([]searchrepo.PostHit, 0, len(resp.Hits.Hits))
	for _, h := range resp.Hits.Hits {
		out = append(out, searchrepo.PostHit{
			PostID:     uuidMust(h.Source["post_id"]),
			AuthorID:   uuidMust(h.Source["author_id"]),
			AuthorName: str(h.Source["author_name"]),
			Body:       str(h.Source["body"]),
			Score:      h.Score,
		})
	}
	return out, nil
}

type esResponse struct {
	Hits struct {
		Hits []struct {
			Score  float64                `json:"_score"`
			Source map[string]interface{} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (c *Client) searchPeople(ctx context.Context, index string, body map[string]any) ([]searchrepo.PersonHit, error) {
	raw, err := c.post(ctx, "/"+index+"/_search", body)
	if err != nil {
		return nil, err
	}
	var resp esResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	out := make([]searchrepo.PersonHit, 0, len(resp.Hits.Hits))
	for _, h := range resp.Hits.Hits {
		out = append(out, searchrepo.PersonHit{
			UserID:   uuidMust(h.Source["user_id"]),
			Slug:     str(h.Source["slug"]),
			FullName: str(h.Source["full_name"]),
			Headline: str(h.Source["headline"]),
			Location: str(h.Source["location"]),
			Score:    h.Score,
		})
	}
	return out, nil
}

func (c *Client) post(ctx context.Context, path string, body map[string]any) ([]byte, error) {
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+path, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("elasticsearch %s: %s", res.Status, string(b))
	}
	return b, nil
}

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func uuidMust(v any) uuid.UUID {
	id, _ := uuid.Parse(str(v))
	return id
}
