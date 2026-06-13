// Package tests contains integration-style tests for tessera-data.
package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/feel-coding/tessera-data/internal/handler"
)

// TestHealthReturns200 verifies that the /health handler responds with 200 OK.
func TestHealthReturns200(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.Health(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
