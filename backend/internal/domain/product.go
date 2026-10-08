package domain

import (
	"time"
)

// Product represents the master catalog item from TikTok Shop / FastMoss
type Product struct {
	ID             string    `json:"id"`
	SourceID       string    `json:"source_id"`
	Name           string    `json:"name"`
	ImageURL       string    `json:"image_url"`
	ProductURL     string    `json:"product_url"`
	CommissionRate float64   `json:"commission_rate"`
	CategoryID     *string   `json:"category_id,omitempty"`
	CategoryName   string    `json:"category_name,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
