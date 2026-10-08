package domain

import (
	"context"
	"time"
)

// CrawlStatus represents the lifecycle state of a scraper crawl execution
type CrawlStatus string

const (
	CrawlStatusRunning CrawlStatus = "RUNNING"
	CrawlStatusSuccess CrawlStatus = "SUCCESS"
	CrawlStatusPartial CrawlStatus = "PARTIAL"
	CrawlStatusFailed  CrawlStatus = "FAILED"
)

// CrawlLog tracks the audit trail of ingestion cycles
type CrawlLog struct {
	ID                    string      `json:"id"`
	CrawlRound            string      `json:"crawl_round"`
	Status                CrawlStatus `json:"status"`
	TotalPagesRequested   int         `json:"total_pages_requested"`
	TotalPagesSuccess     int         `json:"total_pages_success"`
	RawProductsScraped    int         `json:"raw_products_scraped"`
	FilteredProductsSaved int         `json:"filtered_products_saved"`
	HTTPErrorCode         *int        `json:"http_error_code,omitempty"`
	ErrorMessage          *string     `json:"error_message,omitempty"`
	StartedAt             time.Time   `json:"started_at"`
	FinishedAt            *time.Time  `json:"finished_at,omitempty"`
}

// CrawlLogRepository defines persistence operations for crawl logs
type CrawlLogRepository interface {
	GetByID(ctx context.Context, id string) (*CrawlLog, error)
	UpdateStatus(ctx context.Context, id string, status CrawlStatus, finishedAt time.Time, errMsg *string) error
}
