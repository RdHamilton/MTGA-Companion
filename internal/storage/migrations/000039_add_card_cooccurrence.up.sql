-- Card co-occurrence table for tracking which cards appear together in decks
-- Used for synergy analysis based on real deck data

CREATE TABLE IF NOT EXISTS card_cooccurrence (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_a_arena_id INTEGER NOT NULL,
    card_b_arena_id INTEGER NOT NULL,
    format TEXT NOT NULL DEFAULT 'all',
    count INTEGER NOT NULL DEFAULT 0,
    pmi_score REAL DEFAULT 0.0,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(card_a_arena_id, card_b_arena_id, format)
);

-- Index for efficient lookups
CREATE INDEX IF NOT EXISTS idx_cooccurrence_card_a ON card_cooccurrence(card_a_arena_id, format);
CREATE INDEX IF NOT EXISTS idx_cooccurrence_card_b ON card_cooccurrence(card_b_arena_id, format);
CREATE INDEX IF NOT EXISTS idx_cooccurrence_format ON card_cooccurrence(format);
CREATE INDEX IF NOT EXISTS idx_cooccurrence_pmi ON card_cooccurrence(pmi_score DESC);

-- Deck source tracking for co-occurrence data
CREATE TABLE IF NOT EXISTS cooccurrence_sources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_type TEXT NOT NULL,
    source_id TEXT NOT NULL,
    format TEXT NOT NULL DEFAULT 'all',
    deck_count INTEGER NOT NULL DEFAULT 0,
    card_count INTEGER NOT NULL DEFAULT 0,
    last_synced TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source_type, source_id, format)
);

-- Card frequency table for PMI calculation
-- Tracks how often each card appears across all analyzed decks
CREATE TABLE IF NOT EXISTS card_frequency (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_arena_id INTEGER NOT NULL,
    format TEXT NOT NULL DEFAULT 'all',
    deck_count INTEGER NOT NULL DEFAULT 0,
    total_decks INTEGER NOT NULL DEFAULT 0,
    frequency REAL GENERATED ALWAYS AS (CAST(deck_count AS REAL) / NULLIF(total_decks, 0)) STORED,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(card_arena_id, format)
);

CREATE INDEX IF NOT EXISTS idx_frequency_card ON card_frequency(card_arena_id, format);
CREATE INDEX IF NOT EXISTS idx_frequency_format ON card_frequency(format);
