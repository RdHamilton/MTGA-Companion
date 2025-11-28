package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// MetadataStaleness contains counts of fresh/stale metadata.
type MetadataStaleness struct {
	Total     int
	Fresh     int
	Stale     int
	VeryStale int
}

// StaleCard represents a card with stale metadata.
type StaleCard struct {
	ArenaID     string
	SetCode     string
	LastUpdated string
}

// SetRarityCount represents card counts for a set and rarity.
type SetRarityCount struct {
	SetCode string
	SetName string
	Rarity  string
	Total   int
}

// CardSetInfo represents the set/rarity info for a card.
type CardSetInfo struct {
	ArenaID string
	SetCode string
	Rarity  string
}

// SetCardRepository provides methods for managing set cards cached from Scryfall.
type SetCardRepository interface {
	// SaveCard saves a set card to the database.
	SaveCard(ctx context.Context, card *models.SetCard) error

	// SaveCards saves multiple set cards in a batch.
	SaveCards(ctx context.Context, cards []*models.SetCard) error

	// GetCardByArenaID retrieves a card by its Arena ID.
	GetCardByArenaID(ctx context.Context, arenaID string) (*models.SetCard, error)

	// GetCardsBySet retrieves all cards for a given set code.
	GetCardsBySet(ctx context.Context, setCode string) ([]*models.SetCard, error)

	// SearchCards searches for cards by name across all cached sets.
	// Optionally filter by set codes. Returns up to limit results.
	SearchCards(ctx context.Context, query string, setCodes []string, limit int) ([]*models.SetCard, error)

	// IsSetCached checks if a set has been cached.
	IsSetCached(ctx context.Context, setCode string) (bool, error)

	// GetCachedSets returns a list of all cached set codes.
	GetCachedSets(ctx context.Context) ([]string, error)

	// DeleteSet removes all cards for a given set (for cache invalidation).
	DeleteSet(ctx context.Context, setCode string) error

	// Staleness tracking methods
	// GetMetadataStaleness returns counts of fresh, stale, and very stale cards.
	GetMetadataStaleness(ctx context.Context, staleAgeSeconds, veryStaleAgeSeconds int) (*MetadataStaleness, error)

	// GetStaleCards returns cards with stale metadata, ordered by oldest first.
	GetStaleCards(ctx context.Context, staleAgeSeconds, limit int) ([]*StaleCard, error)

	// Set completion methods
	// GetSetRarityCounts returns card counts grouped by set and rarity, with set names.
	GetSetRarityCounts(ctx context.Context) ([]*SetRarityCount, error)

	// GetAllCardSetInfo returns arena_id, set_code, and rarity for all cards.
	GetAllCardSetInfo(ctx context.Context) ([]*CardSetInfo, error)
}

type setCardRepository struct {
	db *sql.DB
}

// NewSetCardRepository creates a new set card repository.
func NewSetCardRepository(db *sql.DB) SetCardRepository {
	return &setCardRepository{db: db}
}

