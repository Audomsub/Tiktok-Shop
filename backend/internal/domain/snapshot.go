package domain

import (
	"context"
	"time"
)

// ProductSnapshot represents an immutable point-in-time financial & sales capture
type ProductSnapshot struct {
	ID              string    `json:"id"`
	ProductID       string    `json:"product_id"`
	CrawlLogID      *string   `json:"crawl_log_id,omitempty"`
	SnapshotTime    time.Time `json:"snapshot_time"`
	Price           float64   `json:"price"`
	CommissionRate  float64   `json:"commission_rate"`
	TotalSales      int       `json:"total_sales"`
	DeltaSales      int       `json:"delta_sales"`
	VelocityPerHour float64   `json:"velocity_per_hour"`
	WinningScore    float64   `json:"winning_score"`
	CreatedAt       time.Time `json:"created_at"`
}

// CalculatedSnapshot carries computed delta, velocity, and winning score updates
type CalculatedSnapshot struct {
	SnapshotID      string
	DeltaSales      int
	VelocityPerHour float64
	WinningScore    float64
}

// SnapshotRepository defines database persistence operations for product snapshots
type SnapshotRepository interface {
	GetSnapshotsByCrawlLogID(ctx context.Context, crawlLogID string) ([]*ProductSnapshot, error)
	GetPreviousSnapshots(ctx context.Context, productIDs []string, beforeTime time.Time, maxLookbackHours float64) (map[string]*ProductSnapshot, error)
	BulkUpdateCalculations(ctx context.Context, calculations []*CalculatedSnapshot) error
}

// VelocityUsecase defines the business logic contract for sales velocity tracking
type VelocityUsecase interface {
	ComputeVelocity(current *ProductSnapshot, previous *ProductSnapshot) (deltaSales int, velocity float64)
}
