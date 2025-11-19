-- Remove win rate prediction fields from draft_sessions table
DROP INDEX IF EXISTS idx_draft_sessions_predicted_win_rate;

ALTER TABLE draft_sessions DROP COLUMN predicted_at;
ALTER TABLE draft_sessions DROP COLUMN prediction_factors;
ALTER TABLE draft_sessions DROP COLUMN predicted_win_rate_max;
ALTER TABLE draft_sessions DROP COLUMN predicted_win_rate_min;
ALTER TABLE draft_sessions DROP COLUMN predicted_win_rate;
