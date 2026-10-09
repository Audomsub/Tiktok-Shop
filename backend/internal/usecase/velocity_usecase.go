package usecase

import (
	"math"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

const (
	// MaxLookbackHours defines the maximum historical lookback window (24 hours)
	MaxLookbackHours = 24.0
	// MinElapsedHoursThreshold avoids divide-by-zero on microsecond differences
	MinElapsedHoursThreshold = 0.001
)

type velocityUsecase struct{}

// NewVelocityUsecase creates an instance of domain.VelocityUsecase
func NewVelocityUsecase() domain.VelocityUsecase {
	return &velocityUsecase{}
}

// ComputeVelocity evaluates sales delta and hourly velocity between current and previous snapshots
func (u *velocityUsecase) ComputeVelocity(current *domain.ProductSnapshot, previous *domain.ProductSnapshot) (int, float64) {
	if current == nil {
		return 0, 0.0
	}

	// Cold start (No prior snapshot): Treat as Baseline Snapshot
	if previous == nil {
		return 0, 0.0
	}

	// Invalid chronological order
	if !current.SnapshotTime.After(previous.SnapshotTime) {
		return 0, 0.0
	}

	// Calculate elapsed hours between consecutive snapshots
	elapsedHours := current.SnapshotTime.Sub(previous.SnapshotTime).Hours()

	// Lookback Window check: if gap exceeds 24 hours, reset as baseline to prevent diluted velocity
	if elapsedHours > MaxLookbackHours {
		return 0, 0.0
	}

	// Compute Delta Sales
	deltaSales := current.TotalSales - previous.TotalSales

	// Anomaly / refund correction guard: clamp negative delta to zero
	if deltaSales < 0 {
		deltaSales = 0
	}

	// Guard against division by near-zero duration
	if elapsedHours < MinElapsedHoursThreshold {
		return deltaSales, 0.0
	}

	// Calculate hourly velocity rounded to 2 decimal places
	rawVelocity := float64(deltaSales) / elapsedHours
	roundedVelocity := math.Round(rawVelocity*100) / 100

	return deltaSales, roundedVelocity
}
