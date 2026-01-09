-- ML Suggestion Engine Schema
-- Tracks card combination statistics and ML-derived synergy scores

-- Card combination statistics - tracks how card pairs perform together
CREATE TABLE IF NOT EXISTS card_combination_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_id_1 INTEGER NOT NULL,
    card_id_2 INTEGER NOT NULL,
    deck_id TEXT,
    format TEXT DEFAULT 'Standard',

    -- Co-occurrence metrics
    games_together INTEGER DEFAULT 0,
    games_card1_only INTEGER DEFAULT 0,
    games_card2_only INTEGER DEFAULT 0,

    -- Win rate metrics
    wins_together INTEGER DEFAULT 0,
    wins_card1_only INTEGER DEFAULT 0,
    wins_card2_only INTEGER DEFAULT 0,

    -- Derived scores (calculated)
    synergy_score REAL DEFAULT 0.0,
    confidence_score REAL DEFAULT 0.0,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Ensure card_id_1 < card_id_2 for uniqueness
    UNIQUE(card_id_1, card_id_2, deck_id, format)
);

CREATE INDEX IF NOT EXISTS idx_combo_stats_card1 ON card_combination_stats(card_id_1);
CREATE INDEX IF NOT EXISTS idx_combo_stats_card2 ON card_combination_stats(card_id_2);
CREATE INDEX IF NOT EXISTS idx_combo_stats_deck ON card_combination_stats(deck_id);
CREATE INDEX IF NOT EXISTS idx_combo_stats_format ON card_combination_stats(format);
CREATE INDEX IF NOT EXISTS idx_combo_stats_synergy ON card_combination_stats(synergy_score DESC);

-- ML suggestions - stores generated suggestions with explanations
CREATE TABLE IF NOT EXISTS ml_suggestions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    deck_id TEXT NOT NULL,
    suggestion_type TEXT NOT NULL, -- 'add', 'remove', 'swap'

    -- Card references
    card_id INTEGER,
    card_name TEXT,
    swap_for_card_id INTEGER,
    swap_for_card_name TEXT,

    -- Scoring
    confidence REAL DEFAULT 0.0,
    expected_win_rate_change REAL DEFAULT 0.0,

    -- Explanation
    title TEXT NOT NULL,
    description TEXT,
    reasoning TEXT, -- JSON array of reasons
    evidence TEXT,  -- JSON object with supporting data

    -- Status
    is_dismissed BOOLEAN DEFAULT FALSE,
    was_applied BOOLEAN DEFAULT FALSE,
    outcome_win_rate_change REAL,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    applied_at TIMESTAMP,
    outcome_recorded_at TIMESTAMP,

    FOREIGN KEY (deck_id) REFERENCES decks(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_ml_suggestions_deck ON ml_suggestions(deck_id);
CREATE INDEX IF NOT EXISTS idx_ml_suggestions_type ON ml_suggestions(suggestion_type);
CREATE INDEX IF NOT EXISTS idx_ml_suggestions_confidence ON ml_suggestions(confidence DESC);
CREATE INDEX IF NOT EXISTS idx_ml_suggestions_active ON ml_suggestions(deck_id, is_dismissed);

-- Card affinity scores - pre-computed synergy between any two cards
CREATE TABLE IF NOT EXISTS card_affinity (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_id_1 INTEGER NOT NULL,
    card_id_2 INTEGER NOT NULL,
    format TEXT DEFAULT 'Standard',

    -- Affinity metrics
    affinity_score REAL DEFAULT 0.0,
    sample_size INTEGER DEFAULT 0,
    confidence REAL DEFAULT 0.0,

    -- Source of affinity
    source TEXT DEFAULT 'historical', -- 'historical', 'keyword', 'tribal', 'external'

    -- Timestamps
    computed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(card_id_1, card_id_2, format)
);

CREATE INDEX IF NOT EXISTS idx_affinity_card1 ON card_affinity(card_id_1);
CREATE INDEX IF NOT EXISTS idx_affinity_card2 ON card_affinity(card_id_2);
CREATE INDEX IF NOT EXISTS idx_affinity_score ON card_affinity(affinity_score DESC);

-- User play patterns - aggregated statistics for personalization
CREATE TABLE IF NOT EXISTS user_play_patterns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id TEXT NOT NULL,

    -- Archetype preferences
    preferred_archetype TEXT,
    aggro_affinity REAL DEFAULT 0.0,
    midrange_affinity REAL DEFAULT 0.0,
    control_affinity REAL DEFAULT 0.0,
    combo_affinity REAL DEFAULT 0.0,

    -- Color preferences
    color_preferences TEXT, -- JSON: {"W": 0.3, "U": 0.2, "B": 0.1, ...}

    -- Play style metrics
    avg_game_length REAL DEFAULT 0.0,
    aggression_score REAL DEFAULT 0.0,
    interaction_score REAL DEFAULT 0.0,

    -- Sample sizes
    total_matches INTEGER DEFAULT 0,
    total_decks INTEGER DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(account_id)
);

CREATE INDEX IF NOT EXISTS idx_play_patterns_account ON user_play_patterns(account_id);

-- ML model metadata - tracks model versions and performance
CREATE TABLE IF NOT EXISTS ml_model_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    model_name TEXT NOT NULL,
    model_version TEXT NOT NULL,

    -- Training info
    training_samples INTEGER DEFAULT 0,
    training_date TIMESTAMP,

    -- Performance metrics
    accuracy REAL,
    precision_score REAL,
    recall REAL,
    f1_score REAL,

    -- Model state
    is_active BOOLEAN DEFAULT FALSE,
    model_data BLOB, -- Serialized model weights if needed

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(model_name, model_version)
);
