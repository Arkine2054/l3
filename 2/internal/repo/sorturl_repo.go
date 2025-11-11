package repo

import (
	"database/sql"

	"gitlab.com/arkine/l3/2/internal/models"
)

type ShortURLRepo struct {
	db *sql.DB
}

func NewShortURLRepo(db *sql.DB) *ShortURLRepo {
	return &ShortURLRepo{db: db}
}

func (r *ShortURLRepo) Create(alias, target string) (int, error) {
	var id int
	err := r.db.QueryRow(
		`INSERT INTO short_urls(alias, target) VALUES($1,$2) RETURNING id`, alias, target,
	).Scan(&id)
	return id, err
}

func (r *ShortURLRepo) FindByAlias(alias string) (*models.ShortURL, error) {
	var s models.ShortURL
	err := r.db.QueryRow(`SELECT id, alias, target, created_at FROM short_urls WHERE alias=$1`, alias).
		Scan(&s.ID, &s.Alias, &s.Target, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}
