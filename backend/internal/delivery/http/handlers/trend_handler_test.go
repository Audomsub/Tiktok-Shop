package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
	"github.com/go-chi/chi/v5"
)

type mockTrendUsecase struct {
	getTrendsFunc func(ctx context.Context, productID string, days int) (*domain.ProductTrendsResponse, error)
}

func (m *mockTrendUsecase) GetProductTrends(ctx context.Context, productID string, days int) (*domain.ProductTrendsResponse, error) {
	if m.getTrendsFunc != nil {
		return m.getTrendsFunc(ctx, productID, days)
	}
	return nil, nil
}

func TestTrendHandler_GetProductTrends(t *testing.T) {
	t.Run("Returns 200 with trend response on valid product ID", func(t *testing.T) {
		mockUsecase := &mockTrendUsecase{
			getTrendsFunc: func(ctx context.Context, productID string, days int) (*domain.ProductTrendsResponse, error) {
				if productID != "prod-123" {
					t.Errorf("expected productID 'prod-123', got '%s'", productID)
				}
				if days != 14 {
					t.Errorf("expected days 14, got %d", days)
				}

				return &domain.ProductTrendsResponse{
					ProductID:   productID,
					ProductName: "Sample Product",
					Days:        days,
					TotalPoints: 1,
					Points: []*domain.TrendPoint{
						{
							SnapshotTime:    time.Now(),
							Price:           300.0,
							CommissionRate:  20.0,
							TotalSales:      500,
							VelocityPerHour: 10.0,
							ExpectedReturn:  60.0,
							WinningScore:    85.0,
						},
					},
				}, nil
			},
		}

		handler := NewTrendHandler(mockUsecase)

		r := chi.NewRouter()
		r.Get("/api/v1/products/{id}/trends", handler.GetProductTrends)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/products/prod-123/trends?days=14", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp domain.ProductTrendsResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed decoding json: %v", err)
		}

		if resp.ProductID != "prod-123" || resp.TotalPoints != 1 {
			t.Errorf("unexpected response content: %+v", resp)
		}
	})

	t.Run("Returns 404 when product is not found", func(t *testing.T) {
		mockUsecase := &mockTrendUsecase{
			getTrendsFunc: func(ctx context.Context, productID string, days int) (*domain.ProductTrendsResponse, error) {
				return nil, domain.ErrProductNotFound
			},
		}

		handler := NewTrendHandler(mockUsecase)

		r := chi.NewRouter()
		r.Get("/api/v1/products/{id}/trends", handler.GetProductTrends)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/products/non-existent/trends", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("Returns 500 when usecase returns an unexpected error", func(t *testing.T) {
		mockUsecase := &mockTrendUsecase{
			getTrendsFunc: func(ctx context.Context, productID string, days int) (*domain.ProductTrendsResponse, error) {
				return nil, errors.New("db error")
			},
		}

		handler := NewTrendHandler(mockUsecase)

		r := chi.NewRouter()
		r.Get("/api/v1/products/{id}/trends", handler.GetProductTrends)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/products/prod-1/trends", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}
	})
}
