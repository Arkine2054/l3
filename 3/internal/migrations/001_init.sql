-- +goose Up
CREATE TABLE IF NOT EXISTS comments (
                                        id SERIAL PRIMARY KEY,
                                        parent_id INT REFERENCES comments(id) ON DELETE CASCADE,
                                        content TEXT NOT NULL,
                                        author TEXT,
                                        created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_parent_id ON comments(parent_id);
CREATE INDEX IF NOT EXISTS idx_created_at ON comments(created_at);

-- +goose Down
DROP TABLE IF EXISTS comments;
