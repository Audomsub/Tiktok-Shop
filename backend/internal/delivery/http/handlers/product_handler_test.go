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
)

type mockProductUsecase struct {
	getCatalogFunc func(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error)
}

func (m *mockProductUsecase) GetCatalog(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error) {
	if m.getCatalogFunc != nil {
		return m.getCatalogFunc(ctx, filter)
	}
	return nil, nil
}

func TestProductHandler_GetCatalog(t *testing.T) {
	t.Run("Returns 200 with catalog response on valid query", func(t *testing.T) {
		var passedFilter domain.ProductCatalogFilter

		mockUsecase := &mockProductUsecase{
			getCatalogFunc: func(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error) {
				passedFilter = filter
				return &domain.ProductCatalogResponse{
					Page:       filter.Page,
					Limit:      filter.Limit,
					TotalCount: 1,
					TotalPages: 1,
					Items: []*domain.ProductCatalogItem{
						{
							ID:           "prod-1",
							Name:         "Wireless Earbuds",
							Price:        499.0,
							WinningScore: 88.0,
							CreatedAt:    time.Now(),
						},
					},
				}, nil
			},
		}

		handler := NewProductHandler(mockUsecase)
		url := "/api/v1/products?q=earbuds&category_id=cat-1&min_price=100&max_price=1000&min_commission=15&sort_by=velocity&sort_order=asc&page=2&limit=50"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rec := httptest.NewRecorder()

		handler.GetCatalog(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		if passedFilter.Query != "earbuds" {
			t.Errorf("expected Query 'earbuds', got '%s'", passedFilter.Query)
		}
		if passedFilter.CategoryID != "cat-1" {
			t.Errorf("expected CategoryID 'cat-1', got '%s'", passedFilter.CategoryID)
		}
		if passedFilter.MinPrice == nil || *passedFilter.MinPrice != 100.0 {
			t.Errorf("expected MinPrice 100.0, got %v", passedFilter.MinPrice)
		}
		if passedFilter.MaxPrice == nil || *passedFilter.MaxPrice != 1000.0 {
			t.Errorf("expected MaxPrice 1000.0, got %v", passedFilter.MaxPrice)
		}
		if passedFilter.MinCommission == nil || *passedFilter.MinCommission != 15.0 {
			t.Errorf("expected MinCommission 15.0, got %v", passedFilter.MinCommission)
		}
		if passedFilter.SortBy != "velocity" {
			t.Errorf("expected SortBy 'velocity', got '%s'", passedFilter.SortBy)
		}
		if passedFilter.SortOrder != "asc" {
			t.Errorf("expected SortOrder 'asc', got '%s'", passedFilter.SortOrder)
		}
		if passedFilter.Page != 2 {
			t.Errorf("expected Page 2, got %d", passedFilter.Page)
		}
		if passedFilter.Limit != 50 {
			t.Errorf("expected Limit 50, got %d", passedFilter.Limit)
		}

		var resp domain.ProductCatalogResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed decoding json: %v", err)
		}

		if resp.TotalCount != 1 || len(resp.Items) != 1 {
			t.Errorf("unexpected response content: %+v", resp)
		}
	})

	t.Run("Returns 500 when usecase returns an error", func(t *testing.T) {
		mockUsecase := &mockProductUsecase{
			getCatalogFunc: func(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error) {
				return nil, errors.New("query failed")
			},
		}

		handler := NewProductHandler(mockUsecase)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
		rec := httptest.NewRecorder()

		handler.GetCatalog(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}
	})
}
