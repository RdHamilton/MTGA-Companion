-- Add URL columns to draft_card_ratings for Scryfall ID extraction
-- 17Lands provides image URLs that contain Scryfall card UUIDs
-- These can be used to reliably join with Scryfall data for Arena-exclusive sets

ALTER TABLE draft_card_ratings ADD COLUMN url TEXT;
ALTER TABLE draft_card_ratings ADD COLUMN url_back TEXT;
