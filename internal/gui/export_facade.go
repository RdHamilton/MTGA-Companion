package gui

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/ramonehamilton/MTGA-Companion/internal/events"
	"github.com/ramonehamilton/MTGA-Companion/internal/export"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logprocessor"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// ExportFacade handles exporting and importing data
type ExportFacade struct {
	services        *Services
	eventDispatcher *events.EventDispatcher
}

// NewExportFacade creates a new ExportFacade
func NewExportFacade(services *Services, eventDispatcher *events.EventDispatcher) *ExportFacade {
	return &ExportFacade{
		services:        services,
		eventDispatcher: eventDispatcher,
	}
}

// ExportToJSON exports all match data to a JSON file.
// Uses the Builder pattern for cleaner configuration.
func (e *ExportFacade) ExportToJSON(ctx context.Context) error {
	if e.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	// Prompt user to select save location
	filePath, err := wailsruntime.SaveFileDialog(ctx, wailsruntime.SaveDialogOptions{
		DefaultFilename: fmt.Sprintf("mtga-matches-%s.json", time.Now().Format("2006-01-02")),
		Title:           "Export Matches to JSON",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to open save dialog: %v", err)}
	}
	if filePath == "" {
		// User cancelled
		return nil
	}

	// Get all matches
	matches, err := e.services.Storage.GetMatches(ctx, models.StatsFilter{})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to get matches: %v", err)}
	}

	// Export to JSON using builder pattern
	err = export.NewExportBuilder().
		WithFormat(export.FormatJSON).
		WithFilePath(filePath).
		WithPrettyJSON(true).
		WithOverwrite(true).
		Export(matches)
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to export to JSON: %v", err)}
	}

	log.Printf("Successfully exported %d matches to %s", len(matches), filePath)
	return nil
}

// ExportToCSV exports all match data to a CSV file.
// Uses the Builder pattern for cleaner configuration.
func (e *ExportFacade) ExportToCSV(ctx context.Context) error {
	if e.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	// Prompt user to select save location
	filePath, err := wailsruntime.SaveFileDialog(ctx, wailsruntime.SaveDialogOptions{
		DefaultFilename: fmt.Sprintf("mtga-matches-%s.csv", time.Now().Format("2006-01-02")),
		Title:           "Export Matches to CSV",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "CSV Files (*.csv)", Pattern: "*.csv"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to open save dialog: %v", err)}
	}
	if filePath == "" {
		// User cancelled
		return nil
	}

	// Get all matches
	matches, err := e.services.Storage.GetMatches(ctx, models.StatsFilter{})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to get matches: %v", err)}
	}

	// Export to CSV using builder pattern
	err = export.NewExportBuilder().
		WithFormat(export.FormatCSV).
		WithFilePath(filePath).
		WithOverwrite(true).
		Export(matches)
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to export to CSV: %v", err)}
	}

	log.Printf("Successfully exported %d matches to %s", len(matches), filePath)
	return nil
}

// ImportFromFile imports match data from a JSON file.
func (e *ExportFacade) ImportFromFile(ctx context.Context) error {
	if e.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	// Prompt user to select file
	filePath, err := wailsruntime.OpenFileDialog(ctx, wailsruntime.OpenDialogOptions{
		Title: "Import Matches from JSON",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to open file dialog: %v", err)}
	}
	if filePath == "" {
		// User cancelled
		return nil
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to read file: %v", err)}
	}

	// Parse JSON
	var matches []*models.Match
	if err := json.Unmarshal(data, &matches); err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to parse JSON: %v", err)}
	}

	// Import matches (this would need a service method to handle duplicate checking)
	imported := 0
	for _, match := range matches {
		// Save each match (skip duplicates)
		if err := e.services.Storage.SaveMatch(ctx, match); err != nil {
			log.Printf("Warning: Failed to import match %s: %v", match.ID, err)
			continue
		}
		imported++
	}

	log.Printf("Successfully imported %d/%d matches from %s", imported, len(matches), filePath)
	return nil
}

// ClearAllData clears all match history from the database.
func (e *ExportFacade) ClearAllData(ctx context.Context) error {
	if e.services == nil || e.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	// Show confirmation dialog
	selection, err := wailsruntime.MessageDialog(ctx, wailsruntime.MessageDialogOptions{
		Type:          wailsruntime.QuestionDialog,
		Title:         "Clear All Data",
		Message:       "⚠️ WARNING: This will permanently delete all match history and statistics.\n\nThis action cannot be undone.\n\nAre you sure you want to continue?",
		DefaultButton: "No",
		Buttons:       []string{"Yes, Delete All Data", "No"},
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to show confirmation dialog: %v", err)}
	}
	if selection != "Yes, Delete All Data" {
		// User cancelled or clicked No
		return nil
	}

	// Delete all matches
	if err := e.services.Storage.ClearAllMatches(ctx); err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to clear data: %v", err)}
	}

	log.Println("Successfully cleared all match history")
	return nil
}

