package handler

import (
	"net/http"
	"strconv"

	"github.com/feel-coding/tessera-data/internal/httpx"
	"github.com/feel-coding/tessera-data/internal/service"
)

const defaultWindowMinutes = 60

// VelocityHandler handles cross-user velocity check endpoints.
type VelocityHandler struct {
	svc *service.VelocityService
}

// NewVelocityHandler creates a VelocityHandler backed by svc.
func NewVelocityHandler(svc *service.VelocityService) *VelocityHandler {
	return &VelocityHandler{svc: svc}
}

// Check handles GET /velocity/check?ip=...&card_bin=...&window_minutes=60.
// It returns distinct-user counts for the given IP and card BIN within the
// requested time window.
func (h *VelocityHandler) Check(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	ip := q.Get("ip")
	bin := q.Get("card_bin")

	if ip == "" && bin == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "at least one of ip or card_bin is required", nil)
		return
	}

	window := defaultWindowMinutes
	if wStr := q.Get("window_minutes"); wStr != "" {
		n, err := strconv.Atoi(wStr)
		if err != nil || n < 1 {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "window_minutes must be a positive integer", nil)
			return
		}
		window = n
	}

	result, err := h.svc.Check(r.Context(), ip, bin, window)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "VELOCITY_ERROR", err.Error(), nil)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}
