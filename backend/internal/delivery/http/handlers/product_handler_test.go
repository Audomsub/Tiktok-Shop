package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

type mockProductUsecase struct {
	getCatalogFunc       func(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error)
	exportCatalogCSVFunc func(ctx context.Context, filter domain.ProductCatalogFilter, w io.Writer) error
}

func (m *mockProductUsecase) GetCatalog(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error) {
	if m.getCatalogFunc != nil {
		return m.getCatalogFunc(ctx, filter)
	}
	return nil, nil
}

func (m *mockProductUsecase) ExportCatalogCSV(ctx context.Context, filter domain.ProductCatalogFilter, w io.Writer) error {
	if m.exportCatalogCSVFunc != nil {
		return m.exportCatalogCSVFunc(ctx, filter, w)
	}
	return nil
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

func TestProductHandler_ExportCatalog(t *testing.T) {
	t.Run("Streams CSV with proper headers and passes filter parameters", func(t *testing.T) {
		var passedFilter domain.ProductCatalogFilter
		mockUsecase := &mockProductUsecase{
			exportCatalogCSVFunc: func(ctx context.Context, filter domain.ProductCatalogFilter, w io.Writer) error {
				passedFilter = filter
				_, _ = w.Write([]byte("mock-csv-data"))
				return nil
			},
		}

		handler := NewProductHandler(mockUsecase)
		url := "/api/v1/products/export?q=lipstick&category_id=cat-beauty&min_price=150&max_price=800&min_commission=20&sort_by=winning_score&sort_order=desc"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rec := httptest.NewRecorder()

		handler.ExportCatalog(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		contentType := rec.Header().Get("Content-Type")
		if contentType != "text/csv; charset=utf-8" {
			t.Errorf("expected Content-Type text/csv; charset=utf-8, got %q", contentType)
		}

		contentDisp := rec.Header().Get("Content-Disposition")
		expectedDisp := `attachment; filename="tiktok_winning_products.csv"`
		if contentDisp != expectedDisp {
			t.Errorf("expected Content-Disposition %q, got %q", expectedDisp, contentDisp)
		}

		if passedFilter.Query != "lipstick" {
			t.Errorf("expected query lipstick, got %q", passedFilter.Query)
		}
		if passedFilter.CategoryID != "cat-beauty" {
			t.Errorf("expected category_id cat-beauty, got %q", passedFilter.CategoryID)
		}
		if passedFilter.MinPrice == nil || *passedFilter.MinPrice != 150.0 {
			t.Errorf("expected min_price 150.0, got %v", passedFilter.MinPrice)
		}
		if passedFilter.MaxPrice == nil || *passedFilter.MaxPrice != 800.0 {
			t.Errorf("expected max_price 800.0, got %v", passedFilter.MaxPrice)
		}
		if passedFilter.MinCommission == nil || *passedFilter.MinCommission != 20.0 {
			t.Errorf("expected min_commission 20.0, got %v", passedFilter.MinCommission)
		}
		if passedFilter.SortBy != "winning_score" {
			t.Errorf("expected sort_by winning_score, got %q", passedFilter.SortBy)
		}
		if passedFilter.SortOrder != "desc" {
			t.Errorf("expected sort_order desc, got %q", passedFilter.SortOrder)
		}

		if rec.Body.String() != "mock-csv-data" {
			t.Errorf("expected body 'mock-csv-data', got %q", rec.Body.String())
		}
	})
}
