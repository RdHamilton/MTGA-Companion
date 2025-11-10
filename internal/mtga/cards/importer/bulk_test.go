package importer

import (
	"testing"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
)

func TestDefaultBulkImportOptions(t *testing.T) {
	opts := DefaultBulkImportOptions()

	if opts.BatchSize != 500 {
		t.Errorf("Expected batch size 500, got %d", opts.BatchSize)
	}

	if opts.MaxAge.Hours() != 24 {
		t.Errorf("Expected max age 24h, got %v", opts.MaxAge)
	}

	if opts.ForceDownload {
		t.Error("Expected ForceDownload to be false by default")
	}

	if opts.Verbose {
		t.Error("Expected Verbose to be false by default")
	}
}

func TestConvertToStorageCard(t *testing.T) {
	arenaID := 12345
	card := &scryfall.Card{
		ID:              "test-id",
		ArenaID:         &arenaID,
		Name:            "Test Card",
		ManaCost:        "{1}{U}",
		CMC:             2.0,
		TypeLine:        "Creature - Test",
		OracleText:      "Test oracle text",
		Colors:          []string{"U"},
		ColorIdentity:   []string{"U"},
		Rarity:          "rare",
		SetCode:         "TST",
		CollectorNumber: "001",
		Layout:          "normal",
		ReleasedAt:      "2024-01-01",
		Legalities: scryfall.Legalities{
			Standard: "legal",
			Modern:   "legal",
			Legacy:   "legal",
		},
	}

	storageCard := ConvertToStorageCard(card)

	if storageCard.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got %s", storageCard.ID)
	}

	if storageCard.ArenaID == nil || *storageCard.ArenaID != 12345 {
		t.Error("Expected ArenaID to be 12345")
	}

	if storageCard.Name != "Test Card" {
		t.Errorf("Expected name 'Test Card', got %s", storageCard.Name)
	}

	if storageCard.CMC != 2.0 {
		t.Errorf("Expected CMC 2.0, got %.1f", storageCard.CMC)
	}

	if len(storageCard.Colors) != 1 || storageCard.Colors[0] != "U" {
		t.Errorf("Expected colors [U], got %v", storageCard.Colors)
	}

	if storageCard.SetCode != "TST" {
		t.Errorf("Expected set code 'TST', got %s", storageCard.SetCode)
	}

	// Check legalities conversion
	if storageCard.Legalities["standard"] != "legal" {
		t.Errorf("Expected standard legality 'legal', got %s", storageCard.Legalities["standard"])
	}

	if storageCard.Legalities["modern"] != "legal" {
		t.Errorf("Expected modern legality 'legal', got %s", storageCard.Legalities["modern"])
	}
}

func TestLegalitiesToMap(t *testing.T) {
	legalities := scryfall.Legalities{
		Standard:      "legal",
		Future:        "not_legal",
		Historic:      "legal",
		Modern:        "legal",
		Legacy:        "legal",
		Vintage:       "legal",
		Commander:     "legal",
		Brawl:         "legal",
		HistoricBrawl: "legal",
		Alchemy:       "not_legal",
	}

	legalityMap := legalitiesToMap(legalities)

	testCases := []struct {
		key      string
		expected string
	}{
		{"standard", "legal"},
		{"future", "not_legal"},
		{"historic", "legal"},
		{"modern", "legal"},
		{"legacy", "legal"},
		{"vintage", "legal"},
		{"commander", "legal"},
		{"brawl", "legal"},
		{"historicbrawl", "legal"},
		{"alchemy", "not_legal"},
	}

	for _, tc := range testCases {
		if legalityMap[tc.key] != tc.expected {
			t.Errorf("Expected %s legality '%s', got '%s'", tc.key, tc.expected, legalityMap[tc.key])
		}
	}
}
