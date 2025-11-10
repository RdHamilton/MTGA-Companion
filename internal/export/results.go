package export

import (
	"context"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// ResultBreakdownExportRow represents a single result reason breakdown for CSV export.
type ResultBreakdownExportRow struct {
	ResultType         string  `csv:"result_type" json:"result_type"`                     // "win" or "loss"
	Reason             string  `csv:"reason" json:"reason"`                               // "normal", "concede", etc.
	Count              int     `csv:"count" json:"count"`                                 // Number of matches
	Percentage         float64 `csv:"percentage" json:"percentage"`                       // Percentage of total
	TotalForResultType int     `csv:"total_for_result_type" json:"total_for_result_type"` // Total wins or losses
}

// ResultBreakdownExportJSON represents complete result breakdown in JSON format.
type ResultBreakdownExportJSON struct {
	Wins       *ResultTypeBreakdownJSON `json:"wins"`
	Losses     *ResultTypeBreakdownJSON `json:"losses"`
	ExportedAt string                   `json:"exported_at"`
	Filter     *FilterJSON              `json:"filter,omitempty"`
}

// ResultTypeBreakdownJSON represents breakdown for wins or losses in JSON format.
type ResultTypeBreakdownJSON struct {
	Total              int               `json:"total"`
	Normal             *ReasonDetailJSON `json:"normal,omitempty"`
	Concede            *ReasonDetailJSON `json:"concede,omitempty"`
	Timeout            *ReasonDetailJSON `json:"timeout,omitempty"`
	Disconnect         *ReasonDetailJSON `json:"disconnect,omitempty"`
	OpponentConcede    *ReasonDetailJSON `json:"opponent_concede,omitempty"`
	OpponentTimeout    *ReasonDetailJSON `json:"opponent_timeout,omitempty"`
	OpponentDisconnect *ReasonDetailJSON `json:"opponent_disconnect,omitempty"`
	Draw               *ReasonDetailJSON `json:"draw,omitempty"`
	Other              *ReasonDetailJSON `json:"other,omitempty"`
}

// ReasonDetailJSON represents details for a specific reason.
type ReasonDetailJSON struct {
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// FilterJSON represents the filter applied to the data.
type FilterJSON struct {
	StartDate *string `json:"start_date,omitempty"`
	EndDate   *string `json:"end_date,omitempty"`
	Format    *string `json:"format,omitempty"`
}

// ExportResultBreakdown exports result breakdown data to the specified format.
func ExportResultBreakdown(ctx context.Context, service *storage.Service, filter models.StatsFilter, opts Options) error {
	// Get matches for analysis
	matches, err := service.GetMatches(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get matches: %w", err)
	}

	if len(matches) == 0 {
		return fmt.Errorf("no matches found for the specified filter")
	}

	// Calculate breakdowns
	winBreakdown := calculateResultBreakdown(matches, true)
	lossBreakdown := calculateResultBreakdown(matches, false)

	if winBreakdown.Total == 0 && lossBreakdown.Total == 0 {
		return fmt.Errorf("no result data found")
	}

	switch opts.Format {
	case FormatCSV:
		return exportResultBreakdownCSV(winBreakdown, lossBreakdown, opts)
	case FormatJSON:
		return exportResultBreakdownJSON(winBreakdown, lossBreakdown, filter, opts)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

// resultBreakdown represents a breakdown of match results by reason.
type resultBreakdown struct {
	Normal             int
	Concede            int
	Timeout            int
	Draw               int
	Disconnect         int
	OpponentConcede    int
	OpponentTimeout    int
	OpponentDisconnect int
	Other              int
	Total              int
}

// calculateResultBreakdown calculates a breakdown of match results by reason.
func calculateResultBreakdown(matches []*models.Match, isWin bool) resultBreakdown {
	breakdown := resultBreakdown{}

	for _, match := range matches {
		// Filter by win/loss
		if isWin && match.Result != "win" {
			continue
		}
		if !isWin && match.Result != "loss" {
			continue
		}

		breakdown.Total++

		if match.ResultReason == nil {
			breakdown.Normal++
			continue
		}

		reason := *match.ResultReason
		switch reason {
		case "normal":
			breakdown.Normal++
		case "concede":
			breakdown.Concede++
		case "timeout":
			breakdown.Timeout++
		case "disconnect":
			breakdown.Disconnect++
		case "opponent_concede":
			breakdown.OpponentConcede++
		case "opponent_timeout":
			breakdown.OpponentTimeout++
		case "opponent_disconnect":
			breakdown.OpponentDisconnect++
		case "draw":
			breakdown.Draw++
		default:
			breakdown.Other++
		}
	}

	return breakdown
}

// exportResultBreakdownCSV exports result breakdown to CSV format (one row per reason).
func exportResultBreakdownCSV(winBreakdown, lossBreakdown resultBreakdown, opts Options) error {
	var rows []ResultBreakdownExportRow

	// Add win breakdown rows
	if winBreakdown.Total > 0 {
		rows = append(rows, breakdownToRows(winBreakdown, "win")...)
	}

	// Add loss breakdown rows
	if lossBreakdown.Total > 0 {
		rows = append(rows, breakdownToRows(lossBreakdown, "loss")...)
	}

	if len(rows) == 0 {
		return fmt.Errorf("no data to export")
	}

	exporter := NewExporter(opts)
	return exporter.Export(rows)
}

// breakdownToRows converts a result breakdown to CSV rows.
func breakdownToRows(breakdown resultBreakdown, resultType string) []ResultBreakdownExportRow {
	var rows []ResultBreakdownExportRow

	addRow := func(reason string, count int) {
		if count > 0 {
			percentage := float64(count) / float64(breakdown.Total) * 100
			rows = append(rows, ResultBreakdownExportRow{
				ResultType:         resultType,
				Reason:             reason,
				Count:              count,
				Percentage:         percentage,
				TotalForResultType: breakdown.Total,
			})
		}
	}

	addRow("normal", breakdown.Normal)
	addRow("concede", breakdown.Concede)
	addRow("timeout", breakdown.Timeout)
	addRow("disconnect", breakdown.Disconnect)
	addRow("opponent_concede", breakdown.OpponentConcede)
	addRow("opponent_timeout", breakdown.OpponentTimeout)
	addRow("opponent_disconnect", breakdown.OpponentDisconnect)
	addRow("draw", breakdown.Draw)
	addRow("other", breakdown.Other)

	return rows
}

// exportResultBreakdownJSON exports result breakdown to JSON format.
func exportResultBreakdownJSON(winBreakdown, lossBreakdown resultBreakdown, filter models.StatsFilter, opts Options) error {
	jsonData := ResultBreakdownExportJSON{
		ExportedAt: time.Now().Format("2006-01-02T15:04:05Z07:00"),
	}

	// Add filter info if present
	if filter.StartDate != nil || filter.EndDate != nil || filter.Format != nil {
		filterJSON := &FilterJSON{}
		if filter.StartDate != nil {
			dateStr := filter.StartDate.Format("2006-01-02")
			filterJSON.StartDate = &dateStr
		}
		if filter.EndDate != nil {
			dateStr := filter.EndDate.Format("2006-01-02")
			filterJSON.EndDate = &dateStr
		}
		if filter.Format != nil {
			filterJSON.Format = filter.Format
		}
		jsonData.Filter = filterJSON
	}

	// Wins breakdown
	if winBreakdown.Total > 0 {
		jsonData.Wins = breakdownToJSON(winBreakdown)
	}

	// Losses breakdown
	if lossBreakdown.Total > 0 {
		jsonData.Losses = breakdownToJSON(lossBreakdown)
	}

	exporter := NewExporter(opts)
	return exporter.Export(jsonData)
}

// breakdownToJSON converts a result breakdown to JSON format.
func breakdownToJSON(breakdown resultBreakdown) *ResultTypeBreakdownJSON {
	json := &ResultTypeBreakdownJSON{
		Total: breakdown.Total,
	}

	addDetail := func(count int) *ReasonDetailJSON {
		if count == 0 {
			return nil
		}
		percentage := float64(count) / float64(breakdown.Total) * 100
		return &ReasonDetailJSON{
			Count:      count,
			Percentage: percentage,
		}
	}

	json.Normal = addDetail(breakdown.Normal)
	json.Concede = addDetail(breakdown.Concede)
	json.Timeout = addDetail(breakdown.Timeout)
	json.Disconnect = addDetail(breakdown.Disconnect)
	json.OpponentConcede = addDetail(breakdown.OpponentConcede)
	json.OpponentTimeout = addDetail(breakdown.OpponentTimeout)
	json.OpponentDisconnect = addDetail(breakdown.OpponentDisconnect)
	json.Draw = addDetail(breakdown.Draw)
	json.Other = addDetail(breakdown.Other)

	return json
}
