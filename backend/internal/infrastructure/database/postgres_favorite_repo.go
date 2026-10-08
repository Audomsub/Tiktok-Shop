package database

import (
	"context"
	"fmt"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresFavoriteRepo implements domain.FavoriteRepository
type PostgresFavoriteRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresFavoriteRepo constructs a new PostgresFavoriteRepo instance
func NewPostgresFavoriteRepo(pool *pgxpool.Pool) domain.FavoriteRepository {
	return &PostgresFavoriteRepo{pool: pool}
}

// Add inserts a product into the user's favorites idempotently
func (r *PostgresFavoriteRepo) Add(ctx context.Context, userID string, productID string) error {
	if userID == "" || productID == "" {
		return fmt.Errorf("user_id and product_id cannot be empty")
	}

	query := `
		INSERT INTO favorite_products (user_id, product_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, product_id) DO NOTHING;
	`

	_, err := r.pool.Exec(ctx, query, userID, productID)
	if err != nil {
		return fmt.Errorf("failed adding product to favorites: %w", err)
	}
	return nil
}

// Remove deletes a product from the user's favorites
func (r *PostgresFavoriteRepo) Remove(ctx context.Context, userID string, productID string) error {
	if userID == "" || productID == "" {
		return fmt.Errorf("user_id and product_id cannot be empty")
	}

	query := `
		DELETE FROM favorite_products
		WHERE user_id = $1 AND product_id = $2;
	`

	_, err := r.pool.Exec(ctx, query, userID, productID)
	if err != nil {
		return fmt.Errorf("failed removing product from favorites: %w", err)
	}
	return nil
}

// ListByUser retrieves all favorited products for the given user, joined with master product details and latest snapshots
func (r *PostgresFavoriteRepo) ListByUser(ctx context.Context, userID string) ([]*domain.FavoriteItem, error) {
	if userID == "" {
		return []*domain.FavoriteItem{}, nil
	}

	query := `
		SELECT 
			f.product_id,
			p.source_id,
			p.name,
			COALESCE(p.product_url, '') AS product_url,
			COALESCE(p.image_url, '') AS image_url,
			p.category_id,
			COALESCE(c.name, 'Uncategorized') AS category_name,
			COALESCE(snap.price, 0.00) AS price,
			p.commission_rate,
			COALESCE(snap.total_sales, 0) AS total_sales,
			COALESCE(snap.delta_sales, 0) AS delta_sales,
			COALESCE(snap.velocity_per_hour, 0.00) AS velocity_per_hour,
			COALESCE(snap.winning_score, 0.00) AS winning_score,
			COALESCE(snap.snapshot_time, f.created_at) AS snapshot_time,
			f.created_at AS favorited_at
		FROM favorite_products f
		INNER JOIN products p ON f.product_id = p.id
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN LATERAL (
			SELECT price, commission_rate, total_sales, delta_sales, velocity_per_hour, winning_score, snapshot_time
			FROM product_snapshots
			WHERE product_id = p.id
			ORDER BY snapshot_time DESC
			LIMIT 1
		) snap ON true
		WHERE f.user_id = $1
		ORDER BY f.created_at DESC;
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed querying user favorites: %w", err)
	}
	defer rows.Close()

	items := make([]*domain.FavoriteItem, 0)
	for rows.Next() {
		item := &domain.FavoriteItem{}
		err := rows.Scan(
			&item.ProductID,
			&item.SourceID,
			&item.Name,
			&item.ProductURL,
			&item.ImageURL,
			&item.CategoryID,
			&item.CategoryName,
			&item.Price,
			&item.CommissionRate,
			&item.TotalSales,
			&item.DeltaSales,
			&item.VelocityPerHour,
			&item.WinningScore,
			&item.SnapshotTime,
			&item.FavoritedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning favorite item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating favorite rows: %w", err)
	}

	return items, nil
}

// GetUserFavoriteProductIDs returns a set map of all product IDs favorited by the user
func (r *PostgresFavoriteRepo) GetUserFavoriteProductIDs(ctx context.Context, userID string) (map[string]bool, error) {
	favMap := make(map[string]bool)
	if userID == "" {
		return favMap, nil
	}

	query := `
		SELECT product_id
		FROM favorite_products
		WHERE user_id = $1;
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed querying user favorite product IDs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var productID string
		if err := rows.Scan(&productID); err != nil {
			return nil, fmt.Errorf("failed scanning favorite product ID: %w", err)
		}
		favMap[productID] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating favorite product ID rows: %w", err)
	}

	return favMap, nil
}
