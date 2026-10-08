package database

import (
	"context"
	"fmt"
	"strings"
	"sync"
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

// BulkUpdateCalculations persists computed delta sales, velocity, and scores using a worker pool and chunked bulk SQL
func (r *PostgresSnapshotRepo) BulkUpdateCalculations(ctx context.Context, calculations []*domain.CalculatedSnapshot) error {
	if len(calculations) == 0 {
		return nil
	}

	const chunkSize = 100
	const numWorkers = 5

	// 1. Partition calculations into chunks
	var chunks [][]*domain.CalculatedSnapshot
	for i := 0; i < len(calculations); i += chunkSize {
		end := i + chunkSize
		if end > len(calculations) {
			end = len(calculations)
		}
		chunks = append(chunks, calculations[i:end])
	}

	// 2. Set up Worker Pool
	chunkChan := make(chan []*domain.CalculatedSnapshot, len(chunks))
	errChan := make(chan error, len(chunks))

	for _, chunk := range chunks {
		chunkChan <- chunk
	}
	close(chunkChan)

	var wg sync.WaitGroup
	workers := numWorkers
	if len(chunks) < workers {
		workers = len(chunks)
	}

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chunk := range chunkChan {
				if err := r.executeChunkUpdate(ctx, chunk); err != nil {
					errChan <- err
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errChan)

	// Return first error encountered if any
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// executeChunkUpdate builds and executes a single bulk UPDATE ... FROM (VALUES ...) statement
func (r *PostgresSnapshotRepo) executeChunkUpdate(ctx context.Context, chunk []*domain.CalculatedSnapshot) error {
	if len(chunk) == 0 {
		return nil
	}

	// Dynamically build VALUES ($1::uuid, $2::int, $3::numeric, $4::numeric), ($5::uuid, ...)
	var valClauses []string
	var args []interface{}
	argIdx := 1

	for _, item := range chunk {
		valClauses = append(valClauses, fmt.Sprintf("($%d::uuid, $%d::int, $%d::numeric, $%d::numeric)", argIdx, argIdx+1, argIdx+2, argIdx+3))
		args = append(args, item.SnapshotID, item.DeltaSales, item.VelocityPerHour, item.WinningScore)
		argIdx += 4
	}

	query := fmt.Sprintf(`
		UPDATE product_snapshots AS s
		SET delta_sales = u.delta_sales,
		    velocity_per_hour = u.velocity_per_hour,
		    winning_score = u.winning_score
		FROM (VALUES %s) AS u(id, delta_sales, velocity_per_hour, winning_score)
		WHERE s.id = u.id;
	`, strings.Join(valClauses, ", "))

	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed executing chunk bulk update: %w", err)
	}

	return nil
}

