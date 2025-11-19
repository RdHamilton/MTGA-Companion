-- Add win rate prediction fields to draft_sessions table
ALTER TABLE draft_sessions ADD COLUMN predicted_win_rate REAL;
ALTER TABLE draft_sessions ADD COLUMN predicted_win_rate_min REAL;
ALTER TABLE draft_sessions ADD COLUMN predicted_win_rate_max REAL;
ALTER TABLE draft_sessions ADD COLUMN prediction_factors TEXT; -- JSON string with breakdown
ALTER TABLE draft_sessions ADD COLUMN predicted_at TIMESTAMP;

-- Index for querying predictions
CREATE INDEX idx_draft_sessions_predicted_win_rate ON draft_sessions(predicted_win_rate) WHERE predicted_win_rate IS NOT NULL;
