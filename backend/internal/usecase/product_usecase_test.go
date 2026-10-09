package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"testing"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

type mockProductRepo struct {
	listCatalogFunc   func(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error)
	streamCatalogFunc func(ctx context.Context, filter domain.ProductCatalogFilter, onRow func(item *domain.ProductCatalogItem) error) error
}

func (m *mockProductRepo) ListCatalog(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error) {
	if m.listCatalogFunc != nil {
		return m.listCatalogFunc(ctx, filter)
	}
	return nil, nil
}

func (m *mockProductRepo) StreamCatalog(ctx context.Context, filter domain.ProductCatalogFilter, onRow func(item *domain.ProductCatalogItem) error) error {
	if m.streamCatalogFunc != nil {
		return m.streamCatalogFunc(ctx, filter, onRow)
	}
	return nil
}

func TestProductUsecase_GetCatalog(t *testing.T) {
	t.Run("Successfully returns catalog with expected return calculation", func(t *testing.T) {
		mockRepo := &mockProductRepo{
			listCatalogFunc: func(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error) {
				return &domain.ProductCatalogResponse{
					Page:       filter.Page,
					Limit:      filter.Limit,
					TotalCount: 1,
					TotalPages: 1,
					Items: []*domain.ProductCatalogItem{
						{
							ID:             "p1",
							Name:           "Bluetooth Speaker",
							Price:          800.0,
							CommissionRate: 15.0,
							WinningScore:   85.0,
							CreatedAt:      time.Now(),
						},
					},
				}, nil
			},
		}

		u := NewProductUsecase(mockRepo)
		resp, err := u.GetCatalog(context.Background(), domain.ProductCatalogFilter{Page: 1, Limit: 20})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.TotalCount != 1 {
			t.Fatalf("expected TotalCount 1, got %d", resp.TotalCount)
		}
		if len(resp.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(resp.Items))
		}

		item := resp.Items[0]
		expected := 120.0 // 800 * 0.15
		if item.ExpectedReturn != expected {
			t.Errorf("expected expected return %f, got %f", expected, item.ExpectedReturn)
		}
	})

	t.Run("Sanitizes invalid pagination parameters", func(t *testing.T) {
		var receivedFilter domain.ProductCatalogFilter
		mockRepo := &mockProductRepo{
			listCatalogFunc: func(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error) {
				receivedFilter = filter
				return &domain.ProductCatalogResponse{TotalCount: 0, Items: []*domain.ProductCatalogItem{}}, nil
			},
		}

		u := NewProductUsecase(mockRepo)
		_, err := u.GetCatalog(context.Background(), domain.ProductCatalogFilter{Page: -5, Limit: 500})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if receivedFilter.Page != 1 {
			t.Errorf("expected page 1, got %d", receivedFilter.Page)
		}
		if receivedFilter.Limit != 100 {
			t.Errorf("expected limit 100, got %d", receivedFilter.Limit)
		}
	})

	t.Run("Propagates database errors", func(t *testing.T) {
		mockRepo := &mockProductRepo{
			listCatalogFunc: func(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error) {
				return nil, errors.New("db query failed")
			},
		}

		u := NewProductUsecase(mockRepo)
		_, err := u.GetCatalog(context.Background(), domain.ProductCatalogFilter{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("Enriches IsFavorited when user is authenticated with favorites", func(t *testing.T) {
		mockRepo := &mockProductRepo{
			listCatalogFunc: func(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error) {
				return &domain.ProductCatalogResponse{
					TotalCount: 2,
					Items: []*domain.ProductCatalogItem{
						{ID: "prod-fav", Name: "Favorited Item"},
						{ID: "prod-regular", Name: "Regular Item"},
					},
				}, nil
			},
		}

		favRepo := newInMemoryFavoriteRepo()
		_ = favRepo.Add(context.Background(), "user-alice", "prod-fav")

		u := NewProductUsecase(mockRepo, favRepo)

		// Context with user alice
		ctxAlice := context.WithValue(context.Background(), domain.UserIDContextKey, "user-alice")
		respAlice, err := u.GetCatalog(ctxAlice, domain.ProductCatalogFilter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !respAlice.Items[0].IsFavorited {
			t.Errorf("expected prod-fav to have IsFavorited=true for user-alice")
		}
		if respAlice.Items[1].IsFavorited {
			t.Errorf("expected prod-regular to have IsFavorited=false for user-alice")
		}

		// Context without user (unauthenticated)
		respAnon, err := u.GetCatalog(context.Background(), domain.ProductCatalogFilter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if respAnon.Items[0].IsFavorited || respAnon.Items[1].IsFavorited {
			t.Errorf("expected anonymous user to have all IsFavorited=false")
		}
	})
}

func TestProductUsecase_ExportCatalogCSV(t *testing.T) {
	t.Run("Streams CSV with UTF-8 BOM, required headers, and correct values", func(t *testing.T) {
		mockRepo := &mockProductRepo{
			streamCatalogFunc: func(ctx context.Context, filter domain.ProductCatalogFilter, onRow func(item *domain.ProductCatalogItem) error) error {
				item := &domain.ProductCatalogItem{
					ID:              "p-thai-1",
					Name:            "เสื้อยืดผ้าฝ้ายระบายอากาศ TikTok Viral",
					CategoryName:    "Fashion",
					Price:           290.00,
					CommissionRate:  20.00,
					TotalSales:      1500,
					VelocityPerHour: 12.50,
					WinningScore:    88.5,
					ProductURL:      "https://shop.tiktok.com/p1",
					CreatedAt:       time.Now(),
				}
				return onRow(item)
			},
		}

		u := NewProductUsecase(mockRepo)
		var buf bytes.Buffer
		err := u.ExportCatalogCSV(context.Background(), domain.ProductCatalogFilter{}, &buf)
		if err != nil {
			t.Fatalf("unexpected export error: %v", err)
		}

		data := buf.Bytes()
		// 1. Check UTF-8 BOM (\xEF\xBB\xBF)
		if len(data) < 3 || data[0] != 0xEF || data[1] != 0xBB || data[2] != 0xBF {
			t.Fatalf("expected UTF-8 BOM at beginning of file, got %v", data[:3])
		}

		// Parse CSV without BOM
		r := csv.NewReader(bytes.NewReader(data[3:]))
		records, err := r.ReadAll()
		if err != nil {
			t.Fatalf("failed reading CSV content: %v", err)
		}

		if len(records) != 2 {
			t.Fatalf("expected 2 records (1 header + 1 row), got %d", len(records))
		}

		// 2. Check 10 required header columns
		expectedHeaders := []string{
			"Product ID",
			"Name",
			"Category",
			"Price",
			"Commission Rate",
			"Expected Return THB",
			"Total Sales",
			"Velocity Per Hour",
			"Winning Score",
			"TikTok Product URL",
		}
		for i, h := range expectedHeaders {
			if records[0][i] != h {
				t.Errorf("header col %d mismatch: expected %q, got %q", i, h, records[0][i])
			}
		}

		// 3. Check data row
		row := records[1]
		if row[0] != "p-thai-1" {
			t.Errorf("expected product id p-thai-1, got %q", row[0])
		}
		if row[1] != "เสื้อยืดผ้าฝ้ายระบายอากาศ TikTok Viral" {
			t.Errorf("expected Thai product name, got %q", row[1])
		}
		if row[2] != "Fashion" {
			t.Errorf("expected Category Fashion, got %q", row[2])
		}
		if row[3] != "290.00" {
			t.Errorf("expected Price 290.00, got %q", row[3])
		}
		if row[4] != "20.00%" {
			t.Errorf("expected Commission Rate 20.00%%, got %q", row[4])
		}
		if row[5] != "58.00" { // 290 * 0.20 = 58.00
			t.Errorf("expected Expected Return 58.00, got %q", row[5])
		}
		if row[6] != "1500" {
			t.Errorf("expected Total Sales 1500, got %q", row[6])
		}
		if row[7] != "12.50" {
			t.Errorf("expected Velocity 12.50, got %q", row[7])
		}
		if row[8] != "88.50" {
			t.Errorf("expected Winning Score 88.50, got %q", row[8])
		}
		if row[9] != "https://shop.tiktok.com/p1" {
			t.Errorf("expected URL https://shop.tiktok.com/p1, got %q", row[9])
		}
	})

	t.Run("Propagates stream error from repository", func(t *testing.T) {
		mockRepo := &mockProductRepo{
			streamCatalogFunc: func(ctx context.Context, filter domain.ProductCatalogFilter, onRow func(item *domain.ProductCatalogItem) error) error {
				return errors.New("stream db disconnected")
			},
		}

		u := NewProductUsecase(mockRepo)
		var buf bytes.Buffer
		err := u.ExportCatalogCSV(context.Background(), domain.ProductCatalogFilter{}, &buf)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
