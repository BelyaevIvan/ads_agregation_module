package model

import "time"

type Listing struct {
	ID           string    `json:"id"`
	SourceID     *string   `json:"-"`
	OriginalText *string   `json:"original_text,omitempty"`
	PostURL      string    `json:"post_url,omitempty"`
	PostedAt     *time.Time `json:"posted_at"`
	Brand        *string   `json:"brand"`
	Model        *string   `json:"model"`
	Category     *string   `json:"category"`
	Color        *string   `json:"color"`
	Price        *float64  `json:"price"`
	City         *string   `json:"city"`
	Condition    *string   `json:"condition"`
	SizeRus      []string  `json:"size_rus"`
	SizeUS       []string  `json:"size_us"`
	SizeEU       []string  `json:"size_eu"`
	IsHidden     bool      `json:"is_hidden,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type ListingListItem struct {
	ID            string     `json:"id"`
	Brand         *string    `json:"brand"`
	Model         *string    `json:"model"`
	Category      *string    `json:"category"`
	Color         *string    `json:"color"`
	Price         *float64   `json:"price"`
	City          *string    `json:"city"`
	Condition     *string    `json:"condition"`
	SizeRus       []string   `json:"size_rus"`
	SizeUS        []string   `json:"size_us"`
	SizeEU        []string   `json:"size_eu"`
	CoverPhotoURL *string    `json:"cover_photo_url"`
	Platform      *string    `json:"platform"`
	PostedAt      *time.Time `json:"posted_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

type ListingDetail struct {
	Listing
	Source *SourceInfo    `json:"source"`
	Photos []ListingPhoto `json:"photos"`
}

type SourceInfo struct {
	Platform   string  `json:"platform"`
	Title      *string `json:"title"`
	ExternalID string  `json:"external_id"`
}

type ListingPhoto struct {
	ID        string    `json:"id,omitempty"`
	URL       string    `json:"url"`
	IsCover   bool      `json:"is_cover"`
	SortOrder int       `json:"sort_order"`
}

// Admin listing includes is_hidden and source_title
type AdminListingItem struct {
	ListingListItem
	SourceTitle *string `json:"source_title"`
	IsHidden    bool    `json:"is_hidden"`
}

type ListingSearchParams struct {
	Q             string
	Brands        []string
	Categories    []string
	Cities        []string
	Condition     string
	SizeRus       []string
	SizeEU        []string
	SizeUS        []string
	PriceMin      *float64
	PriceMax      *float64
	IncludeNoSize  bool
	IncludeNoPrice bool
	IncludeNoCity  bool
	Platforms     []string
	Sort          string
	Limit         int
	Offset        int
}

type AdminListingSearchParams struct {
	Q         string
	Status    string // "active", "hidden", ""
	Platforms []string
	Sort      string
	Limit     int
	Offset    int
}

type VisibilityRequest struct {
	IsHidden *bool `json:"is_hidden"`
}

type EditTextRequest struct {
	OriginalText string `json:"original_text"`
}
