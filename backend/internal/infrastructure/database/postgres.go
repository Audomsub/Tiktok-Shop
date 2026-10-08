package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresHealthRepo implements domain.HealthRepository using pgxpool
type PostgresHealthRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresHealthRepo creates a new instance of PostgresHealthRepo
func NewPostgresHealthRepo(pool *pgxpool.Pool) domain.HealthRepository {
	return &PostgresHealthRepo{pool: pool}
}

// Ping checks if the PostgreSQL database is reachable and accepting queries
func (r *PostgresHealthRepo) Ping(ctx context.Context) error {
	if r.pool == nil {
		return fmt.Errorf("database connection pool is not initialized")
	}
	return r.pool.Ping(ctx)
}

// NewPostgresPool initializes a production-tuned pgx connection pool
func NewPostgresPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database configuration: %w", err)
	}

	// Performance & Connection Pool Tuning (Supabase Best Practices)
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnIdleTime = 5 * time.Minute
	config.MaxConnLifetime = 1 * time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection pool: %w", err)
	}

	// Ping database on startup to confirm credentials and network connectivity
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to connect to database during startup ping: %w", err)
	}

	return pool, nil
}
