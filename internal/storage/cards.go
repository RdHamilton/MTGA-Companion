package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
)

// Card represents a card in the local cache with Scryfall data.
type Card struct {
	ID              string
	ArenaID         *int
	Name            string
	ManaCost        string
	CMC             float64
	TypeLine        string
	OracleText      string
	Colors          []string
	ColorIdentity   []string
	Rarity          string
	SetCode         string
	CollectorNumber string
	Power           string
	Toughness       string
	Loyalty         string
	ImageURIs       *scryfall.ImageURIs
	Layout          string
	CardFaces       []scryfall.CardFace
	Legalities      map[string]string
	ReleasedAt      string
	CachedAt        time.Time
	LastUpdated     time.Time
}

// Set represents a Magic set in the local cache.
type Set struct {
	Code        string
	Name        string
	ReleasedAt  string
	CardCount   int
	SetType     string
	IconSVGURI  string
	CachedAt    time.Time
	LastUpdated time.Time
}

// SaveCard saves or updates a card in the database.
func (s *Service) SaveCard(ctx context.Context, card *Card) error {
	// Serialize JSON fields
	colorsJSON, err := json.Marshal(card.Colors)
	if err != nil {
		return fmt.Errorf("failed to marshal colors: %w", err)
	}

	colorIdentityJSON, err := json.Marshal(card.ColorIdentity)
	if err != nil {
		return fmt.Errorf("failed to marshal color_identity: %w", err)
	}

	imageURIsJSON, err := json.Marshal(card.ImageURIs)
	if err != nil {
		return fmt.Errorf("failed to marshal image_uris: %w", err)
	}

	cardFacesJSON, err := json.Marshal(card.CardFaces)
	if err != nil {
		return fmt.Errorf("failed to marshal card_faces: %w", err)
	}

	legalitiesJSON, err := json.Marshal(card.Legalities)
	if err != nil {
		return fmt.Errorf("failed to marshal legalities: %w", err)
	}

	query := `
		INSERT INTO cards (
			id, arena_id, name, mana_cost, cmc, type_line, oracle_text,
			colors, color_identity, rarity, set_code, collector_number,
			power, toughness, loyalty, image_uris, layout, card_faces,
			legalities, released_at, last_updated
		) VALUES (
			?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, CURRENT_TIMESTAMP
		)
		ON CONFLICT(id) DO UPDATE SET
			arena_id = excluded.arena_id,
			name = excluded.name,
			mana_cost = excluded.mana_cost,
			cmc = excluded.cmc,
			type_line = excluded.type_line,
			oracle_text = excluded.oracle_text,
			colors = excluded.colors,
			color_identity = excluded.color_identity,
			rarity = excluded.rarity,
			set_code = excluded.set_code,
			collector_number = excluded.collector_number,
			power = excluded.power,
			toughness = excluded.toughness,
			loyalty = excluded.loyalty,
			image_uris = excluded.image_uris,
			layout = excluded.layout,
			card_faces = excluded.card_faces,
			legalities = excluded.legalities,
			released_at = excluded.released_at,
			last_updated = CURRENT_TIMESTAMP
	`

	_, err = s.db.Conn().ExecContext(ctx, query,
		card.ID, card.ArenaID, card.Name, card.ManaCost, card.CMC, card.TypeLine, card.OracleText,
		string(colorsJSON), string(colorIdentityJSON), card.Rarity, card.SetCode, card.CollectorNumber,
		card.Power, card.Toughness, card.Loyalty, string(imageURIsJSON), card.Layout, string(cardFacesJSON),
		string(legalitiesJSON), card.ReleasedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save card: %w", err)
	}

	return nil
}

