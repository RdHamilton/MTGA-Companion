package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// EDHRECRepository handles EDHREC synergy data operations.
type EDHRECRepository interface {
	// Synergy operations
	UpsertSynergy(ctx context.Context, synergy *models.EDHRECSynergy) error
	BulkUpsertSynergies(ctx context.Context, synergies []*models.EDHRECSynergy) error
	GetSynergiesForCard(ctx context.Context, cardName string, limit int) ([]*models.EDHRECSynergy, error)
	GetSynergyScore(ctx context.Context, cardName, synergyCardName string) (float64, error)
	GetTopSynergies(ctx context.Context, cardName string, limit int) ([]*models.EDHRECSynergy, error)
	DeleteSynergiesForCard(ctx context.Context, cardName string) error

	// Metadata operations
	UpsertMetadata(ctx context.Context, metadata *models.EDHRECCardMetadata) error
	GetMetadata(ctx context.Context, cardName string) (*models.EDHRECCardMetadata, error)
	GetMetadataCount(ctx context.Context) (int, error)

	// Theme operations
	UpsertThemeCard(ctx context.Context, themeCard *models.EDHRECThemeCard) error
	BulkUpsertThemeCards(ctx context.Context, themeCards []*models.EDHRECThemeCard) error
	GetCardsForTheme(ctx context.Context, themeName string, limit int) ([]*models.EDHRECThemeCard, error)
	GetThemesForCard(ctx context.Context, cardName string) ([]*models.EDHRECThemeCard, error)
	DeleteThemeCards(ctx context.Context, themeName string) error

	// Utility
	GetSynergyCount(ctx context.Context) (int, error)
	ClearAll(ctx context.Context) error
}

// edhrecRepo implements EDHRECRepository.
type edhrecRepo struct {
	db *sql.DB
}

// NewEDHRECRepository creates a new EDHREC repository.
func NewEDHRECRepository(db *sql.DB) EDHRECRepository {
	return &edhrecRepo{db: db}
}

// UpsertSynergy inserts or updates a synergy relationship.
func (r *edhrecRepo) UpsertSynergy(ctx context.Context, synergy *models.EDHRECSynergy) error {
	query := `
		INSERT INTO edhrec_synergy (card_name, synergy_card_name, synergy_score, inclusion_count, num_decks, lift, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(card_name, synergy_card_name) DO UPDATE SET
			synergy_score = excluded.synergy_score,
			inclusion_count = excluded.inclusion_count,
			num_decks = excluded.num_decks,
			lift = excluded.lift,
			last_updated = CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(ctx, query,
		synergy.CardName, synergy.SynergyCardName, synergy.SynergyScore,
		synergy.InclusionCount, synergy.NumDecks, synergy.Lift,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert synergy: %w", err)
	}

	return nil
}

// BulkUpsertSynergies inserts or updates multiple synergy relationships.
func (r *edhrecRepo) BulkUpsertSynergies(ctx context.Context, synergies []*models.EDHRECSynergy) error {
	if len(synergies) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO edhrec_synergy (card_name, synergy_card_name, synergy_score, inclusion_count, num_decks, lift, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(card_name, synergy_card_name) DO UPDATE SET
			synergy_score = excluded.synergy_score,
			inclusion_count = excluded.inclusion_count,
			num_decks = excluded.num_decks,
			lift = excluded.lift,
			last_updated = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, synergy := range synergies {
		_, err := stmt.ExecContext(ctx,
			synergy.CardName, synergy.SynergyCardName, synergy.SynergyScore,
			synergy.InclusionCount, synergy.NumDecks, synergy.Lift,
		)
		if err != nil {
			return fmt.Errorf("failed to insert synergy for %s -> %s: %w",
				synergy.CardName, synergy.SynergyCardName, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// GetSynergiesForCard gets all synergies for a card.
func (r *edhrecRepo) GetSynergiesForCard(ctx context.Context, cardName string, limit int) ([]*models.EDHRECSynergy, error) {
	query := `
		SELECT id, card_name, synergy_card_name, synergy_score, inclusion_count, num_decks, lift, last_updated
		FROM edhrec_synergy
		WHERE LOWER(card_name) = LOWER(?)
		ORDER BY synergy_score DESC
	`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.QueryContext(ctx, query, cardName)
	if err != nil {
		return nil, fmt.Errorf("failed to get synergies: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var synergies []*models.EDHRECSynergy
	for rows.Next() {
		var s models.EDHRECSynergy
		if err := rows.Scan(
			&s.ID, &s.CardName, &s.SynergyCardName, &s.SynergyScore,
			&s.InclusionCount, &s.NumDecks, &s.Lift, &s.LastUpdated,
		); err != nil {
			return nil, fmt.Errorf("failed to scan synergy: %w", err)
		}
		synergies = append(synergies, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating synergies: %w", err)
	}

	return synergies, nil
}

// GetSynergyScore gets the synergy score between two cards.
func (r *edhrecRepo) GetSynergyScore(ctx context.Context, cardName, synergyCardName string) (float64, error) {
	query := `
		SELECT synergy_score
		FROM edhrec_synergy
		WHERE (LOWER(card_name) = LOWER(?) AND LOWER(synergy_card_name) = LOWER(?))
		   OR (LOWER(card_name) = LOWER(?) AND LOWER(synergy_card_name) = LOWER(?))
	`

	var score float64
	err := r.db.QueryRowContext(ctx, query, cardName, synergyCardName, synergyCardName, cardName).Scan(&score)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get synergy score: %w", err)
	}

	return score, nil
}

// GetTopSynergies gets the top synergies for a card.
func (r *edhrecRepo) GetTopSynergies(ctx context.Context, cardName string, limit int) ([]*models.EDHRECSynergy, error) {
	return r.GetSynergiesForCard(ctx, cardName, limit)
}

// DeleteSynergiesForCard deletes all synergies for a card.
func (r *edhrecRepo) DeleteSynergiesForCard(ctx context.Context, cardName string) error {
	query := `DELETE FROM edhrec_synergy WHERE LOWER(card_name) = LOWER(?)`

	_, err := r.db.ExecContext(ctx, query, cardName)
	if err != nil {
		return fmt.Errorf("failed to delete synergies: %w", err)
	}

	return nil
}

// UpsertMetadata inserts or updates card metadata.
func (r *edhrecRepo) UpsertMetadata(ctx context.Context, metadata *models.EDHRECCardMetadata) error {
	query := `
		INSERT INTO edhrec_card_metadata (card_name, sanitized_name, num_decks, salt_score, color_identity, last_updated)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(card_name) DO UPDATE SET
			sanitized_name = excluded.sanitized_name,
			num_decks = excluded.num_decks,
			salt_score = excluded.salt_score,
			color_identity = excluded.color_identity,
			last_updated = CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(ctx, query,
		metadata.CardName, metadata.SanitizedName, metadata.NumDecks,
		metadata.SaltScore, metadata.ColorIdentity,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert metadata: %w", err)
	}

	return nil
}

