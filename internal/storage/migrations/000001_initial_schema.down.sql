-- Rollback initial schema
-- Drops all tables in reverse order of dependencies

DROP TABLE IF EXISTS draft_events;
DROP TABLE IF EXISTS rank_history;
DROP TABLE IF EXISTS collection_history;
DROP TABLE IF EXISTS collection;
DROP TABLE IF EXISTS deck_cards;
DROP TABLE IF EXISTS decks;
DROP TABLE IF EXISTS player_stats;
DROP TABLE IF EXISTS games;
DROP TABLE IF EXISTS matches;
