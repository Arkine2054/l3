package models

import "time"

type ShortURL struct {
	ID        int       `json:"id"`
	Alias     string    `json:"alias"`
	Target    string    `json:"target"`
	CreatedAt time.Time `json:"created_at"`
}

type Click struct {
	ID         int       `json:"id"`
	ShortURLID int       `json:"short_url_id"`
	UserAgent  string    `json:"user_agent"`
	IP         string    `json:"ip"`
	Referer    string    `json:"referer"`
	CreatedAt  time.Time `json:"created_at"`
}