// SaveCard saves a set card to the database.
func (r *setCardRepository) SaveCard(ctx context.Context, card *models.SetCard) error {
	// Marshal arrays to JSON
	typesJSON, err := json.Marshal(card.Types)
	if err != nil {
		return err
	}

	colorsJSON, err := json.Marshal(card.Colors)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO set_cards (
			set_code, arena_id, scryfall_id, name, mana_cost, cmc, types, colors,
			rarity, text, power, toughness, image_url, image_url_small, image_url_art, fetched_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(set_code, arena_id) DO UPDATE SET
			scryfall_id = excluded.scryfall_id,
			name = excluded.name,
			mana_cost = excluded.mana_cost,
			cmc = excluded.cmc,
			types = excluded.types,
			colors = excluded.colors,
			rarity = excluded.rarity,
			text = excluded.text,
			power = excluded.power,
			toughness = excluded.toughness,
			image_url = excluded.image_url,
			image_url_small = excluded.image_url_small,
			image_url_art = excluded.image_url_art,
			fetched_at = excluded.fetched_at
	`
	result, err := r.db.ExecContext(ctx, query,
		card.SetCode,
		card.ArenaID,
		card.ScryfallID,
		card.Name,
		card.ManaCost,
		card.CMC,
		string(typesJSON),
		string(colorsJSON),
		card.Rarity,
		card.Text,
		card.Power,
		card.Toughness,
		card.ImageURL,
		card.ImageURLSmall,
		card.ImageURLArt,
		card.FetchedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil {
		card.ID = int(id)
	}

	return nil
}

// SaveCards saves multiple set cards in a batch.
func (r *setCardRepository) SaveCards(ctx context.Context, cards []*models.SetCard) error {
	// Start a transaction for batch insert
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO set_cards (
			set_code, arena_id, scryfall_id, name, mana_cost, cmc, types, colors,
			rarity, text, power, toughness, image_url, image_url_small, image_url_art, fetched_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(set_code, arena_id) DO UPDATE SET
			scryfall_id = excluded.scryfall_id,
			name = excluded.name,
			mana_cost = excluded.mana_cost,
			cmc = excluded.cmc,
			types = excluded.types,
			colors = excluded.colors,
			rarity = excluded.rarity,
			text = excluded.text,
			power = excluded.power,
			toughness = excluded.toughness,
			image_url = excluded.image_url,
			image_url_small = excluded.image_url_small,
			image_url_art = excluded.image_url_art,
			fetched_at = excluded.fetched_at
	`)
	if err != nil {
		return err
	}
	defer func() {
		_ = stmt.Close()
	}()

	for _, card := range cards {
		typesJSON, err := json.Marshal(card.Types)
		if err != nil {
			return err
		}

		colorsJSON, err := json.Marshal(card.Colors)
		if err != nil {
			return err
		}

		_, err = stmt.ExecContext(ctx,
			card.SetCode,
			card.ArenaID,
			card.ScryfallID,
			card.Name,
			card.ManaCost,
			card.CMC,
			string(typesJSON),
			string(colorsJSON),
			card.Rarity,
			card.Text,
			card.Power,
			card.Toughness,
			card.ImageURL,
			card.ImageURLSmall,
			card.ImageURLArt,
			card.FetchedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetCardByArenaID retrieves a card by its Arena ID.
func (r *setCardRepository) GetCardByArenaID(ctx context.Context, arenaID string) (*models.SetCard, error) {
	query := `
		SELECT id, set_code, arena_id, scryfall_id, name, mana_cost, cmc, types, colors,
			   rarity, text, power, toughness, image_url, image_url_small, image_url_art, fetched_at
		FROM set_cards
		WHERE arena_id = ?
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, arenaID)

	card := &models.SetCard{}
	var typesJSON, colorsJSON string

	err := row.Scan(
		&card.ID,
		&card.SetCode,
		&card.ArenaID,
		&card.ScryfallID,
		&card.Name,
		&card.ManaCost,
		&card.CMC,
		&typesJSON,
		&colorsJSON,
		&card.Rarity,
		&card.Text,
		&card.Power,
		&card.Toughness,
		&card.ImageURL,
		&card.ImageURLSmall,
		&card.ImageURLArt,
		&card.FetchedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse JSON arrays
	if err := json.Unmarshal([]byte(typesJSON), &card.Types); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(colorsJSON), &card.Colors); err != nil {
		return nil, err
	}

	return card, nil
}

// GetCardsBySet retrieves all cards for a given set code.
func (r *setCardRepository) GetCardsBySet(ctx context.Context, setCode string) ([]*models.SetCard, error) {
	query := `
		SELECT id, set_code, arena_id, scryfall_id, name, mana_cost, cmc, types, colors,
			   rarity, text, power, toughness, image_url, image_url_small, image_url_art, fetched_at
		FROM set_cards
		WHERE set_code = ?
		ORDER BY name
	`
	rows, err := r.db.QueryContext(ctx, query, setCode)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	cards := []*models.SetCard{}
	for rows.Next() {
		card := &models.SetCard{}
		var typesJSON, colorsJSON string

		err := rows.Scan(
			&card.ID,
			&card.SetCode,
			&card.ArenaID,
			&card.ScryfallID,
			&card.Name,
			&card.ManaCost,
			&card.CMC,
			&typesJSON,
			&colorsJSON,
			&card.Rarity,
			&card.Text,
			&card.Power,
			&card.Toughness,
			&card.ImageURL,
			&card.ImageURLSmall,
			&card.ImageURLArt,
			&card.FetchedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON arrays
		if err := json.Unmarshal([]byte(typesJSON), &card.Types); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(colorsJSON), &card.Colors); err != nil {
			return nil, err
		}

		cards = append(cards, card)
	}

	return cards, rows.Err()
}

// IsSetCached checks if a set has been cached.
func (r *setCardRepository) IsSetCached(ctx context.Context, setCode string) (bool, error) {
	query := `SELECT COUNT(*) FROM set_cards WHERE set_code = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, setCode).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetCachedSets returns a list of all cached set codes.
func (r *setCardRepository) GetCachedSets(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT set_code FROM set_cards ORDER BY set_code`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	sets := []string{}
	for rows.Next() {
		var setCode string
		if err := rows.Scan(&setCode); err != nil {
			return nil, err
		}
		sets = append(sets, setCode)
	}

	return sets, rows.Err()
}

// DeleteSet removes all cards for a given set (for cache invalidation).
func (r *setCardRepository) DeleteSet(ctx context.Context, setCode string) error {
	query := `DELETE FROM set_cards WHERE set_code = ?`
	_, err := r.db.ExecContext(ctx, query, setCode)
	return err
}

// SearchCards searches for cards by name across all cached sets.
// If setCodes is empty, searches all sets. Returns up to limit results.
func (r *setCardRepository) SearchCards(ctx context.Context, query string, setCodes []string, limit int) ([]*models.SetCard, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > 200 {
		limit = 200 // Max limit to prevent performance issues
	}

	var sqlQuery string
	var args []interface{}

	if len(setCodes) > 0 {
		// Build placeholders for IN clause
		placeholders := make([]string, len(setCodes))
		for i, code := range setCodes {
			placeholders[i] = "?"
			args = append(args, code)
		}
		sqlQuery = `
			SELECT id, set_code, arena_id, scryfall_id, name, mana_cost, cmc, types, colors,
				   rarity, text, power, toughness, image_url, image_url_small, image_url_art, fetched_at
			FROM set_cards
			WHERE name LIKE ? AND set_code IN (` + joinStrings(placeholders, ",") + `)
			ORDER BY name
			LIMIT ?
		`
		// Prepend the name query, append the limit
		args = append([]interface{}{"%" + query + "%"}, args...)
		args = append(args, limit)
	} else {
		sqlQuery = `
			SELECT id, set_code, arena_id, scryfall_id, name, mana_cost, cmc, types, colors,
				   rarity, text, power, toughness, image_url, image_url_small, image_url_art, fetched_at
			FROM set_cards
			WHERE name LIKE ?
			ORDER BY name
			LIMIT ?
		`
		args = []interface{}{"%" + query + "%", limit}
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	cards := []*models.SetCard{}
	for rows.Next() {
		card := &models.SetCard{}
		var typesJSON, colorsJSON string

		err := rows.Scan(
			&card.ID,
			&card.SetCode,
			&card.ArenaID,
			&card.ScryfallID,
			&card.Name,
			&card.ManaCost,
			&card.CMC,
			&typesJSON,
			&colorsJSON,
			&card.Rarity,
			&card.Text,
			&card.Power,
			&card.Toughness,
			&card.ImageURL,
			&card.ImageURLSmall,
			&card.ImageURLArt,
			&card.FetchedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON arrays
		if err := json.Unmarshal([]byte(typesJSON), &card.Types); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(colorsJSON), &card.Colors); err != nil {
			return nil, err
		}

		cards = append(cards, card)
	}

	return cards, rows.Err()
}

// joinStrings joins strings with a separator (simple helper to avoid importing strings package).
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// GetMetadataStaleness returns counts of fresh, stale, and very stale cards.
func (r *setCardRepository) GetMetadataStaleness(ctx context.Context, staleAgeSeconds, veryStaleAgeSeconds int) (*MetadataStaleness, error) {
	query := `
		SELECT
			COUNT(*) as total,
			COALESCE(SUM(CASE WHEN fetched_at >= datetime('now', '-' || ? || ' seconds') THEN 1 ELSE 0 END), 0) as fresh,
			COALESCE(SUM(CASE WHEN fetched_at < datetime('now', '-' || ? || ' seconds')
				AND fetched_at >= datetime('now', '-' || ? || ' seconds') THEN 1 ELSE 0 END), 0) as stale,
			COALESCE(SUM(CASE WHEN fetched_at < datetime('now', '-' || ? || ' seconds') THEN 1 ELSE 0 END), 0) as very_stale
		FROM set_cards
		WHERE fetched_at IS NOT NULL
	`

	var result MetadataStaleness
	err := r.db.QueryRowContext(ctx, query,
		staleAgeSeconds,
		staleAgeSeconds,
		veryStaleAgeSeconds,
		veryStaleAgeSeconds,
	).Scan(&result.Total, &result.Fresh, &result.Stale, &result.VeryStale)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetStaleCards returns cards with stale metadata, ordered by oldest first.
func (r *setCardRepository) GetStaleCards(ctx context.Context, staleAgeSeconds, limit int) ([]*StaleCard, error) {
	query := `
		SELECT arena_id, set_code, fetched_at
		FROM set_cards
		WHERE fetched_at < datetime('now', '-' || ? || ' seconds')
		ORDER BY fetched_at ASC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, staleAgeSeconds, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var cards []*StaleCard
	for rows.Next() {
		card := &StaleCard{}
		if err := rows.Scan(&card.ArenaID, &card.SetCode, &card.LastUpdated); err != nil {
			continue
		}
		cards = append(cards, card)
	}

	return cards, rows.Err()
}

// GetSetRarityCounts returns card counts grouped by set and rarity, with set names.
func (r *setCardRepository) GetSetRarityCounts(ctx context.Context) ([]*SetRarityCount, error) {
	query := `
		SELECT
			sc.set_code,
			COALESCE(st.name, UPPER(sc.set_code)) as set_name,
			sc.rarity,
			COUNT(*) as total
		FROM set_cards sc
		LEFT JOIN sets st ON sc.set_code = st.code
		GROUP BY sc.set_code, sc.rarity
		ORDER BY sc.set_code, sc.rarity
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var counts []*SetRarityCount
	for rows.Next() {
		count := &SetRarityCount{}
		if err := rows.Scan(&count.SetCode, &count.SetName, &count.Rarity, &count.Total); err != nil {
			return nil, err
		}
		counts = append(counts, count)
	}

	return counts, rows.Err()
}

// GetAllCardSetInfo returns arena_id, set_code, and rarity for all cards.
func (r *setCardRepository) GetAllCardSetInfo(ctx context.Context) ([]*CardSetInfo, error) {
	query := `
		SELECT arena_id, set_code, rarity
		FROM set_cards
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var cardInfos []*CardSetInfo
	for rows.Next() {
		card := &CardSetInfo{}
		if err := rows.Scan(&card.ArenaID, &card.SetCode, &card.Rarity); err != nil {
			return nil, err
		}
		cardInfos = append(cardInfos, card)
	}

	return cardInfos, rows.Err()
}
