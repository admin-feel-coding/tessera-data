// Package handler contains HTTP handlers for tessera-data endpoints.
package handler

import (
	"net/http"

	"github.com/feel-coding/tessera-data/internal/httpx"
)

// Health handles GET /health.
func Health(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "tessera-data",
		"version": "0.1.0",
	})
}
