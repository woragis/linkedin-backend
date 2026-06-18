package realm

import (
	"context"
	"net/http"
	"strings"
)

// ID identifies which database realm a request uses.
type ID string

const (
	Volume ID = "volume"
	Live   ID = "live"
)

const Header = "X-App-Realm"

type ctxKey struct{}

// FromHeader parses the realm header. Default is Live.
func FromHeader(value string) ID {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(Volume), "vol", "lab":
		return Volume
	default:
		return Live
	}
}

func WithContext(ctx context.Context, id ID) context.Context {
	if id == "" {
		id = Live
	}
	return context.WithValue(ctx, ctxKey{}, id)
}

func FromContext(ctx context.Context) ID {
	if v, ok := ctx.Value(ctxKey{}).(ID); ok && v != "" {
		return v
	}
	return Live
}

// Middleware stores the resolved realm on the request context.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := FromHeader(r.Header.Get(Header))
		next.ServeHTTP(w, r.WithContext(WithContext(r.Context(), id)))
	})
}

// CachePrefix returns the Redis key prefix for feed cache isolation.
func CachePrefix(id ID) string {
	if id == Volume {
		return "volume:"
	}
	return "live:"
}
