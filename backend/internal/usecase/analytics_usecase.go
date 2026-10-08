package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

type analyticsUsecase struct {
	snapshotRepo    domain.SnapshotRepository
	crawlLogRepo    domain.CrawlLogRepository
	velocityUsecase domain.VelocityUsecase
	scoringUsecase  domain.ScoringUsecase
}

// NewAnalyticsUsecase constructs a new AnalyticsUsecase implementation
func NewAnalyticsUsecase(
	snapshotRepo domain.SnapshotRepository,
	crawlLogRepo domain.CrawlLogRepository,
	velocityUsecase domain.VelocityUsecase,
	scoringUsecase domain.ScoringUsecase,
) domain.AnalyticsUsecase {
	return &analyticsUsecase{
		snapshotRepo:    snapshotRepo,
		crawlLogRepo:    crawlLogRepo,
		velocityUsecase: velocityUsecase,
		scoringUsecase:  scoringUsecase,
	}
}

// ProcessCrawlRound orchestrates end-to-end batch velocity and scoring computation for a crawl round
func (u *analyticsUsecase) ProcessCrawlRound(ctx context.Context, crawlLogID string) (*domain.BatchProcessResult, error) {
	startTime := time.Now()

	// 1. Verify existence of crawl log
	crawlLog, err := u.crawlLogRepo.GetByID(ctx, crawlLogID)
	if err != nil {
		return nil, fmt.Errorf("crawl log verification failed: %w", err)
	}

	// 2. Fetch all snapshots captured during this crawl round
	snapshots, err := u.snapshotRepo.GetSnapshotsByCrawlLogID(ctx, crawlLogID)
	if err != nil {
		errMsg := err.Error()
		_ = u.crawlLogRepo.UpdateStatus(ctx, crawlLogID, domain.CrawlStatusFailed, time.Now().UTC(), &errMsg)
		return nil, fmt.Errorf("failed retrieving snapshots for crawl log %s: %w", crawlLogID, err)
	}

	if len(snapshots) == 0 {
		now := time.Now().UTC()
		_ = u.crawlLogRepo.UpdateStatus(ctx, crawlLogID, domain.CrawlStatusSuccess, now, nil)
		return &domain.BatchProcessResult{
			CrawlLogID:     crawlLogID,
			ProcessedCount: 0,
			DurationMs:     time.Since(startTime).Milliseconds(),
			Status:         string(domain.CrawlStatusSuccess),
		}, nil
	}

	// 3. Extract unique product IDs and snapshot reference time
	productIDs := make([]string, len(snapshots))
	for i, s := range snapshots {
		productIDs[i] = s.ProductID
	}
	snapshotTime := snapshots[0].SnapshotTime

	// 4. Retrieve preceding snapshots within 24-hour Lookback Window
	previousSnapshots, err := u.snapshotRepo.GetPreviousSnapshots(ctx, productIDs, snapshotTime, MaxLookbackHours)
	if err != nil {
		errMsg := err.Error()
		_ = u.crawlLogRepo.UpdateStatus(ctx, crawlLogID, domain.CrawlStatusFailed, time.Now().UTC(), &errMsg)
		return nil, fmt.Errorf("failed retrieving previous snapshots: %w", err)
	}

	// 5. Compute sales delta and hourly velocity per product
	scoringItems := make([]*domain.ScoringItem, len(snapshots))
	for i, curr := range snapshots {
		prev := previousSnapshots[curr.ProductID]
		delta, velocity := u.velocityUsecase.ComputeVelocity(curr, prev)

		scoringItems[i] = &domain.ScoringItem{
			SnapshotID:      curr.ID,
			ProductID:       curr.ProductID,
			Price:           curr.Price,
			CommissionRate:  curr.CommissionRate,
			VelocityPerHour: velocity,
		}
		// Temporarily keep delta on current snapshot for mapping
		curr.DeltaSales = delta
		curr.VelocityPerHour = velocity
	}

	// 6. Run batch Min-Max Normalization and Winning Score calculation
	scoredItems := u.scoringUsecase.CalculateBatchScores(scoringItems)

	// 7. Map calculated results into CalculatedSnapshot structures
	calculations := make([]*domain.CalculatedSnapshot, len(scoredItems))
	for i, scored := range scoredItems {
		calculations[i] = &domain.CalculatedSnapshot{
			SnapshotID:      scored.SnapshotID,
			DeltaSales:      snapshots[i].DeltaSales,
			VelocityPerHour: scored.VelocityPerHour,
			WinningScore:    scored.WinningScore,
		}
	}

	// 8. Persist calculations using Worker Pool & Chunked Bulk Update
	if err := u.snapshotRepo.BulkUpdateCalculations(ctx, calculations); err != nil {
		errMsg := err.Error()
		_ = u.crawlLogRepo.UpdateStatus(ctx, crawlLogID, domain.CrawlStatusFailed, time.Now().UTC(), &errMsg)
		return nil, fmt.Errorf("failed persisting calculated snapshot batch: %w", err)
	}

	// 9. Update crawl log status to SUCCESS
	finishedAt := time.Now().UTC()
	finalStatus := domain.CrawlStatusSuccess
	if crawlLog.Status == domain.CrawlStatusPartial {
		finalStatus = domain.CrawlStatusPartial
	}

	if err := u.crawlLogRepo.UpdateStatus(ctx, crawlLogID, finalStatus, finishedAt, nil); err != nil {
		return nil, fmt.Errorf("failed updating crawl log final status: %w", err)
	}

	duration := time.Since(startTime).Milliseconds()
	return &domain.BatchProcessResult{
		CrawlLogID:     crawlLogID,
		ProcessedCount: len(calculations),
		DurationMs:     duration,
		Status:         string(finalStatus),
	}, nil
}
