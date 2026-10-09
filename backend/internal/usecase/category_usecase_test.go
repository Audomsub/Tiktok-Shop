package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

type mockCategoryRepo struct {
	listAllFunc func(ctx context.Context) ([]*domain.Category, error)
}

func (m *mockCategoryRepo) ListAll(ctx context.Context) ([]*domain.Category, error) {
	if m.listAllFunc != nil {
		return m.listAllFunc(ctx)
	}
	return nil, nil
}

func TestCategoryUsecase_GetCategories(t *testing.T) {
	t.Run("Successfully returns categories", func(t *testing.T) {
		mockRepo := &mockCategoryRepo{
			listAllFunc: func(ctx context.Context) ([]*domain.Category, error) {
				return []*domain.Category{
					{ID: "c1", Name: "Beauty", Slug: "beauty", CreatedAt: time.Now()},
					{ID: "c2", Name: "Electronics", Slug: "electronics", CreatedAt: time.Now()},
				}, nil
			},
		}

		u := NewCategoryUsecase(mockRepo)
		cats, err := u.GetCategories(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cats) != 2 {
			t.Fatalf("expected 2 categories, got %d", len(cats))
		}
		if cats[0].Name != "Beauty" || cats[1].Name != "Electronics" {
			t.Errorf("unexpected categories returned: %+v", cats)
		}
	})

	t.Run("Returns empty slice when repo returns nil", func(t *testing.T) {
		mockRepo := &mockCategoryRepo{
			listAllFunc: func(ctx context.Context) ([]*domain.Category, error) {
				return nil, nil
			},
		}

		u := NewCategoryUsecase(mockRepo)
		cats, err := u.GetCategories(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cats == nil || len(cats) != 0 {
			t.Errorf("expected empty slice, got %+v", cats)
		}
	})

	t.Run("Propagates repository errors", func(t *testing.T) {
		mockRepo := &mockCategoryRepo{
			listAllFunc: func(ctx context.Context) ([]*domain.Category, error) {
				return nil, errors.New("db query failed")
			},
		}

		u := NewCategoryUsecase(mockRepo)
		_, err := u.GetCategories(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
