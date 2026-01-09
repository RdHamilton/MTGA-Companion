package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	_ "modernc.org/sqlite"
)

func setupMLSuggestionTestDB(t *testing.T) *sql.DB {
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

		CREATE TABLE IF NOT EXISTS card_combination_stats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			card_id_1 INTEGER NOT NULL,
			card_id_2 INTEGER NOT NULL,
			deck_id TEXT,
			format TEXT DEFAULT 'Standard',
			games_together INTEGER DEFAULT 0,
			games_card1_only INTEGER DEFAULT 0,
			games_card2_only INTEGER DEFAULT 0,
			wins_together INTEGER DEFAULT 0,
			wins_card1_only INTEGER DEFAULT 0,
			wins_card2_only INTEGER DEFAULT 0,
			synergy_score REAL DEFAULT 0.0,
			confidence_score REAL DEFAULT 0.0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(card_id_1, card_id_2, deck_id, format),
			CHECK(card_id_1 < card_id_2)
		);

		CREATE TABLE IF NOT EXISTS ml_suggestions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			deck_id TEXT NOT NULL,
			suggestion_type TEXT NOT NULL,
			card_id INTEGER,
			card_name TEXT,
			swap_for_card_id INTEGER,
			swap_for_card_name TEXT,
			confidence REAL DEFAULT 0.0,
			expected_win_rate_change REAL DEFAULT 0.0,
			title TEXT NOT NULL,
			description TEXT,
			reasoning TEXT,
			evidence TEXT,
			is_dismissed BOOLEAN DEFAULT FALSE,
			was_applied BOOLEAN DEFAULT FALSE,
			outcome_win_rate_change REAL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			applied_at TIMESTAMP,
			outcome_recorded_at TIMESTAMP,
			FOREIGN KEY (deck_id) REFERENCES decks(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS user_play_patterns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			account_id TEXT NOT NULL,
			preferred_archetype TEXT,
			aggro_affinity REAL DEFAULT 0.0,
			midrange_affinity REAL DEFAULT 0.0,
			control_affinity REAL DEFAULT 0.0,
			combo_affinity REAL DEFAULT 0.0,
			color_preferences TEXT,
			avg_game_length REAL DEFAULT 0.0,
			aggression_score REAL DEFAULT 0.0,
			interaction_score REAL DEFAULT 0.0,
			total_matches INTEGER DEFAULT 0,
			total_decks INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(account_id)
		);

		CREATE TABLE IF NOT EXISTS ml_model_metadata (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			model_name TEXT NOT NULL,
			model_version TEXT NOT NULL,
			training_samples INTEGER DEFAULT 0,
			training_date TIMESTAMP,
			accuracy REAL,
			precision_score REAL,
			recall REAL,
			f1_score REAL,
			is_active BOOLEAN DEFAULT FALSE,
			model_data BLOB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(model_name, model_version)
		);

		CREATE TABLE IF NOT EXISTS card_affinity (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			card_id_1 INTEGER NOT NULL,
			card_id_2 INTEGER NOT NULL,
			format TEXT DEFAULT 'Standard',
			affinity_score REAL DEFAULT 0.0,
			sample_size INTEGER DEFAULT 0,
			confidence REAL DEFAULT 0.0,
			source TEXT DEFAULT 'historical',
			computed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(card_id_1, card_id_2, format),
			CHECK(card_id_1 < card_id_2)
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

func TestMLSuggestionRepository_CreateSuggestion(t *testing.T) {
	db := setupMLSuggestionTestDB(t)
	defer db.Close()

	repo := NewMLSuggestionRepository(db)
	ctx := context.Background()

	suggestion := &models.MLSuggestion{
		DeckID:                "deck-1",
		SuggestionType:        models.MLSuggestionTypeAdd,
		CardID:                12345,
		CardName:              "Lightning Bolt",
		Confidence:            0.85,
		ExpectedWinRateChange: 2.5,
		Title:                 "Add Lightning Bolt",
		Description:           "This card has strong synergy",
		CreatedAt:             time.Now(),
	}

	err := repo.CreateSuggestion(ctx, suggestion)
	if err != nil {
		t.Fatalf("Failed to create ML suggestion: %v", err)
	}

	if suggestion.ID == 0 {
		t.Error("Expected suggestion ID to be set after creation")
	}
}

func TestMLSuggestionRepository_GetSuggestionsByDeck(t *testing.T) {
	db := setupMLSuggestionTestDB(t)
	defer db.Close()

	repo := NewMLSuggestionRepository(db)
	ctx := context.Background()

	// Create test suggestions
	s1 := &models.MLSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.MLSuggestionTypeAdd,
		CardName:       "Card 1",
		Confidence:     0.8,
		Title:          "Add Card 1",
		CreatedAt:      time.Now(),
	}
	s2 := &models.MLSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.MLSuggestionTypeRemove,
		CardName:       "Card 2",
		Confidence:     0.6,
		Title:          "Remove Card 2",
		CreatedAt:      time.Now(),
	}

	_ = repo.CreateSuggestion(ctx, s1)
	_ = repo.CreateSuggestion(ctx, s2)

	suggestions, err := repo.GetSuggestionsByDeck(ctx, "deck-1")
	if err != nil {
		t.Fatalf("Failed to get suggestions: %v", err)
	}

	if len(suggestions) != 2 {
		t.Errorf("Expected 2 suggestions, got %d", len(suggestions))
	}
}

