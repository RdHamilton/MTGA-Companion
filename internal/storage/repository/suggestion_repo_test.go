package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	_ "modernc.org/sqlite"
)

func setupSuggestionTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create required tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS decks (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS improvement_suggestions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			deck_id TEXT NOT NULL,
			suggestion_type TEXT NOT NULL,
			priority TEXT DEFAULT 'medium',
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			evidence TEXT,
			card_references TEXT,
			is_dismissed BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (deck_id) REFERENCES decks(id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Insert test deck
	_, err = db.Exec(`INSERT INTO decks (id, name) VALUES ('deck-1', 'Test Deck')`)
	if err != nil {
		t.Fatalf("Failed to insert test deck: %v", err)
	}

	return db
}

func TestSuggestionRepository_CreateSuggestion(t *testing.T) {
	db := setupSuggestionTestDB(t)
	defer db.Close()

	repo := NewSuggestionRepository(db)
	ctx := context.Background()

	suggestion := &models.ImprovementSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.SuggestionTypeCurve,
		Priority:       models.SuggestionPriorityHigh,
		Title:          "Curve Too High",
		Description:    "Consider adding more low-cost spells",
	}

	err := repo.CreateSuggestion(ctx, suggestion)
	if err != nil {
		t.Fatalf("Failed to create suggestion: %v", err)
	}

	if suggestion.ID == 0 {
		t.Error("Expected suggestion ID to be set after creation")
	}
}

func TestSuggestionRepository_GetSuggestionsByDeck(t *testing.T) {
	db := setupSuggestionTestDB(t)
	defer db.Close()

	repo := NewSuggestionRepository(db)
	ctx := context.Background()

	// Create multiple suggestions
	s1 := &models.ImprovementSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.SuggestionTypeCurve,
		Priority:       models.SuggestionPriorityHigh,
		Title:          "Curve Issue",
		Description:    "Description 1",
	}
	s2 := &models.ImprovementSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.SuggestionTypeMana,
		Priority:       models.SuggestionPriorityMedium,
		Title:          "Mana Issue",
		Description:    "Description 2",
	}

	_ = repo.CreateSuggestion(ctx, s1)
	_ = repo.CreateSuggestion(ctx, s2)

	// Get all suggestions
	suggestions, err := repo.GetSuggestionsByDeck(ctx, "deck-1")
	if err != nil {
		t.Fatalf("Failed to get suggestions: %v", err)
	}

	if len(suggestions) != 2 {
		t.Errorf("Expected 2 suggestions, got %d", len(suggestions))
	}
}

func TestSuggestionRepository_GetActiveSuggestions(t *testing.T) {
	db := setupSuggestionTestDB(t)
	defer db.Close()

	repo := NewSuggestionRepository(db)
	ctx := context.Background()

	// Create suggestions
	s1 := &models.ImprovementSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.SuggestionTypeCurve,
		Priority:       models.SuggestionPriorityHigh,
		Title:          "Active Suggestion",
		Description:    "Description 1",
	}
	s2 := &models.ImprovementSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.SuggestionTypeMana,
		Priority:       models.SuggestionPriorityMedium,
		Title:          "To Be Dismissed",
		Description:    "Description 2",
	}

	_ = repo.CreateSuggestion(ctx, s1)
	_ = repo.CreateSuggestion(ctx, s2)

	// Dismiss one
	_ = repo.DismissSuggestion(ctx, s2.ID)

	// Get active only
	suggestions, err := repo.GetActiveSuggestions(ctx, "deck-1")
	if err != nil {
		t.Fatalf("Failed to get active suggestions: %v", err)
	}

	if len(suggestions) != 1 {
		t.Errorf("Expected 1 active suggestion, got %d", len(suggestions))
	}

	if suggestions[0].Title != "Active Suggestion" {
		t.Errorf("Expected 'Active Suggestion', got '%s'", suggestions[0].Title)
	}
}

func TestSuggestionRepository_GetSuggestionByID(t *testing.T) {
	db := setupSuggestionTestDB(t)
	defer db.Close()

	repo := NewSuggestionRepository(db)
	ctx := context.Background()

	// Create a suggestion
	suggestion := &models.ImprovementSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.SuggestionTypeCurve,
		Priority:       models.SuggestionPriorityHigh,
		Title:          "Test Suggestion",
		Description:    "Test Description",
	}
	_ = repo.CreateSuggestion(ctx, suggestion)

	// Get by ID
	retrieved, err := repo.GetSuggestionByID(ctx, suggestion.ID)
	if err != nil {
		t.Fatalf("Failed to get suggestion by ID: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected suggestion, got nil")
	}

	if retrieved.Title != "Test Suggestion" {
		t.Errorf("Expected title 'Test Suggestion', got '%s'", retrieved.Title)
	}
}