// GetCardByArenaID retrieves a card by its Arena ID.
func (s *Service) GetCardByArenaID(ctx context.Context, arenaID int) (*Card, error) {
	query := `
		SELECT id, arena_id, name, mana_cost, cmc, type_line, oracle_text,
			colors, color_identity, rarity, set_code, collector_number,
			power, toughness, loyalty, image_uris, layout, card_faces,
			legalities, released_at, cached_at, last_updated
		FROM cards
		WHERE arena_id = ?
	`

	var card Card
	var colorsJSON, colorIdentityJSON, imageURIsJSON, cardFacesJSON, legalitiesJSON string

	err := s.db.Conn().QueryRowContext(ctx, query, arenaID).Scan(
		&card.ID, &card.ArenaID, &card.Name, &card.ManaCost, &card.CMC, &card.TypeLine, &card.OracleText,
		&colorsJSON, &colorIdentityJSON, &card.Rarity, &card.SetCode, &card.CollectorNumber,
		&card.Power, &card.Toughness, &card.Loyalty, &imageURIsJSON, &card.Layout, &cardFacesJSON,
		&legalitiesJSON, &card.ReleasedAt, &card.CachedAt, &card.LastUpdated,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get card by arena ID: %w", err)
	}

	// Deserialize JSON fields
	if err := json.Unmarshal([]byte(colorsJSON), &card.Colors); err != nil {
		return nil, fmt.Errorf("failed to unmarshal colors: %w", err)
	}

	if err := json.Unmarshal([]byte(colorIdentityJSON), &card.ColorIdentity); err != nil {
		return nil, fmt.Errorf("failed to unmarshal color_identity: %w", err)
	}

	if imageURIsJSON != "" && imageURIsJSON != "null" {
		if err := json.Unmarshal([]byte(imageURIsJSON), &card.ImageURIs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal image_uris: %w", err)
		}
	}

	if cardFacesJSON != "" && cardFacesJSON != "null" && cardFacesJSON != "[]" {
		if err := json.Unmarshal([]byte(cardFacesJSON), &card.CardFaces); err != nil {
			return nil, fmt.Errorf("failed to unmarshal card_faces: %w", err)
		}
	}

	if err := json.Unmarshal([]byte(legalitiesJSON), &card.Legalities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal legalities: %w", err)
	}

	return &card, nil
}

// GetCard retrieves a card by its Scryfall ID.
func (s *Service) GetCard(ctx context.Context, id string) (*Card, error) {
	query := `
		SELECT id, arena_id, name, mana_cost, cmc, type_line, oracle_text,
			colors, color_identity, rarity, set_code, collector_number,
			power, toughness, loyalty, image_uris, layout, card_faces,
			legalities, released_at, cached_at, last_updated
		FROM cards
		WHERE id = ?
	`

	var card Card
	var colorsJSON, colorIdentityJSON, imageURIsJSON, cardFacesJSON, legalitiesJSON string

	err := s.db.Conn().QueryRowContext(ctx, query, id).Scan(
		&card.ID, &card.ArenaID, &card.Name, &card.ManaCost, &card.CMC, &card.TypeLine, &card.OracleText,
		&colorsJSON, &colorIdentityJSON, &card.Rarity, &card.SetCode, &card.CollectorNumber,
		&card.Power, &card.Toughness, &card.Loyalty, &imageURIsJSON, &card.Layout, &cardFacesJSON,
		&legalitiesJSON, &card.ReleasedAt, &card.CachedAt, &card.LastUpdated,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get card: %w", err)
	}

	// Deserialize JSON fields
	if err := json.Unmarshal([]byte(colorsJSON), &card.Colors); err != nil {
		return nil, fmt.Errorf("failed to unmarshal colors: %w", err)
	}

	if err := json.Unmarshal([]byte(colorIdentityJSON), &card.ColorIdentity); err != nil {
		return nil, fmt.Errorf("failed to unmarshal color_identity: %w", err)
	}

	if imageURIsJSON != "" && imageURIsJSON != "null" {
		if err := json.Unmarshal([]byte(imageURIsJSON), &card.ImageURIs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal image_uris: %w", err)
		}
	}

	if cardFacesJSON != "" && cardFacesJSON != "null" && cardFacesJSON != "[]" {
		if err := json.Unmarshal([]byte(cardFacesJSON), &card.CardFaces); err != nil {
			return nil, fmt.Errorf("failed to unmarshal card_faces: %w", err)
		}
	}

	if err := json.Unmarshal([]byte(legalitiesJSON), &card.Legalities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal legalities: %w", err)
	}

	return &card, nil
}

