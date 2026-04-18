package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"brandhunt/api-service/internal/model"

	"github.com/lib/pq"
)

type ListingRepo struct {
	db *sql.DB
}

func NewListingRepo(db *sql.DB) *ListingRepo {
	return &ListingRepo{db: db}
}

func (r *ListingRepo) Search(p *model.ListingSearchParams) ([]model.ListingListItem, int, error) {
	var (
		conditions []string
		args       []any
		argIdx     = 1
	)

	conditions = append(conditions, "l.is_hidden = FALSE")

	if p.Q != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(l.brand ILIKE $%d OR l.model ILIKE $%d OR l.original_text ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, "%"+p.Q+"%")
		argIdx++
	}

	// Array filters for simple string columns
	type simpleFilter struct {
		col  string
		vals []string
	}
	for _, f := range []simpleFilter{
		{"l.brand", p.Brands},
		{"l.category", p.Categories},
		{"l.city", p.Cities},
		{"s.platform", p.Platforms},
	} {
		if len(f.vals) > 0 {
			conditions = append(conditions, fmt.Sprintf("%s = ANY($%d)", f.col, argIdx))
			args = append(args, pq.Array(f.vals))
			argIdx++
		}
	}

	if p.Condition != "" {
		conditions = append(conditions, fmt.Sprintf("l.condition = $%d", argIdx))
		args = append(args, p.Condition)
		argIdx++
	}

	// Size array filters — check intersection
	type arrayFilter struct {
		col  string
		vals []string
	}
	for _, f := range []arrayFilter{
		{"l.size_rus", p.SizeRus},
		{"l.size_eu", p.SizeEU},
		{"l.size_us", p.SizeUS},
	} {
		if len(f.vals) > 0 {
			if p.IncludeNoSize {
				conditions = append(conditions, fmt.Sprintf("(%s && $%d OR %s IS NULL)", f.col, argIdx, f.col))
			} else {
				conditions = append(conditions, fmt.Sprintf("%s && $%d", f.col, argIdx))
			}
			args = append(args, pq.Array(f.vals))
			argIdx++
		}
	}

	// Price range
	if p.PriceMin != nil {
		if p.IncludeNoPrice {
			conditions = append(conditions, fmt.Sprintf("(l.price >= $%d OR l.price IS NULL)", argIdx))
		} else {
			conditions = append(conditions, fmt.Sprintf("l.price >= $%d", argIdx))
		}
		args = append(args, *p.PriceMin)
		argIdx++
	}
	if p.PriceMax != nil {
		if p.IncludeNoPrice {
			conditions = append(conditions, fmt.Sprintf("(l.price <= $%d OR l.price IS NULL)", argIdx))
		} else {
			conditions = append(conditions, fmt.Sprintf("l.price <= $%d", argIdx))
		}
		args = append(args, *p.PriceMax)
		argIdx++
	}

	// City include_no_city
	if len(p.Cities) > 0 && p.IncludeNoCity {
		// Re-do city filter to include NULL
		// Remove the previously added city condition and replace
		newConds := make([]string, 0, len(conditions))
		for _, c := range conditions {
			if !strings.Contains(c, "l.city = ANY") {
				newConds = append(newConds, c)
			}
		}
		// Find the arg index for city — it's already in args, find position
		// Simpler: just rebuild with OR NULL
		for i, a := range args {
			if arr, ok := a.(pq.StringArray); ok {
				found := false
				for _, v := range arr {
					for _, cv := range p.Cities {
						if v == cv {
							found = true
							break
						}
					}
					if found {
						break
					}
				}
				if found {
					newConds = append(newConds, fmt.Sprintf("(l.city = ANY($%d) OR l.city IS NULL)", i+1))
					break
				}
			}
		}
		conditions = newConds
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	orderBy := "ORDER BY l.posted_at DESC NULLS LAST"
	switch p.Sort {
	case "price_asc":
		orderBy = "ORDER BY l.price ASC NULLS LAST"
	case "price_desc":
		orderBy = "ORDER BY l.price DESC NULLS LAST"
	}

	baseQuery := fmt.Sprintf(`
		FROM listings l
		LEFT JOIN sources s ON l.source_id = s.id
		%s`, where)

	// Count
	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count listings: %w", err)
	}

	// Items
	dataQuery := fmt.Sprintf(`
		SELECT l.id, l.brand, l.model, l.category, l.color, l.price, l.city, l.condition,
		       l.size_rus, l.size_us, l.size_eu,
		       (SELECT photo_url FROM listing_photos WHERE listing_id = l.id AND is_cover = TRUE LIMIT 1),
		       s.platform, l.posted_at, l.created_at
		%s %s LIMIT $%d OFFSET $%d`,
		baseQuery, orderBy, argIdx, argIdx+1)
	args = append(args, p.Limit, p.Offset)

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query listings: %w", err)
	}
	defer rows.Close()

	var items []model.ListingListItem
	for rows.Next() {
		var it model.ListingListItem
		err := rows.Scan(
			&it.ID, &it.Brand, &it.Model, &it.Category, &it.Color, &it.Price, &it.City, &it.Condition,
			pq.Array(&it.SizeRus), pq.Array(&it.SizeUS), pq.Array(&it.SizeEU),
			&it.CoverPhotoURL, &it.Platform, &it.PostedAt, &it.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan listing: %w", err)
		}
		if it.SizeRus == nil {
			it.SizeRus = []string{}
		}
		if it.SizeUS == nil {
			it.SizeUS = []string{}
		}
		if it.SizeEU == nil {
			it.SizeEU = []string{}
		}
		rewritePhotoURLPtr(it.CoverPhotoURL)
		items = append(items, it)
	}
	if items == nil {
		items = []model.ListingListItem{}
	}
	return items, total, rows.Err()
}

