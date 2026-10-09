package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

type trendUsecase struct {
	repo domain.TrendRepository
}

// NewTrendUsecase constructs a new TrendUsecase instance
func NewTrendUsecase(repo domain.TrendRepository) domain.TrendUsecase {
	return &trendUsecase{repo: repo}
}

// GetProductTrends retrieves chronological time-series snapshot data for a product
func (u *trendUsecase) GetProductTrends(ctx context.Context, productID string, days int) (*domain.ProductTrendsResponse, error) {
	// Sanitize days lookback parameter
	if days <= 0 {
		days = 7
	} else if days > 30 {
		days = 30
	}

	// 1. Verify that product exists
	product, err := u.repo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed verifying product existence: %w", err)
	}
	if product == nil {
		return nil, domain.ErrProductNotFound
	}

	// 2. Fetch snapshots since lookback cutoff
	since := time.Now().AddDate(0, 0, -days)
	snapshots, err := u.repo.GetProductSnapshotsSince(ctx, productID, since)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving trend snapshots: %w", err)
	}

	points := make([]*domain.TrendPoint, 0, len(snapshots))
	for _, snap := range snapshots {
		expectedReturn := snap.Price * (snap.CommissionRate / 100.0)
		points = append(points, &domain.TrendPoint{
			SnapshotTime:    snap.SnapshotTime,
			Price:           snap.Price,
			CommissionRate:  snap.CommissionRate,
			TotalSales:      snap.TotalSales,
			DeltaSales:      snap.DeltaSales,
			VelocityPerHour: snap.VelocityPerHour,
			ExpectedReturn:  math.Round(expectedReturn*100) / 100,
			WinningScore:    snap.WinningScore,
		})
	}

	return &domain.ProductTrendsResponse{
		ProductID:   product.ID,
		ProductName: product.Name,
		Days:        days,
		TotalPoints: len(points),
		Points:      points,
	}, nil
}
