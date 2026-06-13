package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/feel-coding/tessera-data/internal/httpx"
	"github.com/feel-coding/tessera-data/internal/service"
	"github.com/feel-coding/tessera-data/internal/store"
)

// DevicesHandler handles device fingerprint endpoints.
type DevicesHandler struct {
	pool *pgxpool.Pool
}

// NewDevices creates a DevicesHandler backed by pool.
func NewDevices(pool *pgxpool.Pool) *DevicesHandler {
	return &DevicesHandler{pool: pool}
}

// GetFingerprint handles GET /devices/{id}/fingerprint.
func (h *DevicesHandler) GetFingerprint(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "id")
	if deviceID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "device id is required", nil)
		return
	}

	result, err := service.GetDeviceFingerprint(r.Context(), h.pool, deviceID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "device not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve device fingerprint", nil)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}
