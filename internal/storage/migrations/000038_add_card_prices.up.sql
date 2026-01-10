-- Add price fields to set_cards table for Scryfall pricing data

ALTER TABLE set_cards ADD COLUMN price_usd REAL DEFAULT NULL;
ALTER TABLE set_cards ADD COLUMN price_usd_foil REAL DEFAULT NULL;
ALTER TABLE set_cards ADD COLUMN price_eur REAL DEFAULT NULL;
ALTER TABLE set_cards ADD COLUMN price_eur_foil REAL DEFAULT NULL;
ALTER TABLE set_cards ADD COLUMN price_tix REAL DEFAULT NULL;
ALTER TABLE set_cards ADD COLUMN prices_updated_at TIMESTAMP DEFAULT NULL;

-- Index for efficient price queries
CREATE INDEX IF NOT EXISTS idx_set_cards_prices ON set_cards(price_usd) WHERE price_usd IS NOT NULL;
