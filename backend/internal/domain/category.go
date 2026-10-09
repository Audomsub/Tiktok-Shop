package domain

import (
	"context"
	"time"
)

// Category represents a product category taxonomy
type Category struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

// CategoryRepository defines data access methods for categories
type CategoryRepository interface {
	ListAll(ctx context.Context) ([]*Category, error)
}

// CategoryUsecase handles business logic for category retrieval
type CategoryUsecase interface {
	GetCategories(ctx context.Context) ([]*Category, error)
}
