-- Migration: Add opponent deck analysis tables
-- Features: Opponent deck profiles, matchup statistics, expected cards

-- Add additional columns to deck_performance_history for opponent analysis
ALTER TABLE deck_performance_history ADD COLUMN opponent_color_identity TEXT;
ALTER TABLE deck_performance_history ADD COLUMN opponent_confidence REAL DEFAULT 0;
ALTER TABLE deck_performance_history ADD COLUMN opponent_cards_seen INTEGER DEFAULT 0;

-- Opponent deck profiles - reconstructed from observed cards
CREATE TABLE IF NOT EXISTS opponent_deck_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    match_id TEXT NOT NULL UNIQUE,
    detected_archetype TEXT,
    archetype_confidence REAL DEFAULT 0,
    color_identity TEXT NOT NULL,
    deck_style TEXT,  -- aggro, control, midrange, combo
    cards_observed INTEGER NOT NULL DEFAULT 0,
    estimated_deck_size INTEGER DEFAULT 60,
    observed_card_ids TEXT,  -- JSON array of card IDs
    inferred_card_ids TEXT,  -- JSON array of inferred cards based on archetype
    signature_cards TEXT,  -- JSON array of signature cards found
    format TEXT,  -- Standard, Historic, etc.
    meta_archetype_id INTEGER,  -- Link to meta archetype if matched
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE
);

CREATE INDEX idx_opponent_profiles_match_id ON opponent_deck_profiles(match_id);
CREATE INDEX idx_opponent_profiles_archetype ON opponent_deck_profiles(detected_archetype);
CREATE INDEX idx_opponent_profiles_format ON opponent_deck_profiles(format);

-- Matchup statistics - win rates against each archetype
CREATE TABLE IF NOT EXISTS matchup_statistics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL,
    player_archetype TEXT NOT NULL,
    opponent_archetype TEXT NOT NULL,
    format TEXT NOT NULL,
    total_matches INTEGER DEFAULT 0,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    avg_game_duration INTEGER,  -- in seconds
    last_match_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    UNIQUE(account_id, player_archetype, opponent_archetype, format)
);

CREATE INDEX idx_matchup_stats_account ON matchup_statistics(account_id);
CREATE INDEX idx_matchup_stats_player_archetype ON matchup_statistics(player_archetype);
CREATE INDEX idx_matchup_stats_opponent_archetype ON matchup_statistics(opponent_archetype);
CREATE INDEX idx_matchup_stats_format ON matchup_statistics(format);

-- Expected cards table - what cards are commonly played in each archetype
CREATE TABLE IF NOT EXISTS archetype_expected_cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    archetype_name TEXT NOT NULL,
    format TEXT NOT NULL,
    card_id INTEGER NOT NULL,
    card_name TEXT NOT NULL,
    inclusion_rate REAL DEFAULT 0,  -- 0.0-1.0, how often this card appears in the archetype
    avg_copies REAL DEFAULT 1,
    is_signature BOOLEAN DEFAULT FALSE,
    category TEXT,  -- removal, threat, interaction, wincon
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(archetype_name, format, card_id)
);

CREATE INDEX idx_expected_cards_archetype ON archetype_expected_cards(archetype_name);
CREATE INDEX idx_expected_cards_format ON archetype_expected_cards(format);
