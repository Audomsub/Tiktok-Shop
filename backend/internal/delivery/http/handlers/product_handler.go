package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
)

// ProductHandler handles product catalog requests
type ProductHandler struct {
	usecase domain.ProductUsecase
}

// NewProductHandler constructs a new ProductHandler instance
func NewProductHandler(u domain.ProductUsecase) *ProductHandler {
	return &ProductHandler{usecase: u}
}

// GetCatalog handles public GET /api/v1/products with dynamic query filters
func (h *ProductHandler) GetCatalog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	queryValues := r.URL.Query()

	filter := domain.ProductCatalogFilter{
		Query:      strings.TrimSpace(queryValues.Get("q")),
		CategoryID: strings.TrimSpace(queryValues.Get("category_id")),
		SortBy:     strings.TrimSpace(queryValues.Get("sort_by")),
		SortOrder:  strings.TrimSpace(queryValues.Get("sort_order")),
		Page:       1,
		Limit:      20,
	}

	if minPriceStr := queryValues.Get("min_price"); minPriceStr != "" {
		if val, err := strconv.ParseFloat(minPriceStr, 64); err == nil && val >= 0 {
			filter.MinPrice = &val
		}
	}

	if maxPriceStr := queryValues.Get("max_price"); maxPriceStr != "" {
		if val, err := strconv.ParseFloat(maxPriceStr, 64); err == nil && val >= 0 {
			filter.MaxPrice = &val
		}
	}

	if minCommStr := queryValues.Get("min_commission"); minCommStr != "" {
		if val, err := strconv.ParseFloat(minCommStr, 64); err == nil && val >= 0 {
			filter.MinCommission = &val
		}
	}

	if pageStr := queryValues.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}

	if limitStr := queryValues.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	resp, err := h.usecase.GetCatalog(r.Context(), filter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// ExportCatalog handles public GET /api/v1/products/export with streaming CSV response
func (h *ProductHandler) ExportCatalog(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()

	filter := domain.ProductCatalogFilter{
		Query:      strings.TrimSpace(queryValues.Get("q")),
		CategoryID: strings.TrimSpace(queryValues.Get("category_id")),
		SortBy:     strings.TrimSpace(queryValues.Get("sort_by")),
		SortOrder:  strings.TrimSpace(queryValues.Get("sort_order")),
	}

	if minPriceStr := queryValues.Get("min_price"); minPriceStr != "" {
		if val, err := strconv.ParseFloat(minPriceStr, 64); err == nil && val >= 0 {
			filter.MinPrice = &val
		}
	}

	if maxPriceStr := queryValues.Get("max_price"); maxPriceStr != "" {
		if val, err := strconv.ParseFloat(maxPriceStr, 64); err == nil && val >= 0 {
			filter.MaxPrice = &val
		}
	}

	if minCommStr := queryValues.Get("min_commission"); minCommStr != "" {
		if val, err := strconv.ParseFloat(minCommStr, 64); err == nil && val >= 0 {
			filter.MinCommission = &val
		}
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="tiktok_winning_products.csv"`)
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	if err := h.usecase.ExportCatalogCSV(r.Context(), filter, w); err != nil {
		// Response headers already written; logging is sufficient
		_ = err
	}
}
