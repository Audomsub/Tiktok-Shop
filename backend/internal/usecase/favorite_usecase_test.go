package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

// inMemoryFavoriteRepo simulates database persistence with User-level isolation
type inMemoryFavoriteRepo struct {
	mu        sync.RWMutex
	favorites map[string]map[string]*domain.FavoriteItem // userID -> productID -> FavoriteItem
}

func newInMemoryFavoriteRepo() *inMemoryFavoriteRepo {
	return &inMemoryFavoriteRepo{
		favorites: make(map[string]map[string]*domain.FavoriteItem),
	}
}

func (m *inMemoryFavoriteRepo) Add(ctx context.Context, userID string, productID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	userFavs, ok := m.favorites[userID]
	if !ok {
		userFavs = make(map[string]*domain.FavoriteItem)
		m.favorites[userID] = userFavs
	}

	// Idempotent: ON CONFLICT DO NOTHING
	if _, exists := userFavs[productID]; !exists {
		userFavs[productID] = &domain.FavoriteItem{
			ProductID:       productID,
			Name:            "Product " + productID,
			Price:           200.0,
			CommissionRate:  10.0,
			WinningScore:    85.0,
			SnapshotTime:    time.Now(),
			FavoritedAt:     time.Now(),
		}
	}
	return nil
}

func (m *inMemoryFavoriteRepo) Remove(ctx context.Context, userID string, productID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if userFavs, ok := m.favorites[userID]; ok {
		delete(userFavs, productID)
	}
	return nil
}

func (m *inMemoryFavoriteRepo) ListByUser(ctx context.Context, userID string) ([]*domain.FavoriteItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userFavs, ok := m.favorites[userID]
	if !ok {
		return []*domain.FavoriteItem{}, nil
	}

	var items []*domain.FavoriteItem
	for _, item := range userFavs {
		copyItem := *item
		items = append(items, &copyItem)
	}
	return items, nil
}

func (m *inMemoryFavoriteRepo) GetUserFavoriteProductIDs(ctx context.Context, userID string) (map[string]bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]bool)
	if userFavs, ok := m.favorites[userID]; ok {
		for pID := range userFavs {
			result[pID] = true
		}
	}
	return result, nil
}

func TestFavoriteUsecase_Validation(t *testing.T) {
	repo := newInMemoryFavoriteRepo()
	uc := NewFavoriteUsecase(repo)
	ctx := context.Background()

	t.Run("AddFavorite fails on empty userID", func(t *testing.T) {
		err := uc.AddFavorite(ctx, "", "prod-1")
		if !errors.Is(err, ErrInvalidUserID) {
			t.Fatalf("expected ErrInvalidUserID, got %v", err)
		}
	})

	t.Run("AddFavorite fails on empty productID", func(t *testing.T) {
		err := uc.AddFavorite(ctx, "user-1", "   ")
		if !errors.Is(err, ErrInvalidProductID) {
			t.Fatalf("expected ErrInvalidProductID, got %v", err)
		}
	})

	t.Run("RemoveFavorite fails on empty userID", func(t *testing.T) {
		err := uc.RemoveFavorite(ctx, "", "prod-1")
		if !errors.Is(err, ErrInvalidUserID) {
			t.Fatalf("expected ErrInvalidUserID, got %v", err)
		}
	})

	t.Run("RemoveFavorite fails on empty productID", func(t *testing.T) {
		err := uc.RemoveFavorite(ctx, "user-1", "")
		if !errors.Is(err, ErrInvalidProductID) {
			t.Fatalf("expected ErrInvalidProductID, got %v", err)
		}
	})

	t.Run("GetFavorites fails on empty userID", func(t *testing.T) {
		_, err := uc.GetFavorites(ctx, "  ")
		if !errors.Is(err, ErrInvalidUserID) {
			t.Fatalf("expected ErrInvalidUserID, got %v", err)
		}
	})
}

func TestFavoriteUsecase_Calculations(t *testing.T) {
	repo := newInMemoryFavoriteRepo()
	uc := NewFavoriteUsecase(repo)
	ctx := context.Background()

	user := "user-alice"
	product := "prod-100"

	// Add product
	if err := uc.AddFavorite(ctx, user, product); err != nil {
		t.Fatalf("unexpected error adding favorite: %v", err)
	}

	// Retrieve and verify ExpectedReturn calculation (Price 200 * CommissionRate 10% = 20.00)
	resp, err := uc.GetFavorites(ctx, user)
	if err != nil {
		t.Fatalf("unexpected error getting favorites: %v", err)
	}

	if resp.Total != 1 {
		t.Fatalf("expected 1 favorite item, got %d", resp.Total)
	}

	if resp.Items[0].ExpectedReturn != 20.00 {
		t.Fatalf("expected ExpectedReturn to be 20.00, got %f", resp.Items[0].ExpectedReturn)
	}
}

func TestFavoriteUsecase_UserIsolation(t *testing.T) {
	repo := newInMemoryFavoriteRepo()
	uc := NewFavoriteUsecase(repo)
	ctx := context.Background()

	userA := "user-a-111"
	userB := "user-b-222"
	prod1 := "prod-secret-1"
	prod2 := "prod-common-2"

	// 1. User A favorites prod1 and prod2
	if err := uc.AddFavorite(ctx, userA, prod1); err != nil {
		t.Fatalf("failed User A add prod1: %v", err)
	}
	if err := uc.AddFavorite(ctx, userA, prod2); err != nil {
		t.Fatalf("failed User A add prod2: %v", err)
	}

	// 2. User B currently has no favorites
	respB, err := uc.GetFavorites(ctx, userB)
	if err != nil {
		t.Fatalf("failed User B get favorites: %v", err)
	}
	if respB.Total != 0 || len(respB.Items) != 0 {
		t.Fatalf("User B should not see User A's favorites, got total %d", respB.Total)
	}

	// 3. User B tries to delete User A's favorite (prod1)
	if err := uc.RemoveFavorite(ctx, userB, prod1); err != nil {
		t.Fatalf("unexpected error during User B remove: %v", err)
	}

	// User A's favorite should NOT be affected
	respA, err := uc.GetFavorites(ctx, userA)
	if err != nil {
		t.Fatalf("failed User A get favorites: %v", err)
	}
	if respA.Total != 2 {
		t.Fatalf("User A's favorites should remain 2 after User B delete attempt, got %d", respA.Total)
	}

	// 4. Verify idempotent AddFavorite (adding prod1 again for User A does not duplicate)
	if err := uc.AddFavorite(ctx, userA, prod1); err != nil {
		t.Fatalf("failed idempotent AddFavorite: %v", err)
	}
	respAAfterIdempotent, _ := uc.GetFavorites(ctx, userA)
	if respAAfterIdempotent.Total != 2 {
		t.Fatalf("expected still 2 items after duplicate add, got %d", respAAfterIdempotent.Total)
	}

	// 5. User A removes prod1
	if err := uc.RemoveFavorite(ctx, userA, prod1); err != nil {
		t.Fatalf("failed User A remove prod1: %v", err)
	}
	respAFinal, _ := uc.GetFavorites(ctx, userA)
	if respAFinal.Total != 1 {
		t.Fatalf("expected 1 item after User A delete, got %d", respAFinal.Total)
	}
	if respAFinal.Items[0].ProductID != prod2 {
		t.Fatalf("expected remaining item to be prod2, got %s", respAFinal.Items[0].ProductID)
	}
}
