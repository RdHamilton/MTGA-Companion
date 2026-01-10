package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

func setupEDHRECTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	// Create a temporary file for the test database
	tmpFile, err := os.CreateTemp("", "edhrec_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	db, err := sql.Open("sqlite3", tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS edhrec_synergy (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			card_name TEXT NOT NULL,
			synergy_card_name TEXT NOT NULL,
			synergy_score REAL NOT NULL,
			inclusion_count INTEGER DEFAULT 0,
			num_decks INTEGER DEFAULT 0,
			lift REAL DEFAULT 0.0,
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(card_name, synergy_card_name)
		);
		CREATE INDEX IF NOT EXISTS idx_edhrec_synergy_card ON edhrec_synergy(card_name);

		CREATE TABLE IF NOT EXISTS edhrec_card_metadata (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			card_name TEXT NOT NULL UNIQUE,
			sanitized_name TEXT NOT NULL,
			num_decks INTEGER DEFAULT 0,
			salt_score REAL DEFAULT 0.0,
			color_identity TEXT,
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_edhrec_metadata_name ON edhrec_card_metadata(card_name);

		CREATE TABLE IF NOT EXISTS edhrec_theme_cards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			theme_name TEXT NOT NULL,
			card_name TEXT NOT NULL,
			synergy_score REAL DEFAULT 0.0,
			is_top_card BOOLEAN DEFAULT FALSE,
			is_high_synergy BOOLEAN DEFAULT FALSE,
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(theme_name, card_name)
		);
		CREATE INDEX IF NOT EXISTS idx_edhrec_theme_cards_theme ON edhrec_theme_cards(theme_name);
		CREATE INDEX IF NOT EXISTS idx_edhrec_theme_cards_card ON edhrec_theme_cards(card_name);
	`)
	if err != nil {
		db.Close()
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to create tables: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}

	return db, cleanup
}

func TestEDHRECRepository_UpsertSynergy(t *testing.T) {
	db, cleanup := setupEDHRECTestDB(t)
	defer cleanup()

	repo := NewEDHRECRepository(db)
	ctx := context.Background()

	synergy := &models.EDHRECSynergy{
		CardName:        "Sol Ring",
		SynergyCardName: "Arcane Signet",
		SynergyScore:    0.8,
		InclusionCount:  90,
		NumDecks:        50000,
		Lift:            1.5,
	}

	err := repo.UpsertSynergy(ctx, synergy)
	if err != nil {
		t.Fatalf("UpsertSynergy failed: %v", err)
	}

	// Verify the insert
	synergies, err := repo.GetSynergiesForCard(ctx, "Sol Ring", 10)
	if err != nil {
		t.Fatalf("GetSynergiesForCard failed: %v", err)
	}

	if len(synergies) != 1 {
		t.Fatalf("Expected 1 synergy, got %d", len(synergies))
	}

	if synergies[0].SynergyCardName != "Arcane Signet" {
		t.Errorf("SynergyCardName = %q, want %q", synergies[0].SynergyCardName, "Arcane Signet")
	}

	// Test update
	synergy.SynergyScore = 0.9
	err = repo.UpsertSynergy(ctx, synergy)
	if err != nil {
		t.Fatalf("UpsertSynergy (update) failed: %v", err)
	}

	synergies, err = repo.GetSynergiesForCard(ctx, "Sol Ring", 10)
	if err != nil {
		t.Fatalf("GetSynergiesForCard after update failed: %v", err)
	}

	if synergies[0].SynergyScore != 0.9 {
		t.Errorf("SynergyScore = %f, want %f", synergies[0].SynergyScore, 0.9)
	}
}

func TestEDHRECRepository_BulkUpsertSynergies(t *testing.T) {
	db, cleanup := setupEDHRECTestDB(t)
	defer cleanup()

	repo := NewEDHRECRepository(db)
	ctx := context.Background()

	synergies := []*models.EDHRECSynergy{
		{CardName: "Sol Ring", SynergyCardName: "Arcane Signet", SynergyScore: 0.8},
		{CardName: "Sol Ring", SynergyCardName: "Command Tower", SynergyScore: 0.7},
		{CardName: "Sol Ring", SynergyCardName: "Mana Crypt", SynergyScore: 0.9},
	}

	err := repo.BulkUpsertSynergies(ctx, synergies)
	if err != nil {
		t.Fatalf("BulkUpsertSynergies failed: %v", err)
	}

	result, err := repo.GetSynergiesForCard(ctx, "Sol Ring", 10)
	if err != nil {
		t.Fatalf("GetSynergiesForCard failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 synergies, got %d", len(result))
	}
}

func TestEDHRECRepository_GetSynergyScore(t *testing.T) {
	db, cleanup := setupEDHRECTestDB(t)
	defer cleanup()

	repo := NewEDHRECRepository(db)
	ctx := context.Background()

	// Insert test data
	synergy := &models.EDHRECSynergy{
		CardName:        "Sol Ring",
		SynergyCardName: "Arcane Signet",
		SynergyScore:    0.8,
	}
	err := repo.UpsertSynergy(ctx, synergy)
	if err != nil {
		t.Fatalf("UpsertSynergy failed: %v", err)
	}

	// Test forward lookup
	score, err := repo.GetSynergyScore(ctx, "Sol Ring", "Arcane Signet")
	if err != nil {
		t.Fatalf("GetSynergyScore failed: %v", err)
	}
	if score != 0.8 {
		t.Errorf("Score = %f, want %f", score, 0.8)
	}

	// Test reverse lookup (should also work)
	score, err = repo.GetSynergyScore(ctx, "Arcane Signet", "Sol Ring")
	if err != nil {
		t.Fatalf("GetSynergyScore (reverse) failed: %v", err)
	}
	if score != 0.8 {
		t.Errorf("Score (reverse) = %f, want %f", score, 0.8)
	}

	// Test non-existent pair
	score, err = repo.GetSynergyScore(ctx, "Sol Ring", "Lightning Bolt")
	if err != nil {
		t.Fatalf("GetSynergyScore (non-existent) failed: %v", err)
	}
	if score != 0 {
		t.Errorf("Score (non-existent) = %f, want %f", score, 0.0)
	}
}

func TestEDHRECRepository_UpsertMetadata(t *testing.T) {
	db, cleanup := setupEDHRECTestDB(t)
	defer cleanup()

	repo := NewEDHRECRepository(db)
	ctx := context.Background()

	metadata := &models.EDHRECCardMetadata{
		CardName:      "Sol Ring",
		SanitizedName: "sol-ring",
		NumDecks:      100000,
		SaltScore:     0.5,
		ColorIdentity: "",
	}

	err := repo.UpsertMetadata(ctx, metadata)
	if err != nil {
		t.Fatalf("UpsertMetadata failed: %v", err)
	}

	result, err := repo.GetMetadata(ctx, "Sol Ring")
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected metadata, got nil")
	}
	if result.CardName != "Sol Ring" {
		t.Errorf("CardName = %q, want %q", result.CardName, "Sol Ring")
	}
	if result.NumDecks != 100000 {
		t.Errorf("NumDecks = %d, want %d", result.NumDecks, 100000)
	}
}

func TestEDHRECRepository_GetMetadata_CaseInsensitive(t *testing.T) {
	db, cleanup := setupEDHRECTestDB(t)
	defer cleanup()

	repo := NewEDHRECRepository(db)
	ctx := context.Background()

	metadata := &models.EDHRECCardMetadata{
		CardName:      "Sol Ring",
		SanitizedName: "sol-ring",
		NumDecks:      100000,
	}

	err := repo.UpsertMetadata(ctx, metadata)
	if err != nil {
		t.Fatalf("UpsertMetadata failed: %v", err)
	}

	// Test case-insensitive lookup
	result, err := repo.GetMetadata(ctx, "sol ring")
	if err != nil {
		t.Fatalf("GetMetadata (lowercase) failed: %v", err)
	}
	if result == nil {
		t.Error("Expected metadata, got nil for lowercase lookup")
	}

	result, err = repo.GetMetadata(ctx, "SOL RING")
	if err != nil {
		t.Fatalf("GetMetadata (uppercase) failed: %v", err)
	}
	if result == nil {
		t.Error("Expected metadata, got nil for uppercase lookup")
	}
}

func TestEDHRECRepository_ThemeCards(t *testing.T) {
	db, cleanup := setupEDHRECTestDB(t)
	defer cleanup()

	repo := NewEDHRECRepository(db)
	ctx := context.Background()

	themeCards := []*models.EDHRECThemeCard{
		{ThemeName: "tokens", CardName: "Smothering Tithe", SynergyScore: 0.9, IsTopCard: true, IsHighSynergy: true},
		{ThemeName: "tokens", CardName: "Anointed Procession", SynergyScore: 0.85, IsTopCard: true, IsHighSynergy: true},
		{ThemeName: "tokens", CardName: "Doubling Season", SynergyScore: 0.8, IsTopCard: false, IsHighSynergy: true},
		{ThemeName: "aristocrats", CardName: "Smothering Tithe", SynergyScore: 0.6, IsTopCard: false, IsHighSynergy: false},
	}

	err := repo.BulkUpsertThemeCards(ctx, themeCards)
	if err != nil {
		t.Fatalf("BulkUpsertThemeCards failed: %v", err)
	}

	// Test GetCardsForTheme
	cards, err := repo.GetCardsForTheme(ctx, "tokens", 10)
	if err != nil {
		t.Fatalf("GetCardsForTheme failed: %v", err)
	}
	if len(cards) != 3 {
		t.Errorf("Expected 3 cards for tokens theme, got %d", len(cards))
	}

	// Test GetThemesForCard
	themes, err := repo.GetThemesForCard(ctx, "Smothering Tithe")
	if err != nil {
		t.Fatalf("GetThemesForCard failed: %v", err)
	}
	if len(themes) != 2 {
		t.Errorf("Expected 2 themes for Smothering Tithe, got %d", len(themes))
	}
}

func TestEDHRECRepository_DeleteSynergiesForCard(t *testing.T) {
	db, cleanup := setupEDHRECTestDB(t)
	defer cleanup()

	repo := NewEDHRECRepository(db)
	ctx := context.Background()

	synergies := []*models.EDHRECSynergy{
		{CardName: "Sol Ring", SynergyCardName: "Arcane Signet", SynergyScore: 0.8},
		{CardName: "Sol Ring", SynergyCardName: "Command Tower", SynergyScore: 0.7},
		{CardName: "Mana Crypt", SynergyCardName: "Arcane Signet", SynergyScore: 0.9},
	}

	err := repo.BulkUpsertSynergies(ctx, synergies)
	if err != nil {
		t.Fatalf("BulkUpsertSynergies failed: %v", err)
	}

	// Delete synergies for Sol Ring
	err = repo.DeleteSynergiesForCard(ctx, "Sol Ring")
	if err != nil {
		t.Fatalf("DeleteSynergiesForCard failed: %v", err)
	}

	// Verify Sol Ring synergies are deleted
	result, err := repo.GetSynergiesForCard(ctx, "Sol Ring", 10)
	if err != nil {
		t.Fatalf("GetSynergiesForCard failed: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected 0 synergies for Sol Ring after delete, got %d", len(result))
	}

	// Verify Mana Crypt synergies still exist
	result, err = repo.GetSynergiesForCard(ctx, "Mana Crypt", 10)
	if err != nil {
		t.Fatalf("GetSynergiesForCard (Mana Crypt) failed: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 synergy for Mana Crypt, got %d", len(result))
	}
}

func TestEDHRECRepository_GetSynergyCount(t *testing.T) {
	db, cleanup := setupEDHRECTestDB(t)
	defer cleanup()

	repo := NewEDHRECRepository(db)
	ctx := context.Background()

	// Initially empty
	count, err := repo.GetSynergyCount(ctx)
	if err != nil {
		t.Fatalf("GetSynergyCount failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 synergies, got %d", count)
	}

	// Add some synergies
	synergies := []*models.EDHRECSynergy{
		{CardName: "Sol Ring", SynergyCardName: "Arcane Signet", SynergyScore: 0.8},
		{CardName: "Sol Ring", SynergyCardName: "Command Tower", SynergyScore: 0.7},
	}
	err = repo.BulkUpsertSynergies(ctx, synergies)
	if err != nil {
		t.Fatalf("BulkUpsertSynergies failed: %v", err)
	}

	count, err = repo.GetSynergyCount(ctx)
	if err != nil {
		t.Fatalf("GetSynergyCount failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 synergies, got %d", count)
	}
}

func TestEDHRECRepository_ClearAll(t *testing.T) {
	db, cleanup := setupEDHRECTestDB(t)
	defer cleanup()

	repo := NewEDHRECRepository(db)
	ctx := context.Background()

	// Add data to all tables
	err := repo.UpsertSynergy(ctx, &models.EDHRECSynergy{
		CardName:        "Sol Ring",
		SynergyCardName: "Arcane Signet",
		SynergyScore:    0.8,
	})
	if err != nil {
		t.Fatalf("UpsertSynergy failed: %v", err)
	}

	err = repo.UpsertMetadata(ctx, &models.EDHRECCardMetadata{
		CardName:      "Sol Ring",
		SanitizedName: "sol-ring",
		NumDecks:      100000,
	})
	if err != nil {
		t.Fatalf("UpsertMetadata failed: %v", err)
	}

	err = repo.UpsertThemeCard(ctx, &models.EDHRECThemeCard{
		ThemeName:    "tokens",
		CardName:     "Sol Ring",
		SynergyScore: 0.5,
	})
	if err != nil {
		t.Fatalf("UpsertThemeCard failed: %v", err)
	}

	// Clear all
	err = repo.ClearAll(ctx)
	if err != nil {
		t.Fatalf("ClearAll failed: %v", err)
	}

	// Verify all cleared
	synergyCount, _ := repo.GetSynergyCount(ctx)
	if synergyCount != 0 {
		t.Errorf("Expected 0 synergies after clear, got %d", synergyCount)
	}

	metadataCount, _ := repo.GetMetadataCount(ctx)
	if metadataCount != 0 {
		t.Errorf("Expected 0 metadata after clear, got %d", metadataCount)
	}
}

func TestGetMatchingThemes(t *testing.T) {
	tests := []struct {
		name           string
		cardText       string
		cardType       string
		expectedThemes []string
	}{
		{
			name:           "token creation",
			cardText:       "Create a 1/1 white Soldier creature token.",
			cardType:       "Sorcery",
			expectedThemes: []string{"tokens"},
		},
		{
			name:           "sacrifice theme",
			cardText:       "Sacrifice a creature: Draw a card.",
			cardType:       "Enchantment",
			expectedThemes: []string{"sacrifice"},
		},
		{
			name:           "counters theme",
			cardText:       "Put a +1/+1 counter on target creature.",
			cardType:       "Instant",
			expectedThemes: []string{"counters"},
		},
		{
			name:           "lifegain",
			cardText:       "Gain 3 life.",
			cardType:       "Instant",
			expectedThemes: []string{"lifegain"},
		},
		{
			name:           "artifact type",
			cardText:       "Tap: Add one mana of any color.",
			cardType:       "Artifact",
			expectedThemes: []string{"artifacts"},
		},
		{
			name:           "graveyard recursion",
			cardText:       "Return target creature card from your graveyard to the battlefield.",
			cardType:       "Sorcery",
			expectedThemes: []string{"reanimator", "graveyard"},
		},
		{
			name:           "multiple themes",
			cardText:       "Sacrifice a creature: Create two 1/1 Saproling creature tokens.",
			cardType:       "Enchantment",
			expectedThemes: []string{"tokens", "sacrifice"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			themes := GetMatchingThemes(tt.cardText, tt.cardType)

			for _, expected := range tt.expectedThemes {
				found := false
				for _, theme := range themes {
					if theme == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected theme %q not found in result %v", expected, themes)
				}
			}
		})
	}
}

func TestEDHRECSynergy_Timestamp(t *testing.T) {
	db, cleanup := setupEDHRECTestDB(t)
	defer cleanup()

	repo := NewEDHRECRepository(db)
	ctx := context.Background()

	before := time.Now()

	err := repo.UpsertSynergy(ctx, &models.EDHRECSynergy{
		CardName:        "Sol Ring",
		SynergyCardName: "Arcane Signet",
		SynergyScore:    0.8,
	})
	if err != nil {
		t.Fatalf("UpsertSynergy failed: %v", err)
	}

	after := time.Now()

	synergies, err := repo.GetSynergiesForCard(ctx, "Sol Ring", 10)
	if err != nil {
		t.Fatalf("GetSynergiesForCard failed: %v", err)
	}

	if len(synergies) != 1 {
		t.Fatalf("Expected 1 synergy, got %d", len(synergies))
	}

	// Check that LastUpdated is set and within expected range
	lastUpdated := synergies[0].LastUpdated
	if lastUpdated.Before(before.Add(-time.Second)) || lastUpdated.After(after.Add(time.Second)) {
		t.Errorf("LastUpdated %v is outside expected range [%v, %v]", lastUpdated, before, after)
	}
}
