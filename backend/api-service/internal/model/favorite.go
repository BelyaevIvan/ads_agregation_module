package model

import "time"

type Favorite struct {
	ID        string    `json:"id"`
	ListingID string    `json:"listing_id"`
	SavedAt   time.Time `json:"saved_at"`
}

type FavoriteWithListing struct {
	ID      string          `json:"id"`
	Listing FavoriteListingItem `json:"listing"`
	SavedAt time.Time       `json:"saved_at"`
}

type FavoriteListingItem struct {
	ID            string   `json:"id"`
	Brand         *string  `json:"brand"`
	Model         *string  `json:"model"`
	Price         *float64 `json:"price"`
	City          *string  `json:"city"`
	SizeRus       []string `json:"size_rus"`
	SizeEU        []string `json:"size_eu"`
	SizeUS        []string `json:"size_us"`
	CoverPhotoURL *string  `json:"cover_photo_url"`
	Platform      *string  `json:"platform"`
	IsHidden      bool     `json:"is_hidden"`
}

type AddFavoriteRequest struct {
	ListingID string `json:"listing_id"`
}
