package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/usecase"
)

type mockHealthRepo struct {
	err error
}

func (m *mockHealthRepo) Ping(ctx context.Context) error {
	return m.err
}

func TestHealthUsecase_Success(t *testing.T) {
	mockRepo := &mockHealthRepo{err: nil}
	uc := usecase.NewHealthUsecase(mockRepo)

	status, err := uc.Check(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if status.Status != "ok" || status.Database != "connected" {
		t.Errorf("unexpected status: %+v", status)
	}
}

func TestHealthUsecase_Failure(t *testing.T) {
	mockRepo := &mockHealthRepo{err: errors.New("db down")}
	uc := usecase.NewHealthUsecase(mockRepo)

	status, err := uc.Check(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if status.Status != "degraded" || status.Database != "disconnected" {
		t.Errorf("unexpected status on failure: %+v", status)
	}
}
