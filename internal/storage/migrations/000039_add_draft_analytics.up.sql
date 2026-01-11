-- Draft match results linking drafts to matches
CREATE TABLE IF NOT EXISTS draft_match_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    match_id TEXT NOT NULL,
    result TEXT NOT NULL CHECK(result IN ('win', 'loss')),
    opponent_colors TEXT,
    game_wins INTEGER DEFAULT 0,
    game_losses INTEGER DEFAULT 0,
    match_timestamp TIMESTAMP NOT NULL,
    FOREIGN KEY (session_id) REFERENCES draft_sessions(id) ON DELETE CASCADE,
    UNIQUE(session_id, match_id)
);
CREATE INDEX IF NOT EXISTS idx_draft_match_results_session ON draft_match_results(session_id);
CREATE INDEX IF NOT EXISTS idx_draft_match_results_timestamp ON draft_match_results(match_timestamp);

-- Archetype performance aggregation
CREATE TABLE IF NOT EXISTS draft_archetype_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    set_code TEXT NOT NULL,
    color_combination TEXT NOT NULL,
    archetype_name TEXT NOT NULL,
    matches_played INTEGER DEFAULT 0,
    matches_won INTEGER DEFAULT 0,
    drafts_count INTEGER DEFAULT 0,
    avg_draft_grade REAL,
    last_played_at TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(set_code, color_combination)
);
CREATE INDEX IF NOT EXISTS idx_draft_archetype_stats_set ON draft_archetype_stats(set_code);

-- Community comparison cache
CREATE TABLE IF NOT EXISTS draft_community_comparison (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    set_code TEXT NOT NULL,
    draft_format TEXT NOT NULL,
    user_win_rate REAL NOT NULL,
    community_avg_win_rate REAL NOT NULL,
    percentile_rank REAL,
    sample_size INTEGER NOT NULL DEFAULT 0,
    calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(set_code, draft_format)
);
CREATE INDEX IF NOT EXISTS idx_draft_community_comparison_set ON draft_community_comparison(set_code);

-- Temporal trend tracking
-- Note: set_code uses empty string '' for overall/aggregate stats (SQLite UNIQUE treats NULL as distinct)
CREATE TABLE IF NOT EXISTS draft_temporal_trends (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    period_type TEXT NOT NULL CHECK(period_type IN ('week', 'month')),
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    set_code TEXT NOT NULL DEFAULT '',
    drafts_count INTEGER DEFAULT 0,
    matches_played INTEGER DEFAULT 0,
    matches_won INTEGER DEFAULT 0,
    avg_draft_grade REAL,
    calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(period_type, period_start, set_code)
);
CREATE INDEX IF NOT EXISTS idx_draft_temporal_trends_period ON draft_temporal_trends(period_type, period_start);

-- Draft pattern analysis cache
-- Note: set_code uses empty string '' for overall analysis
CREATE TABLE IF NOT EXISTS draft_pattern_analysis (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    set_code TEXT NOT NULL DEFAULT '',
    color_preference_json TEXT,
    type_preference_json TEXT,
    pick_order_pattern_json TEXT,
    archetype_affinity_json TEXT,
    sample_size INTEGER NOT NULL DEFAULT 0,
    calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(set_code)
);
