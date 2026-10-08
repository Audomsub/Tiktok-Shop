package usecase

import (
	"context"
	"fmt"
	"math"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

type productUsecase struct {
	repo    domain.ProductRepository
	favRepo domain.FavoriteRepository
}

// NewProductUsecase constructs a new ProductUsecase instance
func NewProductUsecase(repo domain.ProductRepository, favRepo ...domain.FavoriteRepository) domain.ProductUsecase {
	var fRepo domain.FavoriteRepository
	if len(favRepo) > 0 {
		fRepo = favRepo[0]
	}
	return &productUsecase{repo: repo, favRepo: fRepo}
}

// GetCatalog queries the catalog using provided filters and calculates expected return per item
func (u *productUsecase) GetCatalog(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error) {
	// Apply default values for pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 20
	} else if filter.Limit > 100 {
		filter.Limit = 100
	}

	result, err := u.repo.ListCatalog(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving product catalog: %w", err)
	}

	if result == nil {
		return &domain.ProductCatalogResponse{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: 0,
			TotalPages: 0,
			Items:      make([]*domain.ProductCatalogItem, 0),
		}, nil
	}

	// Fetch user's favorite product IDs if user is authenticated and favRepo is configured
	var favMap map[string]bool
	if u.favRepo != nil {
		if userID, ok := domain.GetUserIDFromContext(ctx); ok && userID != "" {
			if m, err := u.favRepo.GetUserFavoriteProductIDs(ctx, userID); err == nil {
				favMap = m
			}
		}
	}

	// Calculate Expected Return (THB per unit) = Price * (CommissionRate / 100) and enrich IsFavorited
	for _, item := range result.Items {
		expectedReturn := item.Price * (item.CommissionRate / 100.0)
		item.ExpectedReturn = math.Round(expectedReturn*100) / 100

		if favMap != nil && favMap[item.ID] {
			item.IsFavorited = true
		}
	}

	return result, nil
}
