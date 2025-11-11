package model

import "time"

type Status string

const (
	StatusScheduled Status = "scheduled"
	StatusSent      Status = "sent"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

type Notification struct {
	ID        int64     `db:"id" json:"id"`
	Recipient string    `db:"recipient" json:"recipient"`
	Channel   string    `db:"channel" json:"channel"`
	Message   string    `db:"message" json:"message"`
	SendAt    time.Time `db:"send_at" json:"send_at"`
	Status    Status    `db:"status" json:"status"`
	Attempts  int       `db:"attempts" json:"attempts"`
	LastError *string   `db:"last_error" json:"last_error,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
