-- Drop draft statistics tables
DROP INDEX IF EXISTS idx_draft_colors_staleness;
DROP INDEX IF EXISTS idx_draft_colors_event;
DROP INDEX IF EXISTS idx_draft_colors_expansion;
DROP TABLE IF EXISTS draft_color_ratings;

DROP INDEX IF EXISTS idx_draft_ratings_staleness;
DROP INDEX IF EXISTS idx_draft_ratings_format;
DROP INDEX IF EXISTS idx_draft_ratings_expansion;
DROP INDEX IF EXISTS idx_draft_ratings_arena_id;
DROP TABLE IF EXISTS draft_card_ratings;
