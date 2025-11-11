package models

import "time"

type Comment struct {
	ID        int64      `json:"id"`
	ParentID  *int64     `json:"parent_id,omitempty"`
	Content   string     `json:"content"`
	Author    *string    `json:"author,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	Children  []*Comment `json:"children,omitempty"`
}
