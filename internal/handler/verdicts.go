package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/feel-coding/tessera-data/internal/httpx"
	"github.com/feel-coding/tessera-data/internal/service"
	"github.com/feel-coding/tessera-data/internal/store"
)

const (
	defaultVerdictLimit = 50
	maxVerdictLimit     = 200
)

// VerdictsHandler serves verdict persistence and retrieval endpoints.
type VerdictsHandler struct {
	pool *pgxpool.Pool
}

// NewVerdicts creates a VerdictsHandler backed by pool.
func NewVerdicts(pool *pgxpool.Pool) *VerdictsHandler {
	return &VerdictsHandler{pool: pool}
}

// Save handles POST /verdicts.
func (h *VerdictsHandler) Save(w http.ResponseWriter, r *http.Request) {
	if h.pool == nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, "DB_UNAVAILABLE", "database not connected", nil)
		return
	}
	var in service.SaveVerdictInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body", nil)
		return
	}
	if in.TransactionID == "" || in.Decision == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "transaction_id and decision are required", nil)
		return
	}
	id, err := service.SaveVerdict(r.Context(), h.pool, in)
	if err != nil {
		if err.Error() == "decision must be one of APPROVE, DECLINE, ESCALATE" {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, map[string]any{"id": id, "transaction_id": in.TransactionID})
}

// List handles GET /verdicts.
func (h *VerdictsHandler) List(w http.ResponseWriter, r *http.Request) {
	if h.pool == nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, "DB_UNAVAILABLE", "database not connected", nil)
		return
	}

	limit := defaultVerdictLimit
	if s := r.URL.Query().Get("limit"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "limit must be a positive integer", nil)
			return
		}
		if n > maxVerdictLimit {
			n = maxVerdictLimit
		}
		limit = n
	}

	offset := 0
	if s := r.URL.Query().Get("offset"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "offset must be a non-negative integer", nil)
			return
		}
		offset = n
	}

	verdicts, total, err := service.ListVerdicts(r.Context(), h.pool, limit, offset)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"verdicts": verdicts, "total": total})
}

// GetByTransactionID handles GET /verdicts/{transaction_id}.
func (h *VerdictsHandler) GetByTransactionID(w http.ResponseWriter, r *http.Request) {
	if h.pool == nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, "DB_UNAVAILABLE", "database not connected", nil)
		return
	}
	transactionID := chi.URLParam(r, "transaction_id")
	verdict, err := service.GetVerdictByTransactionID(r.Context(), h.pool, transactionID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "verdict not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, verdict)
}
