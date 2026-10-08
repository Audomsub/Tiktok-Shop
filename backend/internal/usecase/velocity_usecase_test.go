package usecase_test

import (
	"testing"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
	"github.com/Audomsub/Tiktok-Shop/backend/internal/usecase"
)

func TestVelocityUsecase_ComputeVelocity(t *testing.T) {
	uc := usecase.NewVelocityUsecase()
	baseTime := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	t.Run("Cold start with no previous snapshot returns baseline zero", func(t *testing.T) {
		current := &domain.ProductSnapshot{
			ID:           "curr-1",
			ProductID:    "prod-1",
			SnapshotTime: baseTime,
			TotalSales:   5000,
		}

		delta, velocity := uc.ComputeVelocity(current, nil)

		if delta != 0 {
			t.Errorf("expected delta 0 for cold start, got %d", delta)
		}
		if velocity != 0.0 {
			t.Errorf("expected velocity 0.0 for cold start, got %f", velocity)
		}
	})

	t.Run("Standard 6-hour interval computes accurate hourly velocity", func(t *testing.T) {
		prev := &domain.ProductSnapshot{
			ID:           "prev-1",
			ProductID:    "prod-1",
			SnapshotTime: baseTime.Add(-6 * time.Hour),
			TotalSales:   100,
		}
		curr := &domain.ProductSnapshot{
			ID:           "curr-1",
			ProductID:    "prod-1",
			SnapshotTime: baseTime,
			TotalSales:   220,
		}

		delta, velocity := uc.ComputeVelocity(curr, prev)

		expectedDelta := 120
		expectedVelocity := 20.00 // 120 / 6 hours

		if delta != expectedDelta {
			t.Errorf("expected delta %d, got %d", expectedDelta, delta)
		}
		if velocity != expectedVelocity {
			t.Errorf("expected velocity %f, got %f", expectedVelocity, velocity)
		}
	})

	t.Run("12-hour interval rounds to two decimal places", func(t *testing.T) {
		prev := &domain.ProductSnapshot{
			ID:           "prev-1",
			ProductID:    "prod-1",
			SnapshotTime: baseTime.Add(-12 * time.Hour),
			TotalSales:   500,
		}
		curr := &domain.ProductSnapshot{
			ID:           "curr-1",
			ProductID:    "prod-1",
			SnapshotTime: baseTime,
			TotalSales:   750, // delta = 250, 250 / 12 = 20.8333... -> 20.83
		}

		delta, velocity := uc.ComputeVelocity(curr, prev)

		if delta != 250 {
			t.Errorf("expected delta 250, got %d", delta)
		}
		if velocity != 20.83 {
			t.Errorf("expected velocity 20.83, got %f", velocity)
		}
	})

	t.Run("Lookback window exceeded (> 24 hours) resets to baseline zero", func(t *testing.T) {
		prev := &domain.ProductSnapshot{
			ID:           "prev-1",
			ProductID:    "prod-1",
			SnapshotTime: baseTime.Add(-25 * time.Hour), // 25 hours ago
			TotalSales:   100,
		}
		curr := &domain.ProductSnapshot{
			ID:           "curr-1",
			ProductID:    "prod-1",
			SnapshotTime: baseTime,
			TotalSales:   1000,
		}

		delta, velocity := uc.ComputeVelocity(curr, prev)

		if delta != 0 || velocity != 0.0 {
			t.Errorf("expected delta 0 and velocity 0.0 when exceeding 24h lookback, got delta=%d velocity=%f", delta, velocity)
		}
	})

	t.Run("Exactly 24 hours lookback is allowed and calculated", func(t *testing.T) {
		prev := &domain.ProductSnapshot{
			ID:           "prev-1",
			ProductID:    "prod-1",
			SnapshotTime: baseTime.Add(-24 * time.Hour),
			TotalSales:   1000,
		}
		curr := &domain.ProductSnapshot{
			ID:           "curr-1",
			ProductID:    "prod-1",
			SnapshotTime: baseTime,
			TotalSales:   1240, // delta = 240, velocity = 240 / 24 = 10.0
		}

		delta, velocity := uc.ComputeVelocity(curr, prev)

		if delta != 240 {
			t.Errorf("expected delta 240, got %d", delta)
		}
		if velocity != 10.00 {
			t.Errorf("expected velocity 10.00, got %f", velocity)
		}
	})

	t.Run("Negative delta sales (data anomaly) is clamped to zero", func(t *testing.T) {
		prev := &domain.ProductSnapshot{
			ID:           "prev-1",
			ProductID:    "prod-1",
			SnapshotTime: baseTime.Add(-6 * time.Hour),
			TotalSales:   500,
		}
		curr := &domain.ProductSnapshot{
			ID:           "curr-1",
			ProductID:    "prod-1",
			SnapshotTime: baseTime,
			TotalSales:   480, // Sales decreased by 20 (anomaly)
		}

		delta, velocity := uc.ComputeVelocity(curr, prev)

		if delta != 0 {
			t.Errorf("expected clamped delta 0, got %d", delta)
		}
		if velocity != 0.0 {
			t.Errorf("expected velocity 0.0, got %f", velocity)
		}
	})

	t.Run("Zero sales change results in zero velocity", func(t *testing.T) {
		prev := &domain.ProductSnapshot{
			ID:           "prev-1",
			ProductID:    "prod-1",
			SnapshotTime: baseTime.Add(-6 * time.Hour),
			TotalSales:   300,
		}
		curr := &domain.ProductSnapshot{
			ID:           "curr-1",
			ProductID:    "prod-1",
			SnapshotTime: baseTime,
			TotalSales:   300,
		}

		delta, velocity := uc.ComputeVelocity(curr, prev)

		if delta != 0 || velocity != 0.0 {
			t.Errorf("expected delta 0 and velocity 0.0, got delta=%d velocity=%f", delta, velocity)
		}
	})

	t.Run("Chronological inversion returns baseline zero", func(t *testing.T) {
		prev := &domain.ProductSnapshot{
			ID:           "prev-1",
			ProductID:    "prod-1",
			SnapshotTime: baseTime.Add(2 * time.Hour), // Future timestamp
			TotalSales:   100,
		}
		curr := &domain.ProductSnapshot{
			ID:           "curr-1",
			ProductID:    "prod-1",
			SnapshotTime: baseTime,
			TotalSales:   200,
		}

		delta, velocity := uc.ComputeVelocity(curr, prev)

		if delta != 0 || velocity != 0.0 {
			t.Errorf("expected delta 0 and velocity 0.0 on chronological inversion, got delta=%d velocity=%f", delta, velocity)
		}
	})

	t.Run("Nil current snapshot safely returns zero", func(t *testing.T) {
		prev := &domain.ProductSnapshot{
			ID:           "prev-1",
			ProductID:    "prod-1",
			SnapshotTime: baseTime.Add(-6 * time.Hour),
			TotalSales:   100,
		}

		delta, velocity := uc.ComputeVelocity(nil, prev)

		if delta != 0 || velocity != 0.0 {
			t.Errorf("expected delta 0 and velocity 0.0 for nil current, got delta=%d velocity=%f", delta, velocity)
		}
	})
}
