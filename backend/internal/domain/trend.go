package domain

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrProductNotFound indicates a requested product does not exist in the database
	ErrProductNotFound = errors.New("product not found")
)

// TrendPoint represents a single historical snapshot point for trend chart visualization
type TrendPoint struct {
	SnapshotTime    time.Time `json:"snapshot_time"`
	Price           float64   `json:"price"`
	CommissionRate  float64   `json:"commission_rate"`
	TotalSales      int       `json:"total_sales"`
	DeltaSales      int       `json:"delta_sales"`
	VelocityPerHour float64   `json:"velocity_per_hour"`
	ExpectedReturn  float64   `json:"expected_return"`
	WinningScore    float64   `json:"winning_score"`
}

// ProductTrendsResponse represents the API response for GET /api/v1/products/:id/trends
type ProductTrendsResponse struct {
	ProductID   string        `json:"product_id"`
	ProductName string        `json:"product_name"`
	Days        int           `json:"days"`
	TotalPoints int           `json:"total_points"`
	Points      []*TrendPoint `json:"points"`
}

// TrendRepository defines database operations for product trends
type TrendRepository interface {
	GetProductByID(ctx context.Context, productID string) (*Product, error)
	GetProductSnapshotsSince(ctx context.Context, productID string, since time.Time) ([]*ProductSnapshot, error)
}

// TrendUsecase defines business operations for product trends
type TrendUsecase interface {
	GetProductTrends(ctx context.Context, productID string, days int) (*ProductTrendsResponse, error)
}
