-- Add v1.3 Deck Builder support to decks table
-- This migration adds fields to track deck source, draft integration, and performance
-- Note: account_id already exists from migration 000002

-- Add source tracking: "draft", "constructed", "imported"
ALTER TABLE decks ADD COLUMN source TEXT NOT NULL DEFAULT 'constructed'
    CHECK(source IN ('draft', 'constructed', 'imported'));

-- Add draft_event_id for draft-based decks (nullable)
ALTER TABLE decks ADD COLUMN draft_event_id TEXT
    REFERENCES draft_events(id) ON DELETE SET NULL;

-- Add performance tracking fields
ALTER TABLE decks ADD COLUMN matches_played INTEGER NOT NULL DEFAULT 0;
ALTER TABLE decks ADD COLUMN matches_won INTEGER NOT NULL DEFAULT 0;
ALTER TABLE decks ADD COLUMN games_played INTEGER NOT NULL DEFAULT 0;
ALTER TABLE decks ADD COLUMN games_won INTEGER NOT NULL DEFAULT 0;

-- Add from_draft_pick flag to deck_cards to track cards picked during draft
ALTER TABLE deck_cards ADD COLUMN from_draft_pick INTEGER NOT NULL DEFAULT 0
    CHECK(from_draft_pick IN (0, 1));

-- Create indexes for performance (account_id index already exists from migration 000002)
CREATE INDEX IF NOT EXISTS idx_decks_source ON decks(source);
CREATE INDEX IF NOT EXISTS idx_decks_draft_event_id ON decks(draft_event_id);

-- Create deck_tags table for categorizing decks
CREATE TABLE IF NOT EXISTS deck_tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    deck_id TEXT NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(deck_id, tag)
);

CREATE INDEX IF NOT EXISTS idx_deck_tags_deck_id ON deck_tags(deck_id);
CREATE INDEX IF NOT EXISTS idx_deck_tags_tag ON deck_tags(tag);
