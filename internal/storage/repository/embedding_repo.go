package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// EmbeddingRepository handles card embedding operations.
type EmbeddingRepository interface {
	// Embedding operations
	UpsertEmbedding(ctx context.Context, embedding *models.CardEmbedding) error
	GetEmbedding(ctx context.Context, arenaID int) (*models.CardEmbedding, error)
	GetEmbeddings(ctx context.Context, arenaIDs []int) ([]*models.CardEmbedding, error)
	GetAllEmbeddings(ctx context.Context) ([]*models.CardEmbedding, error)
	DeleteEmbedding(ctx context.Context, arenaID int) error
	GetEmbeddingCount(ctx context.Context) (int, error)
	GetOutdatedEmbeddings(ctx context.Context, version int) ([]int, error)

	// Similarity cache operations
	UpsertSimilarity(ctx context.Context, similarity *models.CardSimilarity) error
	BulkUpsertSimilarities(ctx context.Context, similarities []*models.CardSimilarity) error
	GetSimilarCards(ctx context.Context, arenaID int, limit int) ([]*models.SimilarCard, error)
	GetSimilarityBetween(ctx context.Context, arenaID1, arenaID2 int) (float64, error)
	ClearSimilarityCache(ctx context.Context) error
	ClearSimilarityCacheForCard(ctx context.Context, arenaID int) error
}

type embeddingRepo struct {
	db *sql.DB
}

// NewEmbeddingRepository creates a new embedding repository.
func NewEmbeddingRepository(db *sql.DB) EmbeddingRepository {
	return &embeddingRepo{db: db}
}

// UpsertEmbedding inserts or updates a card embedding.
func (r *embeddingRepo) UpsertEmbedding(ctx context.Context, embedding *models.CardEmbedding) error {
	// Convert embedding to JSON
	embeddingJSON, err := json.Marshal(embedding.Embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	query := `
		INSERT INTO card_embeddings (arena_id, card_name, embedding, embedding_version, source, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(arena_id) DO UPDATE SET
			card_name = excluded.card_name,
			embedding = excluded.embedding,
			embedding_version = excluded.embedding_version,
			source = excluded.source,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err = r.db.ExecContext(ctx, query,
		embedding.ArenaID, embedding.CardName, string(embeddingJSON),
		embedding.EmbeddingVersion, embedding.Source,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert embedding: %w", err)
	}

	return nil
}

// GetEmbedding retrieves an embedding by arena ID.
func (r *embeddingRepo) GetEmbedding(ctx context.Context, arenaID int) (*models.CardEmbedding, error) {
	query := `
		SELECT id, arena_id, card_name, embedding, embedding_version, source, created_at, updated_at
		FROM card_embeddings
		WHERE arena_id = ?
	`

	var e models.CardEmbedding
	var embeddingJSON string

	err := r.db.QueryRowContext(ctx, query, arenaID).Scan(
		&e.ID, &e.ArenaID, &e.CardName, &embeddingJSON,
		&e.EmbeddingVersion, &e.Source, &e.CreatedAt, &e.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding: %w", err)
	}

	// Parse embedding JSON
	if err := json.Unmarshal([]byte(embeddingJSON), &e.Embedding); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embedding: %w", err)
	}

	return &e, nil
}

// GetEmbeddings retrieves embeddings for multiple arena IDs.
func (r *embeddingRepo) GetEmbeddings(ctx context.Context, arenaIDs []int) ([]*models.CardEmbedding, error) {
	if len(arenaIDs) == 0 {
		return nil, nil
	}

	// Build placeholders
	placeholders := ""
	args := make([]interface{}, len(arenaIDs))
	for i, id := range arenaIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, arena_id, card_name, embedding, embedding_version, source, created_at, updated_at
		FROM card_embeddings
		WHERE arena_id IN (%s)
	`, placeholders)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get embeddings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var embeddings []*models.CardEmbedding
	for rows.Next() {
		var e models.CardEmbedding
		var embeddingJSON string

		if err := rows.Scan(
			&e.ID, &e.ArenaID, &e.CardName, &embeddingJSON,
			&e.EmbeddingVersion, &e.Source, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan embedding: %w", err)
		}

		if err := json.Unmarshal([]byte(embeddingJSON), &e.Embedding); err != nil {
			return nil, fmt.Errorf("failed to unmarshal embedding: %w", err)
		}

		embeddings = append(embeddings, &e)
	}

	return embeddings, nil
}

