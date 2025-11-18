package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

// DraftRatingsRepository provides methods for managing cached 17Lands ratings.
type DraftRatingsRepository interface {
	// SaveSetRatings saves card and color ratings for a set.
	SaveSetRatings(ctx context.Context, setCode, draftFormat string, cardRatings []seventeenlands.CardRating, colorRatings []seventeenlands.ColorRating) error

	// GetCardRatings retrieves cached card ratings for a set.
	GetCardRatings(ctx context.Context, setCode, draftFormat string) ([]seventeenlands.CardRating, time.Time, error)

	// GetColorRatings retrieves cached color ratings for a set.
	GetColorRatings(ctx context.Context, setCode, draftFormat string) ([]seventeenlands.ColorRating, time.Time, error)

	// GetCardRatingByArenaID retrieves a specific card's rating by Arena ID.
	GetCardRatingByArenaID(ctx context.Context, setCode, draftFormat, arenaID string) (*seventeenlands.CardRating, error)

	// IsSetRatingsCached checks if ratings for a set are cached.
	IsSetRatingsCached(ctx context.Context, setCode, draftFormat string) (bool, error)

	// DeleteSetRatings removes all ratings for a set (for cache invalidation).
	DeleteSetRatings(ctx context.Context, setCode, draftFormat string) error
}

type draftRatingsRepository struct {
	db *sql.DB
}

// NewDraftRatingsRepository creates a new draft ratings repository.
func NewDraftRatingsRepository(db *sql.DB) DraftRatingsRepository {
	return &draftRatingsRepository{db: db}
}

