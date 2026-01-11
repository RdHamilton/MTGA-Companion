package models

import "time"

// EmbeddingSource indicates how the embedding was generated.
type EmbeddingSource string

const (
	EmbeddingSourceCharacteristics EmbeddingSource = "characteristics"
	EmbeddingSourceCooccurrence    EmbeddingSource = "cooccurrence"
	EmbeddingSourceHybrid          EmbeddingSource = "hybrid"
)

// CardEmbedding represents a vector embedding for a card.
type CardEmbedding struct {
	ID               int64           `json:"id" db:"id"`
	ArenaID          int             `json:"arenaId" db:"arena_id"`
	CardName         string          `json:"cardName" db:"card_name"`
	Embedding        []float64       `json:"embedding" db:"-"` // Stored as JSON
	EmbeddingJSON    string          `json:"-" db:"embedding"` // JSON string for DB
	EmbeddingVersion int             `json:"embeddingVersion" db:"embedding_version"`
	Source           EmbeddingSource `json:"source" db:"source"`
	CreatedAt        time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time       `json:"updatedAt" db:"updated_at"`
}

// CardSimilarity represents a pre-computed similarity between two cards.
type CardSimilarity struct {
	ID              int64     `json:"id" db:"id"`
	CardArenaID     int       `json:"cardArenaId" db:"card_arena_id"`
	SimilarArenaID  int       `json:"similarArenaId" db:"similar_arena_id"`
	SimilarityScore float64   `json:"similarityScore" db:"similarity_score"`
	Rank            int       `json:"rank" db:"rank"`
	CreatedAt       time.Time `json:"createdAt" db:"created_at"`
}

// SimilarCard represents a similar card with metadata.
type SimilarCard struct {
	ArenaID         int     `json:"arenaId"`
	CardName        string  `json:"cardName"`
	SimilarityScore float64 `json:"similarityScore"`
	Rank            int     `json:"rank"`
}

// EmbeddingDimensions defines the size of the embedding vector.
const EmbeddingDimensions = 64

// EmbeddingVersion is the current version of the embedding algorithm.
// Increment this when the embedding generation logic changes.
const EmbeddingVersion = 1
