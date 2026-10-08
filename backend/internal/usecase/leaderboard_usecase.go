package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

type leaderboardUsecase struct {
	repo    domain.LeaderboardRepository
	favRepo domain.FavoriteRepository
}

// NewLeaderboardUsecase constructs an instance of domain.LeaderboardUsecase
func NewLeaderboardUsecase(repo domain.LeaderboardRepository, favRepo ...domain.FavoriteRepository) domain.LeaderboardUsecase {
	var fRepo domain.FavoriteRepository
	if len(favRepo) > 0 {
		fRepo = favRepo[0]
	}
	return &leaderboardUsecase{repo: repo, favRepo: fRepo}
}

// GetLeaderboard fetches top winning products and enriches them with ranks and badge indicators
func (u *leaderboardUsecase) GetLeaderboard(ctx context.Context, limit int) (*domain.LeaderboardResponse, error) {
	if limit <= 0 {
		limit = 10
	} else if limit > 50 {
		limit = 50
	}

	// 1. Retrieve the latest successful crawl log
	crawlLog, err := u.repo.GetLatestSuccessfulCrawlLog(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving latest successful crawl log: %w", err)
	}

	if crawlLog == nil {
		return &domain.LeaderboardResponse{
			Total: 0,
			Items: make([]*domain.LeaderboardItem, 0),
		}, nil
	}

	// 2. Retrieve top snapshots for this crawl round
	items, err := u.repo.GetTopSnapshotsByCrawlLogID(ctx, crawlLog.ID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving top snapshots for crawl log %s: %w", crawlLog.ID, err)
	}

	if items == nil {
		items = make([]*domain.LeaderboardItem, 0)
	}

	// 3. Determine top 15% threshold count for viral surge badge
	top15PercentCutoff := int(math.Ceil(float64(len(items)) * 0.15))

	// Fetch user's favorite product IDs if user is authenticated and favRepo is configured
	var favMap map[string]bool
	if u.favRepo != nil {
		if userID, ok := domain.GetUserIDFromContext(ctx); ok && userID != "" {
			if m, err := u.favRepo.GetUserFavoriteProductIDs(ctx, userID); err == nil {
				favMap = m
			}
		}
	}

	// 4. Enrich each item with rank, expected return, and badges
	for i, item := range items {
		item.Rank = i + 1

		// Expected Return (THB per unit) = Price * (CommissionRate / 100)
		expectedReturn := item.Price * (item.CommissionRate / 100.0)
		item.ExpectedReturn = math.Round(expectedReturn*100) / 100

		// Badges:
		// 🔥 Viral Surge: Velocity >= 10 pcs/hr or in top 15% of the round
		isViralSurge := item.VelocityPerHour >= 10.0 || (i < top15PercentCutoff && item.VelocityPerHour > 0)

		// 💎 High Commission: Commission Rate >= 20%
		isHighCommission := item.CommissionRate >= 20.0

		// 💰 High Yield: Expected Return >= 100 THB
		isHighYield := item.ExpectedReturn >= 100.0

		// 🏆 Winning Pick: Winning Score >= 80
		isWinningPick := item.WinningScore >= 80.0

		item.Badges = domain.LeaderboardBadge{
			IsViralSurge:     isViralSurge,
			IsHighCommission: isHighCommission,
			IsHighYield:      isHighYield,
			IsWinningPick:    isWinningPick,
		}

		if favMap != nil && favMap[item.ProductID] {
			item.IsFavorited = true
		}
	}

	updatedAt := crawlLog.StartedAt.Format(time.RFC3339)
	if crawlLog.FinishedAt != nil {
		updatedAt = crawlLog.FinishedAt.Format(time.RFC3339)
	}

	return &domain.LeaderboardResponse{
		CrawlLogID: crawlLog.ID,
		CrawlRound: crawlLog.CrawlRound,
		UpdatedAt:  updatedAt,
		Total:      len(items),
		Items:      items,
	}, nil
}
