package models

import (
	"encoding/json"
	"time"
)

type Item struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:""`
	Role     string `json:"role"`
}

type History struct {
	ID        int64           `json:"id"`
	ItemID    int64           `json:"item_id"`
	UserID    int64           `json:"user_id"`
	Action    string          `json:"action"`
	OldData   json.RawMessage `json:"old_data"`
	NewData   json.RawMessage `json:"new_data"`
	Timestamp time.Time       `json:"timestamp"`
}
