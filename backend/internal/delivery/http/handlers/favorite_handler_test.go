package handlers

import (
	"bytes"
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

type mockFavoriteUsecase struct {
	addFunc    func(ctx context.Context, userID string, productID string) error
	removeFunc func(ctx context.Context, userID string, productID string) error
	getFunc    func(ctx context.Context, userID string) (*domain.FavoriteListResponse, error)
}

func (m *mockFavoriteUsecase) AddFavorite(ctx context.Context, userID string, productID string) error {
	if m.addFunc != nil {
		return m.addFunc(ctx, userID, productID)
	}
	return nil
}

func (m *mockFavoriteUsecase) RemoveFavorite(ctx context.Context, userID string, productID string) error {
	if m.removeFunc != nil {
		return m.removeFunc(ctx, userID, productID)
	}
	return nil
}

func (m *mockFavoriteUsecase) GetFavorites(ctx context.Context, userID string) (*domain.FavoriteListResponse, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, userID)
	}
	return &domain.FavoriteListResponse{Total: 0, Items: []*domain.FavoriteItem{}}, nil
}

func TestFavoriteHandler_AuthCheck(t *testing.T) {
	uc := &mockFavoriteUsecase{}
	h := NewFavoriteHandler(uc)

	t.Run("AddFavorite without auth returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/favorites", bytes.NewBufferString(`{"product_id":"p1"}`))
		rec := httptest.NewRecorder()
		h.AddFavorite(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("RemoveFavorite without auth returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/favorites/p1", nil)
		rec := httptest.NewRecorder()
		h.RemoveFavorite(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("GetFavorites without auth returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/favorites", nil)
		rec := httptest.NewRecorder()
		h.GetFavorites(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})
}

func TestFavoriteHandler_AddFavorite(t *testing.T) {
	t.Run("Success adding favorite", func(t *testing.T) {
		var receivedUser, receivedProduct string
		uc := &mockFavoriteUsecase{
			addFunc: func(ctx context.Context, userID string, productID string) error {
				receivedUser = userID
				receivedProduct = productID
				return nil
			},
		}
		h := NewFavoriteHandler(uc)

		reqBody := `{"product_id":"123e4567-e89b-12d3-a456-426614174000"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/favorites", bytes.NewBufferString(reqBody))
		ctx := context.WithValue(req.Context(), domain.UserIDContextKey, "user-123")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		h.AddFavorite(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		if receivedUser != "user-123" || receivedProduct != "123e4567-e89b-12d3-a456-426614174000" {
			t.Fatalf("handler passed unexpected params: user=%s, product=%s", receivedUser, receivedProduct)
		}
	})

	t.Run("Bad Request on empty product_id", func(t *testing.T) {
		h := NewFavoriteHandler(&mockFavoriteUsecase{})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/favorites", bytes.NewBufferString(`{"product_id":""}`))
		ctx := context.WithValue(req.Context(), domain.UserIDContextKey, "user-123")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		h.AddFavorite(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("Internal server error when usecase fails", func(t *testing.T) {
		uc := &mockFavoriteUsecase{
			addFunc: func(ctx context.Context, userID string, productID string) error {
				return errors.New("db failure")
			},
		}
		h := NewFavoriteHandler(uc)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/favorites", bytes.NewBufferString(`{"product_id":"p1"}`))
		ctx := context.WithValue(req.Context(), domain.UserIDContextKey, "user-123")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		h.AddFavorite(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}
	})
}

func TestFavoriteHandler_RemoveFavorite(t *testing.T) {
	t.Run("Success removing favorite with Chi URLParam", func(t *testing.T) {
		var receivedUser, receivedProduct string
		uc := &mockFavoriteUsecase{
			removeFunc: func(ctx context.Context, userID string, productID string) error {
				receivedUser = userID
				receivedProduct = productID
				return nil
			},
		}
		h := NewFavoriteHandler(uc)

		r := chi.NewRouter()
		r.Delete("/api/v1/favorites/{productId}", h.RemoveFavorite)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/favorites/target-p99", nil)
		ctx := context.WithValue(req.Context(), domain.UserIDContextKey, "user-123")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		if receivedUser != "user-123" || receivedProduct != "target-p99" {
			t.Fatalf("handler passed unexpected params: user=%s, product=%s", receivedUser, receivedProduct)
		}
	})
}

func TestFavoriteHandler_GetFavorites(t *testing.T) {
	t.Run("Success retrieving user favorites", func(t *testing.T) {
		uc := &mockFavoriteUsecase{
			getFunc: func(ctx context.Context, userID string) (*domain.FavoriteListResponse, error) {
				return &domain.FavoriteListResponse{
					Total: 1,
					Items: []*domain.FavoriteItem{
						{
							ProductID:      "p1",
							Name:           "Bookmarked Lipstick",
							WinningScore:   88.5,
							ExpectedReturn: 45.0,
							FavoritedAt:    time.Now(),
						},
					},
				}, nil
			},
		}
		h := NewFavoriteHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/favorites", nil)
		ctx := context.WithValue(req.Context(), domain.UserIDContextKey, "user-123")
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		h.GetFavorites(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp domain.FavoriteListResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed decoding response: %v", err)
		}

		if resp.Total != 1 || len(resp.Items) != 1 || resp.Items[0].ProductID != "p1" {
			t.Fatalf("unexpected response payload: %+v", resp)
		}
	})
}
