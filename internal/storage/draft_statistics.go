package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

// DraftCardRating represents a cached card rating from 17Lands.
type DraftCardRating struct {
	ID        int
	ArenaID   int
	Expansion string
	Format    string
	Colors    string // Comma-separated color filter (e.g., "W,U" or empty for all)

	// Win rate metrics
	GIHWR      float64
	OHWR       float64
	GPWR       float64
	GDWR       float64
	IHDWR      float64
	GIHWRDelta float64
	OHWRDelta  float64
	GDWRDelta  float64
	IHDWRDelta float64

	// Draft metrics
	ALSA float64
	ATA  float64

	// Sample sizes
	GIH int
	OH  int
	GP  int
	GD  int
	IHD int

	// Deck metrics
	GamesPlayed int
	NumDecks    int

	// Metadata
	StartDate   string
	EndDate     string
	CachedAt    time.Time
	LastUpdated time.Time
}

// DraftColorRating represents a cached color combination rating from 17Lands.
type DraftColorRating struct {
	ID               int
	Expansion        string
	EventType        string
	ColorCombination string

	// Metrics
	WinRate     float64
	GamesPlayed int
	NumDecks    int

	// Metadata
	StartDate   string
	EndDate     string
	CachedAt    time.Time
	LastUpdated time.Time
}

