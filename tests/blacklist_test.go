package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/feel-coding/tessera-data/internal/handler"
	"github.com/feel-coding/tessera-data/internal/httpx"
)

// TestCheckBlacklistNoMatch verifies that a clean identifier returns match=false.
func TestCheckBlacklistNoMatch(t *testing.T) {
	pool := connectTestDB(t)
	defer pool.Close()

	blacklistH := handler.NewBlacklist(pool)
	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware("test-key"))
	r.Get("/blacklist/check", blacklistH.Check)
	server := httptest.NewServer(r)
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/blacklist/check?user_id=user_999", nil)
	req.Header.Set("X-Internal-Key", "test-key")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body struct {
		Match bool `json:"match"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Match {
		t.Error("expected match=false for unknown user")
	}
}

// TestCheckBlacklistMatchAfterInsert inserts a blacklist entry and asserts match=true.
func TestCheckBlacklistMatchAfterInsert(t *testing.T) {
	pool := connectTestDB(t)
	defer pool.Close()

	// Insert a test entry idempotently.
	_, err := pool.Exec(context.Background(), `
		INSERT INTO blacklist (kind, value, reason)
		VALUES ('user_id', 'user_001', 'test entry')
		ON CONFLICT (kind, value) DO NOTHING
	`)
	if err != nil {
		t.Fatalf("failed to insert blacklist entry: %v", err)
	}

	blacklistH := handler.NewBlacklist(pool)
	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware("test-key"))
	r.Get("/blacklist/check", blacklistH.Check)
	server := httptest.NewServer(r)
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/blacklist/check?user_id=user_001", nil)
	req.Header.Set("X-Internal-Key", "test-key")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body struct {
		Match bool `json:"match"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !body.Match {
		t.Error("expected match=true for blacklisted user_001")
	}
}
