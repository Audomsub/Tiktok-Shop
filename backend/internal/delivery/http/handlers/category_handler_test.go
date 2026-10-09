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

type mockCategoryUsecase struct {
	getCategoriesFunc func(ctx context.Context) ([]*domain.Category, error)
}

func (m *mockCategoryUsecase) GetCategories(ctx context.Context) ([]*domain.Category, error) {
	if m.getCategoriesFunc != nil {
		return m.getCategoriesFunc(ctx)
	}
	return nil, nil
}

func TestCategoryHandler_GetCategories(t *testing.T) {
	t.Run("Returns 200 with categories on success", func(t *testing.T) {
		mockUsecase := &mockCategoryUsecase{
			getCategoriesFunc: func(ctx context.Context) ([]*domain.Category, error) {
				return []*domain.Category{
					{ID: "c-1", Name: "Electronics", Slug: "electronics", CreatedAt: time.Now()},
				}, nil
			},
		}

		handler := NewCategoryHandler(mockUsecase)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
		rec := httptest.NewRecorder()

		handler.GetCategories(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp []*domain.Category
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed decoding json: %v", err)
		}

		if len(resp) != 1 || resp[0].Name != "Electronics" {
			t.Errorf("unexpected response: %+v", resp)
		}
	})

	t.Run("Returns 500 when usecase fails", func(t *testing.T) {
		mockUsecase := &mockCategoryUsecase{
			getCategoriesFunc: func(ctx context.Context) ([]*domain.Category, error) {
				return nil, errors.New("db failure")
			},
		}

		handler := NewCategoryHandler(mockUsecase)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
		rec := httptest.NewRecorder()

		handler.GetCategories(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}
	})
}
