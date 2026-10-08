package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresTrendRepo implements domain.TrendRepository
type PostgresTrendRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresTrendRepo constructs a new PostgresTrendRepo instance
func NewPostgresTrendRepo(pool *pgxpool.Pool) domain.TrendRepository {
	return &PostgresTrendRepo{pool: pool}
}

// GetProductByID retrieves a product by its UUID primary key
func (r *PostgresTrendRepo) GetProductByID(ctx context.Context, productID string) (*domain.Product, error) {
	query := `
		SELECT id, source_id, name, COALESCE(image_url, ''), COALESCE(product_url, ''), commission_rate, category_id, created_at, updated_at
		FROM products
		WHERE id = $1;
	`

	p := &domain.Product{}
	err := r.pool.QueryRow(ctx, query, productID).Scan(
		&p.ID,
		&p.SourceID,
		&p.Name,
		&p.ImageURL,
		&p.ProductURL,
		&p.CommissionRate,
		&p.CategoryID,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed querying product by id: %w", err)
	}

	return p, nil
}

// GetProductSnapshotsSince queries snapshots for a product since a given timestamp in ascending order
func (r *PostgresTrendRepo) GetProductSnapshotsSince(ctx context.Context, productID string, since time.Time) ([]*domain.ProductSnapshot, error) {
	query := `
		SELECT 
			id, product_id, crawl_log_id, snapshot_time, price, commission_rate,
			total_sales, COALESCE(delta_sales, 0), COALESCE(velocity_per_hour, 0.0),
			COALESCE(winning_score, 0.0), created_at
		FROM product_snapshots
		WHERE product_id = $1 AND snapshot_time >= $2
		ORDER BY snapshot_time ASC;
	`

	rows, err := r.pool.Query(ctx, query, productID, since)
	if err != nil {
		return nil, fmt.Errorf("failed querying snapshots for trend: %w", err)
	}
	defer rows.Close()

	var snapshots []*domain.ProductSnapshot
	for rows.Next() {
		s := &domain.ProductSnapshot{}
		err := rows.Scan(
			&s.ID,
			&s.ProductID,
			&s.CrawlLogID,
			&s.SnapshotTime,
			&s.Price,
			&s.CommissionRate,
			&s.TotalSales,
			&s.DeltaSales,
			&s.VelocityPerHour,
			&s.WinningScore,
			&s.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning snapshot row: %w", err)
		}
		snapshots = append(snapshots, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating snapshot rows: %w", err)
	}

	return snapshots, nil
}
