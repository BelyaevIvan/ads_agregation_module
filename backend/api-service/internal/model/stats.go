package model

type Stats struct {
	TotalListings   int `json:"total_listings"`
	ActiveListings  int `json:"active_listings"`
	HiddenListings  int `json:"hidden_listings"`
	TotalUsers      int `json:"total_users"`
	ActiveSources   int `json:"active_sources"`
	NewListingsToday int `json:"new_listings_today"`
	NewUsersWeek    int `json:"new_users_week"`
}

type ListingsByDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type BrandCount struct {
	Brand string `json:"brand"`
	Count int    `json:"count"`
}

type CityCount struct {
	City  string `json:"city"`
	Count int    `json:"count"`
}