func TestSuggestionRepository_GetSuggestionsByType(t *testing.T) {
	db := setupSuggestionTestDB(t)
	defer db.Close()

	repo := NewSuggestionRepository(db)
	ctx := context.Background()

	// Create suggestions of different types
	s1 := &models.ImprovementSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.SuggestionTypeCurve,
		Priority:       models.SuggestionPriorityHigh,
		Title:          "Curve 1",
		Description:    "Description 1",
	}
	s2 := &models.ImprovementSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.SuggestionTypeMana,
		Priority:       models.SuggestionPriorityMedium,
		Title:          "Mana 1",
		Description:    "Description 2",
	}
	s3 := &models.ImprovementSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.SuggestionTypeCurve,
		Priority:       models.SuggestionPriorityLow,
		Title:          "Curve 2",
		Description:    "Description 3",
	}

	_ = repo.CreateSuggestion(ctx, s1)
	_ = repo.CreateSuggestion(ctx, s2)
	_ = repo.CreateSuggestion(ctx, s3)

	// Get only curve suggestions
	suggestions, err := repo.GetSuggestionsByType(ctx, "deck-1", models.SuggestionTypeCurve)
	if err != nil {
		t.Fatalf("Failed to get suggestions by type: %v", err)
	}

	if len(suggestions) != 2 {
		t.Errorf("Expected 2 curve suggestions, got %d", len(suggestions))
	}

	for _, s := range suggestions {
		if s.SuggestionType != models.SuggestionTypeCurve {
			t.Errorf("Expected type 'curve', got '%s'", s.SuggestionType)
		}
	}
}

func TestSuggestionRepository_DismissSuggestion(t *testing.T) {
	db := setupSuggestionTestDB(t)
	defer db.Close()

	repo := NewSuggestionRepository(db)
	ctx := context.Background()

	// Create a suggestion
	suggestion := &models.ImprovementSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.SuggestionTypeCurve,
		Priority:       models.SuggestionPriorityHigh,
		Title:          "Test",
		Description:    "Description",
	}
	_ = repo.CreateSuggestion(ctx, suggestion)

	// Dismiss it
	err := repo.DismissSuggestion(ctx, suggestion.ID)
	if err != nil {
		t.Fatalf("Failed to dismiss suggestion: %v", err)
	}

	// Verify dismissed
	retrieved, _ := repo.GetSuggestionByID(ctx, suggestion.ID)
	if !retrieved.IsDismissed {
		t.Error("Expected suggestion to be dismissed")
	}
}

func TestSuggestionRepository_UndismissSuggestion(t *testing.T) {
	db := setupSuggestionTestDB(t)
	defer db.Close()

	repo := NewSuggestionRepository(db)
	ctx := context.Background()

	// Create a suggestion
	suggestion := &models.ImprovementSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.SuggestionTypeCurve,
		Priority:       models.SuggestionPriorityHigh,
		Title:          "Test",
		Description:    "Description",
	}
	_ = repo.CreateSuggestion(ctx, suggestion)

	// Dismiss it
	_ = repo.DismissSuggestion(ctx, suggestion.ID)

	// Undismiss it
	err := repo.UndismissSuggestion(ctx, suggestion.ID)
	if err != nil {
		t.Fatalf("Failed to undismiss suggestion: %v", err)
	}

	// Verify not dismissed
	retrieved, _ := repo.GetSuggestionByID(ctx, suggestion.ID)
	if retrieved.IsDismissed {
		t.Error("Expected suggestion to not be dismissed")
	}
}

func TestSuggestionRepository_DeleteSuggestion(t *testing.T) {
	db := setupSuggestionTestDB(t)
	defer db.Close()

	repo := NewSuggestionRepository(db)
	ctx := context.Background()

	// Create a suggestion
	suggestion := &models.ImprovementSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.SuggestionTypeCurve,
		Priority:       models.SuggestionPriorityHigh,
		Title:          "Test",
		Description:    "Description",
	}
	_ = repo.CreateSuggestion(ctx, suggestion)

	// Delete it
	err := repo.DeleteSuggestion(ctx, suggestion.ID)
	if err != nil {
		t.Fatalf("Failed to delete suggestion: %v", err)
	}

	// Verify deleted
	_, err = repo.GetSuggestionByID(ctx, suggestion.ID)
	if err == nil {
		t.Error("Expected error when getting deleted suggestion")
	}
}

