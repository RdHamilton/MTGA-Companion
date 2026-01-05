package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// StandardRepository provides methods for Standard format data.
type StandardRepository interface {
	// Configuration
	GetConfig(ctx context.Context) (*models.StandardConfig, error)
	UpdateConfig(ctx context.Context, config *models.StandardConfig) error

	// Sets
	GetStandardSets(ctx context.Context) ([]*models.StandardSet, error)
	GetUpcomingRotation(ctx context.Context) (*models.UpcomingRotation, error)
	UpdateSetStandardStatus(ctx context.Context, setCode string, isLegal bool, rotationDate *string) error

	// Card legality
	GetCardLegality(ctx context.Context, arenaID string) (*models.CardLegality, error)
	GetCardsLegality(ctx context.Context, arenaIDs []string) (map[string]*models.CardLegality, error)
	UpdateCardLegality(ctx context.Context, arenaID string, legality *models.CardLegality) error

	// Deck validation
	GetRotationAffectedDecks(ctx context.Context) ([]*models.RotationAffectedDeck, error)
}

type standardRepository struct {
	db *sql.DB
}

// NewStandardRepository creates a new Standard repository.
func NewStandardRepository(db *sql.DB) StandardRepository {
	return &standardRepository{db: db}
}

// GetConfig retrieves the Standard configuration.
func (r *standardRepository) GetConfig(ctx context.Context) (*models.StandardConfig, error) {
	query := `
		SELECT id, next_rotation_date, rotation_enabled, updated_at
		FROM standard_config
		WHERE id = 1
	`

	var config models.StandardConfig
	var nextRotationStr string
	var updatedAtStr string

	err := r.db.QueryRowContext(ctx, query).Scan(
		&config.ID, &nextRotationStr, &config.RotationEnabled, &updatedAtStr,
	)
	if err == sql.ErrNoRows {
		// Return default config if not found
		return &models.StandardConfig{
			ID:               1,
			NextRotationDate: time.Date(2027, 1, 23, 0, 0, 0, 0, time.UTC),
			RotationEnabled:  true,
			UpdatedAt:        time.Now(),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get standard config: %w", err)
	}

	// Parse dates
	config.NextRotationDate, _ = time.Parse("2006-01-02", nextRotationStr)
	config.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

	return &config, nil
}

// UpdateConfig updates the Standard configuration.
func (r *standardRepository) UpdateConfig(ctx context.Context, config *models.StandardConfig) error {
	query := `
		INSERT INTO standard_config (id, next_rotation_date, rotation_enabled, updated_at)
		VALUES (1, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			next_rotation_date = excluded.next_rotation_date,
			rotation_enabled = excluded.rotation_enabled,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(ctx, query,
		config.NextRotationDate.Format("2006-01-02"),
		config.RotationEnabled,
	)
	if err != nil {
		return fmt.Errorf("failed to update standard config: %w", err)
	}

	return nil
}

// GetStandardSets retrieves all Standard-legal sets with rotation info.
func (r *standardRepository) GetStandardSets(ctx context.Context) ([]*models.StandardSet, error) {
	config, err := r.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT code, name, released_at, rotation_date, is_standard_legal, icon_svg_uri, card_count
		FROM sets
		WHERE is_standard_legal = TRUE
		ORDER BY released_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get standard sets: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var sets []*models.StandardSet
	now := time.Now()

	for rows.Next() {
		var set models.StandardSet
		err := rows.Scan(
			&set.Code, &set.Name, &set.ReleasedAt, &set.RotationDate,
			&set.IsStandardLegal, &set.IconSVGURI, &set.CardCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan standard set: %w", err)
		}

		// Calculate days until rotation
		if set.RotationDate != nil && config.RotationEnabled {
			rotationDate, err := time.Parse("2006-01-02", *set.RotationDate)
			if err == nil {
				days := int(math.Ceil(rotationDate.Sub(now).Hours() / 24))
				if days > 0 {
					set.DaysUntilRotation = &days
					set.IsRotatingSoon = days <= 90
				}
			}
		}

		sets = append(sets, &set)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating standard sets: %w", err)
	}

	return sets, nil
}

// GetUpcomingRotation returns information about the next Standard rotation.
func (r *standardRepository) GetUpcomingRotation(ctx context.Context) (*models.UpcomingRotation, error) {
	config, err := r.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	if !config.RotationEnabled {
		return &models.UpcomingRotation{
			NextRotationDate:  config.NextRotationDate.Format("2006-01-02"),
			DaysUntilRotation: -1, // Indicates rotation is disabled
			RotatingSets:      []models.StandardSet{},
			RotatingCardCount: 0,
			AffectedDecks:     0,
		}, nil
	}

	// Get sets that will rotate at next rotation
	query := `
		SELECT code, name, released_at, rotation_date, is_standard_legal, icon_svg_uri, card_count
		FROM sets
		WHERE is_standard_legal = TRUE
		  AND rotation_date IS NOT NULL
		  AND date(rotation_date) <= date(?)
		ORDER BY released_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, config.NextRotationDate.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to get rotating sets: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var rotatingSets []models.StandardSet
	totalRotatingCards := 0
	now := time.Now()

	for rows.Next() {
		var set models.StandardSet
		err := rows.Scan(
			&set.Code, &set.Name, &set.ReleasedAt, &set.RotationDate,
			&set.IsStandardLegal, &set.IconSVGURI, &set.CardCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rotating set: %w", err)
		}

		if set.RotationDate != nil {
			rotationDate, _ := time.Parse("2006-01-02", *set.RotationDate)
			days := int(math.Ceil(rotationDate.Sub(now).Hours() / 24))
			if days > 0 {
				set.DaysUntilRotation = &days
				set.IsRotatingSoon = days <= 90
			}
		}

		totalRotatingCards += set.CardCount
		rotatingSets = append(rotatingSets, set)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rotating sets: %w", err)
	}

	// Count affected decks
	affectedDecks, err := r.countAffectedDecks(ctx, rotatingSets)
	if err != nil {
		// Non-fatal error, just log and continue
		affectedDecks = 0
	}

	daysUntilRotation := int(math.Ceil(config.NextRotationDate.Sub(now).Hours() / 24))

	return &models.UpcomingRotation{
		NextRotationDate:  config.NextRotationDate.Format("2006-01-02"),
		DaysUntilRotation: daysUntilRotation,
		RotatingSets:      rotatingSets,
		RotatingCardCount: totalRotatingCards,
		AffectedDecks:     affectedDecks,
	}, nil
}

// countAffectedDecks counts how many Standard decks have cards from rotating sets.
func (r *standardRepository) countAffectedDecks(ctx context.Context, rotatingSets []models.StandardSet) (int, error) {
	if len(rotatingSets) == 0 {
		return 0, nil
	}

	// Build set code list for query
	setCodes := make([]interface{}, len(rotatingSets))
	placeholders := ""
	for i, set := range rotatingSets {
		setCodes[i] = set.Code
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?"
	}

	query := fmt.Sprintf(`
		SELECT COUNT(DISTINCT d.id)
		FROM decks d
		JOIN deck_cards dc ON dc.deck_id = d.id
		JOIN set_cards sc ON sc.arena_id = CAST(dc.card_id AS TEXT)
		WHERE d.format = 'Standard'
		  AND sc.set_code IN (%s)
	`, placeholders)

	var count int
	err := r.db.QueryRowContext(ctx, query, setCodes...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count affected decks: %w", err)
	}

	return count, nil
}

// UpdateSetStandardStatus updates a set's Standard legality status.
func (r *standardRepository) UpdateSetStandardStatus(ctx context.Context, setCode string, isLegal bool, rotationDate *string) error {
	query := `
		UPDATE sets
		SET is_standard_legal = ?, rotation_date = ?
		WHERE code = ?
	`

	_, err := r.db.ExecContext(ctx, query, isLegal, rotationDate, setCode)
	if err != nil {
		return fmt.Errorf("failed to update set standard status: %w", err)
	}

	return nil
}

// GetCardLegality retrieves the legality of a card by Arena ID.
func (r *standardRepository) GetCardLegality(ctx context.Context, arenaID string) (*models.CardLegality, error) {
	query := `
		SELECT legalities
		FROM set_cards
		WHERE arena_id = ?
	`

	var legalitiesJSON sql.NullString
	err := r.db.QueryRowContext(ctx, query, arenaID).Scan(&legalitiesJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get card legality: %w", err)
	}

	if !legalitiesJSON.Valid || legalitiesJSON.String == "" {
		return nil, nil
	}

	var legality models.CardLegality
	if err := json.Unmarshal([]byte(legalitiesJSON.String), &legality); err != nil {
		return nil, fmt.Errorf("failed to unmarshal legalities: %w", err)
	}

	return &legality, nil
}

// GetCardsLegality retrieves legalities for multiple cards.
func (r *standardRepository) GetCardsLegality(ctx context.Context, arenaIDs []string) (map[string]*models.CardLegality, error) {
	if len(arenaIDs) == 0 {
		return make(map[string]*models.CardLegality), nil
	}

	// Build placeholders for IN clause
	placeholders := ""
	args := make([]interface{}, len(arenaIDs))
	for i, id := range arenaIDs {
		args[i] = id
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?"
	}

	query := fmt.Sprintf(`
		SELECT arena_id, legalities
		FROM set_cards
		WHERE arena_id IN (%s) AND legalities IS NOT NULL
	`, placeholders)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get cards legality: %w", err)
	}
	defer func() { _ = rows.Close() }()

	result := make(map[string]*models.CardLegality)
	for rows.Next() {
		var arenaID string
		var legalitiesJSON string

		if err := rows.Scan(&arenaID, &legalitiesJSON); err != nil {
			return nil, fmt.Errorf("failed to scan card legality: %w", err)
		}

		var legality models.CardLegality
		if err := json.Unmarshal([]byte(legalitiesJSON), &legality); err != nil {
			continue // Skip invalid JSON
		}

		result[arenaID] = &legality
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating card legalities: %w", err)
	}

	return result, nil
}

