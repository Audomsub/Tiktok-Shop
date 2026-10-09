package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

var (
	ErrInvalidUserID    = errors.New("invalid or empty user_id")
	ErrInvalidProductID = errors.New("invalid or empty product_id")
)

type favoriteUsecase struct {
	repo domain.FavoriteRepository
}

// NewFavoriteUsecase constructs a new FavoriteUsecase instance
func NewFavoriteUsecase(repo domain.FavoriteRepository) domain.FavoriteUsecase {
	return &favoriteUsecase{repo: repo}
}

// AddFavorite adds a product to user's favorites
func (u *favoriteUsecase) AddFavorite(ctx context.Context, userID string, productID string) error {
	cleanUserID := strings.TrimSpace(userID)
	if cleanUserID == "" {
		return ErrInvalidUserID
	}

	cleanProductID := strings.TrimSpace(productID)
	if cleanProductID == "" {
		return ErrInvalidProductID
	}

	if err := u.repo.Add(ctx, cleanUserID, cleanProductID); err != nil {
		return fmt.Errorf("failed to add favorite: %w", err)
	}

	return nil
}

// RemoveFavorite removes a product from user's favorites
func (u *favoriteUsecase) RemoveFavorite(ctx context.Context, userID string, productID string) error {
	cleanUserID := strings.TrimSpace(userID)
	if cleanUserID == "" {
		return ErrInvalidUserID
	}

	cleanProductID := strings.TrimSpace(productID)
	if cleanProductID == "" {
		return ErrInvalidProductID
	}

	if err := u.repo.Remove(ctx, cleanUserID, cleanProductID); err != nil {
		return fmt.Errorf("failed to remove favorite: %w", err)
	}

	return nil
}

// GetFavorites retrieves all favorited products for the user with calculated returns
func (u *favoriteUsecase) GetFavorites(ctx context.Context, userID string) (*domain.FavoriteListResponse, error) {
	cleanUserID := strings.TrimSpace(userID)
	if cleanUserID == "" {
		return nil, ErrInvalidUserID
	}

	items, err := u.repo.ListByUser(ctx, cleanUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve favorites: %w", err)
	}

	if items == nil {
		items = make([]*domain.FavoriteItem, 0)
	}

	for _, item := range items {
		// Calculate Expected Return (THB per unit) = Price * (CommissionRate / 100)
		expectedReturn := item.Price * (item.CommissionRate / 100.0)
		item.ExpectedReturn = math.Round(expectedReturn*100) / 100
	}

	return &domain.FavoriteListResponse{
		Total: len(items),
		Items: items,
	}, nil
}
