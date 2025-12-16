package utils

import (
	"errors"
	"time"
)

func ValidateFutureTime(value interface{}) error {
	t, ok := value.(time.Time)
	if !ok {
		return errors.New("invalid time value")
	}

	if t.IsZero() {
		return errors.New("send_at is required")
	}

	if t.Before(time.Now()) {
		return errors.New("send_at must be in the future")
	}

	return nil
}
