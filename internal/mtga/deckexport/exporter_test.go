package deckexport

import (
	"strings"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// mockCardService provides test card data
type mockCardService struct {
	cards map[int]*cards.Card
}

func (m *mockCardService) GetCard(id int) (*cards.Card, error) {
	if card, ok := m.cards[id]; ok {
		return card, nil
	}
	return nil, nil
}

func newMockCardService() *mockCardService {
	return &mockCardService{
		cards: map[int]*cards.Card{
			1: {
				ArenaID:         1,
				Name:            "Lightning Bolt",
				SetCode:         "m21",
				CollectorNumber: "123",
				TypeLine:        "Instant",
			},
			2: {
				ArenaID:         2,
				Name:            "Shock",
				SetCode:         "m21",
				CollectorNumber: "124",
				TypeLine:        "Instant",
			},
			3: {
				ArenaID:         3,
				Name:            "Mountain",
				SetCode:         "m21",
				CollectorNumber: "275",
				TypeLine:        "Basic Land — Mountain",
			},
			4: {
				ArenaID:         4,
				Name:            "Duress",
				SetCode:         "m21",
				CollectorNumber: "95",
				TypeLine:        "Sorcery",
			},
		},
	}
}

func createTestDeck() (*models.Deck, []*models.DeckCard) {
	deck := &models.Deck{
		ID:         "test-deck-123",
		Name:       "Test Deck",
		Format:     "Standard",
		Source:     "constructed",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	deckCards := []*models.DeckCard{
		{DeckID: deck.ID, CardID: 1, Quantity: 4, Board: "main"},
		{DeckID: deck.ID, CardID: 2, Quantity: 3, Board: "main"},
		{DeckID: deck.ID, CardID: 3, Quantity: 20, Board: "main"},
		{DeckID: deck.ID, CardID: 4, Quantity: 2, Board: "sideboard"},
	}

	return deck, deckCards
}

func TestNewExporter(t *testing.T) {
	mockService := newMockCardService()
	exporter := NewExporter(mockService)

	if exporter == nil {
		t.Fatal("NewExporter returned nil")
	}
	if exporter.cardProvider == nil {
		t.Error("cardProvider should not be nil")
	}
}

func TestExport_ArenaFormat(t *testing.T) {
	mockService := newMockCardService()
	exporter := NewExporter(mockService)
	deck, deckCards := createTestDeck()

	options := &ExportOptions{
		Format:         FormatArena,
		IncludeHeaders: true,
	}

	result, err := exporter.Export(deck, deckCards, options)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Check basic properties
	if result.Format != FormatArena {
		t.Errorf("Format = %v, want %v", result.Format, FormatArena)
	}
	if result.Filename == "" {
		t.Error("Filename should not be empty")
	}
	if !strings.HasSuffix(result.Filename, ".txt") {
		t.Errorf("Filename should end with .txt, got %s", result.Filename)
	}

	// Check content
	content := result.Content
	if !strings.Contains(content, "Deck\n") {
		t.Error("Content should contain 'Deck' header")
	}
	if !strings.Contains(content, "4 Lightning Bolt (M21) 123") {
		t.Error("Content should contain mainboard cards with set info")
	}
	if !strings.Contains(content, "2 Duress (M21) 95") {
		t.Error("Content should contain sideboard cards")
	}
	if !strings.Contains(content, "20 Mountain (M21) 275") {
		t.Error("Content should contain lands")
	}

	// Check sideboard separation (empty line before sideboard)
	lines := strings.Split(content, "\n")
	hasEmptyLineBeforeSideboard := false
	for i, line := range lines {
		if strings.Contains(line, "Duress") && i > 0 && lines[i-1] == "" {
			hasEmptyLineBeforeSideboard = true
			break
		}
	}
	if !hasEmptyLineBeforeSideboard {
		t.Error("Should have empty line before sideboard")
	}
}

func TestExport_PlainTextFormat(t *testing.T) {
	mockService := newMockCardService()
	exporter := NewExporter(mockService)
	deck, deckCards := createTestDeck()

	options := &ExportOptions{
		Format:         FormatPlainText,
		IncludeHeaders: true,
	}

	result, err := exporter.Export(deck, deckCards, options)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	content := result.Content
	if !strings.Contains(content, "// Test Deck") {
		t.Error("Content should contain deck name as comment")
	}
	if !strings.Contains(content, "4x Lightning Bolt") {
		t.Error("Content should use 'x' format")
	}
	if !strings.Contains(content, "Mainboard:") {
		t.Error("Content should have Mainboard header")
	}
	if !strings.Contains(content, "Sideboard:") {
		t.Error("Content should have Sideboard header")
	}
	if !strings.Contains(content, "2x Duress") {
		t.Error("Content should contain sideboard cards")
	}
}

func TestExport_MTGOFormat(t *testing.T) {
	mockService := newMockCardService()
	exporter := NewExporter(mockService)
	deck, deckCards := createTestDeck()

	options := &ExportOptions{
		Format: FormatMTGO,
	}

	result, err := exporter.Export(deck, deckCards, options)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if !strings.HasSuffix(result.Filename, ".dek") {
		t.Errorf("MTGO export should have .dek extension, got %s", result.Filename)
	}

	content := result.Content
	if !strings.Contains(content, "4 Lightning Bolt") {
		t.Error("Content should contain mainboard cards without 'x'")
	}
	if !strings.Contains(content, "SB: 2 Duress") {
		t.Error("Content should prefix sideboard cards with 'SB:'")
	}
	if strings.Contains(content, "4x") {
		t.Error("MTGO format should not use 'x'")
	}
}

func TestExport_MTGGoldfishFormat(t *testing.T) {
	mockService := newMockCardService()
	exporter := NewExporter(mockService)
	deck, deckCards := createTestDeck()

	options := &ExportOptions{
		Format: FormatMTGGoldfish,
	}

	result, err := exporter.Export(deck, deckCards, options)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	content := result.Content
	if !strings.Contains(content, "4 Lightning Bolt") {
		t.Error("Content should contain mainboard cards")
	}
	if !strings.Contains(content, "Sideboard\n") {
		t.Error("Content should have 'Sideboard' header")
	}
	if !strings.Contains(content, "2 Duress") {
		t.Error("Content should contain sideboard cards")
	}
}

func TestExport_WithoutHeaders(t *testing.T) {
	mockService := newMockCardService()
	exporter := NewExporter(mockService)
	deck, deckCards := createTestDeck()

	options := &ExportOptions{
		Format:         FormatArena,
		IncludeHeaders: false,
	}

	result, err := exporter.Export(deck, deckCards, options)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	content := result.Content
	if strings.Contains(content, "Deck\n") {
		t.Error("Content should not contain 'Deck' header when IncludeHeaders is false")
	}
	// Should still have cards
	if !strings.Contains(content, "Lightning Bolt") {
		t.Error("Content should still contain cards")
	}
}

func TestExport_EmptyDeck(t *testing.T) {
	mockService := newMockCardService()
	exporter := NewExporter(mockService)

	deck := &models.Deck{
		ID:   "empty-deck",
		Name: "Empty Deck",
	}
	deckCards := []*models.DeckCard{}

	options := &ExportOptions{
		Format:         FormatArena,
		IncludeHeaders: true,
	}

	result, err := exporter.Export(deck, deckCards, options)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Should succeed and have at least headers
	if !strings.Contains(result.Content, "Deck") {
		t.Error("Content should contain 'Deck' header even for empty deck")
	}
}

func TestExport_MainboardOnly(t *testing.T) {
	mockService := newMockCardService()
	exporter := NewExporter(mockService)

	deck := &models.Deck{
		ID:   "mainboard-only",
		Name: "Mainboard Only",
	}
	deckCards := []*models.DeckCard{
		{DeckID: deck.ID, CardID: 1, Quantity: 4, Board: "main"},
	}

	options := &ExportOptions{
		Format: FormatArena,
	}

	result, err := exporter.Export(deck, deckCards, options)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	content := result.Content
	if !strings.Contains(content, "Lightning Bolt") {
		t.Error("Content should contain mainboard cards")
	}

	// Should not have empty line at end (no sideboard)
	lines := strings.Split(strings.TrimSpace(content), "\n")
	lastLine := lines[len(lines)-1]
	if !strings.Contains(lastLine, "Lightning Bolt") {
		t.Error("Last line should be a card, not empty")
	}
}

func TestExport_NilDeck(t *testing.T) {
	mockService := newMockCardService()
	exporter := NewExporter(mockService)

	_, err := exporter.Export(nil, []*models.DeckCard{}, nil)
	if err == nil {
		t.Error("Export should return error for nil deck")
	}
	if !strings.Contains(err.Error(), "deck is nil") {
		t.Errorf("Error message should mention nil deck, got: %v", err)
	}
}

func TestExport_DefaultOptions(t *testing.T) {
	mockService := newMockCardService()
	exporter := NewExporter(mockService)
	deck, deckCards := createTestDeck()

	// Pass nil options to test defaults
	result, err := exporter.Export(deck, deckCards, nil)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Should default to Arena format with headers
	if result.Format != FormatArena {
		t.Errorf("Default format should be Arena, got %v", result.Format)
	}
	if !strings.Contains(result.Content, "Deck\n") {
		t.Error("Default should include headers")
	}
}

func TestFilterCardsByBoard(t *testing.T) {
	deckCards := []*models.DeckCard{
		{CardID: 1, Board: "main", Quantity: 4},
		{CardID: 2, Board: "main", Quantity: 3},
		{CardID: 3, Board: "sideboard", Quantity: 2},
	}

	mainboard := filterCardsByBoard(deckCards, "main")
	if len(mainboard) != 2 {
		t.Errorf("Mainboard should have 2 cards, got %d", len(mainboard))
	}

	sideboard := filterCardsByBoard(deckCards, "sideboard")
	if len(sideboard) != 1 {
		t.Errorf("Sideboard should have 1 card, got %d", len(sideboard))
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "My Deck",
			expected: "My Deck",
		},
		{
			name:     "with invalid characters",
			input:    "My/Deck\\Name:Test",
			expected: "My_Deck_Name_Test",
		},
		{
			name:     "with quotes and angles",
			input:    `Deck "Name" <Test>`,
			expected: "Deck _Name_ _Test_",
		},
		{
			name:     "empty name",
			input:    "",
			expected: "deck",
		},
		{
			name:     "very long name",
			input:    strings.Repeat("A", 150),
			expected: strings.Repeat("A", 100),
		},
		{
			name:     "with trailing spaces",
			input:    "  My Deck  ",
			expected: "My Deck",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExport_AllFormats(t *testing.T) {
	mockService := newMockCardService()
	exporter := NewExporter(mockService)
	deck, deckCards := createTestDeck()

	formats := []ExportFormat{
		FormatArena,
		FormatPlainText,
		FormatMTGO,
		FormatMTGGoldfish,
	}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			options := &ExportOptions{
				Format: format,
			}

			result, err := exporter.Export(deck, deckCards, options)
			if err != nil {
				t.Fatalf("Export failed for format %s: %v", format, err)
			}

			if result.Content == "" {
				t.Error("Content should not be empty")
			}
			if result.Format != format {
				t.Errorf("Format = %v, want %v", result.Format, format)
			}
			if result.Filename == "" {
				t.Error("Filename should not be empty")
			}

			// All formats should include mainboard cards
			if !strings.Contains(result.Content, "Lightning Bolt") {
				t.Error("Content should contain Lightning Bolt")
			}

			// All formats should include sideboard cards
			if !strings.Contains(result.Content, "Duress") {
				t.Error("Content should contain Duress from sideboard")
			}
		})
	}
}
