package domain

import (
	"context"
	"time"
)

// HealthStatus represents the health status of the service and its dependencies
type HealthStatus struct {
	Status    string    `json:"status"`
	Database  string    `json:"database"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// HealthRepository defines the contract for infrastructure health checks
type HealthRepository interface {
	Ping(ctx context.Context) error
}

// HealthUsecase defines the business contract for checking system health
type HealthUsecase interface {
	Check(ctx context.Context) (*HealthStatus, error)
}
