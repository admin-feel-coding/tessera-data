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

// UsersHandler handles user-related HTTP endpoints.
type UsersHandler struct {
	pool *pgxpool.Pool
}

// NewUsers creates a UsersHandler backed by pool.
func NewUsers(pool *pgxpool.Pool) *UsersHandler {
	return &UsersHandler{pool: pool}
}

// GetHistory handles GET /users/{id}/history.
func (h *UsersHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "user id is required", nil)
		return
	}

	result, err := service.GetUserHistory(r.Context(), h.pool, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "user not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve user history", nil)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}
