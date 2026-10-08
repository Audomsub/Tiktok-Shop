package domain

import (
	"context"
	"time"
)

// Product represents the master catalog item from TikTok Shop / FastMoss
type Product struct {
	ID             string    `json:"id"`
	SourceID       string    `json:"source_id"`
	Name           string    `json:"name"`
	ImageURL       string    `json:"image_url"`
	ProductURL     string    `json:"product_url"`
	CommissionRate float64   `json:"commission_rate"`
	CategoryID     *string   `json:"category_id,omitempty"`
	CategoryName   string    `json:"category_name,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ProductCatalogFilter contains parameters for searching and filtering the product catalog
type ProductCatalogFilter struct {
	Query         string   `json:"q"`
	CategoryID    string   `json:"category_id"`
	MinPrice      *float64 `json:"min_price"`
	MaxPrice      *float64 `json:"max_price"`
	MinCommission *float64 `json:"min_commission"`
	SortBy        string   `json:"sort_by"`    // winning_score, velocity, total_sales, commission_rate, price, created_at
	SortOrder     string   `json:"sort_order"` // asc, desc
	Page          int      `json:"page"`
	Limit         int      `json:"limit"`
}

// ProductCatalogItem represents an individual catalog item paired with its latest snapshot analytics
type ProductCatalogItem struct {
	ID              string    `json:"id"`
	SourceID        string    `json:"source_id"`
	Name            string    `json:"name"`
	ProductURL      string    `json:"product_url"`
	ImageURL        string    `json:"image_url"`
	CategoryID      *string   `json:"category_id,omitempty"`
	CategoryName    string    `json:"category_name"`
	Price           float64   `json:"price"`
	CommissionRate  float64   `json:"commission_rate"`
	TotalSales      int       `json:"total_sales"`
	DeltaSales      int       `json:"delta_sales"`
	VelocityPerHour float64   `json:"velocity_per_hour"`
	ExpectedReturn  float64   `json:"expected_return"`
	WinningScore    float64   `json:"winning_score"`
	SnapshotTime    time.Time `json:"snapshot_time"`
	CreatedAt       time.Time `json:"created_at"`
}

// ProductCatalogResponse represents the paginated result from product catalog search
type ProductCatalogResponse struct {
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	TotalCount int                   `json:"total_count"`
	TotalPages int                   `json:"total_pages"`
	Items      []*ProductCatalogItem `json:"items"`
}

// ProductRepository defines persistence operations for product catalog retrieval
type ProductRepository interface {
	ListCatalog(ctx context.Context, filter ProductCatalogFilter) (*ProductCatalogResponse, error)
}

// ProductUsecase defines business operations for product catalog
type ProductUsecase interface {
	GetCatalog(ctx context.Context, filter ProductCatalogFilter) (*ProductCatalogResponse, error)
}
