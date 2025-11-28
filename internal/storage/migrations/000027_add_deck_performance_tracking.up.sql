-- Migration: Add deck performance tracking tables for ML training
-- This adds tables to track deck performance history over time,
-- archetype classifications, and recommendation feedback.

-- Table: deck_performance_history
-- Records individual match results with deck state snapshots for ML training
CREATE TABLE deck_performance_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL,
    deck_id TEXT NOT NULL,
    match_id TEXT NOT NULL,

    -- Deck state at time of match
    archetype TEXT,                    -- Primary archetype classification
    secondary_archetype TEXT,          -- Secondary archetype if applicable
    archetype_confidence REAL,         -- Confidence score 0.0-1.0
    color_identity TEXT NOT NULL,      -- e.g., "WU", "RG", "WUBRG"
    card_count INTEGER NOT NULL,       -- Number of cards in deck

    -- Match outcome data
    result TEXT NOT NULL CHECK(result IN ('win', 'loss')),
    games_won INTEGER NOT NULL,
    games_lost INTEGER NOT NULL,
    duration_seconds INTEGER,

    -- Context data
    format TEXT NOT NULL,              -- e.g., "Draft", "Constructed", "Limited"
    event_type TEXT,                   -- e.g., "QuickDraft", "PremierDraft", "Ranked"
    opponent_archetype TEXT,           -- If known from detection
    rank_tier TEXT,                    -- Player rank at time of match

    -- Timestamps
    match_timestamp DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    FOREIGN KEY (deck_id) REFERENCES decks(id) ON DELETE CASCADE,
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE
);

-- Indexes for efficient ML training queries
CREATE INDEX idx_deck_perf_history_account ON deck_performance_history(account_id);
CREATE INDEX idx_deck_perf_history_deck ON deck_performance_history(deck_id);
CREATE INDEX idx_deck_perf_history_archetype ON deck_performance_history(archetype);
CREATE INDEX idx_deck_perf_history_format ON deck_performance_history(format);
CREATE INDEX idx_deck_perf_history_timestamp ON deck_performance_history(match_timestamp);
CREATE INDEX idx_deck_perf_history_result ON deck_performance_history(result);

-- Composite index for archetype performance queries
CREATE INDEX idx_deck_perf_history_archetype_result
    ON deck_performance_history(archetype, result, format);

-- Table: deck_archetypes
-- Stores archetype definitions and their performance statistics
CREATE TABLE deck_archetypes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,                -- e.g., "UW Flyers", "BR Sacrifice"
    set_code TEXT,                     -- Set this archetype applies to (NULL for constructed)
    format TEXT NOT NULL,              -- "draft", "constructed", "limited"
    color_identity TEXT NOT NULL,      -- Primary colors

    -- Signature cards that define this archetype (JSON array of card IDs)
    signature_cards TEXT,

    -- Key synergy patterns (JSON array)
    synergy_patterns TEXT,

    -- Performance statistics (updated periodically)
    total_matches INTEGER NOT NULL DEFAULT 0,
    total_wins INTEGER NOT NULL DEFAULT 0,
    avg_win_rate REAL,

    -- Source information
    source TEXT NOT NULL DEFAULT 'system' CHECK(source IN ('system', '17lands', 'user', 'ml')),
    external_id TEXT,                  -- ID from external source (e.g., 17Lands)

    -- Timestamps
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(name, set_code, format)
);

CREATE INDEX idx_deck_archetypes_set ON deck_archetypes(set_code);
CREATE INDEX idx_deck_archetypes_format ON deck_archetypes(format);
CREATE INDEX idx_deck_archetypes_colors ON deck_archetypes(color_identity);

-- Table: recommendation_feedback
-- Tracks user responses to card/deck recommendations for ML training
CREATE TABLE recommendation_feedback (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL,

    -- Recommendation context
    recommendation_type TEXT NOT NULL CHECK(recommendation_type IN ('card_pick', 'deck_card', 'archetype', 'sideboard')),
    recommendation_id TEXT NOT NULL,   -- Unique ID for this recommendation instance

    -- What was recommended
    recommended_card_id INTEGER,       -- Card that was recommended
    recommended_archetype TEXT,        -- Or archetype that was recommended

    -- Context at time of recommendation (JSON)
    context_data TEXT NOT NULL,        -- Deck state, available picks, game state, etc.

    -- User action
    action TEXT NOT NULL CHECK(action IN ('accepted', 'rejected', 'ignored', 'alternate')),
    alternate_choice_id INTEGER,       -- If user picked something else

    -- Outcome tracking (updated after match if applicable)
    outcome_match_id TEXT,             -- Match where this recommendation was used
    outcome_result TEXT CHECK(outcome_result IN ('win', 'loss')),

    -- Confidence and scoring
    recommendation_score REAL,         -- Original recommendation score
    recommendation_rank INTEGER,       -- Position in recommendation list

    -- Timestamps
    recommended_at DATETIME NOT NULL,
    responded_at DATETIME,
    outcome_recorded_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

-- Indexes for feedback analysis
CREATE INDEX idx_rec_feedback_account ON recommendation_feedback(account_id);
CREATE INDEX idx_rec_feedback_type ON recommendation_feedback(recommendation_type);
CREATE INDEX idx_rec_feedback_action ON recommendation_feedback(action);
CREATE INDEX idx_rec_feedback_timestamp ON recommendation_feedback(recommended_at);
CREATE INDEX idx_rec_feedback_card ON recommendation_feedback(recommended_card_id);

-- Composite index for acceptance rate calculations
CREATE INDEX idx_rec_feedback_type_action
    ON recommendation_feedback(recommendation_type, action, recommended_at);

-- Table: archetype_card_weights
-- Stores which cards are associated with which archetypes and their weights
CREATE TABLE archetype_card_weights (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    archetype_id INTEGER NOT NULL,
    card_id INTEGER NOT NULL,

    -- Weight indicates how strongly this card indicates the archetype
    weight REAL NOT NULL DEFAULT 1.0,  -- 0.0-10.0, higher = stronger indicator
    is_signature INTEGER NOT NULL DEFAULT 0,  -- 1 if this is a signature/key card

    -- Source of this weight
    source TEXT NOT NULL DEFAULT 'system' CHECK(source IN ('system', '17lands', 'user', 'ml')),

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (archetype_id) REFERENCES deck_archetypes(id) ON DELETE CASCADE,
    UNIQUE(archetype_id, card_id)
);

CREATE INDEX idx_archetype_card_weights_archetype ON archetype_card_weights(archetype_id);
CREATE INDEX idx_archetype_card_weights_card ON archetype_card_weights(card_id);
