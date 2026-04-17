package repository

import (
	"database/sql"

	"brandhunt/api-service/internal/model"
)

type SourceRepo struct {
	db *sql.DB
}

func NewSourceRepo(db *sql.DB) *SourceRepo {
	return &SourceRepo{db: db}
}

func (r *SourceRepo) List(limit, offset int) ([]model.Source, int, error) {
	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM sources`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
		SELECT s.id, s.platform, s.external_id, s.title, s.is_active, s.added_at,
		       (SELECT COUNT(*) FROM listings WHERE source_id = s.id)
		FROM sources s
		ORDER BY s.added_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.Source
	for rows.Next() {
		var s model.Source
		if err := rows.Scan(&s.ID, &s.Platform, &s.ExternalID, &s.Title, &s.IsActive, &s.AddedAt, &s.ListingsCount); err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	if items == nil {
		items = []model.Source{}
	}
	return items, total, rows.Err()
}

func (r *SourceRepo) Create(platform, externalID string, title *string) (*model.Source, error) {
	s := &model.Source{}
	err := r.db.QueryRow(`
		INSERT INTO sources (platform, external_id, title) VALUES ($1, $2, $3)
		RETURNING id, platform, external_id, title, is_active, added_at`,
		platform, externalID, title,
	).Scan(&s.ID, &s.Platform, &s.ExternalID, &s.Title, &s.IsActive, &s.AddedAt)
	return s, err
}

func (r *SourceRepo) ExistsByPlatformAndExtID(platform, externalID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM sources WHERE platform = $1 AND external_id = $2)`,
		platform, externalID,
	).Scan(&exists)
	return exists, err
}

func (r *SourceRepo) Exists(id string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM sources WHERE id = $1)`, id).Scan(&exists)
	return exists, err
}

func (r *SourceRepo) SetActive(id string, isActive bool) error {
	res, err := r.db.Exec(`UPDATE sources SET is_active = $2 WHERE id = $1`, id, isActive)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
