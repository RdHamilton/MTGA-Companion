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
	SaveSetRatings(ctx context.Context, setCode, draftFormat string, cardRatings []seventeenlands.CardRating, colorRatings []seventeenlands.ColorRating, dataSource string) error

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

	// Retention/Statistics methods
	GetAllSnapshots(ctx context.Context) ([]*SnapshotInfo, error)
	DeleteSnapshotsBatch(ctx context.Context, ids []int) error
	GetSnapshotCountByExpansion(ctx context.Context) (map[string]int, error)
	GetOldestSnapshotDate(ctx context.Context) (time.Time, error)
	GetNewestSnapshotDate(ctx context.Context) (time.Time, error)

	// Trend methods
	GetCardWinRateTrend(ctx context.Context, arenaID int, expansion string, days int) ([]*TrendPoint, error)
	GetExpansionCardIDs(ctx context.Context, expansion string, days int) ([]int, error)
	GetCardRatingHistory(ctx context.Context, arenaID int, expansion string) ([]*CardRatingSnapshot, error)
	GetPeriodAverages(ctx context.Context, expansion string, startDate, endDate time.Time) (map[int]*PeriodAverage, error)

	// Lookup methods
	GetSetCodeByArenaID(ctx context.Context, arenaID string) (string, error)

	// Staleness tracking methods
	GetStatisticsStaleness(ctx context.Context, staleAgeSeconds int) (*StatisticsStaleness, error)
	GetStaleSets(ctx context.Context, staleAgeSeconds int) ([]string, error)
	GetStaleStats(ctx context.Context, staleAgeSeconds int) ([]*StaleStatItem, error)
}

// StatisticsStaleness contains counts of fresh/stale statistics.
type StatisticsStaleness struct {
	Total     int
	Fresh     int
	Stale     int
	StaleSets []string
}

// StaleStatItem represents a set with stale statistics.
type StaleStatItem struct {
	SetCode     string
	Format      string
	LastUpdated string
}

// SnapshotInfo contains information about a draft statistics snapshot.
type SnapshotInfo struct {
	ID          int
	ArenaID     int
	Expansion   string
	Format      string
	Colors      string
	StartDate   string
	EndDate     string
	CachedAt    time.Time
	LastUpdated time.Time
}

// TrendPoint represents a single point in a trend analysis.
type TrendPoint struct {
	Date        time.Time
	GIHWR       float64
	OHWR        float64
	ALSA        float64
	ATA         float64
	SampleSize  int
	GamesPlayed int
}

// CardRatingSnapshot represents a point-in-time card rating for history.
type CardRatingSnapshot struct {
	ID          int
	ArenaID     int
	Expansion   string
	Format      string
	Colors      string
	GIHWR       float64
	OHWR        float64
	GPWR        float64
	GDWR        float64
	IHDWR       float64
	ALSA        float64
	ATA         float64
	GIH         int
	GamesPlayed int
	NumDecks    int
	StartDate   string
	EndDate     string
	CachedAt    time.Time
	LastUpdated time.Time
}

// PeriodAverage represents average stats for a card during a time period.
type PeriodAverage struct {
	ArenaID    int
	AvgGIHWR   float64
	TotalGIH   int
	SampleSize int
}

type draftRatingsRepository struct {
	db *sql.DB
}

// NewDraftRatingsRepository creates a new draft ratings repository.
func NewDraftRatingsRepository(db *sql.DB) DraftRatingsRepository {
	return &draftRatingsRepository{db: db}
}

