package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/feel-coding/tessera-data/internal/httpx"
	"github.com/feel-coding/tessera-data/internal/service"
)

// BlacklistHandler handles blacklist check endpoints.
type BlacklistHandler struct {
	pool *pgxpool.Pool
}

// NewBlacklist creates a BlacklistHandler backed by pool.
func NewBlacklist(pool *pgxpool.Pool) *BlacklistHandler {
	return &BlacklistHandler{pool: pool}
}

// Check handles GET /blacklist/check.
func (h *BlacklistHandler) Check(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userID := q.Get("user_id")
	email := q.Get("email")
	cardBin := q.Get("card_bin")

	if userID == "" && email == "" && cardBin == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "at least one of user_id, email, or card_bin is required", nil)
		return
	}

	result, err := service.CheckBlacklist(r.Context(), h.pool, userID, email, cardBin)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to check blacklist", nil)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}
