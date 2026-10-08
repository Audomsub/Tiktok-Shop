package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/delivery/http/handlers"
	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

type mockAnalyticsUsecase struct {
	result *domain.BatchProcessResult
	err    error
}

func (m *mockAnalyticsUsecase) ProcessCrawlRound(ctx context.Context, crawlLogID string) (*domain.BatchProcessResult, error) {
	return m.result, m.err
}

func TestJobHandler_ComputeScores(t *testing.T) {
	apiKey := "test-secret-key"

	t.Run("Rejects request with missing or incorrect API Key", func(t *testing.T) {
		mockUsecase := &mockAnalyticsUsecase{}
		handler := handlers.NewJobHandler(mockUsecase, apiKey)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/jobs/compute-scores?crawl_log_id=log-1", nil)
		// Missing X-API-Key header
		rr := httptest.NewRecorder()

		handler.ComputeScores(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d", rr.Code)
		}
	})

	t.Run("Rejects request with missing crawl_log_id", func(t *testing.T) {
		mockUsecase := &mockAnalyticsUsecase{}
		handler := handlers.NewJobHandler(mockUsecase, apiKey)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/jobs/compute-scores", nil)
		req.Header.Set("X-API-Key", apiKey)
		rr := httptest.NewRecorder()

		handler.ComputeScores(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", rr.Code)
		}
	})

	t.Run("Processes valid request via query parameter", func(t *testing.T) {
		mockUsecase := &mockAnalyticsUsecase{
			result: &domain.BatchProcessResult{
				CrawlLogID:     "log-123",
				ProcessedCount: 200,
				DurationMs:     450,
				Status:         "SUCCESS",
			},
		}
		handler := handlers.NewJobHandler(mockUsecase, apiKey)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/jobs/compute-scores?crawl_log_id=log-123", nil)
		req.Header.Set("X-API-Key", apiKey)
		rr := httptest.NewRecorder()

		handler.ComputeScores(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rr.Code)
		}

		var res domain.BatchProcessResult
		if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed decoding response: %v", err)
		}

		if res.ProcessedCount != 200 || res.Status != "SUCCESS" {
			t.Errorf("unexpected response content: %+v", res)
		}
	})

	t.Run("Processes valid request via JSON body", func(t *testing.T) {
		mockUsecase := &mockAnalyticsUsecase{
			result: &domain.BatchProcessResult{
				CrawlLogID:     "log-456",
				ProcessedCount: 50,
				DurationMs:     120,
				Status:         "SUCCESS",
			},
		}
		handler := handlers.NewJobHandler(mockUsecase, apiKey)

		body := bytes.NewBufferString(`{"crawl_log_id":"log-456"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/jobs/compute-scores", body)
		req.Header.Set("X-API-Key", apiKey)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ComputeScores(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rr.Code)
		}
	})

	t.Run("Returns 500 when usecase fails", func(t *testing.T) {
		mockUsecase := &mockAnalyticsUsecase{
			err: errors.New("database connection failed"),
		}
		handler := handlers.NewJobHandler(mockUsecase, apiKey)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/jobs/compute-scores?crawl_log_id=log-fail", nil)
		req.Header.Set("X-API-Key", apiKey)
		rr := httptest.NewRecorder()

		handler.ComputeScores(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 Internal Server Error, got %d", rr.Code)
		}
	})
}
