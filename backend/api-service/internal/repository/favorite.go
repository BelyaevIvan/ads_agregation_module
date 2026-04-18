package repository

import (
	"database/sql"

	"brandhunt/api-service/internal/model"

	"github.com/lib/pq"
)

type FavoriteRepo struct {
	db *sql.DB
}

func NewFavoriteRepo(db *sql.DB) *FavoriteRepo {
	return &FavoriteRepo{db: db}
}

func (r *FavoriteRepo) List(userID string, limit, offset int) ([]model.FavoriteWithListing, int, error) {
	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM favorites WHERE user_id = $1`, userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
		SELECT f.id, f.saved_at,
		       l.id, l.brand, l.model, l.price, l.city,
		       l.size_rus, l.size_eu, l.size_us,
		       (SELECT photo_url FROM listing_photos WHERE listing_id = l.id AND is_cover = TRUE LIMIT 1),
		       s.platform, l.is_hidden
		FROM favorites f
		JOIN listings l ON f.listing_id = l.id
		LEFT JOIN sources s ON l.source_id = s.id
		WHERE f.user_id = $1
		ORDER BY f.saved_at DESC
		LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.FavoriteWithListing
	for rows.Next() {
		var f model.FavoriteWithListing
		err := rows.Scan(
			&f.ID, &f.SavedAt,
			&f.Listing.ID, &f.Listing.Brand, &f.Listing.Model, &f.Listing.Price, &f.Listing.City,
			pq.Array(&f.Listing.SizeRus), pq.Array(&f.Listing.SizeEU), pq.Array(&f.Listing.SizeUS),
			&f.Listing.CoverPhotoURL, &f.Listing.Platform, &f.Listing.IsHidden,
		)
		if err != nil {
			return nil, 0, err
		}
		if f.Listing.SizeRus == nil {
			f.Listing.SizeRus = []string{}
		}
		if f.Listing.SizeEU == nil {
			f.Listing.SizeEU = []string{}
		}
		if f.Listing.SizeUS == nil {
			f.Listing.SizeUS = []string{}
		}
		rewritePhotoURLPtr(f.Listing.CoverPhotoURL)
		items = append(items, f)
	}
	if items == nil {
		items = []model.FavoriteWithListing{}
	}
	return items, total, rows.Err()
}

// Add inserts a favorite. Returns the favorite and whether it was newly created.
func (r *FavoriteRepo) Add(userID, listingID string) (*model.Favorite, bool, error) {
	var f model.Favorite
	// Try insert; on conflict return existing
	err := r.db.QueryRow(`
		INSERT INTO favorites (user_id, listing_id) VALUES ($1, $2)
		ON CONFLICT (user_id, listing_id) DO NOTHING
		RETURNING id, listing_id, saved_at`, userID, listingID,
	).Scan(&f.ID, &f.ListingID, &f.SavedAt)

	if err == sql.ErrNoRows {
		// Conflict — already exists, fetch existing
		err = r.db.QueryRow(
			`SELECT id, listing_id, saved_at FROM favorites WHERE user_id = $1 AND listing_id = $2`,
			userID, listingID,
		).Scan(&f.ID, &f.ListingID, &f.SavedAt)
		if err != nil {
			return nil, false, err
		}
		return &f, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &f, true, nil
}

func (r *FavoriteRepo) Remove(userID, listingID string) (bool, error) {
	res, err := r.db.Exec(
		`DELETE FROM favorites WHERE user_id = $1 AND listing_id = $2`, userID, listingID)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}
