package domain

import (
	"context"
	"time"
)

// AddFavoriteRequest represents the request payload to add a product to bookmarks
type AddFavoriteRequest struct {
	ProductID string `json:"product_id"`
}

// FavoriteItem represents an individual bookmarked product paired with its latest market metrics
type FavoriteItem struct {
	ProductID       string    `json:"product_id"`
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
	FavoritedAt     time.Time `json:"favorited_at"`
}

// FavoriteListResponse represents the response containing bookmarked products
type FavoriteListResponse struct {
	Total int             `json:"total"`
	Items []*FavoriteItem `json:"items"`
}

// FavoriteRepository defines persistence operations for user bookmark management
type FavoriteRepository interface {
	Add(ctx context.Context, userID string, productID string) error
	Remove(ctx context.Context, userID string, productID string) error
	ListByUser(ctx context.Context, userID string) ([]*FavoriteItem, error)
	GetUserFavoriteProductIDs(ctx context.Context, userID string) (map[string]bool, error)
}

// FavoriteUsecase defines business operations for user bookmarks
type FavoriteUsecase interface {
	AddFavorite(ctx context.Context, userID string, productID string) error
	RemoveFavorite(ctx context.Context, userID string, productID string) error
	GetFavorites(ctx context.Context, userID string) (*FavoriteListResponse, error)
}
