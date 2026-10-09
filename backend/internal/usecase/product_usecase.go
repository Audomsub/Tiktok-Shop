package usecase

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

type productUsecase struct {
	repo    domain.ProductRepository
	favRepo domain.FavoriteRepository
}

// NewProductUsecase constructs a new ProductUsecase instance
func NewProductUsecase(repo domain.ProductRepository, favRepo ...domain.FavoriteRepository) domain.ProductUsecase {
	var fRepo domain.FavoriteRepository
	if len(favRepo) > 0 {
		fRepo = favRepo[0]
	}
	return &productUsecase{repo: repo, favRepo: fRepo}
}

// GetCatalog queries the catalog using provided filters and calculates expected return per item
func (u *productUsecase) GetCatalog(ctx context.Context, filter domain.ProductCatalogFilter) (*domain.ProductCatalogResponse, error) {
	// Apply default values for pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 20
	} else if filter.Limit > 100 {
		filter.Limit = 100
	}

	result, err := u.repo.ListCatalog(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving product catalog: %w", err)
	}

	if result == nil {
		return &domain.ProductCatalogResponse{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalCount: 0,
			TotalPages: 0,
			Items:      make([]*domain.ProductCatalogItem, 0),
		}, nil
	}

	// Fetch user's favorite product IDs if user is authenticated and favRepo is configured
	var favMap map[string]bool
	if u.favRepo != nil {
		if userID, ok := domain.GetUserIDFromContext(ctx); ok && userID != "" {
			if m, err := u.favRepo.GetUserFavoriteProductIDs(ctx, userID); err == nil {
				favMap = m
			}
		}
	}

	// Calculate Expected Return (THB per unit) = Price * (CommissionRate / 100) and enrich IsFavorited
	for _, item := range result.Items {
		expectedReturn := item.Price * (item.CommissionRate / 100.0)
		item.ExpectedReturn = math.Round(expectedReturn*100) / 100

		if favMap != nil && favMap[item.ID] {
			item.IsFavorited = true
		}
	}

	return result, nil
}

// ExportCatalogCSV streams matching catalog items as CSV with UTF-8 BOM encoding
func (u *productUsecase) ExportCatalogCSV(ctx context.Context, filter domain.ProductCatalogFilter, w io.Writer) error {
	// 1. Write UTF-8 Byte Order Mark (\xEF\xBB\xBF) so Excel/Numbers display Thai characters cleanly
	if _, err := w.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return fmt.Errorf("failed writing UTF-8 BOM: %w", err)
	}

	csvWriter := csv.NewWriter(w)

	// 2. Write CSV Header row
	header := []string{
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
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed writing CSV header: %w", err)
	}
	csvWriter.Flush()

	flusher, hasFlusher := w.(http.Flusher)
	if hasFlusher {
		flusher.Flush()
	}

	// 3. Stream data rows without buffering entire dataset in memory
	err := u.repo.StreamCatalog(ctx, filter, func(item *domain.ProductCatalogItem) error {
		expectedReturn := item.Price * (item.CommissionRate / 100.0)
		item.ExpectedReturn = math.Round(expectedReturn*100) / 100

		record := []string{
			item.ID,
			item.Name,
			item.CategoryName,
			strconv.FormatFloat(item.Price, 'f', 2, 64),
			strconv.FormatFloat(item.CommissionRate, 'f', 2, 64) + "%",
			strconv.FormatFloat(item.ExpectedReturn, 'f', 2, 64),
			strconv.Itoa(item.TotalSales),
			strconv.FormatFloat(item.VelocityPerHour, 'f', 2, 64),
			strconv.FormatFloat(item.WinningScore, 'f', 2, 64),
			item.ProductURL,
		}

		if err := csvWriter.Write(record); err != nil {
			return err
		}
		csvWriter.Flush()
		if hasFlusher {
			flusher.Flush()
		}
		return csvWriter.Error()
	})

	if err != nil {
		return fmt.Errorf("failed during catalog CSV stream: %w", err)
	}

	csvWriter.Flush()
	return csvWriter.Error()
}
