package usecase

import (
	"math"
	"sort"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

const (
	WeightVelocity       = 0.50
	WeightCommissionRate = 0.30
	WeightExpectedReturn = 0.20
)

type scoringUsecase struct{}

// NewScoringUsecase constructs an instance of domain.ScoringUsecase
func NewScoringUsecase() domain.ScoringUsecase {
	return &scoringUsecase{}
}

// ComputeExpectedReturn calculates projected affiliate commission in THB per unit
func (u *scoringUsecase) ComputeExpectedReturn(price float64, commissionRate float64) float64 {
	if price <= 0 || commissionRate <= 0 {
		return 0.0
	}
	expected := price * (commissionRate / 100.0)
	return math.Round(expected*100) / 100
}

// CalculateBatchScores processes the batch through percentile clipping, min-max normalization, and weighted scoring
func (u *scoringUsecase) CalculateBatchScores(items []*domain.ScoringItem) []*domain.ScoringItem {
	if len(items) == 0 {
		return items
	}

	// 1. Populate Expected Return for all items
	velocities := make([]float64, len(items))
	commissions := make([]float64, len(items))
	returns := make([]float64, len(items))

	for i, item := range items {
		item.ExpectedReturn = u.ComputeExpectedReturn(item.Price, item.CommissionRate)
		velocities[i] = item.VelocityPerHour
		commissions[i] = item.CommissionRate
		returns[i] = item.ExpectedReturn
	}

	// Single item handling
	if len(items) == 1 {
		items[0].NormalizedVelocity = 50.0
		items[0].NormalizedCommission = 50.0
		items[0].NormalizedReturn = 50.0
		items[0].WinningScore = 50.0
		return items
	}

	// 2. Compute 99th percentile threshold for outlier clipping
	p99Velocity := calculatePercentile(velocities, 0.99)
	p99Commission := calculatePercentile(commissions, 0.99)
	p99Return := calculatePercentile(returns, 0.99)

	// 3. Find clipped min & max across the batch
	minV, maxV := findMinMaxClipped(velocities, p99Velocity)
	minC, maxC := findMinMaxClipped(commissions, p99Commission)
	minR, maxR := findMinMaxClipped(returns, p99Return)

	// 4. Normalize and calculate weighted Winning Score for each item
	for _, item := range items {
		clippedV := math.Min(item.VelocityPerHour, p99Velocity)
		clippedC := math.Min(item.CommissionRate, p99Commission)
		clippedR := math.Min(item.ExpectedReturn, p99Return)

		item.NormalizedVelocity = normalize(clippedV, minV, maxV)
		item.NormalizedCommission = normalize(clippedC, minC, maxC)
		item.NormalizedReturn = normalize(clippedR, minR, maxR)

		rawScore := (WeightVelocity * item.NormalizedVelocity) +
			(WeightCommissionRate * item.NormalizedCommission) +
			(WeightExpectedReturn * item.NormalizedReturn)

		// Clamp score to [0.00, 100.00] and round to 2 decimal places
		clampedScore := math.Max(0.0, math.Min(100.0, rawScore))
		item.WinningScore = math.Round(clampedScore*100) / 100
	}

	return items
}

// calculatePercentile sorts a copy of values and extracts the percentile value
func calculatePercentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	idx := int(math.Ceil(p*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// findMinMaxClipped finds minimum and maximum values after capping at p99
func findMinMaxClipped(values []float64, cap float64) (float64, float64) {
	if len(values) == 0 {
		return 0.0, 0.0
	}
	min := math.Min(values[0], cap)
	max := min

	for _, v := range values {
		clipped := math.Min(v, cap)
		if clipped < min {
			min = clipped
		}
		if clipped > max {
			max = clipped
		}
	}
	return min, max
}

// normalize applies min-max scaling to project a value onto [0, 100]
func normalize(val, min, max float64) float64 {
	if max <= min {
		if max == 0.0 {
			return 0.0
		}
		return 50.0
	}
	norm := ((val - min) / (max - min)) * 100.0
	if norm < 0.0 {
		return 0.0
	}
	if norm > 100.0 {
		return 100.0
	}
	return math.Round(norm*100) / 100
}
