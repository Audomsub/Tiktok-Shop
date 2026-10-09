package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/delivery/http/handlers"
	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

// mockHealthUsecase implements domain.HealthUsecase for testing
type mockHealthUsecase struct {
	status *domain.HealthStatus
	err    error
}

func (m *mockHealthUsecase) Check(ctx context.Context) (*domain.HealthStatus, error) {
	return m.status, m.err
}

func TestHealthHandler_Healthy(t *testing.T) {
	mockUsecase := &mockHealthUsecase{
		status: &domain.HealthStatus{
			Status:    "ok",
			Database:  "connected",
			Timestamp: time.Now().UTC(),
			Version:   "1.0.0",
		},
		err: nil,
	}

	handler := handlers.NewHealthHandler(mockUsecase)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.Check(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rr.Code)
	}

	var res domain.HealthStatus
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	if res.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", res.Status)
	}
	if res.Database != "connected" {
		t.Errorf("expected database 'connected', got '%s'", res.Database)
	}
}

func TestHealthHandler_Degraded(t *testing.T) {
	mockUsecase := &mockHealthUsecase{
		status: &domain.HealthStatus{
			Status:    "degraded",
			Database:  "disconnected",
			Timestamp: time.Now().UTC(),
			Version:   "1.0.0",
		},
		err: errors.New("connection refused"),
	}

	handler := handlers.NewHealthHandler(mockUsecase)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.Check(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503 Service Unavailable, got %d", rr.Code)
	}

	var res domain.HealthStatus
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	if res.Status != "degraded" {
		t.Errorf("expected status 'degraded', got '%s'", res.Status)
	}
	if res.Database != "disconnected" {
		t.Errorf("expected database 'disconnected', got '%s'", res.Database)
	}
}
