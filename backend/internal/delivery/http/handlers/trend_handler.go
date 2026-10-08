package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
	"github.com/go-chi/chi/v5"
)

// TrendHandler handles HTTP requests for historical product trend analytics
type TrendHandler struct {
	usecase domain.TrendUsecase
}

// NewTrendHandler constructs a new TrendHandler instance
func NewTrendHandler(u domain.TrendUsecase) *TrendHandler {
	return &TrendHandler{usecase: u}
}

// GetProductTrends handles GET /api/v1/products/{id}/trends?days=7
func (h *TrendHandler) GetProductTrends(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	productID := strings.TrimSpace(chi.URLParam(r, "id"))
	if productID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "product id is required"})
		return
	}

	days := 7
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if parsed, err := strconv.Atoi(daysStr); err == nil && parsed > 0 {
			days = parsed
		}
	}

	resp, err := h.usecase.GetProductTrends(r.Context(), productID, days)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "product not found"})
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
