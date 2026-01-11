-- MTGZone archetype definitions table
-- Stores archetype information extracted from deck guides and tier lists

CREATE TABLE IF NOT EXISTS mtgzone_archetypes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    format TEXT NOT NULL,
    tier TEXT,
    description TEXT,
    play_style TEXT,
    source_url TEXT,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, format)
);

CREATE INDEX IF NOT EXISTS idx_mtgzone_archetypes_format ON mtgzone_archetypes(format);
CREATE INDEX IF NOT EXISTS idx_mtgzone_archetypes_tier ON mtgzone_archetypes(tier);

-- MTGZone archetype core cards table
-- Stores which cards are essential for each archetype

CREATE TABLE IF NOT EXISTS mtgzone_archetype_cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    archetype_id INTEGER NOT NULL,
    card_name TEXT NOT NULL,
    role TEXT NOT NULL, -- 'core', 'flex', 'sideboard'
    copies INTEGER DEFAULT 4,
    importance TEXT, -- 'essential', 'important', 'optional'
    notes TEXT,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (archetype_id) REFERENCES mtgzone_archetypes(id) ON DELETE CASCADE,
    UNIQUE(archetype_id, card_name)
);

CREATE INDEX IF NOT EXISTS idx_mtgzone_archetype_cards_archetype ON mtgzone_archetype_cards(archetype_id);
CREATE INDEX IF NOT EXISTS idx_mtgzone_archetype_cards_card ON mtgzone_archetype_cards(card_name);
CREATE INDEX IF NOT EXISTS idx_mtgzone_archetype_cards_role ON mtgzone_archetype_cards(role);

-- MTGZone synergy annotations table
-- Stores synergy relationships with explanations from articles

CREATE TABLE IF NOT EXISTS mtgzone_synergies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_a TEXT NOT NULL,
    card_b TEXT NOT NULL,
    reason TEXT NOT NULL,
    source_url TEXT,
    archetype_context TEXT, -- Which archetype this synergy belongs to
    confidence REAL DEFAULT 0.5,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(card_a, card_b, archetype_context)
);

CREATE INDEX IF NOT EXISTS idx_mtgzone_synergies_card_a ON mtgzone_synergies(card_a);
CREATE INDEX IF NOT EXISTS idx_mtgzone_synergies_card_b ON mtgzone_synergies(card_b);

-- MTGZone articles metadata table
-- Tracks which articles have been processed

CREATE TABLE IF NOT EXISTS mtgzone_articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    article_type TEXT, -- 'deck_guide', 'tier_list', 'set_review', 'strategy'
    format TEXT,
    archetype TEXT,
    published_at TIMESTAMP,
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    cards_mentioned TEXT -- JSON array of card names mentioned
);

CREATE INDEX IF NOT EXISTS idx_mtgzone_articles_type ON mtgzone_articles(article_type);
CREATE INDEX IF NOT EXISTS idx_mtgzone_articles_format ON mtgzone_articles(format);
