-- Drop old draft ratings tables if they exist (from migration 000007)
DROP TABLE IF EXISTS draft_card_ratings;
DROP TABLE IF EXISTS draft_color_ratings;

-- Create draft_card_ratings table for caching 17Lands card performance data
CREATE TABLE IF NOT EXISTS draft_card_ratings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    set_code TEXT NOT NULL,
    draft_format TEXT NOT NULL,
    arena_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    color TEXT,
    rarity TEXT,
    gihwr REAL, -- Games In Hand Win Rate (%)
    ohwr REAL,  -- Opening Hand Win Rate (%)
    alsa REAL,  -- Average Last Seen At (pick number)
    ata REAL,   -- Average Taken At (pick number)
    gih_count INTEGER, -- Games In Hand count (sample size)
    cached_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(set_code, draft_format, arena_id)
);

CREATE INDEX idx_draft_card_ratings_set ON draft_card_ratings(set_code, draft_format);
CREATE INDEX idx_draft_card_ratings_arena_id ON draft_card_ratings(arena_id);
CREATE INDEX idx_draft_card_ratings_gihwr ON draft_card_ratings(gihwr DESC);

-- Create draft_color_ratings table for caching 17Lands color combination performance
CREATE TABLE IF NOT EXISTS draft_color_ratings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    set_code TEXT NOT NULL,
    draft_format TEXT NOT NULL,
    color_combination TEXT NOT NULL, -- e.g., "W", "UB", "WUG"
    win_rate REAL,
    games_played INTEGER,
    cached_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(set_code, draft_format, color_combination)
);

CREATE INDEX idx_draft_color_ratings_set ON draft_color_ratings(set_code, draft_format);
CREATE INDEX idx_draft_color_ratings_win_rate ON draft_color_ratings(win_rate DESC);
