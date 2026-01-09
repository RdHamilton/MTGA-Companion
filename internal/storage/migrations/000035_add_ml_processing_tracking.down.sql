-- Remove ML processing tracking from matches table

DROP INDEX IF EXISTS idx_matches_processed_for_ml;
DROP INDEX IF EXISTS idx_card_individual_stats_format;
DROP TABLE IF EXISTS card_individual_stats;

-- SQLite doesn't support DROP COLUMN directly, recreate table without processed_for_ml
CREATE TABLE IF NOT EXISTS matches_new (
    id TEXT PRIMARY KEY,
    account_id INTEGER,
    event_id TEXT NOT NULL,
    event_name TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    duration_seconds INTEGER,
    player_wins INTEGER NOT NULL,
    opponent_wins INTEGER NOT NULL,
    player_team_id INTEGER NOT NULL,
    deck_id TEXT,
    rank_before TEXT,
    rank_after TEXT,
    format TEXT NOT NULL,
    result TEXT NOT NULL CHECK(result IN ('win', 'loss')),
    result_reason TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    opponent_name TEXT,
    opponent_id TEXT,
    notes TEXT DEFAULT '',
    rating INTEGER DEFAULT 0
);

INSERT INTO matches_new SELECT
    id, account_id, event_id, event_name, timestamp, duration_seconds,
    player_wins, opponent_wins, player_team_id, deck_id,
    rank_before, rank_after, format, result, result_reason, created_at,
    opponent_name, opponent_id, notes, rating
FROM matches;

DROP TABLE matches;
ALTER TABLE matches_new RENAME TO matches;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_matches_timestamp ON matches(timestamp);
CREATE INDEX IF NOT EXISTS idx_matches_event_id ON matches(event_id);
CREATE INDEX IF NOT EXISTS idx_matches_format ON matches(format);
CREATE INDEX IF NOT EXISTS idx_matches_result ON matches(result);
CREATE INDEX IF NOT EXISTS idx_matches_account_id ON matches(account_id);
CREATE INDEX IF NOT EXISTS idx_matches_opponent_id ON matches(opponent_id);
CREATE INDEX IF NOT EXISTS idx_matches_opponent_name ON matches(opponent_name);
