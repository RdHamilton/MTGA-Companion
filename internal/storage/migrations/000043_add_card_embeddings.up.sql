-- Card embeddings for semantic similarity
CREATE TABLE IF NOT EXISTS card_embeddings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    arena_id INTEGER NOT NULL UNIQUE,
    card_name TEXT NOT NULL,
    embedding TEXT NOT NULL,  -- JSON array of float64 values
    embedding_version INTEGER NOT NULL DEFAULT 1,
    source TEXT DEFAULT 'characteristics',  -- 'characteristics', 'cooccurrence', 'hybrid'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_card_embeddings_arena_id ON card_embeddings(arena_id);
CREATE INDEX IF NOT EXISTS idx_card_embeddings_name ON card_embeddings(card_name);
CREATE INDEX IF NOT EXISTS idx_card_embeddings_version ON card_embeddings(embedding_version);

-- Pre-computed similarity cache for top-k similar cards
CREATE TABLE IF NOT EXISTS card_similarity_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_arena_id INTEGER NOT NULL,
    similar_arena_id INTEGER NOT NULL,
    similarity_score REAL NOT NULL,
    rank INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(card_arena_id, similar_arena_id)
);

CREATE INDEX IF NOT EXISTS idx_similarity_cache_card ON card_similarity_cache(card_arena_id);
CREATE INDEX IF NOT EXISTS idx_similarity_cache_score ON card_similarity_cache(similarity_score DESC);
