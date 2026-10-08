package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresSnapshotRepo implements domain.SnapshotRepository
type PostgresSnapshotRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresSnapshotRepo constructs a new PostgresSnapshotRepo
func NewPostgresSnapshotRepo(pool *pgxpool.Pool) domain.SnapshotRepository {
	return &PostgresSnapshotRepo{pool: pool}
}

// GetSnapshotsByCrawlLogID retrieves all product snapshots captured in a given crawl round
func (r *PostgresSnapshotRepo) GetSnapshotsByCrawlLogID(ctx context.Context, crawlLogID string) ([]*domain.ProductSnapshot, error) {
	query := `
		SELECT 
			id, product_id, crawl_log_id, snapshot_time, price, commission_rate, 
			total_sales, delta_sales, velocity_per_hour, winning_score, created_at
		FROM product_snapshots
		WHERE crawl_log_id = $1
		ORDER BY created_at ASC;
	`

	rows, err := r.pool.Query(ctx, query, crawlLogID)
	if err != nil {
		return nil, fmt.Errorf("failed to query snapshots by crawl_log_id: %w", err)
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
			return nil, fmt.Errorf("failed to scan product snapshot row: %w", err)
		}
		snapshots = append(snapshots, s)
	}

	return snapshots, rows.Err()
}

// GetPreviousSnapshots finds the latest previous snapshot for each product within the Lookback Window
func (r *PostgresSnapshotRepo) GetPreviousSnapshots(
	ctx context.Context,
	productIDs []string,
	beforeTime time.Time,
	maxLookbackHours float64,
) (map[string]*domain.ProductSnapshot, error) {
	if len(productIDs) == 0 {
		return make(map[string]*domain.ProductSnapshot), nil
	}

	lookbackInterval := fmt.Sprintf("%f hours", maxLookbackHours)
	query := `
		SELECT DISTINCT ON (product_id) 
			id, product_id, crawl_log_id, snapshot_time, price, commission_rate, 
			total_sales, delta_sales, velocity_per_hour, winning_score, created_at
		FROM product_snapshots
		WHERE product_id = ANY($1)
		  AND snapshot_time < $2
		  AND snapshot_time >= $2 - $3::interval
		ORDER BY product_id, snapshot_time DESC;
	`

	rows, err := r.pool.Query(ctx, query, productIDs, beforeTime, lookbackInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to query previous snapshots: %w", err)
	}
	defer rows.Close()

	results := make(map[string]*domain.ProductSnapshot)
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
			return nil, fmt.Errorf("failed to scan previous snapshot row: %w", err)
		}
		results[s.ProductID] = s
	}

	return results, rows.Err()
}

// BulkUpdateCalculations updates computed metrics back into the database
func (r *PostgresSnapshotRepo) BulkUpdateCalculations(ctx context.Context, calculations []*domain.CalculatedSnapshot) error {
	// Implemented as part of Ticket 04 (Worker Pool Bulk Updater)
	return nil
}