// UpdateCardLegality updates the legality of a card.
func (r *standardRepository) UpdateCardLegality(ctx context.Context, arenaID string, legality *models.CardLegality) error {
	legalitiesJSON, err := json.Marshal(legality)
	if err != nil {
		return fmt.Errorf("failed to marshal legalities: %w", err)
	}

	query := `
		UPDATE set_cards
		SET legalities = ?
		WHERE arena_id = ?
	`

	_, err = r.db.ExecContext(ctx, query, string(legalitiesJSON), arenaID)
	if err != nil {
		return fmt.Errorf("failed to update card legality: %w", err)
	}

	return nil
}

// GetRotationAffectedDecks returns all Standard decks affected by the upcoming rotation.
func (r *standardRepository) GetRotationAffectedDecks(ctx context.Context) ([]*models.RotationAffectedDeck, error) {
	config, err := r.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	if !config.RotationEnabled {
		return []*models.RotationAffectedDeck{}, nil
	}

	// Get Standard decks with rotating cards
	query := `
		WITH rotating_sets AS (
			SELECT code, name, rotation_date
			FROM sets
			WHERE is_standard_legal = TRUE
			  AND rotation_date IS NOT NULL
			  AND date(rotation_date) <= date(?)
		),
		deck_rotating_cards AS (
			SELECT
				d.id as deck_id,
				d.name as deck_name,
				d.format,
				sc.arena_id,
				sc.name as card_name,
				sc.set_code,
				rs.name as set_name,
				rs.rotation_date,
				dc.quantity
			FROM decks d
			JOIN deck_cards dc ON dc.deck_id = d.id
			JOIN set_cards sc ON sc.arena_id = CAST(dc.card_id AS TEXT)
			JOIN rotating_sets rs ON rs.code = sc.set_code
			WHERE d.format = 'Standard'
		)
		SELECT
			deck_id,
			deck_name,
			format,
			SUM(quantity) as rotating_card_count,
			(SELECT SUM(quantity) FROM deck_cards WHERE deck_id = drc.deck_id) as total_cards,
			GROUP_CONCAT(arena_id || '|' || card_name || '|' || set_code || '|' || set_name || '|' || rotation_date || '|' || quantity, ';;') as rotating_cards_data
		FROM deck_rotating_cards drc
		GROUP BY deck_id, deck_name, format
		ORDER BY rotating_card_count DESC
	`

	rows, err := r.db.QueryContext(ctx, query, config.NextRotationDate.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to get rotation affected decks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	now := time.Now()
	var affectedDecks []*models.RotationAffectedDeck

	for rows.Next() {
		var deck models.RotationAffectedDeck
		var rotatingCardsData sql.NullString

		err := rows.Scan(
			&deck.DeckID, &deck.DeckName, &deck.Format,
			&deck.RotatingCardCount, &deck.TotalCards, &rotatingCardsData,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan affected deck: %w", err)
		}

		if deck.TotalCards > 0 {
			deck.PercentAffected = float64(deck.RotatingCardCount) / float64(deck.TotalCards) * 100
		}

		// Parse rotating cards data
		if rotatingCardsData.Valid && rotatingCardsData.String != "" {
			deck.RotatingCards = parseRotatingCardsData(rotatingCardsData.String, now)
		}

		affectedDecks = append(affectedDecks, &deck)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating affected decks: %w", err)
	}

	return affectedDecks, nil
}

// parseRotatingCardsData parses the concatenated rotating cards data from SQL.
func parseRotatingCardsData(data string, now time.Time) []models.RotatingCard {
	var cards []models.RotatingCard

	// Split by ;; to get individual cards
	cardEntries := splitString(data, ";;")
	for _, entry := range cardEntries {
		parts := splitString(entry, "|")
		if len(parts) < 5 {
			continue
		}

		card := models.RotatingCard{
			CardName:     parts[1],
			SetCode:      parts[2],
			SetName:      parts[3],
			RotationDate: parts[4],
		}

		// Parse arena_id as CardID
		if arenaID := parts[0]; arenaID != "" {
			// Convert arena_id string to int if possible
			var cardID int
			if _, err := fmt.Sscanf(arenaID, "%d", &cardID); err == nil {
				card.CardID = cardID
			}
		}

		// Calculate days until rotation
		if rotationDate, err := time.Parse("2006-01-02", card.RotationDate); err == nil {
			card.DaysUntilRotation = int(math.Ceil(rotationDate.Sub(now).Hours() / 24))
		}

		cards = append(cards, card)
	}

	return cards
}

// splitString splits a string by delimiter without using strings package.
func splitString(s, sep string) []string {
	var result []string
	sepLen := len(sep)

	for {
		idx := -1
		for i := 0; i <= len(s)-sepLen; i++ {
			if s[i:i+sepLen] == sep {
				idx = i
				break
			}
		}

		if idx == -1 {
			result = append(result, s)
			break
		}

		result = append(result, s[:idx])
		s = s[idx+sepLen:]
	}

	return result
}
