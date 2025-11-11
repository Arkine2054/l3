-- +goose Up
CREATE TABLE IF NOT EXISTS images (
                                      id SERIAL PRIMARY KEY,
                                      orig_filename TEXT NOT NULL,
                                      stored_path TEXT NOT NULL,
                                      processed_path TEXT,
                                      thumb_path TEXT,
                                      status TEXT NOT NULL DEFAULT 'pending',
                                      format TEXT,
                                      created_at TIMESTAMP DEFAULT NOW(),
                                      updated_at TIMESTAMP DEFAULT NOW()
);
-- make sure updated_at is updated on change (simple trigger optional)
-- +goose Down
DROP TABLE IF EXISTS images;