// GetAllEmbeddings retrieves all embeddings.
func (r *embeddingRepo) GetAllEmbeddings(ctx context.Context) ([]*models.CardEmbedding, error) {
	query := `
		SELECT id, arena_id, card_name, embedding, embedding_version, source, created_at, updated_at
		FROM card_embeddings
		ORDER BY arena_id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all embeddings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var embeddings []*models.CardEmbedding
	for rows.Next() {
		var e models.CardEmbedding
		var embeddingJSON string

		if err := rows.Scan(
			&e.ID, &e.ArenaID, &e.CardName, &embeddingJSON,
			&e.EmbeddingVersion, &e.Source, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan embedding: %w", err)
		}

		if err := json.Unmarshal([]byte(embeddingJSON), &e.Embedding); err != nil {
			return nil, fmt.Errorf("failed to unmarshal embedding: %w", err)
		}

		embeddings = append(embeddings, &e)
	}

	return embeddings, nil
}

// DeleteEmbedding deletes an embedding by arena ID.
func (r *embeddingRepo) DeleteEmbedding(ctx context.Context, arenaID int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM card_embeddings WHERE arena_id = ?", arenaID)
	if err != nil {
		return fmt.Errorf("failed to delete embedding: %w", err)
	}
	return nil
}

// GetEmbeddingCount returns the total number of embeddings.
func (r *embeddingRepo) GetEmbeddingCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM card_embeddings").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get embedding count: %w", err)
	}
	return count, nil
}

// GetOutdatedEmbeddings returns arena IDs of embeddings with version < specified version.
func (r *embeddingRepo) GetOutdatedEmbeddings(ctx context.Context, version int) ([]int, error) {
	query := `SELECT arena_id FROM card_embeddings WHERE embedding_version < ?`

	rows, err := r.db.QueryContext(ctx, query, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get outdated embeddings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var arenaIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan arena ID: %w", err)
		}
		arenaIDs = append(arenaIDs, id)
	}

	return arenaIDs, nil
}

// UpsertSimilarity inserts or updates a similarity record.
func (r *embeddingRepo) UpsertSimilarity(ctx context.Context, similarity *models.CardSimilarity) error {
	query := `
		INSERT INTO card_similarity_cache (card_arena_id, similar_arena_id, similarity_score, rank)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(card_arena_id, similar_arena_id) DO UPDATE SET
			similarity_score = excluded.similarity_score,
			rank = excluded.rank
	`

	_, err := r.db.ExecContext(ctx, query,
		similarity.CardArenaID, similarity.SimilarArenaID,
		similarity.SimilarityScore, similarity.Rank,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert similarity: %w", err)
	}

	return nil
}

// BulkUpsertSimilarities inserts or updates multiple similarity records.
func (r *embeddingRepo) BulkUpsertSimilarities(ctx context.Context, similarities []*models.CardSimilarity) error {
	if len(similarities) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO card_similarity_cache (card_arena_id, similar_arena_id, similarity_score, rank)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(card_arena_id, similar_arena_id) DO UPDATE SET
			similarity_score = excluded.similarity_score,
			rank = excluded.rank
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, s := range similarities {
		_, err := stmt.ExecContext(ctx, s.CardArenaID, s.SimilarArenaID, s.SimilarityScore, s.Rank)
		if err != nil {
			return fmt.Errorf("failed to insert similarity: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// GetSimilarCards retrieves the most similar cards for a given card.
func (r *embeddingRepo) GetSimilarCards(ctx context.Context, arenaID int, limit int) ([]*models.SimilarCard, error) {
	query := `
		SELECT c.similar_arena_id, e.card_name, c.similarity_score, c.rank
		FROM card_similarity_cache c
		JOIN card_embeddings e ON c.similar_arena_id = e.arena_id
		WHERE c.card_arena_id = ?
		ORDER BY c.rank ASC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, arenaID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get similar cards: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var similarCards []*models.SimilarCard
	for rows.Next() {
		var s models.SimilarCard
		if err := rows.Scan(&s.ArenaID, &s.CardName, &s.SimilarityScore, &s.Rank); err != nil {
			return nil, fmt.Errorf("failed to scan similar card: %w", err)
		}
		similarCards = append(similarCards, &s)
	}

	return similarCards, nil
}

// GetSimilarityBetween returns the similarity score between two cards.
func (r *embeddingRepo) GetSimilarityBetween(ctx context.Context, arenaID1, arenaID2 int) (float64, error) {
	query := `
		SELECT similarity_score
		FROM card_similarity_cache
		WHERE (card_arena_id = ? AND similar_arena_id = ?)
		   OR (card_arena_id = ? AND similar_arena_id = ?)
		LIMIT 1
	`

	var score float64
	err := r.db.QueryRowContext(ctx, query, arenaID1, arenaID2, arenaID2, arenaID1).Scan(&score)

	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get similarity: %w", err)
	}

	return score, nil
}

// ClearSimilarityCache removes all cached similarities.
func (r *embeddingRepo) ClearSimilarityCache(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM card_similarity_cache")
	if err != nil {
		return fmt.Errorf("failed to clear similarity cache: %w", err)
	}
	return nil
}

// ClearSimilarityCacheForCard removes cached similarities for a specific card.
func (r *embeddingRepo) ClearSimilarityCacheForCard(ctx context.Context, arenaID int) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM card_similarity_cache WHERE card_arena_id = ? OR similar_arena_id = ?",
		arenaID, arenaID,
	)
	if err != nil {
		return fmt.Errorf("failed to clear similarity cache for card: %w", err)
	}
	return nil
}
