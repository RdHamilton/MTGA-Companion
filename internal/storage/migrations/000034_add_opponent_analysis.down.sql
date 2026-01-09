-- Rollback: Remove opponent deck analysis tables

-- Drop tables
DROP TABLE IF EXISTS archetype_expected_cards;
DROP TABLE IF EXISTS matchup_statistics;
DROP TABLE IF EXISTS opponent_deck_profiles;

-- Remove columns from deck_performance_history (SQLite doesn't support DROP COLUMN easily)
-- Using ALTER TABLE to rename and recreate
CREATE TABLE deck_performance_history_backup AS SELECT
    id, account_id, deck_id, match_id, archetype, secondary_archetype,
    archetype_confidence, color_identity, card_count, result, games_won,
    games_lost, duration_seconds, format, event_type, opponent_archetype,
    rank_tier, match_timestamp, created_at
FROM deck_performance_history;

DROP TABLE deck_performance_history;

CREATE TABLE deck_performance_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL,
    deck_id TEXT NOT NULL,
    match_id TEXT NOT NULL,
    archetype TEXT,
    secondary_archetype TEXT,
    archetype_confidence REAL,
    color_identity TEXT NOT NULL,
    card_count INTEGER NOT NULL,
    result TEXT NOT NULL,
    games_won INTEGER DEFAULT 0,
    games_lost INTEGER DEFAULT 0,
    duration_seconds INTEGER,
    format TEXT NOT NULL,
    event_type TEXT,
    opponent_archetype TEXT,
    rank_tier TEXT,
    match_timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    FOREIGN KEY (deck_id) REFERENCES decks(id) ON DELETE CASCADE,
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE
);

INSERT INTO deck_performance_history SELECT * FROM deck_performance_history_backup;
DROP TABLE deck_performance_history_backup;

CREATE INDEX idx_perf_history_account ON deck_performance_history(account_id);
CREATE INDEX idx_perf_history_deck ON deck_performance_history(deck_id);
CREATE INDEX idx_perf_history_match ON deck_performance_history(match_id);
CREATE INDEX idx_perf_history_archetype ON deck_performance_history(archetype);
CREATE INDEX idx_perf_history_opponent_archetype ON deck_performance_history(opponent_archetype);