func TestSuggestionRepository_DeleteSuggestionsByDeck(t *testing.T) {
	db := setupSuggestionTestDB(t)
	defer db.Close()

	repo := NewSuggestionRepository(db)
	ctx := context.Background()

	// Create multiple suggestions
	s1 := &models.ImprovementSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.SuggestionTypeCurve,
		Priority:       models.SuggestionPriorityHigh,
		Title:          "Test 1",
		Description:    "Description 1",
	}
	s2 := &models.ImprovementSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.SuggestionTypeMana,
		Priority:       models.SuggestionPriorityMedium,
		Title:          "Test 2",
		Description:    "Description 2",
	}

	_ = repo.CreateSuggestion(ctx, s1)
	_ = repo.CreateSuggestion(ctx, s2)

	// Delete all for deck
	err := repo.DeleteSuggestionsByDeck(ctx, "deck-1")
	if err != nil {
		t.Fatalf("Failed to delete suggestions by deck: %v", err)
	}

	// Verify deletion
	suggestions, err := repo.GetSuggestionsByDeck(ctx, "deck-1")
	if err != nil {
		t.Fatalf("Failed to get suggestions: %v", err)
	}

	if len(suggestions) != 0 {
		t.Errorf("Expected 0 suggestions after deletion, got %d", len(suggestions))
	}
}

func TestSuggestionRepository_DeleteActiveSuggestionsByDeck(t *testing.T) {
	db := setupSuggestionTestDB(t)
	defer db.Close()

	repo := NewSuggestionRepository(db)
	ctx := context.Background()

	// Create suggestions
	s1 := &models.ImprovementSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.SuggestionTypeCurve,
		Priority:       models.SuggestionPriorityHigh,
		Title:          "Active",
		Description:    "Description 1",
	}
	s2 := &models.ImprovementSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.SuggestionTypeMana,
		Priority:       models.SuggestionPriorityMedium,
		Title:          "Dismissed",
		Description:    "Description 2",
	}

	_ = repo.CreateSuggestion(ctx, s1)
	_ = repo.CreateSuggestion(ctx, s2)

	// Dismiss one
	_ = repo.DismissSuggestion(ctx, s2.ID)

	// Delete active only
	err := repo.DeleteActiveSuggestionsByDeck(ctx, "deck-1")
	if err != nil {
		t.Fatalf("Failed to delete active suggestions: %v", err)
	}

	// Verify: dismissed should remain
	suggestions, err := repo.GetSuggestionsByDeck(ctx, "deck-1")
	if err != nil {
		t.Fatalf("Failed to get suggestions: %v", err)
	}

	if len(suggestions) != 1 {
		t.Errorf("Expected 1 suggestion (dismissed), got %d", len(suggestions))
	}

	if !suggestions[0].IsDismissed {
		t.Error("Expected remaining suggestion to be dismissed")
	}
}

func TestSuggestionRepository_PriorityOrdering(t *testing.T) {
	db := setupSuggestionTestDB(t)
	defer db.Close()

	repo := NewSuggestionRepository(db)
	ctx := context.Background()

	// Create suggestions with different priorities (in reverse order)
	s1 := &models.ImprovementSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.SuggestionTypeCurve,
		Priority:       models.SuggestionPriorityLow,
		Title:          "Low",
		Description:    "Description 1",
	}
	s2 := &models.ImprovementSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.SuggestionTypeMana,
		Priority:       models.SuggestionPriorityHigh,
		Title:          "High",
		Description:    "Description 2",
	}
	s3 := &models.ImprovementSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.SuggestionTypeSequencing,
		Priority:       models.SuggestionPriorityMedium,
		Title:          "Medium",
		Description:    "Description 3",
	}

	_ = repo.CreateSuggestion(ctx, s1)
	_ = repo.CreateSuggestion(ctx, s2)
	_ = repo.CreateSuggestion(ctx, s3)

	// Get suggestions (should be ordered by priority)
	suggestions, err := repo.GetSuggestionsByDeck(ctx, "deck-1")
	if err != nil {
		t.Fatalf("Failed to get suggestions: %v", err)
	}

	if len(suggestions) != 3 {
		t.Fatalf("Expected 3 suggestions, got %d", len(suggestions))
	}

	// Check order: high, medium, low
	if suggestions[0].Priority != models.SuggestionPriorityHigh {
		t.Errorf("Expected first priority to be 'high', got '%s'", suggestions[0].Priority)
	}
	if suggestions[1].Priority != models.SuggestionPriorityMedium {
		t.Errorf("Expected second priority to be 'medium', got '%s'", suggestions[1].Priority)
	}
	if suggestions[2].Priority != models.SuggestionPriorityLow {
		t.Errorf("Expected third priority to be 'low', got '%s'", suggestions[2].Priority)
	}
}
