package database

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresProductRepo implements domain.ProductRepository
type PostgresProductRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresProductRepo constructs a new PostgresProductRepo instance
func NewPostgresProductRepo(pool *pgxpool.Pool) domain.ProductRepository {
	return &PostgresProductRepo{pool: pool}
}

// ListCatalog queries products and their latest snapshot metrics with filtering, sorting, and pagination
func (r *PostgresProductRepo) ListCatalog(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error) {
	// 1. Sanitize pagination bounds
	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	// 2. Build dynamic WHERE conditions and parameterized arguments
	var whereConditions []string
	var args []any
	argIndex := 1

	// Keyword search on product name (ILIKE case-insensitive)
	if strings.TrimSpace(filter.Query) != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("p.name ILIKE $%d", argIndex))
		args = append(args, "%"+strings.TrimSpace(filter.Query)+"%")
		argIndex++
	}

	// Category filter
	if strings.TrimSpace(filter.CategoryID) != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("p.category_id = $%d", argIndex))
		args = append(args, strings.TrimSpace(filter.CategoryID))
		argIndex++
	}

	// Price range filters
	if filter.MinPrice != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("COALESCE(snap.price, 0) >= $%d", argIndex))
		args = append(args, *filter.MinPrice)
		argIndex++
	}
	if filter.MaxPrice != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("COALESCE(snap.price, 0) <= $%d", argIndex))
		args = append(args, *filter.MaxPrice)
		argIndex++
	}

	// Minimum commission rate filter
	if filter.MinCommission != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.commission_rate >= $%d", argIndex))
		args = append(args, *filter.MinCommission)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 3. Count total matching rows
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM products p
		LEFT JOIN LATERAL (
			SELECT price, commission_rate, total_sales, delta_sales, velocity_per_hour, winning_score, snapshot_time
			FROM product_snapshots
			WHERE product_id = p.id
			ORDER BY snapshot_time DESC
			LIMIT 1
		) snap ON true
		%s;
	`, whereClause)

	var totalCount int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("failed counting catalog items: %w", err)
	}

	totalPages := 0
	if totalCount > 0 {
		totalPages = int(math.Ceil(float64(totalCount) / float64(limit)))
	}

	if totalCount == 0 {
		return &domain.ProductCatalogResponse{
			Page:       page,
			Limit:      limit,
			TotalCount: 0,
			TotalPages: 0,
			Items:      make([]*domain.ProductCatalogItem, 0),
		}, nil
	}

	// 4. Determine safe ORDER BY clause (Whitelisted columns to prevent SQL injection)
	sortColumn := "COALESCE(snap.winning_score, 0)"
	switch strings.ToLower(filter.SortBy) {
	case "velocity":
		sortColumn = "COALESCE(snap.velocity_per_hour, 0)"
	case "total_sales":
		sortColumn = "COALESCE(snap.total_sales, 0)"
	case "commission_rate":
		sortColumn = "p.commission_rate"
	case "price":
		sortColumn = "COALESCE(snap.price, 0)"
	case "created_at":
		sortColumn = "p.created_at"
	case "winning_score":
		sortColumn = "COALESCE(snap.winning_score, 0)"
	}

	sortOrder := "DESC"
	if strings.ToLower(filter.SortOrder) == "asc" {
		sortOrder = "ASC"
	}

	orderByClause := fmt.Sprintf("ORDER BY %s %s, p.created_at DESC", sortColumn, sortOrder)

	// 5. Query page items
	dataQuery := fmt.Sprintf(`
		SELECT 
			p.id,
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
			COALESCE(snap.snapshot_time, p.created_at) AS snapshot_time,
			p.created_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN LATERAL (
			SELECT price, commission_rate, total_sales, delta_sales, velocity_per_hour, winning_score, snapshot_time
			FROM product_snapshots
			WHERE product_id = p.id
			ORDER BY snapshot_time DESC
			LIMIT 1
		) snap ON true
		%s
		%s
		LIMIT $%d OFFSET $%d;
	`, whereClause, orderByClause, argIndex, argIndex+1)

	queryArgs := append(args, limit, offset)

	rows, err := r.pool.Query(ctx, dataQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed querying product catalog items: %w", err)
	}
	defer rows.Close()

	var items []*domain.ProductCatalogItem
	for rows.Next() {
		item := &domain.ProductCatalogItem{}
		err := rows.Scan(
			&item.ID,
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
			&item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning product catalog item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating product catalog rows: %w", err)
	}

	return &domain.ProductCatalogResponse{
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
		TotalPages: totalPages,
		Items:      items,
	}, nil
}
