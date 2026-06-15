package middleware

import (
	"os"
	"strings"
)

type Config struct {
	CORSOrigins          []string
	CORSAllowCredentials bool
}

func LoadConfigFromEnv() Config {
	origins := []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	if v := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS")); v != "" {
		origins = nil
		for _, part := range strings.Split(v, ",") {
			if o := strings.TrimSpace(part); o != "" {
				origins = append(origins, o)
			}
		}
	}
	allowCreds := true
	if v := strings.TrimSpace(os.Getenv("CORS_ALLOW_CREDENTIALS")); v == "0" || strings.EqualFold(v, "false") {
		allowCreds = false
	}
	return Config{
		CORSOrigins:          origins,
		CORSAllowCredentials: allowCreds,
	}
}
