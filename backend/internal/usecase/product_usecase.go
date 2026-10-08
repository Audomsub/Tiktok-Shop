package usecase

import (
	"context"
	"fmt"
	"math"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

type productUsecase struct {
	repo domain.ProductRepository
}

// NewProductUsecase constructs a new ProductUsecase instance
func NewProductUsecase(repo domain.ProductRepository) domain.ProductUsecase {
	return &productUsecase{repo: repo}
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

	// Calculate Expected Return (THB per unit) = Price * (CommissionRate / 100)
	for _, item := range result.Items {
		expectedReturn := item.Price * (item.CommissionRate / 100.0)
		item.ExpectedReturn = math.Round(expectedReturn*100) / 100
	}

	return result, nil
}