// ImportLogFileResult represents the result of importing a log file.
type ImportLogFileResult struct {
	FileName      string `json:"fileName"`
	EntriesRead   int    `json:"entriesRead"`
	MatchesStored int    `json:"matchesStored"`
	GamesStored   int    `json:"gamesStored"`
	DecksStored   int    `json:"decksStored"`
	RanksStored   int    `json:"ranksStored"`
	QuestsStored  int    `json:"questsStored"`
	DraftsStored  int    `json:"draftsStored"`
	PicksStored   int    `json:"picksStored"`
}

// ExportDraftsToJSON exports draft session data to a JSON file.
// Uses the Builder pattern for flexible export configuration.
func (e *ExportFacade) ExportDraftsToJSON(ctx context.Context) error {
	if e.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	// Prompt user to select save location
	filePath, err := wailsruntime.SaveFileDialog(ctx, wailsruntime.SaveDialogOptions{
		DefaultFilename: fmt.Sprintf("mtga-drafts-%s.json", time.Now().Format("2006-01-02")),
		Title:           "Export Drafts to JSON",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to open save dialog: %v", err)}
	}
	if filePath == "" {
		// User cancelled
		return nil
	}

	// Get all draft sessions
	activeSessions, err := e.services.Storage.DraftRepo().GetActiveSessions(ctx)
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to get active drafts: %v", err)}
	}

	completedSessions, err := e.services.Storage.DraftRepo().GetCompletedSessions(ctx, 1000)
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to get completed drafts: %v", err)}
	}

	allSessions := append(activeSessions, completedSessions...)

	// Export to JSON using builder pattern
	err = export.NewExportBuilder().
		WithFormat(export.FormatJSON).
		WithFilePath(filePath).
		WithPrettyJSON(true).
		WithOverwrite(true).
		Export(allSessions)
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to export drafts to JSON: %v", err)}
	}

	log.Printf("Successfully exported %d draft sessions to %s", len(allSessions), filePath)
	return nil
}

// ExportDraftsToCSV exports draft session data to a CSV file.
// Uses the Builder pattern for flexible export configuration.
func (e *ExportFacade) ExportDraftsToCSV(ctx context.Context) error {
	if e.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	// Prompt user to select save location
	filePath, err := wailsruntime.SaveFileDialog(ctx, wailsruntime.SaveDialogOptions{
		DefaultFilename: fmt.Sprintf("mtga-drafts-%s.csv", time.Now().Format("2006-01-02")),
		Title:           "Export Drafts to CSV",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "CSV Files (*.csv)", Pattern: "*.csv"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to open save dialog: %v", err)}
	}
	if filePath == "" {
		// User cancelled
		return nil
	}

	// Get all draft sessions
	activeSessions, err := e.services.Storage.DraftRepo().GetActiveSessions(ctx)
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to get active drafts: %v", err)}
	}

	completedSessions, err := e.services.Storage.DraftRepo().GetCompletedSessions(ctx, 1000)
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to get completed drafts: %v", err)}
	}

	allSessions := append(activeSessions, completedSessions...)

	// Export to CSV using builder pattern
	err = export.NewExportBuilder().
		WithFormat(export.FormatCSV).
		WithFilePath(filePath).
		WithOverwrite(true).
		Export(allSessions)
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to export drafts to CSV: %v", err)}
	}

	log.Printf("Successfully exported %d draft sessions to %s", len(allSessions), filePath)
	return nil
}

