-- Metadata table for storing application configuration and state
CREATE TABLE IF NOT EXISTS metadata (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster lookups by update time
CREATE INDEX IF NOT EXISTS idx_metadata_updated_at ON metadata(updated_at);
