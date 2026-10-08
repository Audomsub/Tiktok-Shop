package database

import (
	"context"
	"fmt"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresCategoryRepo implements domain.CategoryRepository
type PostgresCategoryRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresCategoryRepo constructs a new PostgresCategoryRepo instance
func NewPostgresCategoryRepo(pool *pgxpool.Pool) domain.CategoryRepository {
	return &PostgresCategoryRepo{pool: pool}
}

// ListAll retrieves all categories sorted alphabetically by name
func (r *PostgresCategoryRepo) ListAll(ctx context.Context) ([]*domain.Category, error) {
	query := `
		SELECT id, name, slug, created_at
		FROM categories
		ORDER BY name ASC;
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed querying categories: %w", err)
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		cat := &domain.Category{}
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Slug, &cat.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed scanning category row: %w", err)
		}
		categories = append(categories, cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during category rows iteration: %w", err)
	}

	return categories, nil
}
