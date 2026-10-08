package usecase_test

import (
	"fmt"
	"testing"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
	"github.com/Audomsub/Tiktok-Shop/backend/internal/usecase"
)

func TestScoringUsecase_ComputeExpectedReturn(t *testing.T) {
	uc := usecase.NewScoringUsecase()

	tests := []struct {
		name       string
		price      float64
		commission float64
		expected   float64
	}{
		{"Standard 500 THB at 20%", 500.0, 20.0, 100.0},
		{"Decimal commission 15.5% on 200 THB", 200.0, 15.5, 31.0},
		{"Zero price returns zero", 0.0, 20.0, 0.0},
		{"Negative price returns zero", -100.0, 20.0, 0.0},
		{"Zero commission returns zero", 500.0, 0.0, 0.0},
		{"Rounding to 2 decimal places", 199.0, 12.33, 24.54},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := uc.ComputeExpectedReturn(tt.price, tt.commission)
			if got != tt.expected {
				t.Errorf("expected expected return %f, got %f", tt.expected, got)
			}
		})
	}
}

func TestScoringUsecase_CalculateBatchScores(t *testing.T) {
	uc := usecase.NewScoringUsecase()

	t.Run("Empty slice returns empty without panic", func(t *testing.T) {
		res := uc.CalculateBatchScores([]*domain.ScoringItem{})
		if len(res) != 0 {
			t.Errorf("expected empty slice, got %d items", len(res))
		}
	})

	t.Run("Single item batch returns baseline 50.0", func(t *testing.T) {
		item := &domain.ScoringItem{
			SnapshotID:      "snap-1",
			ProductID:       "prod-1",
			Price:           300.0,
			CommissionRate:  15.0,
			VelocityPerHour: 25.0,
		}
		res := uc.CalculateBatchScores([]*domain.ScoringItem{item})
		if len(res) != 1 {
			t.Fatalf("expected 1 item, got %d", len(res))
		}
		if res[0].WinningScore != 50.0 {
			t.Errorf("expected score 50.0 for single item, got %f", res[0].WinningScore)
		}
	})

	t.Run("Standard multi-item batch computes proper ranking within 0 to 100", func(t *testing.T) {
		items := []*domain.ScoringItem{
			{SnapshotID: "1", ProductID: "p1", Price: 100.0, CommissionRate: 10.0, VelocityPerHour: 5.0},
			{SnapshotID: "2", ProductID: "p2", Price: 300.0, CommissionRate: 20.0, VelocityPerHour: 30.0},
			{SnapshotID: "3", ProductID: "p3", Price: 800.0, CommissionRate: 35.0, VelocityPerHour: 80.0},
		}

		res := uc.CalculateBatchScores(items)

		// Item 3 has top velocity, commission, and return -> should have score 100
		if res[2].WinningScore != 100.0 {
			t.Errorf("expected top item to receive score 100.0, got %f", res[2].WinningScore)
		}

		// Item 1 has lowest across all -> should have score 0.0
		if res[0].WinningScore != 0.0 {
			t.Errorf("expected lowest item to receive score 0.0, got %f", res[0].WinningScore)
		}

		// Item 2 should fall in the middle
		if res[1].WinningScore <= 0.0 || res[1].WinningScore >= 100.0 {
			t.Errorf("expected middle item to be strictly between 0 and 100, got %f", res[1].WinningScore)
		}
	})

	t.Run("Outlier clipping at 99th percentile prevents distortion", func(t *testing.T) {
		// Create 100 items: 99 items with velocity 10 to 50, and 1 extreme outlier with 10,000
		items := make([]*domain.ScoringItem, 100)
		for i := 0; i < 99; i++ {
			items[i] = &domain.ScoringItem{
				SnapshotID:      fmt.Sprintf("s-%d", i),
				ProductID:       fmt.Sprintf("p-%d", i),
				Price:           200.0,
				CommissionRate:  15.0,
				VelocityPerHour: float64(10 + (i % 40)), // 10 to 49 pcs/hr
			}
		}
		// Extreme viral outlier
		items[99] = &domain.ScoringItem{
			SnapshotID:      "s-outlier",
			ProductID:       "p-outlier",
			Price:           200.0,
			CommissionRate:  15.0,
			VelocityPerHour: 10000.0,
		}

		res := uc.CalculateBatchScores(items)

		// Check that typical items are not flattened to near-zero velocity
		maxNormalVelocity := res[98].NormalizedVelocity
		if maxNormalVelocity <= 10.0 {
			t.Errorf("expected normal items to retain meaningful normalized velocity, got %f", maxNormalVelocity)
		}

		// Ensure all winning scores stay strictly within [0, 100]
		for _, item := range res {
			if item.WinningScore < 0.0 || item.WinningScore > 100.0 {
				t.Fatalf("winning score out of bounds: %f", item.WinningScore)
			}
		}
	})

	t.Run("Identical values across all items does not divide by zero", func(t *testing.T) {
		items := []*domain.ScoringItem{
			{SnapshotID: "1", ProductID: "p1", Price: 200.0, CommissionRate: 15.0, VelocityPerHour: 20.0},
			{SnapshotID: "2", ProductID: "p2", Price: 200.0, CommissionRate: 15.0, VelocityPerHour: 20.0},
			{SnapshotID: "3", ProductID: "p3", Price: 200.0, CommissionRate: 15.0, VelocityPerHour: 20.0},
		}

		res := uc.CalculateBatchScores(items)

		for _, item := range res {
			if item.WinningScore != 50.0 {
				t.Errorf("expected score 50.0 for uniform distribution, got %f", item.WinningScore)
			}
		}
	})
}
