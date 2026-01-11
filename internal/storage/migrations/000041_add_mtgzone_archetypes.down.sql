DROP INDEX IF EXISTS idx_mtgzone_articles_format;
DROP INDEX IF EXISTS idx_mtgzone_articles_type;
DROP TABLE IF EXISTS mtgzone_articles;

DROP INDEX IF EXISTS idx_mtgzone_synergies_card_b;
DROP INDEX IF EXISTS idx_mtgzone_synergies_card_a;
DROP TABLE IF EXISTS mtgzone_synergies;

DROP INDEX IF EXISTS idx_mtgzone_archetype_cards_role;
DROP INDEX IF EXISTS idx_mtgzone_archetype_cards_card;
DROP INDEX IF EXISTS idx_mtgzone_archetype_cards_archetype;
DROP TABLE IF EXISTS mtgzone_archetype_cards;

DROP INDEX IF EXISTS idx_mtgzone_archetypes_tier;
DROP INDEX IF EXISTS idx_mtgzone_archetypes_format;
DROP TABLE IF EXISTS mtgzone_archetypes;
