package domain

// ScoringItem holds input snapshot metrics and calculated score outputs
type ScoringItem struct {
	SnapshotID           string  `json:"snapshot_id"`
	ProductID            string  `json:"product_id"`
	Price                float64 `json:"price"`
	CommissionRate       float64 `json:"commission_rate"`
	VelocityPerHour      float64 `json:"velocity_per_hour"`
	ExpectedReturn       float64 `json:"expected_return"`
	NormalizedVelocity   float64 `json:"normalized_velocity"`
	NormalizedCommission float64 `json:"normalized_commission"`
	NormalizedReturn     float64 `json:"normalized_return"`
	WinningScore         float64 `json:"winning_score"`
}

// ScoringUsecase defines the business contract for winning score calculations
type ScoringUsecase interface {
	ComputeExpectedReturn(price float64, commissionRate float64) float64
	CalculateBatchScores(items []*ScoringItem) []*ScoringItem
}
