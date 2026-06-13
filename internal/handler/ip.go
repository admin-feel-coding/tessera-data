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

// IPHandler handles IP risk signal endpoints.
type IPHandler struct {
	pool *pgxpool.Pool
}

// NewIP creates an IPHandler backed by pool.
func NewIP(pool *pgxpool.Pool) *IPHandler {
	return &IPHandler{pool: pool}
}

// GetRisk handles GET /ip/{ip}/risk.
func (h *IPHandler) GetRisk(w http.ResponseWriter, r *http.Request) {
	ip := chi.URLParam(r, "ip")
	if ip == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "ip is required", nil)
		return
	}

	result, err := service.GetIPRisk(r.Context(), h.pool, ip)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "ip not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve ip risk", nil)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}
