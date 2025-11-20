-- Create dataset_metadata table to track 17Lands data sources and freshness
CREATE TABLE IF NOT EXISTS dataset_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    set_code TEXT NOT NULL,
    draft_format TEXT NOT NULL,
    data_source TEXT NOT NULL, -- "s3" or "web_api"
    last_updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    total_cards INTEGER,       -- Number of cards with ratings
    total_games INTEGER,       -- Total games analyzed (if available from S3)
    dataset_version TEXT,      -- Version/date of S3 dataset
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(set_code, draft_format)
);

CREATE INDEX idx_dataset_metadata_set ON dataset_metadata(set_code, draft_format);
CREATE INDEX idx_dataset_metadata_updated ON dataset_metadata(last_updated_at DESC);
CREATE INDEX idx_dataset_metadata_source ON dataset_metadata(data_source);

-- Add data_source column to existing draft_card_ratings table
ALTER TABLE draft_card_ratings ADD COLUMN data_source TEXT DEFAULT 'api';

-- Add data_source column to existing draft_color_ratings table
ALTER TABLE draft_color_ratings ADD COLUMN data_source TEXT DEFAULT 'api';
