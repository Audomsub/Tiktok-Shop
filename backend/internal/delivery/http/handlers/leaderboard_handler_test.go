package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

type mockLeaderboardUsecase struct {
	getLeaderboardFunc func(ctx context.Context, limit int) (*domain.LeaderboardResponse, error)
}

func (m *mockLeaderboardUsecase) GetLeaderboard(ctx context.Context, limit int) (*domain.LeaderboardResponse, error) {
	if m.getLeaderboardFunc != nil {
		return m.getLeaderboardFunc(ctx, limit)
	}
	return nil, nil
}

func TestLeaderboardHandler_GetLeaderboard(t *testing.T) {
	t.Run("Returns 200 with leaderboard items on success", func(t *testing.T) {
		mockUsecase := &mockLeaderboardUsecase{
			getLeaderboardFunc: func(ctx context.Context, limit int) (*domain.LeaderboardResponse, error) {
				if limit != 10 {
					t.Errorf("expected default limit 10, got %d", limit)
				}
				return &domain.LeaderboardResponse{
					CrawlLogID: "log-1",
					CrawlRound: "18:00",
					UpdatedAt:  "2026-10-08T18:05:00Z",
					Total:      1,
					Items: []*domain.LeaderboardItem{
						{
							Rank:            1,
							ProductID:       "p-1",
							Name:            "Trending Product",
							WinningScore:    95.0,
							Price:           300.0,
							CommissionRate:  20.0,
							ExpectedReturn:  60.0,
							VelocityPerHour: 15.0,
							Badges: domain.LeaderboardBadge{
								IsViralSurge:     true,
								IsHighCommission: true,
								IsWinningPick:    true,
							},
						},
					},
				}, nil
			},
		}

		handler := NewLeaderboardHandler(mockUsecase)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/leaderboard", nil)
		rec := httptest.NewRecorder()

		handler.GetLeaderboard(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp domain.LeaderboardResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed decoding json response: %v", err)
		}

		if resp.Total != 1 {
			t.Errorf("expected total 1, got %d", resp.Total)
		}
		if len(resp.Items) != 1 || resp.Items[0].Name != "Trending Product" {
			t.Errorf("unexpected items in response")
		}
	})

	t.Run("Correctly parses custom limit query parameter", func(t *testing.T) {
		var receivedLimit int
		mockUsecase := &mockLeaderboardUsecase{
			getLeaderboardFunc: func(ctx context.Context, limit int) (*domain.LeaderboardResponse, error) {
				receivedLimit = limit
				return &domain.LeaderboardResponse{Total: 0, Items: []*domain.LeaderboardItem{}}, nil
			},
		}

		handler := NewLeaderboardHandler(mockUsecase)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/leaderboard?limit=25", nil)
		rec := httptest.NewRecorder()

		handler.GetLeaderboard(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		if receivedLimit != 25 {
			t.Errorf("expected received limit 25, got %d", receivedLimit)
		}
	})

	t.Run("Returns 500 when usecase returns an error", func(t *testing.T) {
		mockUsecase := &mockLeaderboardUsecase{
			getLeaderboardFunc: func(ctx context.Context, limit int) (*domain.LeaderboardResponse, error) {
				return nil, errors.New("db failure")
			},
		}

		handler := NewLeaderboardHandler(mockUsecase)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/leaderboard", nil)
		rec := httptest.NewRecorder()

		handler.GetLeaderboard(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", rec.Code)
		}
	})
}
