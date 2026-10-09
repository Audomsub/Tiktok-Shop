package usecase

import (
	"context"
	"fmt"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

type categoryUsecase struct {
	repo domain.CategoryRepository
}

// NewCategoryUsecase constructs a new CategoryUsecase instance
func NewCategoryUsecase(repo domain.CategoryRepository) domain.CategoryUsecase {
	return &categoryUsecase{repo: repo}
}

// GetCategories returns all product categories
func (u *categoryUsecase) GetCategories(ctx context.Context) ([]*domain.Category, error) {
	categories, err := u.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving categories: %w", err)
	}
	if categories == nil {
		categories = make([]*domain.Category, 0)
	}
	return categories, nil
}
