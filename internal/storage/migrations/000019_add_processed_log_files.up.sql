-- Track which log files have been processed to prevent duplicate processing
CREATE TABLE IF NOT EXISTS processed_log_files (
    filename TEXT PRIMARY KEY,
    processed_at TIMESTAMP NOT NULL,
    entry_count INTEGER DEFAULT 0,
    matches_found INTEGER DEFAULT 0,
    file_size_bytes INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for quick lookup by processed date
CREATE INDEX IF NOT EXISTS idx_processed_log_files_processed_at ON processed_log_files(processed_at);
