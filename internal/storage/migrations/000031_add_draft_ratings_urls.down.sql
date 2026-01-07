-- Remove URL columns from draft_card_ratings
-- SQLite doesn't support DROP COLUMN directly, so we recreate the table

CREATE TABLE draft_card_ratings_temp (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    set_code TEXT NOT NULL,
    draft_format TEXT NOT NULL,
    arena_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    color TEXT,
    rarity TEXT,
    gihwr REAL,
    ohwr REAL,
    alsa REAL,
    ata REAL,
    gih_count INTEGER,
    cached_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(set_code, draft_format, arena_id)
);

INSERT INTO draft_card_ratings_temp (id, set_code, draft_format, arena_id, name, color, rarity, gihwr, ohwr, alsa, ata, gih_count, cached_at)
SELECT id, set_code, draft_format, arena_id, name, color, rarity, gihwr, ohwr, alsa, ata, gih_count, cached_at
FROM draft_card_ratings;

DROP TABLE draft_card_ratings;
ALTER TABLE draft_card_ratings_temp RENAME TO draft_card_ratings;

-- Recreate indexes
CREATE INDEX idx_draft_card_ratings_set ON draft_card_ratings(set_code, draft_format);
CREATE INDEX idx_draft_card_ratings_arena_id ON draft_card_ratings(arena_id);
CREATE INDEX idx_draft_card_ratings_gihwr ON draft_card_ratings(gihwr DESC);
