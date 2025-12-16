package model

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	"gitlab.com/arkine/l3/1/internal/utils"
)

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

func (n Notification) Validate() error {
	return validation.ValidateStruct(&n,
		validation.Field(&n.Recipient,
			validation.Required,
			is.Email,
		),
		validation.Field(&n.Channel,
			validation.Required,
			validation.In("email", "telegram", "simulated"),
		),
		validation.Field(&n.Message,
			validation.Required,
			validation.Length(1, 1000),
		),
		validation.Field(&n.SendAt,
			validation.Required,
			validation.By(utils.ValidateFutureTime),
		),
	)
}
