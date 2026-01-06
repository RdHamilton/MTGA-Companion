-- Rollback: Remove in-game play tracking tables

-- Drop indexes first
DROP INDEX IF EXISTS idx_opponent_cards_match_id;
DROP INDEX IF EXISTS idx_opponent_cards_game_id;
DROP INDEX IF EXISTS idx_game_snapshots_match_id;
DROP INDEX IF EXISTS idx_game_snapshots_game_id;
DROP INDEX IF EXISTS idx_game_plays_turn;
DROP INDEX IF EXISTS idx_game_plays_match_id;
DROP INDEX IF EXISTS idx_game_plays_game_id;

-- Drop tables
DROP TABLE IF EXISTS opponent_cards_observed;
DROP TABLE IF EXISTS game_state_snapshots;
DROP TABLE IF EXISTS game_plays;
