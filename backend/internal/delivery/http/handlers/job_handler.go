package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

// JobHandler handles ingestion and scoring job requests
type JobHandler struct {
	analyticsUsecase domain.AnalyticsUsecase
	internalAPIKey   string
}

// NewJobHandler constructs a new JobHandler instance
func NewJobHandler(u domain.AnalyticsUsecase, internalAPIKey string) *JobHandler {
	return &JobHandler{
		analyticsUsecase: u,
		internalAPIKey:   internalAPIKey,
	}
}

// ComputeScoresRequest represents JSON request body for scoring jobs
type ComputeScoresRequest struct {
	CrawlLogID string `json:"crawl_log_id"`
}

// ComputeScores processes batch score calculations for a crawl round
func (h *JobHandler) ComputeScores(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Authenticate internal webhook via X-API-Key
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" || apiKey != h.internalAPIKey {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized: invalid or missing X-API-Key"})
		return
	}

	// 2. Extract crawl_log_id from query parameter or JSON body
	crawlLogID := strings.TrimSpace(r.URL.Query().Get("crawl_log_id"))
	if crawlLogID == "" && r.Body != nil {
		var req ComputeScoresRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			crawlLogID = strings.TrimSpace(req.CrawlLogID)
		}
	}

	if crawlLogID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "bad request: crawl_log_id is required"})
		return
	}

	// 3. Process crawl round scoring
	result, err := h.analyticsUsecase.ProcessCrawlRound(r.Context(), crawlLogID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}
