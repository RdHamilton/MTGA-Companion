-- Rollback opponent tracking and game-level result reasons

-- SQLite doesn't support DROP COLUMN directly
-- We need to recreate tables without the new columns

-- Recreate games table without result_reason
CREATE TABLE IF NOT EXISTS games_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    match_id TEXT NOT NULL,
    game_number INTEGER NOT NULL,
    result TEXT NOT NULL CHECK(result IN ('win', 'loss')),
    duration_seconds INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
    UNIQUE(match_id, game_number)
);

INSERT INTO games_new (id, match_id, game_number, result, duration_seconds, created_at)
SELECT id, match_id, game_number, result, duration_seconds, created_at FROM games;

DROP TABLE games;
ALTER TABLE games_new RENAME TO games;

CREATE INDEX IF NOT EXISTS idx_games_match_id ON games(match_id);

-- Recreate matches table without opponent fields
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
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO matches_new SELECT
    id, account_id, event_id, event_name, timestamp, duration_seconds,
    player_wins, opponent_wins, player_team_id, deck_id,
    rank_before, rank_after, format, result, result_reason, created_at
FROM matches;

DROP TABLE matches;
ALTER TABLE matches_new RENAME TO matches;

CREATE INDEX IF NOT EXISTS idx_matches_timestamp ON matches(timestamp);
CREATE INDEX IF NOT EXISTS idx_matches_event_id ON matches(event_id);
CREATE INDEX IF NOT EXISTS idx_matches_format ON matches(format);
CREATE INDEX IF NOT EXISTS idx_matches_result ON matches(result);
CREATE INDEX IF NOT EXISTS idx_matches_account_id ON matches(account_id);
