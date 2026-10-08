package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresLeaderboardRepo implements domain.LeaderboardRepository
type PostgresLeaderboardRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresLeaderboardRepo constructs a new PostgresLeaderboardRepo instance
func NewPostgresLeaderboardRepo(pool *pgxpool.Pool) domain.LeaderboardRepository {
	return &PostgresLeaderboardRepo{pool: pool}
}

// GetLatestSuccessfulCrawlLog finds the most recent successfully completed crawl log
func (r *PostgresLeaderboardRepo) GetLatestSuccessfulCrawlLog(ctx context.Context) (*domain.CrawlLog, error) {
	query := `
		SELECT 
			id, crawl_round, status, total_pages_requested, total_pages_success,
			raw_products_scraped, filtered_products_saved, http_error_code,
			error_message, started_at, finished_at
		FROM crawl_logs
		WHERE status = 'SUCCESS'
		ORDER BY started_at DESC
		LIMIT 1;
	`

	log := &domain.CrawlLog{}
	err := r.pool.QueryRow(ctx, query).Scan(
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No successful crawl logs found yet
		}
		return nil, fmt.Errorf("failed to query latest successful crawl log: %w", err)
	}

	return log, nil
}

// GetTopSnapshotsByCrawlLogID retrieves the highest ranked products for a given crawl round
func (r *PostgresLeaderboardRepo) GetTopSnapshotsByCrawlLogID(ctx context.Context, crawlLogID string, limit int) ([]*domain.LeaderboardItem, error) {
	if limit <= 0 {
		limit = 10
	} else if limit > 50 {
		limit = 50
	}

	query := `
		SELECT 
			s.product_id,
			p.name,
			COALESCE(p.product_url, '') AS product_url,
			COALESCE(p.image_url, '') AS image_url,
			p.category_id,
			COALESCE(c.name, 'Uncategorized') AS category_name,
			s.price,
			s.total_sales,
			COALESCE(s.delta_sales, 0) AS delta_sales,
			COALESCE(s.velocity_per_hour, 0.00) AS velocity_per_hour,
			s.commission_rate,
			COALESCE(s.winning_score, 0.00) AS winning_score,
			s.snapshot_time
		FROM product_snapshots s
		INNER JOIN products p ON s.product_id = p.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE s.crawl_log_id = $1
		ORDER BY s.winning_score DESC
		LIMIT $2;
	`

	rows, err := r.pool.Query(ctx, query, crawlLogID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top snapshots by crawl_log_id: %w", err)
	}
	defer rows.Close()

	var items []*domain.LeaderboardItem
	for rows.Next() {
		item := &domain.LeaderboardItem{}
		err := rows.Scan(
			&item.ProductID,
			&item.Name,
			&item.ProductURL,
			&item.ImageURL,
			&item.CategoryID,
			&item.CategoryName,
			&item.Price,
			&item.TotalSales,
			&item.DeltaSales,
			&item.VelocityPerHour,
			&item.CommissionRate,
			&item.WinningScore,
			&item.SnapshotTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning leaderboard row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating leaderboard rows: %w", err)
	}

	return items, nil
}
