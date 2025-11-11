CREATE TABLE IF NOT EXISTS short_urls (
                                          id SERIAL PRIMARY KEY,
                                          alias TEXT UNIQUE NOT NULL,
                                          target TEXT NOT NULL,
                                          created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS clicks (
                                      id SERIAL PRIMARY KEY,
                                      short_url_id INT REFERENCES short_urls(id) ON DELETE CASCADE,
                                      user_agent TEXT,
                                      ip TEXT,
                                      referer TEXT,
                                      created_at TIMESTAMP DEFAULT NOW()
);