// SearchCards searches for cards by name (case-insensitive partial match).
func (s *Service) SearchCards(ctx context.Context, query string) ([]*Card, error) {
	sqlQuery := `
		SELECT id, arena_id, name, mana_cost, cmc, type_line, oracle_text,
			colors, color_identity, rarity, set_code, collector_number,
			power, toughness, loyalty, image_uris, layout, card_faces,
			legalities, released_at, cached_at, last_updated
		FROM cards
		WHERE name LIKE ?
		ORDER BY name
		LIMIT 100
	`

	rows, err := s.db.Conn().QueryContext(ctx, sqlQuery, "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search cards: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var cards []*Card
	for rows.Next() {
		var card Card
		var colorsJSON, colorIdentityJSON, imageURIsJSON, cardFacesJSON, legalitiesJSON string

		err := rows.Scan(
			&card.ID, &card.ArenaID, &card.Name, &card.ManaCost, &card.CMC, &card.TypeLine, &card.OracleText,
			&colorsJSON, &colorIdentityJSON, &card.Rarity, &card.SetCode, &card.CollectorNumber,
			&card.Power, &card.Toughness, &card.Loyalty, &imageURIsJSON, &card.Layout, &cardFacesJSON,
			&legalitiesJSON, &card.ReleasedAt, &card.CachedAt, &card.LastUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}

		// Deserialize JSON fields
		if err := json.Unmarshal([]byte(colorsJSON), &card.Colors); err != nil {
			return nil, fmt.Errorf("failed to unmarshal colors: %w", err)
		}

		if err := json.Unmarshal([]byte(colorIdentityJSON), &card.ColorIdentity); err != nil {
			return nil, fmt.Errorf("failed to unmarshal color_identity: %w", err)
		}

		if imageURIsJSON != "" && imageURIsJSON != "null" {
			if err := json.Unmarshal([]byte(imageURIsJSON), &card.ImageURIs); err != nil {
				return nil, fmt.Errorf("failed to unmarshal image_uris: %w", err)
			}
		}

		if cardFacesJSON != "" && cardFacesJSON != "null" && cardFacesJSON != "[]" {
			if err := json.Unmarshal([]byte(cardFacesJSON), &card.CardFaces); err != nil {
				return nil, fmt.Errorf("failed to unmarshal card_faces: %w", err)
			}
		}

		if err := json.Unmarshal([]byte(legalitiesJSON), &card.Legalities); err != nil {
			return nil, fmt.Errorf("failed to unmarshal legalities: %w", err)
		}

		cards = append(cards, &card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cards: %w", err)
	}

	return cards, nil
}

// GetCardsBySet retrieves all cards for a given set code.
func (s *Service) GetCardsBySet(ctx context.Context, setCode string) ([]*Card, error) {
	query := `
		SELECT id, arena_id, name, mana_cost, cmc, type_line, oracle_text,
			colors, color_identity, rarity, set_code, collector_number,
			power, toughness, loyalty, image_uris, layout, card_faces,
			legalities, released_at, cached_at, last_updated
		FROM cards
		WHERE set_code = ?
		ORDER BY collector_number
	`

	rows, err := s.db.Conn().QueryContext(ctx, query, setCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get cards by set: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var cards []*Card
	for rows.Next() {
		var card Card
		var colorsJSON, colorIdentityJSON, imageURIsJSON, cardFacesJSON, legalitiesJSON string

		err := rows.Scan(
			&card.ID, &card.ArenaID, &card.Name, &card.ManaCost, &card.CMC, &card.TypeLine, &card.OracleText,
			&colorsJSON, &colorIdentityJSON, &card.Rarity, &card.SetCode, &card.CollectorNumber,
			&card.Power, &card.Toughness, &card.Loyalty, &imageURIsJSON, &card.Layout, &cardFacesJSON,
			&legalitiesJSON, &card.ReleasedAt, &card.CachedAt, &card.LastUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}

		// Deserialize JSON fields
		if err := json.Unmarshal([]byte(colorsJSON), &card.Colors); err != nil {
			return nil, fmt.Errorf("failed to unmarshal colors: %w", err)
		}

		if err := json.Unmarshal([]byte(colorIdentityJSON), &card.ColorIdentity); err != nil {
			return nil, fmt.Errorf("failed to unmarshal color_identity: %w", err)
		}

		if imageURIsJSON != "" && imageURIsJSON != "null" {
			if err := json.Unmarshal([]byte(imageURIsJSON), &card.ImageURIs); err != nil {
				return nil, fmt.Errorf("failed to unmarshal image_uris: %w", err)
			}
		}

		if cardFacesJSON != "" && cardFacesJSON != "null" && cardFacesJSON != "[]" {
			if err := json.Unmarshal([]byte(cardFacesJSON), &card.CardFaces); err != nil {
				return nil, fmt.Errorf("failed to unmarshal card_faces: %w", err)
			}
		}

		if err := json.Unmarshal([]byte(legalitiesJSON), &card.Legalities); err != nil {
			return nil, fmt.Errorf("failed to unmarshal legalities: %w", err)
		}

		cards = append(cards, &card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cards: %w", err)
	}

	return cards, nil
}

// GetStaleCards retrieves cards that haven't been updated in the specified duration.
func (s *Service) GetStaleCards(ctx context.Context, olderThan time.Duration) ([]*Card, error) {
	// Calculate seconds for SQLite datetime modifier
	seconds := int64(olderThan.Seconds())

	query := `
		SELECT id, arena_id, name, mana_cost, cmc, type_line, oracle_text,
			colors, color_identity, rarity, set_code, collector_number,
			power, toughness, loyalty, image_uris, layout, card_faces,
			legalities, released_at, cached_at, last_updated
		FROM cards
		WHERE unixepoch(last_updated) <= unixepoch('now', '-' || ? || ' seconds')
		ORDER BY last_updated ASC
	`

	rows, err := s.db.Conn().QueryContext(ctx, query, seconds)
	if err != nil {
		return nil, fmt.Errorf("failed to get stale cards: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var cards []*Card
	for rows.Next() {
		var card Card
		var colorsJSON, colorIdentityJSON, imageURIsJSON, cardFacesJSON, legalitiesJSON string

		err := rows.Scan(
			&card.ID, &card.ArenaID, &card.Name, &card.ManaCost, &card.CMC, &card.TypeLine, &card.OracleText,
			&colorsJSON, &colorIdentityJSON, &card.Rarity, &card.SetCode, &card.CollectorNumber,
			&card.Power, &card.Toughness, &card.Loyalty, &imageURIsJSON, &card.Layout, &cardFacesJSON,
			&legalitiesJSON, &card.ReleasedAt, &card.CachedAt, &card.LastUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}

		// Deserialize JSON fields
		if err := json.Unmarshal([]byte(colorsJSON), &card.Colors); err != nil {
			return nil, fmt.Errorf("failed to unmarshal colors: %w", err)
		}

		if err := json.Unmarshal([]byte(colorIdentityJSON), &card.ColorIdentity); err != nil {
			return nil, fmt.Errorf("failed to unmarshal color_identity: %w", err)
		}

		if imageURIsJSON != "" && imageURIsJSON != "null" {
			if err := json.Unmarshal([]byte(imageURIsJSON), &card.ImageURIs); err != nil {
				return nil, fmt.Errorf("failed to unmarshal image_uris: %w", err)
			}
		}

		if cardFacesJSON != "" && cardFacesJSON != "null" && cardFacesJSON != "[]" {
			if err := json.Unmarshal([]byte(cardFacesJSON), &card.CardFaces); err != nil {
				return nil, fmt.Errorf("failed to unmarshal card_faces: %w", err)
			}
		}

		if err := json.Unmarshal([]byte(legalitiesJSON), &card.Legalities); err != nil {
			return nil, fmt.Errorf("failed to unmarshal legalities: %w", err)
		}

		cards = append(cards, &card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cards: %w", err)
	}

	return cards, nil
}

// SaveSet saves or updates a set in the database.
func (s *Service) SaveSet(ctx context.Context, set *Set) error {
	query := `
		INSERT INTO sets (
			code, name, released_at, card_count, set_type, icon_svg_uri, last_updated
		) VALUES (
			?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP
		)
		ON CONFLICT(code) DO UPDATE SET
			name = excluded.name,
			released_at = excluded.released_at,
			card_count = excluded.card_count,
			set_type = excluded.set_type,
			icon_svg_uri = excluded.icon_svg_uri,
			last_updated = CURRENT_TIMESTAMP
	`

	_, err := s.db.Conn().ExecContext(ctx, query,
		set.Code, set.Name, set.ReleasedAt, set.CardCount, set.SetType, set.IconSVGURI,
	)
	if err != nil {
		return fmt.Errorf("failed to save set: %w", err)
	}

	return nil
}

// GetSet retrieves a set by its code.
func (s *Service) GetSet(ctx context.Context, code string) (*Set, error) {
	query := `
		SELECT code, name, released_at, card_count, set_type, icon_svg_uri, cached_at, last_updated
		FROM sets
		WHERE code = ?
	`

	var set Set
	err := s.db.Conn().QueryRowContext(ctx, query, code).Scan(
		&set.Code, &set.Name, &set.ReleasedAt, &set.CardCount, &set.SetType, &set.IconSVGURI,
		&set.CachedAt, &set.LastUpdated,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get set: %w", err)
	}

	return &set, nil
}

// GetAllSets retrieves all sets ordered by release date (newest first).
func (s *Service) GetAllSets(ctx context.Context) ([]*Set, error) {
	query := `
		SELECT code, name, released_at, card_count, set_type, icon_svg_uri, cached_at, last_updated
		FROM sets
		ORDER BY released_at DESC
	`

	rows, err := s.db.Conn().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all sets: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var sets []*Set
	for rows.Next() {
		var set Set
		err := rows.Scan(
			&set.Code, &set.Name, &set.ReleasedAt, &set.CardCount, &set.SetType, &set.IconSVGURI,
			&set.CachedAt, &set.LastUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan set: %w", err)
		}
		sets = append(sets, &set)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sets: %w", err)
	}

	return sets, nil
}

// GetStaleSets retrieves sets that haven't been updated in the specified duration.
func (s *Service) GetStaleSets(ctx context.Context, olderThan time.Duration) ([]*Set, error) {
	// Calculate seconds for SQLite datetime modifier
	seconds := int64(olderThan.Seconds())

	query := `
		SELECT code, name, released_at, card_count, set_type, icon_svg_uri, cached_at, last_updated
		FROM sets
		WHERE unixepoch(last_updated) <= unixepoch('now', '-' || ? || ' seconds')
		ORDER BY last_updated ASC
	`

	rows, err := s.db.Conn().QueryContext(ctx, query, seconds)
	if err != nil {
		return nil, fmt.Errorf("failed to get stale sets: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var sets []*Set
	for rows.Next() {
		var set Set
		err := rows.Scan(
			&set.Code, &set.Name, &set.ReleasedAt, &set.CardCount, &set.SetType, &set.IconSVGURI,
			&set.CachedAt, &set.LastUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan set: %w", err)
		}
		sets = append(sets, &set)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sets: %w", err)
	}

	return sets, nil
}
