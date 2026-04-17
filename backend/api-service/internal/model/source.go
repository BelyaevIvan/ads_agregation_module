package model

import "time"

type Source struct {
	ID            string    `json:"id"`
	Platform      string    `json:"platform"`
	ExternalID    string    `json:"external_id"`
	Title         *string   `json:"title"`
	IsActive      bool      `json:"is_active"`
	ListingsCount int       `json:"listings_count,omitempty"`
	AddedAt       time.Time `json:"added_at"`
}

type CreateSourceRequest struct {
	Platform   string  `json:"platform"`
	ExternalID string  `json:"external_id"`
	Title      *string `json:"title"`
}

type ToggleSourceRequest struct {
	IsActive *bool `json:"is_active"`
}