// SaveCardRatings saves a batch of card ratings to the database.
// Uses INSERT ... ON CONFLICT to update existing entries.
func (s *Service) SaveCardRatings(ctx context.Context, ratings []seventeenlands.CardRating, expansion, format, colors, startDate, endDate string) error {
	if len(ratings) == 0 {
		return nil
	}

	// Prepare batch insert with UPSERT
	query := `
		INSERT INTO draft_card_ratings (
			arena_id, expansion, format, colors,
			gihwr, ohwr, gpwr, gdwr, ihdwr,
			gihwr_delta, ohwr_delta, gdwr_delta, ihdwr_delta,
			alsa, ata,
			gih, oh, gp, gd, ihd,
			games_played, num_decks,
			start_date, end_date, cached_at, last_updated
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(arena_id, expansion, format, colors, start_date, end_date)
		DO UPDATE SET
			gihwr = excluded.gihwr,
			ohwr = excluded.ohwr,
			gpwr = excluded.gpwr,
			gdwr = excluded.gdwr,
			ihdwr = excluded.ihdwr,
			gihwr_delta = excluded.gihwr_delta,
			ohwr_delta = excluded.ohwr_delta,
			gdwr_delta = excluded.gdwr_delta,
			ihdwr_delta = excluded.ihdwr_delta,
			alsa = excluded.alsa,
			ata = excluded.ata,
			gih = excluded.gih,
			oh = excluded.oh,
			gp = excluded.gp,
			gd = excluded.gd,
			ihd = excluded.ihd,
			games_played = excluded.games_played,
			num_decks = excluded.num_decks,
			last_updated = CURRENT_TIMESTAMP
	`

	tx, err := s.db.Conn().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, rating := range ratings {
		// Skip cards without Arena ID
		if rating.MTGAID == 0 {
			continue
		}

		_, err := stmt.ExecContext(ctx,
			rating.MTGAID, expansion, format, colors,
			rating.GIHWR, rating.OHWR, rating.GPWR, rating.GDWR, rating.IHDWR,
			rating.GIHWRDelta, rating.OHWRDelta, rating.GDWRDelta, rating.IHDWRDelta,
			rating.ALSA, rating.ATA,
			rating.GIH, rating.OH, rating.GP, rating.GD, rating.IHD,
			rating.GamesPlayed, rating.NumberDecks,
			startDate, endDate,
		)
		if err != nil {
			return fmt.Errorf("failed to insert rating for card %d: %w", rating.MTGAID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetCardRating retrieves a single card rating.
func (s *Service) GetCardRating(ctx context.Context, arenaID int, expansion, format, colors string) (*DraftCardRating, error) {
	query := `
		SELECT id, arena_id, expansion, format, colors,
		       gihwr, ohwr, gpwr, gdwr, ihdwr,
		       gihwr_delta, ohwr_delta, gdwr_delta, ihdwr_delta,
		       alsa, ata,
		       gih, oh, gp, gd, ihd,
		       games_played, num_decks,
		       start_date, end_date, cached_at, last_updated
		FROM draft_card_ratings
		WHERE arena_id = ? AND expansion = ? AND format = ? AND colors = ?
		ORDER BY last_updated DESC
		LIMIT 1
	`

	var rating DraftCardRating
	err := s.db.Conn().QueryRowContext(ctx, query, arenaID, expansion, format, colors).Scan(
		&rating.ID, &rating.ArenaID, &rating.Expansion, &rating.Format, &rating.Colors,
		&rating.GIHWR, &rating.OHWR, &rating.GPWR, &rating.GDWR, &rating.IHDWR,
		&rating.GIHWRDelta, &rating.OHWRDelta, &rating.GDWRDelta, &rating.IHDWRDelta,
		&rating.ALSA, &rating.ATA,
		&rating.GIH, &rating.OH, &rating.GP, &rating.GD, &rating.IHD,
		&rating.GamesPlayed, &rating.NumDecks,
		&rating.StartDate, &rating.EndDate, &rating.CachedAt, &rating.LastUpdated,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get card rating: %w", err)
	}

	return &rating, nil
}

// GetCardRatingsForSet retrieves all card ratings for a given expansion and format.
func (s *Service) GetCardRatingsForSet(ctx context.Context, expansion, format, colors string) ([]*DraftCardRating, error) {
	query := `
		SELECT id, arena_id, expansion, format, colors,
		       gihwr, ohwr, gpwr, gdwr, ihdwr,
		       gihwr_delta, ohwr_delta, gdwr_delta, ihdwr_delta,
		       alsa, ata,
		       gih, oh, gp, gd, ihd,
		       games_played, num_decks,
		       start_date, end_date, cached_at, last_updated
		FROM draft_card_ratings
		WHERE expansion = ? AND format = ? AND colors = ?
		ORDER BY gihwr DESC
	`

	rows, err := s.db.Conn().QueryContext(ctx, query, expansion, format, colors)
	if err != nil {
		return nil, fmt.Errorf("failed to query card ratings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var ratings []*DraftCardRating
	for rows.Next() {
		var rating DraftCardRating
		err := rows.Scan(
			&rating.ID, &rating.ArenaID, &rating.Expansion, &rating.Format, &rating.Colors,
			&rating.GIHWR, &rating.OHWR, &rating.GPWR, &rating.GDWR, &rating.IHDWR,
			&rating.GIHWRDelta, &rating.OHWRDelta, &rating.GDWRDelta, &rating.IHDWRDelta,
			&rating.ALSA, &rating.ATA,
			&rating.GIH, &rating.OH, &rating.GP, &rating.GD, &rating.IHD,
			&rating.GamesPlayed, &rating.NumDecks,
			&rating.StartDate, &rating.EndDate, &rating.CachedAt, &rating.LastUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card rating: %w", err)
		}
		ratings = append(ratings, &rating)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating card ratings: %w", err)
	}

	return ratings, nil
}

// GetStaleCardRatings returns card ratings older than the specified duration.
func (s *Service) GetStaleCardRatings(ctx context.Context, olderThan time.Duration) ([]*DraftCardRating, error) {
	seconds := int64(olderThan.Seconds())
	query := `
		SELECT DISTINCT expansion, format, colors
		FROM draft_card_ratings
		WHERE unixepoch(last_updated) <= unixepoch('now', '-' || ? || ' seconds')
		ORDER BY last_updated ASC
	`

	rows, err := s.db.Conn().QueryContext(ctx, query, seconds)
	if err != nil {
		return nil, fmt.Errorf("failed to query stale ratings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var ratings []*DraftCardRating
	for rows.Next() {
		var rating DraftCardRating
		err := rows.Scan(&rating.Expansion, &rating.Format, &rating.Colors)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stale rating: %w", err)
		}
		ratings = append(ratings, &rating)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stale ratings: %w", err)
	}

	return ratings, nil
}

// SaveColorRatings saves a batch of color ratings to the database.
func (s *Service) SaveColorRatings(ctx context.Context, ratings []seventeenlands.ColorRating, expansion, eventType, startDate, endDate string) error {
	if len(ratings) == 0 {
		return nil
	}

	query := `
		INSERT INTO draft_color_ratings (
			expansion, event_type, color_combination,
			win_rate, games_played, num_decks,
			start_date, end_date, cached_at, last_updated
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(expansion, event_type, color_combination, start_date, end_date)
		DO UPDATE SET
			win_rate = excluded.win_rate,
			games_played = excluded.games_played,
			num_decks = excluded.num_decks,
			last_updated = CURRENT_TIMESTAMP
	`

	tx, err := s.db.Conn().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, rating := range ratings {
		// Normalize color combination (sort alphabetically)
		colorCombo := normalizeColorCombination(rating.ColorName)

		_, err := stmt.ExecContext(ctx,
			expansion, eventType, colorCombo,
			rating.WinRate, rating.GamesPlayed, rating.NumberDecks,
			startDate, endDate,
		)
		if err != nil {
			return fmt.Errorf("failed to insert color rating %s: %w", colorCombo, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetColorRatings retrieves color ratings for a given expansion and event type.
func (s *Service) GetColorRatings(ctx context.Context, expansion, eventType string) ([]*DraftColorRating, error) {
	query := `
		SELECT id, expansion, event_type, color_combination,
		       win_rate, games_played, num_decks,
		       start_date, end_date, cached_at, last_updated
		FROM draft_color_ratings
		WHERE expansion = ? AND event_type = ?
		ORDER BY win_rate DESC
	`

	rows, err := s.db.Conn().QueryContext(ctx, query, expansion, eventType)
	if err != nil {
		return nil, fmt.Errorf("failed to query color ratings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var ratings []*DraftColorRating
	for rows.Next() {
		var rating DraftColorRating
		err := rows.Scan(
			&rating.ID, &rating.Expansion, &rating.EventType, &rating.ColorCombination,
			&rating.WinRate, &rating.GamesPlayed, &rating.NumDecks,
			&rating.StartDate, &rating.EndDate, &rating.CachedAt, &rating.LastUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan color rating: %w", err)
		}
		ratings = append(ratings, &rating)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating color ratings: %w", err)
	}

	return ratings, nil
}

// GetStaleColorRatings returns color ratings older than the specified duration.
func (s *Service) GetStaleColorRatings(ctx context.Context, olderThan time.Duration) ([]*DraftColorRating, error) {
	seconds := int64(olderThan.Seconds())
	query := `
		SELECT DISTINCT expansion, event_type
		FROM draft_color_ratings
		WHERE unixepoch(last_updated) <= unixepoch('now', '-' || ? || ' seconds')
		ORDER BY last_updated ASC
	`

	rows, err := s.db.Conn().QueryContext(ctx, query, seconds)
	if err != nil {
		return nil, fmt.Errorf("failed to query stale color ratings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var ratings []*DraftColorRating
	for rows.Next() {
		var rating DraftColorRating
		err := rows.Scan(&rating.Expansion, &rating.EventType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stale color rating: %w", err)
		}
		ratings = append(ratings, &rating)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stale color ratings: %w", err)
	}

	return ratings, nil
}

// normalizeColorCombination sorts color letters alphabetically for consistent storage.
// E.g., "RW" -> "RW", "WR" -> "RW", "BGU" -> "BGU"
func normalizeColorCombination(colors string) string {
	// Split into individual color letters
	colorSlice := strings.Split(colors, "")

	// Sort using WUBRG order
	sorted := make([]string, 0, len(colorSlice))
	for _, c := range []string{"W", "U", "B", "R", "G"} {
		for _, color := range colorSlice {
			if color == c {
				sorted = append(sorted, color)
			}
		}
	}

	// If no colors matched WUBRG, return original
	if len(sorted) == 0 {
		return colors
	}

	return strings.Join(sorted, "")
}
