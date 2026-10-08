package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
	"github.com/Audomsub/Tiktok-Shop/backend/internal/usecase"
)

// mockSnapshotRepo implements domain.SnapshotRepository
type mockSnapshotRepo struct {
	snapshots          []*domain.ProductSnapshot
	previousSnapshots  map[string]*domain.ProductSnapshot
	getSnapshotsErr    error
	getPrevErr         error
	bulkUpdateErr      error
	savedCalculations  []*domain.CalculatedSnapshot
}

func (m *mockSnapshotRepo) GetSnapshotsByCrawlLogID(ctx context.Context, crawlLogID string) ([]*domain.ProductSnapshot, error) {
	return m.snapshots, m.getSnapshotsErr
}

func (m *mockSnapshotRepo) GetPreviousSnapshots(ctx context.Context, productIDs []string, beforeTime time.Time, maxLookbackHours float64) (map[string]*domain.ProductSnapshot, error) {
	return m.previousSnapshots, m.getPrevErr
}

func (m *mockSnapshotRepo) BulkUpdateCalculations(ctx context.Context, calculations []*domain.CalculatedSnapshot) error {
	m.savedCalculations = calculations
	return m.bulkUpdateErr
}

// mockCrawlLogRepo implements domain.CrawlLogRepository
type mockCrawlLogRepo struct {
	crawlLog    *domain.CrawlLog
	getErr      error
	updateErr   error
	lastStatus  domain.CrawlStatus
	lastError   *string
}

func (m *mockCrawlLogRepo) GetByID(ctx context.Context, id string) (*domain.CrawlLog, error) {
	return m.crawlLog, m.getErr
}

func (m *mockCrawlLogRepo) UpdateStatus(ctx context.Context, id string, status domain.CrawlStatus, finishedAt time.Time, errMsg *string) error {
	m.lastStatus = status
	m.lastError = errMsg
	return m.updateErr
}

func TestAnalyticsUsecase_ProcessCrawlRound(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	t.Run("Successfully processes and updates batch of snapshots", func(t *testing.T) {
		crawlLog := &domain.CrawlLog{
			ID:         "log-1",
			CrawlRound: "12:00",
			Status:     domain.CrawlStatusRunning,
			StartedAt:  now.Add(-10 * time.Minute),
		}

		currSnapshots := []*domain.ProductSnapshot{
			{ID: "s1", ProductID: "p1", SnapshotTime: now, Price: 300, CommissionRate: 20, TotalSales: 220},
			{ID: "s2", ProductID: "p2", SnapshotTime: now, Price: 500, CommissionRate: 15, TotalSales: 5000}, // New product (Cold start)
		}

		prevSnapshots := map[string]*domain.ProductSnapshot{
			"p1": {ID: "prev-s1", ProductID: "p1", SnapshotTime: now.Add(-6 * time.Hour), Price: 300, CommissionRate: 20, TotalSales: 100},
			// p2 has no previous snapshot
		}

		snapRepo := &mockSnapshotRepo{
			snapshots:         currSnapshots,
			previousSnapshots: prevSnapshots,
		}
		crawlRepo := &mockCrawlLogRepo{crawlLog: crawlLog}
		velUsecase := usecase.NewVelocityUsecase()
		scoreUsecase := usecase.NewScoringUsecase()

		analytics := usecase.NewAnalyticsUsecase(snapRepo, crawlRepo, velUsecase, scoreUsecase)

		result, err := analytics.ProcessCrawlRound(ctx, "log-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.ProcessedCount != 2 {
			t.Errorf("expected 2 processed records, got %d", result.ProcessedCount)
		}
		if crawlRepo.lastStatus != domain.CrawlStatusSuccess {
			t.Errorf("expected crawl log status SUCCESS, got %s", crawlRepo.lastStatus)
		}
		if len(snapRepo.savedCalculations) != 2 {
			t.Fatalf("expected 2 saved calculations, got %d", len(snapRepo.savedCalculations))
		}

		// Verify P1 has calculated delta (220 - 100 = 120) and velocity (120 / 6 = 20)
		p1Calc := snapRepo.savedCalculations[0]
		if p1Calc.DeltaSales != 120 || p1Calc.VelocityPerHour != 20.0 {
			t.Errorf("unexpected calculations for p1: delta=%d, vel=%f", p1Calc.DeltaSales, p1Calc.VelocityPerHour)
		}

		// Verify P2 is treated as cold start baseline (delta = 0, vel = 0)
		p2Calc := snapRepo.savedCalculations[1]
		if p2Calc.DeltaSales != 0 || p2Calc.VelocityPerHour != 0.0 {
			t.Errorf("unexpected calculations for p2 cold start: delta=%d, vel=%f", p2Calc.DeltaSales, p2Calc.VelocityPerHour)
		}
	})

	t.Run("Empty snapshot batch safely marks SUCCESS with zero records", func(t *testing.T) {
		crawlLog := &domain.CrawlLog{ID: "log-empty", Status: domain.CrawlStatusRunning}
		snapRepo := &mockSnapshotRepo{snapshots: []*domain.ProductSnapshot{}}
		crawlRepo := &mockCrawlLogRepo{crawlLog: crawlLog}

		analytics := usecase.NewAnalyticsUsecase(snapRepo, crawlRepo, usecase.NewVelocityUsecase(), usecase.NewScoringUsecase())

		result, err := analytics.ProcessCrawlRound(ctx, "log-empty")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.ProcessedCount != 0 {
			t.Errorf("expected 0 processed count, got %d", result.ProcessedCount)
		}
		if crawlRepo.lastStatus != domain.CrawlStatusSuccess {
			t.Errorf("expected status SUCCESS, got %s", crawlRepo.lastStatus)
		}
	})

	t.Run("Database error updates crawl log to FAILED", func(t *testing.T) {
		crawlLog := &domain.CrawlLog{ID: "log-fail", Status: domain.CrawlStatusRunning}
		snapRepo := &mockSnapshotRepo{
			getSnapshotsErr: errors.New("database timeout"),
		}
		crawlRepo := &mockCrawlLogRepo{crawlLog: crawlLog}

		analytics := usecase.NewAnalyticsUsecase(snapRepo, crawlRepo, usecase.NewVelocityUsecase(), usecase.NewScoringUsecase())

		_, err := analytics.ProcessCrawlRound(ctx, "log-fail")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		if crawlRepo.lastStatus != domain.CrawlStatusFailed {
			t.Errorf("expected status FAILED, got %s", crawlRepo.lastStatus)
		}
	})
}
