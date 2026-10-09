package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

// HealthHandler handles health checking HTTP requests
type HealthHandler struct {
	usecase domain.HealthUsecase
}

// NewHealthHandler creates a new HealthHandler instance
func NewHealthHandler(u domain.HealthUsecase) *HealthHandler {
	return &HealthHandler{usecase: u}
}

// Check responds with the service and database health status
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	status, err := h.usecase.Check(r.Context())

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(status)
}
