package repository

import (
	"database/sql"
	"fmt"
	"sync"

	"brandhunt/api-service/internal/model"
)

type StatsRepo struct {
	db *sql.DB
}

func NewStatsRepo(db *sql.DB) *StatsRepo {
	return &StatsRepo{db: db}
}

func (r *StatsRepo) GetStats() (*model.Stats, error) {
	s := &model.Stats{}
	var wg sync.WaitGroup
	errs := make([]error, 7)

	queries := []struct {
		dest  *int
		query string
	}{
		{&s.TotalListings, `SELECT COUNT(*) FROM listings`},
		{&s.ActiveListings, `SELECT COUNT(*) FROM listings WHERE is_hidden = FALSE`},
		{&s.HiddenListings, `SELECT COUNT(*) FROM listings WHERE is_hidden = TRUE`},
		{&s.TotalUsers, `SELECT COUNT(*) FROM users`},
		{&s.ActiveSources, `SELECT COUNT(*) FROM sources WHERE is_active = TRUE`},
		{&s.NewListingsToday, `SELECT COUNT(*) FROM listings WHERE created_at >= NOW() - INTERVAL '1 day'`},
		{&s.NewUsersWeek, `SELECT COUNT(*) FROM users WHERE created_at >= NOW() - INTERVAL '7 days'`},
	}

	wg.Add(len(queries))
	for i, q := range queries {
		go func(idx int, dest *int, query string) {
			defer wg.Done()
			errs[idx] = r.db.QueryRow(query).Scan(dest)
		}(i, q.dest, q.query)
	}
	wg.Wait()

	for _, e := range errs {
		if e != nil {
			return nil, e
		}
	}
	return s, nil
}

func (r *StatsRepo) ListingsByDay(days int) ([]model.ListingsByDay, error) {
	rows, err := r.db.Query(fmt.Sprintf(`
		SELECT DATE(created_at) AS date, COUNT(*) AS count
		FROM listings
		WHERE created_at >= NOW() - INTERVAL '%d days'
		GROUP BY DATE(created_at)
		ORDER BY date DESC`, days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.ListingsByDay
	for rows.Next() {
		var it model.ListingsByDay
		if err := rows.Scan(&it.Date, &it.Count); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if items == nil {
		items = []model.ListingsByDay{}
	}
	return items, rows.Err()
}

func (r *StatsRepo) TopBrands(limit int) ([]model.BrandCount, error) {
	rows, err := r.db.Query(`
		SELECT brand, COUNT(*) AS count
		FROM listings
		WHERE brand IS NOT NULL AND is_hidden = FALSE
		GROUP BY brand
		ORDER BY count DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.BrandCount
	for rows.Next() {
		var it model.BrandCount
		if err := rows.Scan(&it.Brand, &it.Count); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if items == nil {
		items = []model.BrandCount{}
	}
	return items, rows.Err()
}

func (r *StatsRepo) TopCities(limit int) ([]model.CityCount, error) {
	rows, err := r.db.Query(`
		SELECT city, COUNT(*) AS count
		FROM listings
		WHERE city IS NOT NULL AND is_hidden = FALSE
		GROUP BY city
		ORDER BY count DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.CityCount
	for rows.Next() {
		var it model.CityCount
		if err := rows.Scan(&it.City, &it.Count); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if items == nil {
		items = []model.CityCount{}
	}
	return items, rows.Err()
}
