-- Track decks created within the app for ML training
-- is_app_created: Deck was created/managed by the app (not just imported from MTGA)
-- created_method: How the deck was created (build_around, suggest_decks, manual, imported)
-- seed_card_id: The card used to seed Build Around feature (nullable)

ALTER TABLE decks ADD COLUMN is_app_created BOOLEAN DEFAULT FALSE;
ALTER TABLE decks ADD COLUMN created_method TEXT DEFAULT 'imported';
ALTER TABLE decks ADD COLUMN seed_card_id INTEGER;

-- Index for querying app-created decks for ML training
CREATE INDEX IF NOT EXISTS idx_decks_app_created ON decks(is_app_created) WHERE is_app_created = TRUE;

-- Index for querying by creation method
CREATE INDEX IF NOT EXISTS idx_decks_created_method ON decks(created_method);
