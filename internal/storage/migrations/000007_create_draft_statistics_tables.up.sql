-- Create draft_card_ratings table for 17Lands card performance data
CREATE TABLE IF NOT EXISTS draft_card_ratings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    arena_id INTEGER NOT NULL,
    expansion TEXT NOT NULL,
    format TEXT NOT NULL,              -- PremierDraft, QuickDraft, TradDraft, etc.
    colors TEXT,                        -- Color filter used (null = all colors)

    -- Win rate metrics (stored as decimals, e.g., 0.58 for 58%)
    gihwr REAL,                         -- Games in Hand Win Rate
    ohwr REAL,                          -- Opening Hand Win Rate
    gpwr REAL,                          -- Game Present Win Rate
    gdwr REAL,                          -- Game Drawn Win Rate
    ihdwr REAL,                         -- In Hand Drawn Win Rate

    -- Improvement metrics
    gihwr_delta REAL,                   -- GIH Win Rate Delta (improvement)
    ohwr_delta REAL,                    -- OH Win Rate Delta
    gdwr_delta REAL,                    -- GD Win Rate Delta
    ihdwr_delta REAL,                   -- IHD Win Rate Delta

    -- Draft metrics
    alsa REAL,                          -- Average Last Seen At (pick position)
    ata REAL,                           -- Average Taken At (pick position)

    -- Sample sizes (number of games)
    gih INTEGER,                        -- Games in Hand count
    oh INTEGER,                         -- Opening Hand count
    gp INTEGER,                         -- Game Present count
    gd INTEGER,                         -- Game Drawn count
    ihd INTEGER,                        -- In Hand Drawn count

    -- Deck metrics
    games_played INTEGER,               -- Total games with this card
    num_decks INTEGER,                  -- Number of decks containing this card

    -- Metadata
    start_date TEXT,                    -- YYYY-MM-DD format
    end_date TEXT,                      -- YYYY-MM-DD format
    cached_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Ensure no duplicate entries for same card/expansion/format/date range
    UNIQUE(arena_id, expansion, format, colors, start_date, end_date)
);

-- Create indices for common queries
CREATE INDEX IF NOT EXISTS idx_draft_ratings_arena_id ON draft_card_ratings(arena_id);
CREATE INDEX IF NOT EXISTS idx_draft_ratings_expansion ON draft_card_ratings(expansion);
CREATE INDEX IF NOT EXISTS idx_draft_ratings_format ON draft_card_ratings(expansion, format);
CREATE INDEX IF NOT EXISTS idx_draft_ratings_staleness ON draft_card_ratings(last_updated);

-- Create draft_color_ratings table for 17Lands color combination data
CREATE TABLE IF NOT EXISTS draft_color_ratings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    expansion TEXT NOT NULL,
    event_type TEXT NOT NULL,           -- PremierDraft, QuickDraft, etc.
    color_combination TEXT NOT NULL,    -- W, U, B, R, G, WU, UB, WUG, etc.

    -- Performance metrics
    win_rate REAL,                      -- Overall win rate (decimal)

    -- Game counts
    games_played INTEGER,               -- Total games played
    num_decks INTEGER,                  -- Number of decks with this color combo

    -- Metadata
    start_date TEXT,                    -- YYYY-MM-DD format
    end_date TEXT,                      -- YYYY-MM-DD format
    cached_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Ensure no duplicate entries
    UNIQUE(expansion, event_type, color_combination, start_date, end_date)
);

-- Create indices for color ratings
CREATE INDEX IF NOT EXISTS idx_draft_colors_expansion ON draft_color_ratings(expansion);
CREATE INDEX IF NOT EXISTS idx_draft_colors_event ON draft_color_ratings(expansion, event_type);
CREATE INDEX IF NOT EXISTS idx_draft_colors_staleness ON draft_color_ratings(last_updated);
