package models

import "time"

type ImageRecord struct {
	ID            int64     `json:"id"`
	OrigFilename  string    `json:"orig_filename"`
	StoredPath    string    `json:"stored_path"`
	ProcessedPath *string   `json:"processed_path,omitempty"`
	ThumbPath     *string   `json:"thumb_path,omitempty"`
	Status        string    `json:"status"`
	Format        string    `json:"format"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
