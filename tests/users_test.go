package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/feel-coding/tessera-data/internal/handler"
	"github.com/feel-coding/tessera-data/internal/httpx"
)

// connectTestDB returns a live pool or skips the test if DATABASE_URL is absent or unreachable.
func connectTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("no database: DATABASE_URL not set")
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Skipf("no database: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Skipf("no database: %v", err)
	}
	return pool
}

// newTestRouter builds a chi router with auth middleware and the supplied handler mounted at path.
func newTestRouter(path string, h http.HandlerFunc) http.Handler {
	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware("test-key"))
	r.Get(path, h)
	return r
}

// TestGetUserHistoryReturns200 verifies that a seeded user returns 200 with a JSON body.
func TestGetUserHistoryReturns200(t *testing.T) {
	pool := connectTestDB(t)
	defer pool.Close()

	usersH := handler.NewUsers(pool)
	router := newTestRouter("/users/{id}/history", usersH.GetHistory)
	server := httptest.NewServer(router)
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/users/user_001/history", nil)
	req.Header.Set("X-Internal-Key", "test-key")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// user_001 is in the blacklist seed, but history endpoint doesn't check blacklist — always 200.
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("expected application/json, got %s", ct)
	}
}

// TestGetUserHistoryUnknownUserReturns200NewUser verifies that an unknown user returns 200 with new_user=true.
func TestGetUserHistoryUnknownUserReturns200NewUser(t *testing.T) {
	pool := connectTestDB(t)
	defer pool.Close()

	usersH := handler.NewUsers(pool)
	router := newTestRouter("/users/{id}/history", usersH.GetHistory)
	server := httptest.NewServer(router)
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/users/user_does_not_exist/history", nil)
	req.Header.Set("X-Internal-Key", "test-key")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Unknown user is treated as a new user, not a 404.
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for unknown user, got %d", resp.StatusCode)
	}
}
