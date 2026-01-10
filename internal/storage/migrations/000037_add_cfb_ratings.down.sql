-- Remove ChannelFireball ratings table
DROP INDEX IF EXISTS idx_cfb_ratings_set_code;
DROP INDEX IF EXISTS idx_cfb_ratings_arena_id;
DROP INDEX IF EXISTS idx_cfb_ratings_card_name;
DROP TABLE IF EXISTS cfb_ratings;