func TestMLSuggestionRepository_GetActiveSuggestions(t *testing.T) {
	db := setupMLSuggestionTestDB(t)
	defer db.Close()

	repo := NewMLSuggestionRepository(db)
	ctx := context.Background()

	// Create suggestions - one active, one dismissed
	active := &models.MLSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.MLSuggestionTypeAdd,
		Title:          "Active",
		IsDismissed:    false,
		CreatedAt:      time.Now(),
	}
	dismissed := &models.MLSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.MLSuggestionTypeRemove,
		Title:          "Dismissed",
		IsDismissed:    true,
		CreatedAt:      time.Now(),
	}

	_ = repo.CreateSuggestion(ctx, active)
	_ = repo.CreateSuggestion(ctx, dismissed)

	// Mark second one as dismissed
	_ = repo.DismissSuggestion(ctx, dismissed.ID)

	suggestions, err := repo.GetActiveSuggestions(ctx, "deck-1")
	if err != nil {
		t.Fatalf("Failed to get active suggestions: %v", err)
	}

	if len(suggestions) != 1 {
		t.Errorf("Expected 1 active suggestion, got %d", len(suggestions))
	}
}

func TestMLSuggestionRepository_DismissSuggestion(t *testing.T) {
	db := setupMLSuggestionTestDB(t)
	defer db.Close()

	repo := NewMLSuggestionRepository(db)
	ctx := context.Background()

	suggestion := &models.MLSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.MLSuggestionTypeAdd,
		Title:          "Test",
		CreatedAt:      time.Now(),
	}
	_ = repo.CreateSuggestion(ctx, suggestion)

	err := repo.DismissSuggestion(ctx, suggestion.ID)
	if err != nil {
		t.Fatalf("Failed to dismiss suggestion: %v", err)
	}

	// Verify it's dismissed by getting all suggestions
	suggestions, _ := repo.GetSuggestionsByDeck(ctx, "deck-1")
	var found *models.MLSuggestion
	for _, s := range suggestions {
		if s.ID == suggestion.ID {
			found = s
			break
		}
	}
	if found == nil || !found.IsDismissed {
		t.Error("Expected suggestion to be dismissed")
	}
}

func TestMLSuggestionRepository_ApplySuggestion(t *testing.T) {
	db := setupMLSuggestionTestDB(t)
	defer db.Close()

	repo := NewMLSuggestionRepository(db)
	ctx := context.Background()

	suggestion := &models.MLSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.MLSuggestionTypeSwap,
		Title:          "Swap Test",
		CreatedAt:      time.Now(),
	}
	_ = repo.CreateSuggestion(ctx, suggestion)

	err := repo.ApplySuggestion(ctx, suggestion.ID)
	if err != nil {
		t.Fatalf("Failed to apply suggestion: %v", err)
	}

	// Verify it's applied by getting all suggestions
	suggestions, _ := repo.GetSuggestionsByDeck(ctx, "deck-1")
	var found *models.MLSuggestion
	for _, s := range suggestions {
		if s.ID == suggestion.ID {
			found = s
			break
		}
	}
	if found == nil || !found.WasApplied {
		t.Error("Expected suggestion to be applied")
	}
	if found.AppliedAt == nil {
		t.Error("Expected applied_at to be set")
	}
}

