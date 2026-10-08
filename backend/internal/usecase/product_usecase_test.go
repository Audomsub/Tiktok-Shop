package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

type mockProductRepo struct {
	listCatalogFunc func(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error)
}

func (m *mockProductRepo) ListCatalog(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error) {
	if m.listCatalogFunc != nil {
		return m.listCatalogFunc(ctx, filter)
	}
	return nil, nil
}

func TestProductUsecase_GetCatalog(t *testing.T) {
	t.Run("Successfully returns catalog with expected return calculation", func(t *testing.T) {
		mockRepo := &mockProductRepo{
			listCatalogFunc: func(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error) {
				return &domain.ProductCatalogResponse{
					Page:       filter.Page,
					Limit:      filter.Limit,
					TotalCount: 1,
					TotalPages: 1,
					Items: []*domain.ProductCatalogItem{
						{
							ID:             "p1",
							Name:           "Bluetooth Speaker",
							Price:          800.0,
							CommissionRate: 15.0,
							WinningScore:   85.0,
							CreatedAt:      time.Now(),
						},
					},
				}, nil
			},
		}

		u := NewProductUsecase(mockRepo)
		resp, err := u.GetCatalog(context.Background(), domain.ProductCatalogFilter{Page: 1, Limit: 20})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.TotalCount != 1 {
			t.Fatalf("expected TotalCount 1, got %d", resp.TotalCount)
		}
		if len(resp.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(resp.Items))
		}

		item := resp.Items[0]
		expected := 120.0 // 800 * 0.15
		if item.ExpectedReturn != expected {
			t.Errorf("expected expected return %f, got %f", expected, item.ExpectedReturn)
		}
	})

	t.Run("Sanitizes invalid pagination parameters", func(t *testing.T) {
		var receivedFilter domain.ProductCatalogFilter
		mockRepo := &mockProductRepo{
			listCatalogFunc: func(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error) {
				receivedFilter = filter
				return &domain.ProductCatalogResponse{TotalCount: 0, Items: []*domain.ProductCatalogItem{}}, nil
			},
		}

		u := NewProductUsecase(mockRepo)
		_, err := u.GetCatalog(context.Background(), domain.ProductCatalogFilter{Page: -5, Limit: 500})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if receivedFilter.Page != 1 {
			t.Errorf("expected page 1, got %d", receivedFilter.Page)
		}
		if receivedFilter.Limit != 100 {
			t.Errorf("expected limit 100, got %d", receivedFilter.Limit)
		}
	})

	t.Run("Propagates database errors", func(t *testing.T) {
		mockRepo := &mockProductRepo{
			listCatalogFunc: func(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error) {
				return nil, errors.New("db query failed")
			},
		}

		u := NewProductUsecase(mockRepo)
		_, err := u.GetCatalog(context.Background(), domain.ProductCatalogFilter{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
