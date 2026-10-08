package domain

import (
	"context"
	"time"
)

// LeaderboardBadge represents dynamic analytical badges awarded to winning products
type LeaderboardBadge struct {
	IsViralSurge     bool `json:"is_viral_surge"`     // 🔥 Velocity >= 10 pcs/hr or top 15%
	IsHighCommission bool `json:"is_high_commission"` // 💎 Commission Rate >= 20%
	IsHighYield      bool `json:"is_high_yield"`      // 💰 Expected Return >= 100 THB
	IsWinningPick    bool `json:"is_winning_pick"`    // 🏆 Winning Score >= 80
}

// LeaderboardItem represents an individual enriched product in the winning leaderboard
type LeaderboardItem struct {
	Rank            int              `json:"rank"`
	ProductID       string           `json:"product_id"`
	Name            string           `json:"name"`
	ProductURL      string           `json:"product_url"`
	ImageURL        string           `json:"image_url"`
	CategoryID      *string          `json:"category_id,omitempty"`
	CategoryName    string           `json:"category_name"`
	Price           float64          `json:"price"`
	TotalSales      int              `json:"total_sales"`
	DeltaSales      int              `json:"delta_sales"`
	VelocityPerHour float64          `json:"velocity_per_hour"`
	CommissionRate  float64          `json:"commission_rate"`
	ExpectedReturn  float64          `json:"expected_return"`
	WinningScore    float64          `json:"winning_score"`
	SnapshotTime    time.Time        `json:"snapshot_time"`
	Badges          LeaderboardBadge `json:"badges"`
}

// LeaderboardResponse represents the API response for GET /api/v1/leaderboard
type LeaderboardResponse struct {
	CrawlLogID string             `json:"crawl_log_id"`
	CrawlRound string             `json:"crawl_round"`
	UpdatedAt  string             `json:"updated_at"`
	Total      int                `json:"total"`
	Items      []*LeaderboardItem `json:"items"`
}

// LeaderboardRepository defines database operations required for the leaderboard
type LeaderboardRepository interface {
	GetLatestSuccessfulCrawlLog(ctx context.Context) (*CrawlLog, error)
	GetTopSnapshotsByCrawlLogID(ctx context.Context, crawlLogID string, limit int) ([]*LeaderboardItem, error)
}

// LeaderboardUsecase orchestrates leaderboard retrieval and badge calculations
type LeaderboardUsecase interface {
	GetLeaderboard(ctx context.Context, limit int) (*LeaderboardResponse, error)
}
