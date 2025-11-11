package repo

import (
	"database/sql"

	"gitlab.com/arkine/l3/2/internal/models"
)

type ClickRepo struct {
	db *sql.DB
}

func NewClickRepo(db *sql.DB) *ClickRepo {
	return &ClickRepo{db: db}
}

func (r *ClickRepo) SaveClick(shortURLID int, userAgent, ip, referer string) error {
	_, err := r.db.Exec(`
		INSERT INTO clicks (short_url_id, user_agent, ip, referer)
		VALUES ($1, $2, $3, $4)
	`, shortURLID, userAgent, ip, referer)
	return err
}

func (r *ClickRepo) GetClicks(alias string) ([]models.Click, error) {
	rows, err := r.db.Query(`
		SELECT c.id, c.short_url_id, c.user_agent, c.ip, c.referer, c.created_at
		FROM clicks c
		JOIN short_urls s ON c.short_url_id = s.id
		WHERE s.alias = $1
		ORDER BY c.created_at DESC
	`, alias)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clicks []models.Click
	for rows.Next() {
		var c models.Click
		if err := rows.Scan(&c.ID, &c.ShortURLID, &c.UserAgent, &c.IP, &c.Referer, &c.CreatedAt); err != nil {
			return nil, err
		}
		clicks = append(clicks, c)
	}
	return clicks, nil
}