func TestMLSuggestionRepository_UpsertCombinationStats(t *testing.T) {
	db := setupMLSuggestionTestDB(t)
	defer db.Close()

	repo := NewMLSuggestionRepository(db)
	ctx := context.Background()

	stats := &models.CardCombinationStats{
		CardID1:       100,
		CardID2:       200,
		DeckID:        "deck-1",
		Format:        "Standard",
		GamesTogether: 10,
		WinsTogether:  7,
		SynergyScore:  0.15,
	}

	err := repo.UpsertCombinationStats(ctx, stats)
	if err != nil {
		t.Fatalf("Failed to upsert combination stats: %v", err)
	}

	// Update adds to existing counts (not replaces)
	addStats := &models.CardCombinationStats{
		CardID1:       100,
		CardID2:       200,
		DeckID:        "deck-1",
		Format:        "Standard",
		GamesTogether: 10, // Adding 10 more
		WinsTogether:  7,
	}
	err = repo.UpsertCombinationStats(ctx, addStats)
	if err != nil {
		t.Fatalf("Failed to update combination stats: %v", err)
	}

	// Verify accumulated (10 + 10 = 20)
	result, _ := repo.GetCombinationStats(ctx, 100, 200, "Standard")
	if result.GamesTogether != 20 {
		t.Errorf("Expected 20 games (accumulated), got %d", result.GamesTogether)
	}
}

func TestMLSuggestionRepository_GetTopSynergiesForCard(t *testing.T) {
	db := setupMLSuggestionTestDB(t)
	defer db.Close()

	repo := NewMLSuggestionRepository(db)
	ctx := context.Background()

	// Insert test synergy data (need GamesTogether >= 5 to be returned)
	stats1 := &models.CardCombinationStats{
		CardID1:       100,
		CardID2:       200,
		Format:        "Standard",
		GamesTogether: 10,
		SynergyScore:  0.25,
	}
	stats2 := &models.CardCombinationStats{
		CardID1:       100,
		CardID2:       300,
		Format:        "Standard",
		GamesTogether: 8,
		SynergyScore:  0.15,
	}

	_ = repo.UpsertCombinationStats(ctx, stats1)
	_ = repo.UpsertCombinationStats(ctx, stats2)

	synergies, err := repo.GetTopSynergiesForCard(ctx, 100, "Standard", 10)
	if err != nil {
		t.Fatalf("Failed to get synergies: %v", err)
	}

	if len(synergies) != 2 {
		t.Errorf("Expected 2 synergies, got %d", len(synergies))
	}

	if len(synergies) >= 2 {
		// Should be sorted by synergy score descending
		if synergies[0].SynergyScore < synergies[1].SynergyScore {
			t.Error("Expected synergies to be sorted by score descending")
		}
	}
}

func TestMLSuggestionRepository_SaveAndGetUserPlayPatterns(t *testing.T) {
	db := setupMLSuggestionTestDB(t)
	defer db.Close()

	repo := NewMLSuggestionRepository(db)
	ctx := context.Background()

	patterns := &models.UserPlayPatterns{
		AccountID:          "test-account",
		PreferredArchetype: "Aggro",
		AggroAffinity:      0.8,
		MidrangeAffinity:   0.1,
		ControlAffinity:    0.05,
		ComboAffinity:      0.05,
		TotalMatches:       100,
		TotalDecks:         5,
	}

	err := repo.UpsertUserPlayPatterns(ctx, patterns)
	if err != nil {
		t.Fatalf("Failed to save play patterns: %v", err)
	}

	// Retrieve
	result, err := repo.GetUserPlayPatterns(ctx, "test-account")
	if err != nil {
		t.Fatalf("Failed to get play patterns: %v", err)
	}

	if result.PreferredArchetype != "Aggro" {
		t.Errorf("Expected Aggro archetype, got %s", result.PreferredArchetype)
	}
	if result.TotalMatches != 100 {
		t.Errorf("Expected 100 matches, got %d", result.TotalMatches)
	}
}

