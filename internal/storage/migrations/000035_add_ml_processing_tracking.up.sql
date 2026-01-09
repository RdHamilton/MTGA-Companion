-- Add ML processing tracking to matches table
-- This prevents double-counting when ProcessMatchHistory is called multiple times

ALTER TABLE matches ADD COLUMN processed_for_ml BOOLEAN DEFAULT FALSE;

-- Index for efficient querying of unprocessed matches
CREATE INDEX IF NOT EXISTS idx_matches_processed_for_ml ON matches(processed_for_ml) WHERE processed_for_ml = FALSE;

-- Individual card statistics for accurate synergy calculation
-- Tracks how each card performs independently across all decks
CREATE TABLE IF NOT EXISTS card_individual_stats (
    card_id INTEGER NOT NULL,
    format TEXT NOT NULL,
    total_games INTEGER DEFAULT 0,
    wins INTEGER DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (card_id, format)
);

CREATE INDEX IF NOT EXISTS idx_card_individual_stats_format ON card_individual_stats(format);
