package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

// CategoryHandler handles category taxonomy API requests
type CategoryHandler struct {
	usecase domain.CategoryUsecase
}

// NewCategoryHandler constructs a new CategoryHandler instance
func NewCategoryHandler(u domain.CategoryUsecase) *CategoryHandler {
	return &CategoryHandler{usecase: u}
}

// GetCategories handles GET /api/v1/categories
func (h *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	categories, err := h.usecase.GetCategories(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(categories)
}
