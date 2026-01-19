CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short_path VARCHAR(255) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_short_path ON urls(short_path);
