package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

// LeaderboardHandler handles HTTP requests for winning product leaderboards
type LeaderboardHandler struct {
	usecase domain.LeaderboardUsecase
}

// NewLeaderboardHandler constructs a new LeaderboardHandler instance
func NewLeaderboardHandler(u domain.LeaderboardUsecase) *LeaderboardHandler {
	return &LeaderboardHandler{usecase: u}
}

// GetLeaderboard handles public GET /api/v1/leaderboard requests
func (h *LeaderboardHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	resp, err := h.usecase.GetLeaderboard(r.Context(), limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
