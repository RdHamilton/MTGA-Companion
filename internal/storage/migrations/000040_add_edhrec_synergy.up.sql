-- EDHREC synergy data table
-- Stores synergy scores from EDHREC for Commander format recommendations

CREATE TABLE IF NOT EXISTS edhrec_synergy (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_name TEXT NOT NULL,
    synergy_card_name TEXT NOT NULL,
    synergy_score REAL NOT NULL,
    inclusion_count INTEGER DEFAULT 0,
    num_decks INTEGER DEFAULT 0,
    lift REAL DEFAULT 0.0,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(card_name, synergy_card_name)
);

CREATE INDEX IF NOT EXISTS idx_edhrec_synergy_card ON edhrec_synergy(card_name);
CREATE INDEX IF NOT EXISTS idx_edhrec_synergy_score ON edhrec_synergy(synergy_score DESC);

-- EDHREC card metadata table
-- Stores basic card info from EDHREC (salt score, deck count, etc.)

CREATE TABLE IF NOT EXISTS edhrec_card_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_name TEXT NOT NULL UNIQUE,
    sanitized_name TEXT NOT NULL,
    num_decks INTEGER DEFAULT 0,
    salt_score REAL DEFAULT 0.0,
    color_identity TEXT,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_edhrec_metadata_name ON edhrec_card_metadata(card_name);

-- EDHREC theme cards table
-- Stores cards associated with specific themes (tokens, aristocrats, etc.)

CREATE TABLE IF NOT EXISTS edhrec_theme_cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    theme_name TEXT NOT NULL,
    card_name TEXT NOT NULL,
    synergy_score REAL DEFAULT 0.0,
    is_top_card BOOLEAN DEFAULT FALSE,
    is_high_synergy BOOLEAN DEFAULT FALSE,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(theme_name, card_name)
);

CREATE INDEX IF NOT EXISTS idx_edhrec_theme_cards_theme ON edhrec_theme_cards(theme_name);
CREATE INDEX IF NOT EXISTS idx_edhrec_theme_cards_card ON edhrec_theme_cards(card_name);
