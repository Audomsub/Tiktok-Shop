package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresCrawlLogRepo implements domain.CrawlLogRepository
type PostgresCrawlLogRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresCrawlLogRepo creates an instance of PostgresCrawlLogRepo
func NewPostgresCrawlLogRepo(pool *pgxpool.Pool) domain.CrawlLogRepository {
	return &PostgresCrawlLogRepo{pool: pool}
}

// GetByID retrieves a crawl log entry by its primary key UUID
func (r *PostgresCrawlLogRepo) GetByID(ctx context.Context, id string) (*domain.CrawlLog, error) {
	query := `
		SELECT 
			id, crawl_round, status, total_pages_requested, total_pages_success,
			raw_products_scraped, filtered_products_saved, http_error_code,
			error_message, started_at, finished_at
		FROM crawl_logs
		WHERE id = $1;
	`

	log := &domain.CrawlLog{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&log.ID,
		&log.CrawlRound,
		&log.Status,
		&log.TotalPagesRequested,
		&log.TotalPagesSuccess,
		&log.RawProductsScraped,
		&log.FilteredProductsSaved,
		&log.HTTPErrorCode,
		&log.ErrorMessage,
		&log.StartedAt,
		&log.FinishedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve crawl log %s: %w", id, err)
	}

	return log, nil
}

// UpdateStatus updates the final audit status and completion timestamp of a crawl round
func (r *PostgresCrawlLogRepo) UpdateStatus(
	ctx context.Context,
	id string,
	status domain.CrawlStatus,
	finishedAt time.Time,
	errMsg *string,
) error {
	query := `
		UPDATE crawl_logs
		SET status = $2,
		    finished_at = $3,
		    error_message = $4
		WHERE id = $1;
	`

	_, err := r.pool.Exec(ctx, query, id, string(status), finishedAt, errMsg)
	if err != nil {
		return fmt.Errorf("failed to update crawl log %s status: %w", id, err)
	}

	return nil
}
