package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type FeedCache struct {
	client *redis.Client
	ttl    time.Duration
	prefix string
}

func NewFeedCache(redisURL string, ttl time.Duration, keyPrefix string) (*FeedCache, error) {
	if redisURL == "" {
		return nil, nil
	}
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("redis parse: %w", err)
	}
	if ttl <= 0 {
		ttl = 60 * time.Second
	}
	return &FeedCache{client: redis.NewClient(opt), ttl: ttl, prefix: keyPrefix}, nil
}

func (c *FeedCache) feedKey(userID uuid.UUID, variant string) string {
	return fmt.Sprintf("%sfeed:%s:%s", c.prefix, userID, variant)
}

func (c *FeedCache) Get(ctx context.Context, userID uuid.UUID, variant string, dst any) bool {
	if c == nil || c.client == nil {
		return false
	}
	key := c.feedKey(userID, variant)
	raw, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return false
	}
	return json.Unmarshal(raw, dst) == nil
}

func (c *FeedCache) Set(ctx context.Context, userID uuid.UUID, variant string, value any) {
	if c == nil || c.client == nil {
		return
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return
	}
	key := c.feedKey(userID, variant)
	_ = c.client.Set(ctx, key, raw, c.ttl).Err()
}

func (c *FeedCache) Invalidate(ctx context.Context, userID uuid.UUID) {
	if c == nil || c.client == nil {
		return
	}
	for _, v := range []string{"chronological", "ranked", "treatment", "control"} {
		_ = c.client.Del(ctx, c.feedKey(userID, v)).Err()
	}
}
