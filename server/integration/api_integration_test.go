//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestRegisterLoginProfilePatch(t *testing.T) {
	srv, cleanup := setupIntegrationServer(t)
	defer cleanup()

	email := fmt.Sprintf("integration-%d@test.local", time.Now().UnixNano())
	password := "integration123"

	code, reg := httpJSON(t, srv.URL, http.MethodPost, "/v1/auth/register", map[string]any{
		"email":     email,
		"password":  password,
		"full_name": "Integration User",
	}, nil)
	if code != http.StatusCreated {
		t.Fatalf("register status=%d body=%v", code, reg)
	}
	token, _ := reg["token"].(string)
	if token == "" {
		t.Fatal("missing token")
	}
	auth := authHeader(token)

	code, login := httpJSON(t, srv.URL, http.MethodPost, "/v1/auth/login", map[string]any{
		"email":    email,
		"password": password,
	}, nil)
	if code != http.StatusOK {
		t.Fatalf("login status=%d body=%v", code, login)
	}

	headline := "Integration headline"
	code, profile := httpJSON(t, srv.URL, http.MethodPatch, "/v1/me/profile", map[string]any{
		"headline": headline,
	}, auth)
	if code != http.StatusOK {
		t.Fatalf("patch profile status=%d body=%v", code, profile)
	}
	if profile["headline"] != headline {
		t.Fatalf("expected headline %q, got %v", headline, profile["headline"])
	}
}