func TestMLSuggestionRepository_SaveModelMetadata(t *testing.T) {
	db := setupMLSuggestionTestDB(t)
	defer db.Close()

	repo := NewMLSuggestionRepository(db)
	ctx := context.Background()

	accuracy := 0.85
	meta := &models.MLModelMetadata{
		ModelName:       "synergy-v1",
		ModelVersion:    "1.0.0",
		TrainingSamples: 1000,
		Accuracy:        &accuracy,
		IsActive:        true,
	}

	err := repo.SaveModelMetadata(ctx, meta)
	if err != nil {
		t.Fatalf("Failed to save model metadata: %v", err)
	}

	if meta.ID == 0 {
		t.Error("Expected model ID to be set")
	}

	// Update the same model
	meta.TrainingSamples = 2000
	err = repo.SaveModelMetadata(ctx, meta)
	if err != nil {
		t.Fatalf("Failed to update model metadata: %v", err)
	}
}

func TestCalculateConfidenceScore(t *testing.T) {
	// Formula: 1.0 - 1.0/(1.0+sqrt(sampleSize))
	tests := []struct {
		sampleSize  int
		minExpected float64
		maxExpected float64
	}{
		{1, 0.49, 0.51},    // 1 - 1/(1+1) = 0.5
		{10, 0.75, 0.77},   // 1 - 1/(1+√10) ≈ 0.76
		{100, 0.90, 0.92},  // 1 - 1/(1+10) ≈ 0.909
		{1000, 0.96, 0.98}, // 1 - 1/(1+√1000) ≈ 0.969
	}

	for _, tt := range tests {
		score := CalculateConfidenceScore(tt.sampleSize)
		if score < tt.minExpected || score > tt.maxExpected {
			t.Errorf("CalculateConfidenceScore(%d) = %f, want between %f and %f",
				tt.sampleSize, score, tt.minExpected, tt.maxExpected)
		}
	}
}

