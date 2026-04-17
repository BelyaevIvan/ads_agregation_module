package repository

import (
	"database/sql"
	"errors"

	"brandhunt/api-service/internal/model"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetByEmail(email string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, full_name, phone, tg_link, vk_link, role, created_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Phone, &u.TgLink, &u.VkLink, &u.Role, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

func (r *UserRepo) GetByID(id string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, full_name, phone, tg_link, vk_link, role, created_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Phone, &u.TgLink, &u.VkLink, &u.Role, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

func (r *UserRepo) Create(email, passwordHash string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`INSERT INTO users (email, password_hash) VALUES ($1, $2)
		 RETURNING id, email, password_hash, full_name, phone, tg_link, vk_link, role, created_at`,
		email, passwordHash,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Phone, &u.TgLink, &u.VkLink, &u.Role, &u.CreatedAt)
	return u, err
}

func (r *UserRepo) UpdateProfile(id string, req *model.UpdateProfileRequest) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`UPDATE users SET
			full_name = COALESCE($2, full_name),
			phone     = COALESCE($3, phone),
			tg_link   = COALESCE($4, tg_link),
			vk_link   = COALESCE($5, vk_link)
		 WHERE id = $1
		 RETURNING id, email, password_hash, full_name, phone, tg_link, vk_link, role, created_at`,
		id, req.FullName, req.Phone, req.TgLink, req.VkLink,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Phone, &u.TgLink, &u.VkLink, &u.Role, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return u, err
}
