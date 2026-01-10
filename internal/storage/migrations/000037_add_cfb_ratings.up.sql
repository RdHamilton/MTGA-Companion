-- ChannelFireball card ratings table
-- Stores card ratings from ChannelFireball set reviews
CREATE TABLE IF NOT EXISTS cfb_ratings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_name TEXT NOT NULL,
    set_code TEXT NOT NULL,
    arena_id INTEGER,

    -- CFB Limited Rating (A+, A, A-, B+, B, B-, C+, C, C-, D, F)
    limited_rating TEXT,
    limited_score REAL DEFAULT 0.0,

    -- CFB Constructed Rating (Staple, Playable, Fringe, Unplayable)
    constructed_rating TEXT,
    constructed_score REAL DEFAULT 0.0,

    -- Archetype fit notes
    archetype_fit TEXT,

    -- Commentary/notes from CFB review
    commentary TEXT,

    -- Source information
    source_url TEXT,
    author TEXT,

    -- Metadata
    imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(card_name, set_code)
);

CREATE INDEX IF NOT EXISTS idx_cfb_ratings_set_code ON cfb_ratings(set_code);
CREATE INDEX IF NOT EXISTS idx_cfb_ratings_arena_id ON cfb_ratings(arena_id);
CREATE INDEX IF NOT EXISTS idx_cfb_ratings_card_name ON cfb_ratings(card_name);
