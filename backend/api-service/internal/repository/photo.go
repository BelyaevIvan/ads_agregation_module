package repository

import (
	"database/sql"

	"brandhunt/api-service/internal/model"
)

type PhotoRepo struct {
	db *sql.DB
}

func NewPhotoRepo(db *sql.DB) *PhotoRepo {
	return &PhotoRepo{db: db}
}

func (r *PhotoRepo) GetByIDAndListing(photoID, listingID string) (*model.ListingPhoto, error) {
	p := &model.ListingPhoto{}
	var isCover bool
	err := r.db.QueryRow(
		`SELECT id, photo_url, is_cover, sort_order FROM listing_photos WHERE id = $1 AND listing_id = $2`,
		photoID, listingID,
	).Scan(&p.ID, &p.URL, &isCover, &p.SortOrder)
	p.IsCover = isCover
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func (r *PhotoRepo) Delete(photoID string) error {
	_, err := r.db.Exec(`DELETE FROM listing_photos WHERE id = $1`, photoID)
	return err
}

func (r *PhotoRepo) PromoteNextCover(listingID string) (*string, error) {
	var nextID string
	err := r.db.QueryRow(
		`SELECT id FROM listing_photos WHERE listing_id = $1 ORDER BY sort_order ASC LIMIT 1`,
		listingID,
	).Scan(&nextID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_, err = r.db.Exec(`UPDATE listing_photos SET is_cover = TRUE WHERE id = $1`, nextID)
	if err != nil {
		return nil, err
	}
	return &nextID, nil
}
