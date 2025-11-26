package models

import (
	"database/sql"
	"time"
)

type Event struct {
	ID             int       `db:"id" json:"id"`
	Title          string    `db:"title" json:"title"`
	Date           time.Time `db:"date" json:"date"`
	TotalSeats     int       `db:"total_seats" json:"total_seats"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	AvailableSeats int       `db:"-" json:"available_seats"`
}

type Booking struct {
	ID        int          `db:"id" json:"id"`
	EventID   int          `db:"event_id" json:"event_id"`
	UserName  string       `db:"user_name" json:"user_name"`
	Paid      bool         `db:"paid" json:"paid"`
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt sql.NullTime `db:"updated_at" json:"updated_at"`
}

type EventInfo struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Date       time.Time `json:"date"`
	TotalSeats int       `json:"total_seats"`
	Booked     int       `json:"booked"`
	Available  int       `json:"available"`
}
