package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

type mockTrendRepo struct {
	getProductByIDFunc           func(ctx context.Context, productID string) (*domain.Product, error)
	getProductSnapshotsSinceFunc func(ctx context.Context, productID string, since time.Time) ([]*domain.ProductSnapshot, error)
}

func (m *mockTrendRepo) GetProductByID(ctx context.Context, productID string) (*domain.Product, error) {
	if m.getProductByIDFunc != nil {
		return m.getProductByIDFunc(ctx, productID)
	}
	return nil, nil
}

func (m *mockTrendRepo) GetProductSnapshotsSince(ctx context.Context, productID string, since time.Time) ([]*domain.ProductSnapshot, error) {
	if m.getProductSnapshotsSinceFunc != nil {
		return m.getProductSnapshotsSinceFunc(ctx, productID, since)
	}
	return nil, nil
}

func TestTrendUsecase_GetProductTrends(t *testing.T) {
	now := time.Now()

	t.Run("Successfully returns chronological trend points and expected returns", func(t *testing.T) {
		mockRepo := &mockTrendRepo{
			getProductByIDFunc: func(ctx context.Context, productID string) (*domain.Product, error) {
				return &domain.Product{
					ID:   "p-1",
					Name: "Wireless Earbuds",
				}, nil
			},
			getProductSnapshotsSinceFunc: func(ctx context.Context, productID string, since time.Time) ([]*domain.ProductSnapshot, error) {
				return []*domain.ProductSnapshot{
					{
						SnapshotTime:    now.Add(-48 * time.Hour),
						Price:           500.0,
						CommissionRate:  20.0,
						TotalSales:      1000,
						DeltaSales:      50,
						VelocityPerHour: 8.3,
						WinningScore:    75.0,
					},
					{
						SnapshotTime:    now.Add(-24 * time.Hour),
						Price:           500.0,
						CommissionRate:  20.0,
						TotalSales:      1200,
						DeltaSales:      200,
						VelocityPerHour: 15.0,
						WinningScore:    92.0,
					},
				}, nil
			},
		}

		u := NewTrendUsecase(mockRepo)
		resp, err := u.GetProductTrends(context.Background(), "p-1", 7)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.ProductID != "p-1" {
			t.Errorf("expected product ID 'p-1', got '%s'", resp.ProductID)
		}
		if resp.ProductName != "Wireless Earbuds" {
			t.Errorf("expected product name 'Wireless Earbuds', got '%s'", resp.ProductName)
		}
		if resp.Days != 7 {
			t.Errorf("expected 7 days, got %d", resp.Days)
		}
		if resp.TotalPoints != 2 {
			t.Fatalf("expected 2 points, got %d", resp.TotalPoints)
		}

		point1 := resp.Points[0]
		if point1.ExpectedReturn != 100.0 {
			t.Errorf("point1 expected return: expected 100.0, got %f", point1.ExpectedReturn)
		}

		point2 := resp.Points[1]
		if point2.ExpectedReturn != 100.0 {
			t.Errorf("point2 expected return: expected 100.0, got %f", point2.ExpectedReturn)
		}
	})

	t.Run("Returns ErrProductNotFound when product does not exist", func(t *testing.T) {
		mockRepo := &mockTrendRepo{
			getProductByIDFunc: func(ctx context.Context, productID string) (*domain.Product, error) {
				return nil, nil // Not found
			},
		}

		u := NewTrendUsecase(mockRepo)
		_, err := u.GetProductTrends(context.Background(), "p-non-existent", 7)
		if !errors.Is(err, domain.ErrProductNotFound) {
			t.Fatalf("expected ErrProductNotFound, got %v", err)
		}
	})

	t.Run("Sanitizes invalid days parameter", func(t *testing.T) {
		mockRepo := &mockTrendRepo{
			getProductByIDFunc: func(ctx context.Context, productID string) (*domain.Product, error) {
				return &domain.Product{ID: "p-1", Name: "Test Product"}, nil
			},
			getProductSnapshotsSinceFunc: func(ctx context.Context, productID string, since time.Time) ([]*domain.ProductSnapshot, error) {
				return []*domain.ProductSnapshot{}, nil
			},
		}

		u := NewTrendUsecase(mockRepo)
		// Test days <= 0 defaults to 7
		resp1, err := u.GetProductTrends(context.Background(), "p-1", -5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp1.Days != 7 {
			t.Errorf("expected default 7 days, got %d", resp1.Days)
		}

		// Test days > 30 clamped to 30
		resp2, err := u.GetProductTrends(context.Background(), "p-1", 100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp2.Days != 30 {
			t.Errorf("expected clamped 30 days, got %d", resp2.Days)
		}
	})
}