// SaveSetRatings saves card and color ratings for a set.
func (r *draftRatingsRepository) SaveSetRatings(ctx context.Context, setCode, draftFormat string, cardRatings []seventeenlands.CardRating, colorRatings []seventeenlands.ColorRating) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Save card ratings
	cardStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO draft_card_ratings (
			set_code, draft_format, arena_id, name, color, rarity,
			gihwr, ohwr, alsa, ata, gih_count, cached_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(set_code, draft_format, arena_id) DO UPDATE SET
			name = excluded.name,
			color = excluded.color,
			rarity = excluded.rarity,
			gihwr = excluded.gihwr,
			ohwr = excluded.ohwr,
			alsa = excluded.alsa,
			ata = excluded.ata,
			gih_count = excluded.gih_count,
			cached_at = excluded.cached_at
	`)
	if err != nil {
		return err
	}
	defer func() {
		_ = cardStmt.Close()
	}()

	cachedAt := time.Now()
	for _, card := range cardRatings {
		_, err = cardStmt.ExecContext(ctx,
			setCode,
			draftFormat,
			card.MTGAID,
			card.Name,
			card.Color,
			card.Rarity,
			card.GIHWR,
			card.OHWR,
			card.ALSA,
			card.ATA,
			card.GIH,
			cachedAt,
		)
		if err != nil {
			return err
		}
	}

	// Save color ratings
	colorStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO draft_color_ratings (
			set_code, draft_format, color_combination, win_rate, games_played, cached_at
		) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(set_code, draft_format, color_combination) DO UPDATE SET
			win_rate = excluded.win_rate,
			games_played = excluded.games_played,
			cached_at = excluded.cached_at
	`)
	if err != nil {
		return err
	}
	defer func() {
		_ = colorStmt.Close()
	}()

	for _, color := range colorRatings {
		_, err = colorStmt.ExecContext(ctx,
			setCode,
			draftFormat,
			color.ColorName,
			color.WinRate,
			color.GamesPlayed,
			cachedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetCardRatings retrieves cached card ratings for a set.
func (r *draftRatingsRepository) GetCardRatings(ctx context.Context, setCode, draftFormat string) ([]seventeenlands.CardRating, time.Time, error) {
	query := `
		SELECT arena_id, name, color, rarity, gihwr, ohwr, alsa, ata, gih_count, cached_at
		FROM draft_card_ratings
		WHERE set_code = ? AND draft_format = ?
		ORDER BY gihwr DESC
	`
	rows, err := r.db.QueryContext(ctx, query, setCode, draftFormat)
	if err != nil {
		return nil, time.Time{}, err
	}
	defer func() {
		_ = rows.Close()
	}()

	ratings := []seventeenlands.CardRating{}
	var cachedAt time.Time

	for rows.Next() {
		var rating seventeenlands.CardRating
		err := rows.Scan(
			&rating.MTGAID,
			&rating.Name,
			&rating.Color,
			&rating.Rarity,
			&rating.GIHWR,
			&rating.OHWR,
			&rating.ALSA,
			&rating.ATA,
			&rating.GIH,
			&cachedAt,
		)
		if err != nil {
			return nil, time.Time{}, err
		}
		ratings = append(ratings, rating)
	}

	return ratings, cachedAt, rows.Err()
}

// GetColorRatings retrieves cached color ratings for a set.
func (r *draftRatingsRepository) GetColorRatings(ctx context.Context, setCode, draftFormat string) ([]seventeenlands.ColorRating, time.Time, error) {
	query := `
		SELECT color_combination, win_rate, games_played, cached_at
		FROM draft_color_ratings
		WHERE set_code = ? AND draft_format = ?
		ORDER BY win_rate DESC
	`
	rows, err := r.db.QueryContext(ctx, query, setCode, draftFormat)
	if err != nil {
		return nil, time.Time{}, err
	}
	defer func() {
		_ = rows.Close()
	}()

	ratings := []seventeenlands.ColorRating{}
	var cachedAt time.Time

	for rows.Next() {
		var rating seventeenlands.ColorRating
		err := rows.Scan(
			&rating.ColorName,
			&rating.WinRate,
			&rating.GamesPlayed,
			&cachedAt,
		)
		if err != nil {
			return nil, time.Time{}, err
		}

		// Parse color combination string into individual colors
		rating.Colors = parseColorCombination(rating.ColorName)
		ratings = append(ratings, rating)
	}

	return ratings, cachedAt, rows.Err()
}

// GetCardRatingByArenaID retrieves a specific card's rating by Arena ID.
func (r *draftRatingsRepository) GetCardRatingByArenaID(ctx context.Context, setCode, draftFormat, arenaID string) (*seventeenlands.CardRating, error) {
	query := `
		SELECT arena_id, name, color, rarity, gihwr, ohwr, alsa, ata, gih_count
		FROM draft_card_ratings
		WHERE set_code = ? AND draft_format = ? AND arena_id = ?
	`
	row := r.db.QueryRowContext(ctx, query, setCode, draftFormat, arenaID)

	var rating seventeenlands.CardRating
	err := row.Scan(
		&rating.MTGAID,
		&rating.Name,
		&rating.Color,
		&rating.Rarity,
		&rating.GIHWR,
		&rating.OHWR,
		&rating.ALSA,
		&rating.ATA,
		&rating.GIH,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &rating, nil
}

// IsSetRatingsCached checks if ratings for a set are cached.
func (r *draftRatingsRepository) IsSetRatingsCached(ctx context.Context, setCode, draftFormat string) (bool, error) {
	query := `SELECT COUNT(*) FROM draft_card_ratings WHERE set_code = ? AND draft_format = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, setCode, draftFormat).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// DeleteSetRatings removes all ratings for a set (for cache invalidation).
func (r *draftRatingsRepository) DeleteSetRatings(ctx context.Context, setCode, draftFormat string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Delete card ratings
	_, err = tx.ExecContext(ctx, `DELETE FROM draft_card_ratings WHERE set_code = ? AND draft_format = ?`, setCode, draftFormat)
	if err != nil {
		return err
	}

	// Delete color ratings
	_, err = tx.ExecContext(ctx, `DELETE FROM draft_color_ratings WHERE set_code = ? AND draft_format = ?`, setCode, draftFormat)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// parseColorCombination parses a color combination string into individual colors.
// Examples: "W" -> ["W"], "UB" -> ["U", "B"], "WUG" -> ["W", "U", "G"]
func parseColorCombination(combo string) []string {
	colors := []string{}
	for _, char := range combo {
		color := string(char)
		if color == "W" || color == "U" || color == "B" || color == "R" || color == "G" {
			colors = append(colors, color)
		}
	}
	return colors
}
