package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

// mockLeaderboardRepo is an in-memory test double for domain.LeaderboardRepository
type mockLeaderboardRepo struct {
	getLatestLogFunc func(ctx context.Context) (*domain.CrawlLog, error)
	getTopItemsFunc  func(ctx context.Context, crawlLogID string, limit int) ([]*domain.LeaderboardItem, error)
}

func (m *mockLeaderboardRepo) GetLatestSuccessfulCrawlLog(ctx context.Context) (*domain.CrawlLog, error) {
	if m.getLatestLogFunc != nil {
		return m.getLatestLogFunc(ctx)
	}
	return nil, nil
}

func (m *mockLeaderboardRepo) GetTopSnapshotsByCrawlLogID(ctx context.Context, crawlLogID string, limit int) ([]*domain.LeaderboardItem, error) {
	if m.getTopItemsFunc != nil {
		return m.getTopItemsFunc(ctx, crawlLogID, limit)
	}
	return nil, nil
}

func TestLeaderboardUsecase_GetLeaderboard(t *testing.T) {
	now := time.Now()
	finished := now.Add(5 * time.Minute)

	t.Run("Successfully enriches top products with ranks, expected return, and badges", func(t *testing.T) {
		mockRepo := &mockLeaderboardRepo{
			getLatestLogFunc: func(ctx context.Context) (*domain.CrawlLog, error) {
				return &domain.CrawlLog{
					ID:         "log-100",
					CrawlRound: "12:00",
					Status:     domain.CrawlStatusSuccess,
					StartedAt:  now,
					FinishedAt: &finished,
				}, nil
			},
			getTopItemsFunc: func(ctx context.Context, crawlLogID string, limit int) ([]*domain.LeaderboardItem, error) {
				return []*domain.LeaderboardItem{
					{
						ProductID:       "prod-1",
						Name:            "Wireless Gaming Earbuds",
						Price:           500.0,
						CommissionRate:  25.0,
						VelocityPerHour: 12.0,
						WinningScore:    92.5,
					},
					{
						ProductID:       "prod-2",
						Name:            "Cotton T-Shirt",
						Price:           200.0,
						CommissionRate:  10.0,
						VelocityPerHour: 4.0,
						WinningScore:    65.0,
					},
					{
						ProductID:       "prod-3",
						Name:            "Premium Chef Knife",
						Price:           1000.0,
						CommissionRate:  20.0,
						VelocityPerHour: 2.0,
						WinningScore:    82.0,
					},
				}, nil
			},
		}

		u := NewLeaderboardUsecase(mockRepo)
		resp, err := u.GetLeaderboard(context.Background(), 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Total != 3 {
			t.Fatalf("expected 3 items, got %d", resp.Total)
		}
		if resp.CrawlLogID != "log-100" {
			t.Errorf("expected CrawlLogID 'log-100', got '%s'", resp.CrawlLogID)
		}
		if resp.CrawlRound != "12:00" {
			t.Errorf("expected CrawlRound '12:00', got '%s'", resp.CrawlRound)
		}

		// Item 1 verification
		item1 := resp.Items[0]
		if item1.Rank != 1 {
			t.Errorf("item1 rank: expected 1, got %d", item1.Rank)
		}
		if item1.ExpectedReturn != 125.0 {
			t.Errorf("item1 expected return: expected 125.0, got %f", item1.ExpectedReturn)
		}
		if !item1.Badges.IsViralSurge {
			t.Errorf("item1 expected IsViralSurge=true")
		}
		if !item1.Badges.IsHighCommission {
			t.Errorf("item1 expected IsHighCommission=true")
		}
		if !item1.Badges.IsHighYield {
			t.Errorf("item1 expected IsHighYield=true")
		}
		if !item1.Badges.IsWinningPick {
			t.Errorf("item1 expected IsWinningPick=true")
		}

		// Item 2 verification
		item2 := resp.Items[1]
		if item2.Rank != 2 {
			t.Errorf("item2 rank: expected 2, got %d", item2.Rank)
		}
		if item2.ExpectedReturn != 20.0 {
			t.Errorf("item2 expected return: expected 20.0, got %f", item2.ExpectedReturn)
		}
		if item2.Badges.IsViralSurge {
			t.Errorf("item2 expected IsViralSurge=false")
		}
		if item2.Badges.IsHighCommission {
			t.Errorf("item2 expected IsHighCommission=false")
		}
		if item2.Badges.IsHighYield {
			t.Errorf("item2 expected IsHighYield=false")
		}
		if item2.Badges.IsWinningPick {
			t.Errorf("item2 expected IsWinningPick=false")
		}

		// Item 3 verification
		item3 := resp.Items[2]
		if item3.Rank != 3 {
			t.Errorf("item3 rank: expected 3, got %d", item3.Rank)
		}
		if item3.ExpectedReturn != 200.0 {
			t.Errorf("item3 expected return: expected 200.0, got %f", item3.ExpectedReturn)
		}
		if item3.Badges.IsViralSurge {
			t.Errorf("item3 expected IsViralSurge=false")
		}
		if !item3.Badges.IsHighCommission {
			t.Errorf("item3 expected IsHighCommission=true")
		}
		if !item3.Badges.IsHighYield {
			t.Errorf("item3 expected IsHighYield=true")
		}
		if !item3.Badges.IsWinningPick {
			t.Errorf("item3 expected IsWinningPick=true")
		}
	})

	t.Run("Returns empty response gracefully when no successful crawl log exists", func(t *testing.T) {
		mockRepo := &mockLeaderboardRepo{
			getLatestLogFunc: func(ctx context.Context) (*domain.CrawlLog, error) {
				return nil, nil
			},
		}

		u := NewLeaderboardUsecase(mockRepo)
		resp, err := u.GetLeaderboard(context.Background(), 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Total != 0 {
			t.Errorf("expected 0 items, got %d", resp.Total)
		}
		if len(resp.Items) != 0 {
			t.Errorf("expected empty items slice, got len %d", len(resp.Items))
		}
	})

	t.Run("Propagates database errors appropriately", func(t *testing.T) {
		mockRepo := &mockLeaderboardRepo{
			getLatestLogFunc: func(ctx context.Context) (*domain.CrawlLog, error) {
				return nil, errors.New("connection failed")
			},
		}

		u := NewLeaderboardUsecase(mockRepo)
		_, err := u.GetLeaderboard(context.Background(), 10)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("Enriches IsFavorited on leaderboard items when user has favorites", func(t *testing.T) {
		mockRepo := &mockLeaderboardRepo{
			getLatestLogFunc: func(ctx context.Context) (*domain.CrawlLog, error) {
				return &domain.CrawlLog{
					ID:         "log-200",
					CrawlRound: "15:00",
					Status:     domain.CrawlStatusSuccess,
					StartedAt:  now,
				}, nil
			},
			getTopItemsFunc: func(ctx context.Context, crawlLogID string, limit int) ([]*domain.LeaderboardItem, error) {
				return []*domain.LeaderboardItem{
					{ProductID: "top-fav", Name: "Favorited Leader"},
					{ProductID: "top-other", Name: "Other Leader"},
				}, nil
			},
		}

		favRepo := newInMemoryFavoriteRepo()
		_ = favRepo.Add(context.Background(), "user-bob", "top-fav")

		u := NewLeaderboardUsecase(mockRepo, favRepo)

		// Authenticated user bob
		ctxBob := context.WithValue(context.Background(), domain.UserIDContextKey, "user-bob")
		respBob, err := u.GetLeaderboard(ctxBob, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !respBob.Items[0].IsFavorited {
			t.Errorf("expected top-fav to have IsFavorited=true for user-bob")
		}
		if respBob.Items[1].IsFavorited {
			t.Errorf("expected top-other to have IsFavorited=false for user-bob")
		}

		// Anonymous request
		respAnon, err := u.GetLeaderboard(context.Background(), 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if respAnon.Items[0].IsFavorited || respAnon.Items[1].IsFavorited {
			t.Errorf("expected anonymous user to have all IsFavorited=false")
		}
	})
}