// ImportLogFile imports historical MTGA log file data into the database.
// This allows users to import log files from backups, shared files, or pre-daemon installation.
func (e *ExportFacade) ImportLogFile(ctx context.Context) (*ImportLogFileResult, error) {
	if e.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Prompt user to select log file
	filePath, err := wailsruntime.OpenFileDialog(ctx, wailsruntime.OpenDialogOptions{
		Title: "Import MTGA Log File",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "MTGA Log Files (*.log)", Pattern: "*.log"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to open file dialog: %v", err)}
	}
	if filePath == "" {
		// User cancelled
		return nil, nil
	}

	log.Printf("Importing log file: %s", filePath)

	// Read log file
	reader, err := logreader.NewReader(filePath)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to open log file: %v", err)}
	}
	defer func() {
		_ = reader.Close() // Ignore close error (read-only file)
	}()

	// Read all entries
	entries, err := reader.ReadAll()
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to read log file: %v", err)}
	}

	if len(entries) == 0 {
		return nil, &AppError{Message: "Log file contains no entries"}
	}

	log.Printf("Read %d entries from log file", len(entries))

	// Process entries using log processor
	logProcessor := logprocessor.NewService(e.services.Storage)
	result, err := logProcessor.ProcessLogEntries(ctx, entries)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to process log entries: %v", err)}
	}

	// Extract file name from path
	fileName := filepath.Base(filePath)

	log.Printf("Successfully imported %s: %d entries, %d matches, %d decks, %d quests, %d drafts",
		fileName, len(entries), result.MatchesStored, result.DecksStored, result.QuestsStored, result.DraftsStored)

	// Dispatch events for data refresh
	if result.MatchesStored > 0 || result.GamesStored > 0 {
		e.eventDispatcher.Dispatch(events.Event{
			Type: "stats:updated",
			Data: map[string]interface{}{
				"matches": result.MatchesStored,
				"games":   result.GamesStored,
			},
			Context: ctx,
		})
	}

	if result.DecksStored > 0 {
		e.eventDispatcher.Dispatch(events.Event{
			Type: "deck:updated",
			Data: map[string]interface{}{
				"count": result.DecksStored,
			},
			Context: ctx,
		})
	}

	if result.RanksStored > 0 {
		e.eventDispatcher.Dispatch(events.Event{
			Type: "rank:updated",
			Data: map[string]interface{}{
				"count": result.RanksStored,
			},
			Context: ctx,
		})
	}

	if result.QuestsStored > 0 {
		e.eventDispatcher.Dispatch(events.Event{
			Type: "quest:updated",
			Data: map[string]interface{}{
				"count": result.QuestsStored,
			},
			Context: ctx,
		})
	}

	if result.DraftsStored > 0 {
		e.eventDispatcher.Dispatch(events.Event{
			Type: "draft:updated",
			Data: map[string]interface{}{
				"count": result.DraftsStored,
				"picks": result.DraftPicksStored,
			},
			Context: ctx,
		})
	}

	if result.CollectionCardsAdded > 0 {
		e.eventDispatcher.Dispatch(events.Event{
			Type: "collection:updated",
			Data: map[string]interface{}{
				"newCards":   result.CollectionNewCards,
				"cardsAdded": result.CollectionCardsAdded,
			},
			Context: ctx,
		})
	}

	return &ImportLogFileResult{
		FileName:      fileName,
		EntriesRead:   len(entries),
		MatchesStored: result.MatchesStored,
		GamesStored:   result.GamesStored,
		DecksStored:   result.DecksStored,
		RanksStored:   result.RanksStored,
		QuestsStored:  result.QuestsStored,
		DraftsStored:  result.DraftsStored,
		PicksStored:   result.DraftPicksStored,
	}, nil
}

// REST API friendly methods (no dialogs)

// GetMatchesExportData returns matches data for export without dialog.
func (e *ExportFacade) GetMatchesExportData(ctx context.Context, filter *models.StatsFilter) ([]*models.Match, error) {
	if e.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	var statsFilter models.StatsFilter
	if filter != nil {
		statsFilter = *filter
	}

	matches, err := e.services.Storage.GetMatches(ctx, statsFilter)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get matches: %v", err)}
	}

	return matches, nil
}

// GetDraftsExportData returns drafts data for export without dialog.
func (e *ExportFacade) GetDraftsExportData(ctx context.Context, limit int) ([]*models.DraftSession, error) {
	if e.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	activeSessions, err := e.services.Storage.DraftRepo().GetActiveSessions(ctx)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get active drafts: %v", err)}
	}

	completedSessions, err := e.services.Storage.DraftRepo().GetCompletedSessions(ctx, limit)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get completed drafts: %v", err)}
	}

	return append(activeSessions, completedSessions...), nil
}

// CollectionExportEntry represents a card in the collection for export.
type CollectionExportEntry struct {
	CardID   int `json:"cardId"`
	Quantity int `json:"quantity"`
}

// GetCollectionExportData returns collection data for export without dialog.
func (e *ExportFacade) GetCollectionExportData(ctx context.Context) ([]*CollectionExportEntry, error) {
	if e.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	collectionMap, err := e.services.Storage.CollectionRepo().GetAll(ctx)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get collection: %v", err)}
	}

	// Convert map to slice
	collection := make([]*CollectionExportEntry, 0, len(collectionMap))
	for cardID, quantity := range collectionMap {
		collection = append(collection, &CollectionExportEntry{
			CardID:   cardID,
			Quantity: quantity,
		})
	}

	return collection, nil
}

