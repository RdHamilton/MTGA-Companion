-- Drop indexes first
DROP INDEX IF EXISTS idx_draft_card_ratings_set;
DROP INDEX IF EXISTS idx_draft_card_ratings_arena_id;
DROP INDEX IF EXISTS idx_draft_card_ratings_gihwr;
DROP INDEX IF EXISTS idx_draft_color_ratings_set;
DROP INDEX IF EXISTS idx_draft_color_ratings_win_rate;

-- Drop tables
DROP TABLE IF EXISTS draft_card_ratings;
DROP TABLE IF EXISTS draft_color_ratings;
