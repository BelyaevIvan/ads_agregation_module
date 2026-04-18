package repository

import (
	"database/sql"
	"sync"
)

type FiltersRepo struct {
	db *sql.DB
}

func NewFiltersRepo(db *sql.DB) *FiltersRepo {
	return &FiltersRepo{db: db}
}

// GetDistinctSizes возвращает все уникальные размеры из активных объявлений,
// в трёх размерных сетках. Запросы выполняются параллельно.
func (r *FiltersRepo) GetDistinctSizes() (rus, eu, us []string, err error) {
	queries := []struct {
		dest  *[]string
		query string
	}{
		{&rus, `SELECT DISTINCT unnest(size_rus) FROM listings
		        WHERE is_hidden = FALSE AND size_rus IS NOT NULL`},
		{&eu, `SELECT DISTINCT unnest(size_eu) FROM listings
		       WHERE is_hidden = FALSE AND size_eu IS NOT NULL`},
		{&us, `SELECT DISTINCT unnest(size_us) FROM listings
		       WHERE is_hidden = FALSE AND size_us IS NOT NULL`},
	}

	var wg sync.WaitGroup
	errs := make([]error, len(queries))

	wg.Add(len(queries))
	for i, q := range queries {
		go func(idx int, dest *[]string, query string) {
			defer wg.Done()
			rows, e := r.db.Query(query)
			if e != nil {
				errs[idx] = e
				return
			}
			defer rows.Close()
			for rows.Next() {
				var v sql.NullString
				if e := rows.Scan(&v); e != nil {
					errs[idx] = e
					return
				}
				if v.Valid && v.String != "" {
					*dest = append(*dest, v.String)
				}
			}
			errs[idx] = rows.Err()
		}(i, q.dest, q.query)
	}
	wg.Wait()

	for _, e := range errs {
		if e != nil {
			return nil, nil, nil, e
		}
	}
	if rus == nil {
		rus = []string{}
	}
	if eu == nil {
		eu = []string{}
	}
	if us == nil {
		us = []string{}
	}
	return rus, eu, us, nil
}
