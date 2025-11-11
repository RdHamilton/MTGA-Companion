package refresh

import (
	"context"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// DataType represents different types of cached data.
type DataType int

const (
	DataTypeUnknown DataType = iota
	DataTypeMetadata          // Scryfall metadata
	DataTypeStatistics        // 17Lands statistics
	DataTypeSetInfo           // Set information
	DataTypeBulk              // Bulk data
)

func (dt DataType) String() string {
	switch dt {
	case DataTypeMetadata:
		return "Metadata"
	case DataTypeStatistics:
		return "Statistics"
	case DataTypeSetInfo:
		return "SetInfo"
	case DataTypeBulk:
		return "Bulk"
	default:
		return "Unknown"
	}
}

// StaleAge returns the staleness threshold for each data type.
func (dt DataType) StaleAge() time.Duration {
	switch dt {
	case DataTypeMetadata:
		return 7 * 24 * time.Hour // 7 days
	case DataTypeStatistics:
		return 1 * 24 * time.Hour // 1 day (active sets)
	case DataTypeSetInfo:
		return 30 * 24 * time.Hour // 30 days
	case DataTypeBulk:
		return 7 * 24 * time.Hour // 7 days
	default:
		return 24 * time.Hour
	}
}

// DataFreshness represents the freshness status of cached data.
type DataFreshness struct {
	Type         DataType
	LastUpdated  time.Time
	StaleAge     time.Duration
	Age          time.Duration
	IsFresh      bool
	IsStale      bool
	IsVeryStale  bool // >2x stale age
	ItemID       string
	ItemName     string
}

// StalenessSummary provides an overview of data freshness.
type StalenessSummary struct {
	TotalCards      int
	FreshCards      int
	StaleCards      int
	VeryStaleCards  int
	TotalStats      int
	FreshStats      int
	StaleStats      int
	StaleSets       []string
	NextRefreshDue  time.Time
	RefreshesNeeded int
}

// RefreshItem represents an item that needs refreshing.
type RefreshItem struct {
	Type         DataType
	ArenaID      int
	SetCode      string
	Format       string
	LastUpdated  time.Time
	StaleDays    int
	IsActive     bool
	AccessCount  int
	Priority     int
}

// RefreshPriority defines refresh priority levels.
type RefreshPriority int

const (
	PriorityLow RefreshPriority = iota
	PriorityMedium
	PriorityHigh
)

func (rp RefreshPriority) String() string {
	switch rp {
	case PriorityLow:
		return "Low"
	case PriorityMedium:
		return "Medium"
	case PriorityHigh:
		return "High"
	default:
		return "Unknown"
	}
}

// StalenessTracker tracks data freshness.
type StalenessTracker struct {
	storage *storage.Service
}

// NewStalenessTracker creates a new staleness tracker.
func NewStalenessTracker(storage *storage.Service) *StalenessTracker {
	return &StalenessTracker{
		storage: storage,
	}
}

// GetSummary returns a summary of data staleness.
func (st *StalenessTracker) GetSummary(ctx context.Context) (*StalenessSummary, error) {
	summary := &StalenessSummary{
		StaleSets: []string{},
	}

	// Get card metadata staleness
	metadataSummary, err := st.getMetadataStaleness(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata staleness: %w", err)
	}
	summary.TotalCards = metadataSummary.Total
	summary.FreshCards = metadataSummary.Fresh
	summary.StaleCards = metadataSummary.Stale
	summary.VeryStaleCards = metadataSummary.VeryStale

	// Get statistics staleness
	statsSummary, err := st.getStatisticsStaleness(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics staleness: %w", err)
	}
	summary.TotalStats = statsSummary.Total
	summary.FreshStats = statsSummary.Fresh
	summary.StaleStats = statsSummary.Stale
	summary.StaleSets = statsSummary.StaleSets

	// Calculate next refresh due
	summary.NextRefreshDue = st.calculateNextRefresh()
	summary.RefreshesNeeded = summary.StaleCards + summary.StaleStats

	return summary, nil
}

type metadataStaleness struct {
	Total     int
	Fresh     int
	Stale     int
	VeryStale int
}

// getMetadataStaleness checks Scryfall metadata staleness.
func (st *StalenessTracker) getMetadataStaleness(ctx context.Context) (*metadataStaleness, error) {
	staleAge := DataTypeMetadata.StaleAge()
	veryStaleAge := staleAge * 2

	query := `
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN last_updated >= datetime('now', '-' || ? || ' seconds') THEN 1 ELSE 0 END) as fresh,
			SUM(CASE WHEN last_updated < datetime('now', '-' || ? || ' seconds')
				AND last_updated >= datetime('now', '-' || ? || ' seconds') THEN 1 ELSE 0 END) as stale,
			SUM(CASE WHEN last_updated < datetime('now', '-' || ? || ' seconds') THEN 1 ELSE 0 END) as very_stale
		FROM cards
		WHERE last_updated IS NOT NULL
	`

	var result metadataStaleness
	err := st.storage.GetDB().QueryRowContext(ctx, query,
		int(staleAge.Seconds()),
		int(staleAge.Seconds()),
		int(veryStaleAge.Seconds()),
		int(veryStaleAge.Seconds()),
	).Scan(&result.Total, &result.Fresh, &result.Stale, &result.VeryStale)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

type statisticsStaleness struct {
	Total     int
	Fresh     int
	Stale     int
	StaleSets []string
}

// getStatisticsStaleness checks 17Lands statistics staleness.
func (st *StalenessTracker) getStatisticsStaleness(ctx context.Context) (*statisticsStaleness, error) {
	staleAge := DataTypeStatistics.StaleAge()

	// Count total, fresh, and stale statistics
	countQuery := `
		SELECT
			COUNT(DISTINCT arena_id || '-' || expansion || '-' || format) as total,
			SUM(CASE WHEN last_updated >= datetime('now', '-' || ? || ' seconds') THEN 1 ELSE 0 END) as fresh,
			SUM(CASE WHEN last_updated < datetime('now', '-' || ? || ' seconds') THEN 1 ELSE 0 END) as stale
		FROM draft_card_ratings
		WHERE last_updated IS NOT NULL
	`

	var result statisticsStaleness
	err := st.storage.GetDB().QueryRowContext(ctx, countQuery,
		int(staleAge.Seconds()),
		int(staleAge.Seconds()),
	).Scan(&result.Total, &result.Fresh, &result.Stale)

	if err != nil {
		return nil, err
	}

	// Get stale sets
	setsQuery := `
		SELECT DISTINCT expansion
		FROM draft_card_ratings
		WHERE last_updated < datetime('now', '-' || ? || ' seconds')
		ORDER BY expansion
	`

	rows, err := st.storage.GetDB().QueryContext(ctx, setsQuery, int(staleAge.Seconds()))
	if err != nil {
		return &result, nil // Return counts even if sets query fails
	}
	defer rows.Close()

	for rows.Next() {
		var setCode string
		if err := rows.Scan(&setCode); err != nil {
			continue
		}
		result.StaleSets = append(result.StaleSets, setCode)
	}

	return &result, nil
}

// GetStaleCards returns cards with stale metadata.
func (st *StalenessTracker) GetStaleCards(ctx context.Context, limit int) ([]RefreshItem, error) {
	staleAge := DataTypeMetadata.StaleAge()

	query := `
		SELECT arena_id, set, last_updated
		FROM cards
		WHERE last_updated < datetime('now', '-' || ? || ' seconds')
		ORDER BY last_updated ASC
		LIMIT ?
	`

	rows, err := st.storage.GetDB().QueryContext(ctx, query, int(staleAge.Seconds()), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []RefreshItem
	for rows.Next() {
		var arenaID int
		var setCode string
		var lastUpdatedStr string

		if err := rows.Scan(&arenaID, &setCode, &lastUpdatedStr); err != nil {
			continue
		}

		lastUpdated, _ := time.Parse("2006-01-02 15:04:05", lastUpdatedStr)
		staleDays := int(time.Since(lastUpdated).Hours() / 24)

		items = append(items, RefreshItem{
			Type:        DataTypeMetadata,
			ArenaID:     arenaID,
			SetCode:     setCode,
			LastUpdated: lastUpdated,
			StaleDays:   staleDays,
		})
	}

	return items, nil
}

// GetStaleStats returns sets with stale statistics.
func (st *StalenessTracker) GetStaleStats(ctx context.Context) ([]RefreshItem, error) {
	staleAge := DataTypeStatistics.StaleAge()

	query := `
		SELECT DISTINCT expansion, format, MAX(last_updated) as last_updated
		FROM draft_card_ratings
		WHERE last_updated < datetime('now', '-' || ? || ' seconds')
		GROUP BY expansion, format
		ORDER BY last_updated ASC
	`

	rows, err := st.storage.GetDB().QueryContext(ctx, query, int(staleAge.Seconds()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []RefreshItem
	for rows.Next() {
		var setCode, format, lastUpdatedStr string

		if err := rows.Scan(&setCode, &format, &lastUpdatedStr); err != nil {
			continue
		}

		lastUpdated, _ := time.Parse("2006-01-02 15:04:05", lastUpdatedStr)
		staleDays := int(time.Since(lastUpdated).Hours() / 24)

		items = append(items, RefreshItem{
			Type:        DataTypeStatistics,
			SetCode:     setCode,
			Format:      format,
			LastUpdated: lastUpdated,
			StaleDays:   staleDays,
		})
	}

	return items, nil
}

// calculateNextRefresh determines when the next scheduled refresh should occur.
func (st *StalenessTracker) calculateNextRefresh() time.Time {
	now := time.Now()

	// Next 2 AM (daily active sets refresh)
	next := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
	if now.Hour() >= 2 {
		next = next.Add(24 * time.Hour)
	}

	return next
}

// CheckFreshness checks if data is fresh based on its type and last update time.
func CheckFreshness(dataType DataType, lastUpdated time.Time) DataFreshness {
	staleAge := dataType.StaleAge()
	age := time.Since(lastUpdated)

	return DataFreshness{
		Type:        dataType,
		LastUpdated: lastUpdated,
		StaleAge:    staleAge,
		Age:         age,
		IsFresh:     age < staleAge,
		IsStale:     age >= staleAge && age < staleAge*2,
		IsVeryStale: age >= staleAge*2,
	}
}
