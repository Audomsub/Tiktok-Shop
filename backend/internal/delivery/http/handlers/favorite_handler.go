package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
	"github.com/go-chi/chi/v5"
)

// FavoriteHandler handles user bookmarks / favorite products HTTP endpoints
type FavoriteHandler struct {
	usecase domain.FavoriteUsecase
}

// NewFavoriteHandler constructs a new FavoriteHandler instance
func NewFavoriteHandler(u domain.FavoriteUsecase) *FavoriteHandler {
	return &FavoriteHandler{usecase: u}
}

// AddFavorite handles POST /api/v1/favorites
func (h *FavoriteHandler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := domain.GetUserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	var req domain.AddFavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request payload"})
		return
	}

	cleanProductID := strings.TrimSpace(req.ProductID)
	if cleanProductID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "product_id is required"})
		return
	}

	if err := h.usecase.AddFavorite(r.Context(), userID, cleanProductID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message":    "product added to favorites",
		"product_id": cleanProductID,
	})
}

// RemoveFavorite handles DELETE /api/v1/favorites/{productId}
func (h *FavoriteHandler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := domain.GetUserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	productID := strings.TrimSpace(chi.URLParam(r, "productId"))
	if productID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "productId is required"})
		return
	}

	if err := h.usecase.RemoveFavorite(r.Context(), userID, productID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message":    "product removed from favorites",
		"product_id": productID,
	})
}

// GetFavorites handles GET /api/v1/favorites
func (h *FavoriteHandler) GetFavorites(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := domain.GetUserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	resp, err := h.usecase.GetFavorites(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