func TestCalculateSynergyScore(t *testing.T) {
	tests := []struct {
		name        string
		stats       *models.CardCombinationStats
		expectedMin float64
		expectedMax float64
	}{
		{
			name: "positive synergy",
			stats: &models.CardCombinationStats{
				GamesTogether:  20,
				WinsTogether:   14, // 70%
				GamesCard1Only: 10,
				WinsCard1Only:  5, // 50%
				GamesCard2Only: 10,
				WinsCard2Only:  5, // 50%
			},
			expectedMin: 0.15,
			expectedMax: 0.25,
		},
		{
			name: "not enough data",
			stats: &models.CardCombinationStats{
				GamesTogether: 3, // Less than min required
			},
			expectedMin: 0.0,
			expectedMax: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := CalculateSynergyScore(tt.stats)
			if score < tt.expectedMin || score > tt.expectedMax {
				t.Errorf("CalculateSynergyScore() = %f, want between %f and %f",
					score, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestMLSuggestionRepository_GetCombinationStats(t *testing.T) {
	db := setupMLSuggestionTestDB(t)
	defer db.Close()

	repo := NewMLSuggestionRepository(db)
	ctx := context.Background()

	// Insert test data
	stats := &models.CardCombinationStats{
		CardID1:       100,
		CardID2:       200,
		Format:        "Standard",
		GamesTogether: 15,
		WinsTogether:  10,
	}
	_ = repo.UpsertCombinationStats(ctx, stats)

	// Test getting existing stats
	result, err := repo.GetCombinationStats(ctx, 100, 200, "Standard")
	if err != nil {
		t.Fatalf("Failed to get combination stats: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.GamesTogether != 15 {
		t.Errorf("Expected 15 games together, got %d", result.GamesTogether)
	}

	// Test getting non-existent stats
	result, err = repo.GetCombinationStats(ctx, 999, 1000, "Standard")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != nil {
		t.Error("Expected nil result for non-existent stats")
	}
}

func TestMLSuggestionRepository_RecordSuggestionOutcome(t *testing.T) {
	db := setupMLSuggestionTestDB(t)
	defer db.Close()

	repo := NewMLSuggestionRepository(db)
	ctx := context.Background()

	// Create and apply a suggestion
	suggestion := &models.MLSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.MLSuggestionTypeAdd,
		Title:          "Test Outcome",
		CreatedAt:      time.Now(),
	}
	_ = repo.CreateSuggestion(ctx, suggestion)
	_ = repo.ApplySuggestion(ctx, suggestion.ID)

	// Record outcome
	err := repo.RecordSuggestionOutcome(ctx, suggestion.ID, 3.5)
	if err != nil {
		t.Fatalf("Failed to record outcome: %v", err)
	}

	// Verify outcome was recorded
	suggestions, _ := repo.GetSuggestionsByDeck(ctx, "deck-1")
	var found *models.MLSuggestion
	for _, s := range suggestions {
		if s.ID == suggestion.ID {
			found = s
			break
		}
	}
	if found == nil {
		t.Fatal("Suggestion not found")
	}
	if found.OutcomeWinRateChange == nil || *found.OutcomeWinRateChange != 3.5 {
		t.Error("Expected outcome win rate change to be 3.5")
	}
	if found.OutcomeRecordedAt == nil {
		t.Error("Expected outcome_recorded_at to be set")
	}
}

func TestMLSuggestionRepository_DeleteSuggestionsByDeck(t *testing.T) {
	db := setupMLSuggestionTestDB(t)
	defer db.Close()

	repo := NewMLSuggestionRepository(db)
	ctx := context.Background()

	// Create suggestions
	s1 := &models.MLSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.MLSuggestionTypeAdd,
		Title:          "Delete Test 1",
		CreatedAt:      time.Now(),
	}
	s2 := &models.MLSuggestion{
		DeckID:         "deck-1",
		SuggestionType: models.MLSuggestionTypeRemove,
		Title:          "Delete Test 2",
		CreatedAt:      time.Now(),
	}
	_ = repo.CreateSuggestion(ctx, s1)
	_ = repo.CreateSuggestion(ctx, s2)

	// Verify suggestions exist
	suggestions, _ := repo.GetSuggestionsByDeck(ctx, "deck-1")
	if len(suggestions) != 2 {
		t.Fatalf("Expected 2 suggestions, got %d", len(suggestions))
	}

	// Delete all suggestions for the deck
	err := repo.DeleteSuggestionsByDeck(ctx, "deck-1")
	if err != nil {
		t.Fatalf("Failed to delete suggestions: %v", err)
	}

	// Verify suggestions are deleted
	suggestions, _ = repo.GetSuggestionsByDeck(ctx, "deck-1")
	if len(suggestions) != 0 {
		t.Errorf("Expected 0 suggestions after delete, got %d", len(suggestions))
	}
}

func TestMLSuggestionRepository_CardAffinity(t *testing.T) {
	db := setupMLSuggestionTestDB(t)
	defer db.Close()

	repo := NewMLSuggestionRepository(db)
	ctx := context.Background()

	// Test UpsertCardAffinity
	affinity := &models.CardAffinity{
		CardID1:       100,
		CardID2:       200,
		Format:        "Standard",
		AffinityScore: 0.75,
		SampleSize:    50,
		Confidence:    0.8,
		Source:        "historical",
	}
	err := repo.UpsertCardAffinity(ctx, affinity)
	if err != nil {
		t.Fatalf("Failed to upsert card affinity: %v", err)
	}

	// Test GetCardAffinity
	result, err := repo.GetCardAffinity(ctx, 100, 200, "Standard")
	if err != nil {
		t.Fatalf("Failed to get card affinity: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.AffinityScore != 0.75 {
		t.Errorf("Expected affinity score 0.75, got %f", result.AffinityScore)
	}

	// Test update via upsert
	affinity.AffinityScore = 0.85
	err = repo.UpsertCardAffinity(ctx, affinity)
	if err != nil {
		t.Fatalf("Failed to update card affinity: %v", err)
	}

	result, _ = repo.GetCardAffinity(ctx, 100, 200, "Standard")
	if result.AffinityScore != 0.85 {
		t.Errorf("Expected updated affinity score 0.85, got %f", result.AffinityScore)
	}

	// Test GetCardAffinity for non-existent pair
	result, err = repo.GetCardAffinity(ctx, 999, 1000, "Standard")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != nil {
		t.Error("Expected nil result for non-existent affinity")
	}
}

func TestMLSuggestionRepository_GetTopAffinities(t *testing.T) {
	db := setupMLSuggestionTestDB(t)
	defer db.Close()

	repo := NewMLSuggestionRepository(db)
	ctx := context.Background()

	// Insert multiple affinities for card 100
	affinities := []*models.CardAffinity{
		{CardID1: 100, CardID2: 200, Format: "Standard", AffinityScore: 0.9, SampleSize: 20},
		{CardID1: 100, CardID2: 300, Format: "Standard", AffinityScore: 0.7, SampleSize: 15},
		{CardID1: 100, CardID2: 400, Format: "Standard", AffinityScore: 0.5, SampleSize: 10},
	}
	for _, a := range affinities {
		_ = repo.UpsertCardAffinity(ctx, a)
	}

	// Get top affinities
	results, err := repo.GetTopAffinities(ctx, 100, "Standard", 10)
	if err != nil {
		t.Fatalf("Failed to get top affinities: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("Expected 3 affinities, got %d", len(results))
	}

	// Verify sorted by affinity score descending
	if len(results) >= 2 && results[0].AffinityScore < results[1].AffinityScore {
		t.Error("Expected affinities to be sorted by score descending")
	}

	// Test with limit
	results, err = repo.GetTopAffinities(ctx, 100, "Standard", 2)
	if err != nil {
		t.Fatalf("Failed to get limited affinities: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 affinities with limit, got %d", len(results))
	}
}

func TestMLSuggestionRepository_GetActiveModel(t *testing.T) {
	db := setupMLSuggestionTestDB(t)
	defer db.Close()

	repo := NewMLSuggestionRepository(db)
	ctx := context.Background()

	// Create an active model
	accuracy := 0.9
	activeMeta := &models.MLModelMetadata{
		ModelName:       "synergy-model",
		ModelVersion:    "2.0.0",
		TrainingSamples: 5000,
		Accuracy:        &accuracy,
		IsActive:        true,
	}
	err := repo.SaveModelMetadata(ctx, activeMeta)
	if err != nil {
		t.Fatalf("Failed to save active model: %v", err)
	}

	// Create an inactive model
	inactiveMeta := &models.MLModelMetadata{
		ModelName:       "synergy-model",
		ModelVersion:    "1.0.0",
		TrainingSamples: 1000,
		IsActive:        false,
	}
	_ = repo.SaveModelMetadata(ctx, inactiveMeta)

	// Get active model
	result, err := repo.GetActiveModel(ctx, "synergy-model")
	if err != nil {
		t.Fatalf("Failed to get active model: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil active model")
	}
	if result.ModelVersion != "2.0.0" {
		t.Errorf("Expected version 2.0.0, got %s", result.ModelVersion)
	}
	if !result.IsActive {
		t.Error("Expected model to be active")
	}

	// Test getting non-existent model
	result, err = repo.GetActiveModel(ctx, "non-existent")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != nil {
		t.Error("Expected nil result for non-existent model")
	}
}

func TestMLSuggestionRepository_CalculateAndUpdateSynergyScores(t *testing.T) {
	db := setupMLSuggestionTestDB(t)
	defer db.Close()

	repo := NewMLSuggestionRepository(db)
	ctx := context.Background()

	// Insert combination stats with enough games
	stats := &models.CardCombinationStats{
		CardID1:        100,
		CardID2:        200,
		Format:         "Standard",
		GamesTogether:  20,
		WinsTogether:   14,
		GamesCard1Only: 10,
		WinsCard1Only:  5,
		GamesCard2Only: 10,
		WinsCard2Only:  5,
	}
	_ = repo.UpsertCombinationStats(ctx, stats)

	// Calculate and update synergy scores
	err := repo.CalculateAndUpdateSynergyScores(ctx, 5)
	if err != nil {
		t.Fatalf("Failed to calculate synergy scores: %v", err)
	}

	// Verify synergy score was updated
	result, _ := repo.GetCombinationStats(ctx, 100, 200, "Standard")
	if result.SynergyScore == 0 {
		t.Error("Expected synergy score to be calculated and non-zero")
	}
}

func TestMLSuggestionRepository_GetUserPlayPatterns_NotFound(t *testing.T) {
	db := setupMLSuggestionTestDB(t)
	defer db.Close()

	repo := NewMLSuggestionRepository(db)
	ctx := context.Background()

	// Test getting non-existent patterns
	result, err := repo.GetUserPlayPatterns(ctx, "non-existent-account")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != nil {
		t.Error("Expected nil result for non-existent account")
	}
}

func TestGetPairedCardID(t *testing.T) {
	stats := &models.CardCombinationStats{
		CardID1: 100,
		CardID2: 200,
	}

	// When querying for card 100, should return 200
	paired := GetPairedCardID(stats, 100)
	if paired != 200 {
		t.Errorf("Expected paired card 200, got %d", paired)
	}

	// When querying for card 200, should return 100
	paired = GetPairedCardID(stats, 200)
	if paired != 100 {
		t.Errorf("Expected paired card 100, got %d", paired)
	}
}