// GetMetadata gets metadata for a card.
func (r *edhrecRepo) GetMetadata(ctx context.Context, cardName string) (*models.EDHRECCardMetadata, error) {
	query := `
		SELECT id, card_name, sanitized_name, num_decks, salt_score, color_identity, last_updated
		FROM edhrec_card_metadata
		WHERE LOWER(card_name) = LOWER(?)
	`

	var m models.EDHRECCardMetadata
	err := r.db.QueryRowContext(ctx, query, cardName).Scan(
		&m.ID, &m.CardName, &m.SanitizedName, &m.NumDecks,
		&m.SaltScore, &m.ColorIdentity, &m.LastUpdated,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	return &m, nil
}

// GetMetadataCount gets the count of metadata records.
func (r *edhrecRepo) GetMetadataCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM edhrec_card_metadata").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get metadata count: %w", err)
	}
	return count, nil
}

// UpsertThemeCard inserts or updates a theme card association.
func (r *edhrecRepo) UpsertThemeCard(ctx context.Context, themeCard *models.EDHRECThemeCard) error {
	query := `
		INSERT INTO edhrec_theme_cards (theme_name, card_name, synergy_score, is_top_card, is_high_synergy, last_updated)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(theme_name, card_name) DO UPDATE SET
			synergy_score = excluded.synergy_score,
			is_top_card = excluded.is_top_card,
			is_high_synergy = excluded.is_high_synergy,
			last_updated = CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(ctx, query,
		themeCard.ThemeName, themeCard.CardName, themeCard.SynergyScore,
		themeCard.IsTopCard, themeCard.IsHighSynergy,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert theme card: %w", err)
	}

	return nil
}

// BulkUpsertThemeCards inserts or updates multiple theme card associations.
func (r *edhrecRepo) BulkUpsertThemeCards(ctx context.Context, themeCards []*models.EDHRECThemeCard) error {
	if len(themeCards) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO edhrec_theme_cards (theme_name, card_name, synergy_score, is_top_card, is_high_synergy, last_updated)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(theme_name, card_name) DO UPDATE SET
			synergy_score = excluded.synergy_score,
			is_top_card = excluded.is_top_card,
			is_high_synergy = excluded.is_high_synergy,
			last_updated = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, tc := range themeCards {
		_, err := stmt.ExecContext(ctx,
			tc.ThemeName, tc.CardName, tc.SynergyScore,
			tc.IsTopCard, tc.IsHighSynergy,
		)
		if err != nil {
			return fmt.Errorf("failed to insert theme card: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// GetCardsForTheme gets all cards for a theme.
func (r *edhrecRepo) GetCardsForTheme(ctx context.Context, themeName string, limit int) ([]*models.EDHRECThemeCard, error) {
	query := `
		SELECT id, theme_name, card_name, synergy_score, is_top_card, is_high_synergy, last_updated
		FROM edhrec_theme_cards
		WHERE LOWER(theme_name) = LOWER(?)
		ORDER BY synergy_score DESC
	`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.QueryContext(ctx, query, themeName)
	if err != nil {
		return nil, fmt.Errorf("failed to get theme cards: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var cards []*models.EDHRECThemeCard
	for rows.Next() {
		var tc models.EDHRECThemeCard
		if err := rows.Scan(
			&tc.ID, &tc.ThemeName, &tc.CardName, &tc.SynergyScore,
			&tc.IsTopCard, &tc.IsHighSynergy, &tc.LastUpdated,
		); err != nil {
			return nil, fmt.Errorf("failed to scan theme card: %w", err)
		}
		cards = append(cards, &tc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating theme cards: %w", err)
	}

	return cards, nil
}

// GetThemesForCard gets all themes a card belongs to.
func (r *edhrecRepo) GetThemesForCard(ctx context.Context, cardName string) ([]*models.EDHRECThemeCard, error) {
	query := `
		SELECT id, theme_name, card_name, synergy_score, is_top_card, is_high_synergy, last_updated
		FROM edhrec_theme_cards
		WHERE LOWER(card_name) = LOWER(?)
		ORDER BY synergy_score DESC
	`

	rows, err := r.db.QueryContext(ctx, query, cardName)
	if err != nil {
		return nil, fmt.Errorf("failed to get themes for card: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var themes []*models.EDHRECThemeCard
	for rows.Next() {
		var tc models.EDHRECThemeCard
		if err := rows.Scan(
			&tc.ID, &tc.ThemeName, &tc.CardName, &tc.SynergyScore,
			&tc.IsTopCard, &tc.IsHighSynergy, &tc.LastUpdated,
		); err != nil {
			return nil, fmt.Errorf("failed to scan theme card: %w", err)
		}
		themes = append(themes, &tc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating themes: %w", err)
	}

	return themes, nil
}

// DeleteThemeCards deletes all cards for a theme.
func (r *edhrecRepo) DeleteThemeCards(ctx context.Context, themeName string) error {
	query := `DELETE FROM edhrec_theme_cards WHERE LOWER(theme_name) = LOWER(?)`

	_, err := r.db.ExecContext(ctx, query, themeName)
	if err != nil {
		return fmt.Errorf("failed to delete theme cards: %w", err)
	}

	return nil
}

// GetSynergyCount gets the count of synergy records.
func (r *edhrecRepo) GetSynergyCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM edhrec_synergy").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get synergy count: %w", err)
	}
	return count, nil
}

// ClearAll removes all EDHREC data.
func (r *edhrecRepo) ClearAll(ctx context.Context) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	tables := []string{"edhrec_synergy", "edhrec_card_metadata", "edhrec_theme_cards"}
	for _, table := range tables {
		_, err := tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			return fmt.Errorf("failed to clear %s: %w", table, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// GetMatchingThemes returns themes that match keywords in a card's text or type.
func GetMatchingThemes(cardText, cardType string) []string {
	cardTextLower := strings.ToLower(cardText)
	cardTypeLower := strings.ToLower(cardType)

	var themes []string

	themeKeywords := map[string][]string{
		"tokens":       {"create", "token", "populate"},
		"aristocrats":  {"sacrifice", "dies", "whenever a creature dies"},
		"counters":     {"+1/+1 counter", "-1/-1 counter", "counter on"},
		"blink":        {"exile", "return", "flicker", "enters"},
		"reanimator":   {"graveyard", "return", "reanimate"},
		"spellslinger": {"instant", "sorcery", "cast", "spell"},
		"artifacts":    {"artifact", "equipment", "treasure"},
		"enchantments": {"enchantment", "enchant", "aura"},
		"lifegain":     {"gain", "life", "lifelink"},
		"mill":         {"mill", "library", "graveyard"},
		"voltron":      {"equipment", "aura", "attach", "equipped"},
		"landfall":     {"land", "enters the battlefield", "landfall"},
		"graveyard":    {"graveyard", "dies", "exile from graveyard"},
		"sacrifice":    {"sacrifice", "dies", "when this creature dies"},
		"clones":       {"copy", "clone", "becomes a copy"},
		"equipment":    {"equipment", "equip", "equipped"},
		"wheels":       {"discard", "draw", "wheel"},
		"extra-turns":  {"extra turn", "additional turn"},
		"big-mana":     {"add", "mana", "untap", "double"},
	}

	for theme, keywords := range themeKeywords {
		for _, keyword := range keywords {
			if strings.Contains(cardTextLower, keyword) || strings.Contains(cardTypeLower, keyword) {
				themes = append(themes, theme)
				break
			}
		}
	}

	return themes
}