func (r *ListingRepo) GetByID(id string) (*model.ListingDetail, error) {
	return r.getByID(id, false)
}

// GetByIDAdmin returns the listing regardless of is_hidden. For admin use only.
func (r *ListingRepo) GetByIDAdmin(id string) (*model.ListingDetail, error) {
	return r.getByID(id, true)
}

func (r *ListingRepo) getByID(id string, includeHidden bool) (*model.ListingDetail, error) {
	query := `
		SELECT l.id, l.source_id, l.original_text, l.post_url, l.posted_at,
		       l.brand, l.model, l.category, l.color, l.price, l.city, l.condition,
		       l.size_rus, l.size_us, l.size_eu, l.is_hidden, l.created_at
		FROM listings l WHERE l.id = $1`
	if !includeHidden {
		query += ` AND l.is_hidden = FALSE`
	}

	d := &model.ListingDetail{}
	var sourceID *string
	err := r.db.QueryRow(query, id).Scan(
		&d.ID, &sourceID, &d.OriginalText, &d.PostURL, &d.PostedAt,
		&d.Brand, &d.Model, &d.Category, &d.Color, &d.Price, &d.City, &d.Condition,
		pq.Array(&d.SizeRus), pq.Array(&d.SizeUS), pq.Array(&d.SizeEU), &d.IsHidden, &d.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if d.SizeRus == nil {
		d.SizeRus = []string{}
	}
	if d.SizeUS == nil {
		d.SizeUS = []string{}
	}
	if d.SizeEU == nil {
		d.SizeEU = []string{}
	}

	// Source
	if sourceID != nil {
		src := &model.SourceInfo{}
		err = r.db.QueryRow(
			`SELECT platform, title, external_id FROM sources WHERE id = $1`, *sourceID,
		).Scan(&src.Platform, &src.Title, &src.ExternalID)
		if err == nil {
			d.Source = src
		}
	}

	// Photos
	rows, err := r.db.Query(
		`SELECT id, photo_url, is_cover, sort_order FROM listing_photos WHERE listing_id = $1 ORDER BY sort_order ASC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	d.Photos = []model.ListingPhoto{}
	for rows.Next() {
		var p model.ListingPhoto
		if err := rows.Scan(&p.ID, &p.URL, &p.IsCover, &p.SortOrder); err != nil {
			return nil, err
		}
		p.URL = rewritePhotoURL(p.URL)
		d.Photos = append(d.Photos, p)
	}

	return d, rows.Err()
}

func (r *ListingRepo) Exists(id string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM listings WHERE id = $1)`, id).Scan(&exists)
	return exists, err
}

func (r *ListingRepo) SetVisibility(id string, isHidden bool) error {
	res, err := r.db.Exec(`UPDATE listings SET is_hidden = $2 WHERE id = $1`, id, isHidden)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ListingRepo) UpdateText(id, text string) error {
	res, err := r.db.Exec(`UPDATE listings SET original_text = $2 WHERE id = $1`, id, text)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ListingRepo) AdminSearch(p *model.AdminListingSearchParams) ([]model.AdminListingItem, int, error) {
	var (
		conditions []string
		args       []any
		argIdx     = 1
	)

	if p.Status == "active" {
		conditions = append(conditions, "l.is_hidden = FALSE")
	} else if p.Status == "hidden" {
		conditions = append(conditions, "l.is_hidden = TRUE")
	}

	if p.Q != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(l.brand ILIKE $%d OR l.model ILIKE $%d OR l.original_text ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, "%"+p.Q+"%")
		argIdx++
	}

	if len(p.Platforms) > 0 {
		conditions = append(conditions, fmt.Sprintf("s.platform = ANY($%d)", argIdx))
		args = append(args, pq.Array(p.Platforms))
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	orderBy := "ORDER BY l.posted_at DESC NULLS LAST"
	switch p.Sort {
	case "price_asc":
		orderBy = "ORDER BY l.price ASC NULLS LAST"
	case "price_desc":
		orderBy = "ORDER BY l.price DESC NULLS LAST"
	}

	baseQuery := fmt.Sprintf(`
		FROM listings l
		LEFT JOIN sources s ON l.source_id = s.id
		%s`, where)

	var total int
	err := r.db.QueryRow("SELECT COUNT(*) "+baseQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT l.id, l.brand, l.model, l.category, l.color, l.price, l.city, l.condition,
		       l.size_rus, l.size_us, l.size_eu,
		       (SELECT photo_url FROM listing_photos WHERE listing_id = l.id AND is_cover = TRUE LIMIT 1),
		       s.platform, l.posted_at, l.created_at, s.title, l.is_hidden
		%s %s LIMIT $%d OFFSET $%d`,
		baseQuery, orderBy, argIdx, argIdx+1)
	args = append(args, p.Limit, p.Offset)

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.AdminListingItem
	for rows.Next() {
		var it model.AdminListingItem
		err := rows.Scan(
			&it.ID, &it.Brand, &it.Model, &it.Category, &it.Color, &it.Price, &it.City, &it.Condition,
			pq.Array(&it.SizeRus), pq.Array(&it.SizeUS), pq.Array(&it.SizeEU),
			&it.CoverPhotoURL, &it.Platform, &it.PostedAt, &it.CreatedAt,
			&it.SourceTitle, &it.IsHidden,
		)
		if err != nil {
			return nil, 0, err
		}
		if it.SizeRus == nil {
			it.SizeRus = []string{}
		}
		if it.SizeUS == nil {
			it.SizeUS = []string{}
		}
		if it.SizeEU == nil {
			it.SizeEU = []string{}
		}
		rewritePhotoURLPtr(it.CoverPhotoURL)
		items = append(items, it)
	}
	if items == nil {
		items = []model.AdminListingItem{}
	}
	return items, total, rows.Err()
}
