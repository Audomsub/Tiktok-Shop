package usecase

import (
	"context"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

type healthUsecase struct {
	repo domain.HealthRepository
}

// NewHealthUsecase constructs a new HealthUsecase implementation
func NewHealthUsecase(repo domain.HealthRepository) domain.HealthUsecase {
	return &healthUsecase{repo: repo}
}

// Check evaluates the health of the system and its dependent database
func (u *healthUsecase) Check(ctx context.Context) (*domain.HealthStatus, error) {
	status := &domain.HealthStatus{
		Status:    "ok",
		Database:  "connected",
		Timestamp: time.Now().UTC(),
		Version:   "1.0.0",
	}

	if err := u.repo.Ping(ctx); err != nil {
		status.Status = "degraded"
		status.Database = "disconnected"
		return status, err
	}

	return status, nil
}
