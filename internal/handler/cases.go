package handler

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/feel-coding/tessera-data/internal/httpx"
	"github.com/feel-coding/tessera-data/internal/service"
)

// CasesHandler serves case similarity search and persistence endpoints.
type CasesHandler struct {
	pool *pgxpool.Pool
}

// NewCases creates a CasesHandler backed by pool.
func NewCases(pool *pgxpool.Pool) *CasesHandler {
	return &CasesHandler{pool: pool}
}

// FindSimilar handles POST /cases/similar.
func (h *CasesHandler) FindSimilar(w http.ResponseWriter, r *http.Request) {
	if h.pool == nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, "DB_UNAVAILABLE", "database not connected", nil)
		return
	}
	var in service.SimilarSearchInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body", nil)
		return
	}
	if len(in.Embedding) != 1536 {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "embedding must be 1536 floats", map[string]any{"got_length": len(in.Embedding)})
		return
	}
	cases, err := service.FindSimilarCases(r.Context(), h.pool, in)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"cases": cases})
}

// Save handles POST /cases.
func (h *CasesHandler) Save(w http.ResponseWriter, r *http.Request) {
	if h.pool == nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, "DB_UNAVAILABLE", "database not connected", nil)
		return
	}
	var in service.SaveCaseInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body", nil)
		return
	}
	if len(in.Embedding) != 1536 {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "embedding must be 1536 floats", map[string]any{"got_length": len(in.Embedding)})
		return
	}
	if in.TransactionID == "" || in.Decision == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "transaction_id and decision are required", nil)
		return
	}
	id, err := service.SaveCase(r.Context(), h.pool, in)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, map[string]any{"id": id, "transaction_id": in.TransactionID})
}