// SaveSetRatings saves card and color ratings for a set.
func (r *draftRatingsRepository) SaveSetRatings(ctx context.Context, setCode, draftFormat string, cardRatings []seventeenlands.CardRating, colorRatings []seventeenlands.ColorRating, dataSource string) error {
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
			gihwr, ohwr, alsa, ata, gih_count, data_source, cached_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(set_code, draft_format, arena_id) DO UPDATE SET
			name = excluded.name,
			color = excluded.color,
			rarity = excluded.rarity,
			gihwr = excluded.gihwr,
			ohwr = excluded.ohwr,
			alsa = excluded.alsa,
			ata = excluded.ata,
			gih_count = excluded.gih_count,
			data_source = excluded.data_source,
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
			dataSource,
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

// GetAllSnapshots returns all draft card rating snapshots for retention analysis.
func (r *draftRatingsRepository) GetAllSnapshots(ctx context.Context) ([]*SnapshotInfo, error) {
	query := `
		SELECT id, arena_id, set_code, draft_format, color,
			   cached_at
		FROM draft_card_ratings
		ORDER BY cached_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var snapshots []*SnapshotInfo
	for rows.Next() {
		snapshot := &SnapshotInfo{}
		var cachedAt time.Time
		var color sql.NullString

		err := rows.Scan(
			&snapshot.ID,
			&snapshot.ArenaID,
			&snapshot.Expansion,
			&snapshot.Format,
			&color,
			&cachedAt,
		)
		if err != nil {
			return nil, err
		}

		snapshot.CachedAt = cachedAt
		snapshot.LastUpdated = cachedAt
		if color.Valid {
			snapshot.Colors = color.String
		}

		snapshots = append(snapshots, snapshot)
	}

	return snapshots, rows.Err()
}

// DeleteSnapshotsBatch deletes a batch of snapshots by ID.
func (r *draftRatingsRepository) DeleteSnapshotsBatch(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	// Build IN clause
	query := `DELETE FROM draft_card_ratings WHERE id IN (`
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += ")"

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// GetSnapshotCountByExpansion returns the number of snapshots for each expansion.
func (r *draftRatingsRepository) GetSnapshotCountByExpansion(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT set_code, COUNT(*) as count
		FROM draft_card_ratings
		GROUP BY set_code
		ORDER BY count DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	counts := make(map[string]int)
	for rows.Next() {
		var expansion string
		var count int
		if err := rows.Scan(&expansion, &count); err != nil {
			return nil, err
		}
		counts[expansion] = count
	}

	return counts, rows.Err()
}

// GetOldestSnapshotDate returns the oldest snapshot date.
func (r *draftRatingsRepository) GetOldestSnapshotDate(ctx context.Context) (time.Time, error) {
	query := `SELECT MIN(cached_at) FROM draft_card_ratings`

	var cachedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query).Scan(&cachedAt)
	if err != nil {
		return time.Time{}, err
	}

	if !cachedAt.Valid {
		return time.Time{}, nil
	}

	return cachedAt.Time, nil
}

// GetNewestSnapshotDate returns the newest snapshot date.
func (r *draftRatingsRepository) GetNewestSnapshotDate(ctx context.Context) (time.Time, error) {
	query := `SELECT MAX(cached_at) FROM draft_card_ratings`

	var cachedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query).Scan(&cachedAt)
	if err != nil {
		return time.Time{}, err
	}

	if !cachedAt.Valid {
		return time.Time{}, nil
	}

	return cachedAt.Time, nil
}

// GetCardWinRateTrend returns the win rate trend for a specific card over time.
func (r *draftRatingsRepository) GetCardWinRateTrend(ctx context.Context, arenaID int, expansion string, days int) ([]*TrendPoint, error) {
	query := `
		SELECT cached_at, gihwr, ohwr, alsa, ata, gih_count
		FROM draft_card_ratings
		WHERE arena_id = ?
		  AND set_code = ?
		  AND cached_at >= datetime('now', '-' || ? || ' days')
		ORDER BY cached_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, arenaID, expansion, days)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var points []*TrendPoint
	for rows.Next() {
		point := &TrendPoint{}
		var cachedAt time.Time
		var gihwr, ohwr, alsa, ata sql.NullFloat64
		var gih sql.NullInt64

		err := rows.Scan(&cachedAt, &gihwr, &ohwr, &alsa, &ata, &gih)
		if err != nil {
			return nil, err
		}

		point.Date = cachedAt
		if gihwr.Valid {
			point.GIHWR = gihwr.Float64
		}
		if ohwr.Valid {
			point.OHWR = ohwr.Float64
		}
		if alsa.Valid {
			point.ALSA = alsa.Float64
		}
		if ata.Valid {
			point.ATA = ata.Float64
		}
		if gih.Valid {
			point.SampleSize = int(gih.Int64)
		}

		points = append(points, point)
	}

	return points, rows.Err()
}

// GetExpansionCardIDs returns distinct arena IDs for cards in an expansion within a time range.
func (r *draftRatingsRepository) GetExpansionCardIDs(ctx context.Context, expansion string, days int) ([]int, error) {
	query := `
		SELECT DISTINCT arena_id
		FROM draft_card_ratings
		WHERE set_code = ?
		  AND cached_at >= datetime('now', '-' || ? || ' days')
	`

	rows, err := r.db.QueryContext(ctx, query, expansion, days)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// GetCardRatingHistory returns the complete rating history for a card.
func (r *draftRatingsRepository) GetCardRatingHistory(ctx context.Context, arenaID int, expansion string) ([]*CardRatingSnapshot, error) {
	query := `
		SELECT id, arena_id, set_code, draft_format, color,
			   gihwr, ohwr, alsa, ata, gih_count, cached_at
		FROM draft_card_ratings
		WHERE arena_id = ?
		  AND set_code = ?
		ORDER BY cached_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, arenaID, expansion)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var history []*CardRatingSnapshot
	for rows.Next() {
		rating := &CardRatingSnapshot{}
		var cachedAt time.Time
		var color sql.NullString
		var gihwr, ohwr, alsa, ata sql.NullFloat64
		var gih sql.NullInt64

		err := rows.Scan(
			&rating.ID,
			&rating.ArenaID,
			&rating.Expansion,
			&rating.Format,
			&color,
			&gihwr,
			&ohwr,
			&alsa,
			&ata,
			&gih,
			&cachedAt,
		)
		if err != nil {
			return nil, err
		}

		rating.CachedAt = cachedAt
		rating.LastUpdated = cachedAt
		if color.Valid {
			rating.Colors = color.String
		}
		if gihwr.Valid {
			rating.GIHWR = gihwr.Float64
		}
		if ohwr.Valid {
			rating.OHWR = ohwr.Float64
		}
		if alsa.Valid {
			rating.ALSA = alsa.Float64
		}
		if ata.Valid {
			rating.ATA = ata.Float64
		}
		if gih.Valid {
			rating.GIH = int(gih.Int64)
		}

		history = append(history, rating)
	}

	return history, rows.Err()
}

// GetPeriodAverages returns average GIHWR for cards during a time period.
func (r *draftRatingsRepository) GetPeriodAverages(ctx context.Context, expansion string, startDate, endDate time.Time) (map[int]*PeriodAverage, error) {
	query := `
		SELECT arena_id, AVG(gihwr) as avg_gihwr, SUM(gih_count) as total_gih, COUNT(*) as sample_size
		FROM draft_card_ratings
		WHERE set_code = ?
		  AND cached_at BETWEEN ? AND ?
		GROUP BY arena_id
		HAVING total_gih > 100
	`

	rows, err := r.db.QueryContext(ctx, query, expansion, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	averages := make(map[int]*PeriodAverage)
	for rows.Next() {
		var arenaID int
		var avgGIHWR sql.NullFloat64
		var totalGIH, sampleSize sql.NullInt64

		if err := rows.Scan(&arenaID, &avgGIHWR, &totalGIH, &sampleSize); err != nil {
			return nil, err
		}

		if avgGIHWR.Valid && totalGIH.Valid {
			averages[arenaID] = &PeriodAverage{
				ArenaID:    arenaID,
				AvgGIHWR:   avgGIHWR.Float64,
				TotalGIH:   int(totalGIH.Int64),
				SampleSize: int(sampleSize.Int64),
			}
		}
	}

	return averages, rows.Err()
}

// GetSetCodeByArenaID returns the set code for a card by its Arena ID.
// Returns empty string if not found.
func (r *draftRatingsRepository) GetSetCodeByArenaID(ctx context.Context, arenaID string) (string, error) {
	query := `SELECT DISTINCT set_code FROM draft_card_ratings WHERE arena_id = ? LIMIT 1`
	var setCode string
	err := r.db.QueryRowContext(ctx, query, arenaID).Scan(&setCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return setCode, nil
}

// GetStatisticsStaleness returns counts of fresh and stale draft statistics.
func (r *draftRatingsRepository) GetStatisticsStaleness(ctx context.Context, staleAgeSeconds int) (*StatisticsStaleness, error) {
	countQuery := `
		SELECT
			COUNT(DISTINCT arena_id || '-' || set_code || '-' || draft_format) as total,
			SUM(CASE WHEN cached_at >= datetime('now', '-' || ? || ' seconds') THEN 1 ELSE 0 END) as fresh,
			SUM(CASE WHEN cached_at < datetime('now', '-' || ? || ' seconds') THEN 1 ELSE 0 END) as stale
		FROM draft_card_ratings
		WHERE cached_at IS NOT NULL
	`

	result := &StatisticsStaleness{
		StaleSets: []string{},
	}
	err := r.db.QueryRowContext(ctx, countQuery, staleAgeSeconds, staleAgeSeconds).Scan(
		&result.Total, &result.Fresh, &result.Stale,
	)
	if err != nil {
		return nil, err
	}

	// Get stale sets
	staleSets, err := r.GetStaleSets(ctx, staleAgeSeconds)
	if err != nil {
		// Return counts even if sets query fails
		return result, nil
	}
	result.StaleSets = staleSets

	return result, nil
}

// GetStaleSets returns distinct set codes with stale statistics.
func (r *draftRatingsRepository) GetStaleSets(ctx context.Context, staleAgeSeconds int) ([]string, error) {
	query := `
		SELECT DISTINCT set_code
		FROM draft_card_ratings
		WHERE cached_at < datetime('now', '-' || ? || ' seconds')
		ORDER BY set_code
	`

	rows, err := r.db.QueryContext(ctx, query, staleAgeSeconds)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var sets []string
	for rows.Next() {
		var setCode string
		if err := rows.Scan(&setCode); err != nil {
			continue
		}
		sets = append(sets, setCode)
	}

	return sets, rows.Err()
}

// GetStaleStats returns sets/formats with stale statistics, ordered by oldest first.
func (r *draftRatingsRepository) GetStaleStats(ctx context.Context, staleAgeSeconds int) ([]*StaleStatItem, error) {
	query := `
		SELECT DISTINCT set_code, draft_format, MAX(cached_at) as last_updated
		FROM draft_card_ratings
		WHERE cached_at < datetime('now', '-' || ? || ' seconds')
		GROUP BY set_code, draft_format
		ORDER BY last_updated ASC
	`

	rows, err := r.db.QueryContext(ctx, query, staleAgeSeconds)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []*StaleStatItem
	for rows.Next() {
		item := &StaleStatItem{}
		if err := rows.Scan(&item.SetCode, &item.Format, &item.LastUpdated); err != nil {
			continue
		}
		items = append(items, item)
	}

	return items, rows.Err()
}
