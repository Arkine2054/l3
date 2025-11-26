package models

import "time"

type Kind string

const (
	KindIncome  Kind = "income"
	KindExpense Kind = "expense"
)

type Sale struct {
	ID        int64     `json:"id"`
	Kind      Kind      `json:"kind"`
	Amount    float64   `json:"amount"`
	Category  string    `json:"category"`
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Analytics struct {
	From         *time.Time `json:"from,omitempty"`
	To           *time.Time `json:"to,omitempty"`
	Count        int64      `json:"count"`
	Sum          float64    `json:"sum"`
	Avg          float64    `json:"avg"`
	Median       float64    `json:"median"`
	Percentile90 float64    `json:"percentile_90"`
}