// DeckExportResult represents the result of exporting a deck.
type DeckExportResult struct {
	DeckID   string `json:"deckID"`
	DeckName string `json:"deckName"`
	Format   string `json:"format"`
	Content  string `json:"content"`
}

// GetDeckExportData returns deck data in the requested format without dialog.
func (e *ExportFacade) GetDeckExportData(ctx context.Context, deckID, format string) (*DeckExportResult, error) {
	if e.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	deck, err := e.services.Storage.DeckRepo().GetByID(ctx, deckID)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get deck: %v", err)}
	}

	if deck == nil {
		return nil, &AppError{Message: "Deck not found"}
	}

	// Export to requested format
	var content string
	switch format {
	case "mtga", "arena":
		content = export.ExportToMTGAFormat(deck)
	case "text":
		content = export.ExportToTextFormat(deck)
	case "json":
		jsonBytes, err := json.MarshalIndent(deck, "", "  ")
		if err != nil {
			return nil, &AppError{Message: fmt.Sprintf("Failed to export to JSON: %v", err)}
		}
		content = string(jsonBytes)
	default:
		content = export.ExportToMTGAFormat(deck)
	}

	return &DeckExportResult{
		DeckID:   deckID,
		DeckName: deck.Name,
		Format:   format,
		Content:  content,
	}, nil
}

// ImportResult represents the result of an import operation.
type ImportResult struct {
	Imported int    `json:"imported"`
	Skipped  int    `json:"skipped"`
	Total    int    `json:"total"`
	Message  string `json:"message"`
}

// ImportMatchesData imports matches from data without dialog.
func (e *ExportFacade) ImportMatchesData(ctx context.Context, matches []*models.Match) (*ImportResult, error) {
	if e.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	imported := 0
	skipped := 0
	for _, match := range matches {
		if err := e.services.Storage.SaveMatch(ctx, match); err != nil {
			log.Printf("Warning: Failed to import match %s: %v", match.ID, err)
			skipped++
			continue
		}
		imported++
	}

	return &ImportResult{
		Imported: imported,
		Skipped:  skipped,
		Total:    len(matches),
		Message:  fmt.Sprintf("Imported %d/%d matches", imported, len(matches)),
	}, nil
}

// ClearAllDataWithoutDialog clears all data without confirmation dialog.
func (e *ExportFacade) ClearAllDataWithoutDialog(ctx context.Context) error {
	if e.services == nil || e.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	if err := e.services.Storage.ClearAllMatches(ctx); err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to clear data: %v", err)}
	}

	log.Println("Successfully cleared all match history")
	return nil
}

// ImportLogFileData imports log file from base64 content without dialog.
func (e *ExportFacade) ImportLogFileData(ctx context.Context, base64Content, fileName string) (*ImportLogFileResult, error) {
	if e.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Decode base64 content
	content, err := base64Decode(base64Content)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to decode content: %v", err)}
	}

	// Create temp file for log reader
	tmpFile, err := os.CreateTemp("", "mtga-log-*.log")
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to create temp file: %v", err)}
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmpFile.Write(content); err != nil {
		_ = tmpFile.Close()
		return nil, &AppError{Message: fmt.Sprintf("Failed to write temp file: %v", err)}
	}
	_ = tmpFile.Close()

	// Read log file
	reader, err := logreader.NewReader(tmpPath)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to open log file: %v", err)}
	}
	defer func() {
		_ = reader.Close()
	}()

	entries, err := reader.ReadAll()
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to read log file: %v", err)}
	}

	if len(entries) == 0 {
		return nil, &AppError{Message: "Log file contains no entries"}
	}

	// Process entries
	logProcessor := logprocessor.NewService(e.services.Storage)
	result, err := logProcessor.ProcessLogEntries(ctx, entries)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to process log entries: %v", err)}
	}

	if fileName == "" {
		fileName = "imported-log.log"
	}

	return &ImportLogFileResult{
		FileName:      fileName,
		EntriesRead:   len(entries),
		MatchesStored: result.MatchesStored,
		GamesStored:   result.GamesStored,
		DecksStored:   result.DecksStored,
		RanksStored:   result.RanksStored,
		QuestsStored:  result.QuestsStored,
		DraftsStored:  result.DraftsStored,
		PicksStored:   result.DraftPicksStored,
	}, nil
}

// base64Decode decodes a base64 string.
func base64Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}
