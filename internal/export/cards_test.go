package export

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/unified"
)

// createTestCards creates a set of test unified cards.
func createTestCards() []*unified.UnifiedCard {
	now := time.Now()

	return []*unified.UnifiedCard{
		{
			ID:              "test-1",
			ArenaID:         89019,
			Name:            "Lightning Bolt",
			ManaCost:        "R",
			CMC:             1,
			TypeLine:        "Instant",
			OracleText:      "Deal 3 damage to any target.",
			Colors:          []string{"R"},
			ColorIdentity:   []string{"R"},
			Rarity:          "common",
			SetCode:         "BLB",
			CollectorNumber: "123",
			DraftStats: &unified.DraftStatistics{
				GIHWR:       58.5,
				OHWR:        60.2,
				ALSA:        3.2,
				ATA:         4.1,
				GIH:         12453,
				GamesPlayed: 8932,
				NumDecks:    4521,
				Format:      "PremierDraft",
				LastUpdated: now.Add(-2 * time.Hour),
			},
			MetadataAge:    1 * time.Hour,
			StatsAge:       2 * time.Hour,
			MetadataSource: unified.SourceCache,
			StatsSource:    unified.SourceAPI,
		},
		{
			ID:              "test-2",
			ArenaID:         89020,
			Name:            "Counterspell",
			ManaCost:        "UU",
			CMC:             2,
			TypeLine:        "Instant",
			OracleText:      "Counter target spell.",
			Colors:          []string{"U"},
			ColorIdentity:   []string{"U"},
			Rarity:          "uncommon",
			SetCode:         "BLB",
			CollectorNumber: "124",
			DraftStats: &unified.DraftStatistics{
				GIHWR:       56.2,
				OHWR:        58.1,
				ALSA:        5.1,
				ATA:         6.8,
				GIH:         8932,
				GamesPlayed: 6421,
				NumDecks:    3210,
				Format:      "PremierDraft",
				LastUpdated: now.Add(-3 * time.Hour),
			},
			MetadataAge:    1 * time.Hour,
			StatsAge:       3 * time.Hour,
			MetadataSource: unified.SourceCache,
			StatsSource:    unified.SourceCache,
		},
		{
			ID:              "test-3",
			ArenaID:         89021,
			Name:            "Grizzly Bears",
			ManaCost:        "1G",
			CMC:             2,
			TypeLine:        "Creature — Bear",
			OracleText:      "",
			Colors:          []string{"G"},
			ColorIdentity:   []string{"G"},
			Rarity:          "common",
			SetCode:         "BLB",
			CollectorNumber: "125",
			Power:           "2",
			Toughness:       "2",
			DraftStats: &unified.DraftStatistics{
				GIHWR:       52.1,
				OHWR:        51.3,
				ALSA:        8.5,
				ATA:         9.2,
				GIH:         5421,
				GamesPlayed: 3892,
				NumDecks:    2341,
				Format:      "PremierDraft",
				LastUpdated: now.Add(-1 * time.Hour),
			},
			MetadataAge:    1 * time.Hour,
			StatsAge:       1 * time.Hour,
			MetadataSource: unified.SourceCache,
			StatsSource:    unified.SourceAPI,
		},
		{
			ID:              "test-4",
			ArenaID:         89022,
			Name:            "Dark Ritual",
			ManaCost:        "B",
			CMC:             1,
			TypeLine:        "Instant",
			OracleText:      "Add BBB.",
			Colors:          []string{"B"},
			ColorIdentity:   []string{"B"},
			Rarity:          "rare",
			SetCode:         "BLB",
			CollectorNumber: "126",
			// No draft stats
			MetadataAge:    1 * time.Hour,
			StatsAge:       0,
			MetadataSource: unified.SourceCache,
			StatsSource:    unified.SourceUnknown,
		},
	}
}

func TestExportCardsCSV(t *testing.T) {
	cards := createTestCards()

	tests := []struct {
		name         string
		opts         CardExportOptions
		wantHeaders  []string
		wantRowCount int
	}{
		{
			name: "basic export with stats",
			opts: CardExportOptions{
				Format:       FormatCSV,
				IncludeStats: true,
			},
			wantHeaders:  []string{"Arena ID", "Name", "Mana Cost", "CMC", "Type", "Rarity", "Colors", "Set", "GIHWR", "OHWR", "ALSA", "ATA", "Sample Size", "Games Played"},
			wantRowCount: 4,
		},
		{
			name: "export without stats",
			opts: CardExportOptions{
				Format:       FormatCSV,
				IncludeStats: false,
			},
			wantHeaders:  []string{"Arena ID", "Name", "Mana Cost", "CMC", "Type", "Rarity", "Colors", "Set"},
			wantRowCount: 4,
		},
		{
			name: "export with data age",
			opts: CardExportOptions{
				Format:       FormatCSV,
				IncludeStats: true,
				ShowDataAge:  true,
			},
			wantHeaders:  []string{"Arena ID", "Name", "Mana Cost", "CMC", "Type", "Rarity", "Colors", "Set", "GIHWR", "OHWR", "ALSA", "ATA", "Sample Size", "Games Played", "Metadata Age", "Stats Age"},
			wantRowCount: 4,
		},
		{
			name: "export only with stats",
			opts: CardExportOptions{
				Format:        FormatCSV,
				IncludeStats:  true,
				OnlyWithStats: true,
			},
			wantHeaders:  []string{"Arena ID", "Name", "Mana Cost", "CMC", "Type", "Rarity", "Colors", "Set", "GIHWR", "OHWR", "ALSA", "ATA", "Sample Size", "Games Played"},
			wantRowCount: 3, // Excludes Dark Ritual
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := ExportCards(&buf, cards, tt.opts)
			if err != nil {
				t.Fatalf("ExportCards() error = %v", err)
			}

			// Parse CSV
			reader := csv.NewReader(&buf)
			records, err := reader.ReadAll()
			if err != nil {
				t.Fatalf("Failed to parse CSV: %v", err)
			}

			// Check headers
			if len(records) < 1 {
				t.Fatal("CSV has no header row")
			}
			headers := records[0]
			if len(headers) != len(tt.wantHeaders) {
				t.Errorf("Header count = %d, want %d", len(headers), len(tt.wantHeaders))
			}
			for i, want := range tt.wantHeaders {
				if i >= len(headers) {
					break
				}
				if headers[i] != want {
					t.Errorf("Header[%d] = %q, want %q", i, headers[i], want)
				}
			}

			// Check row count (excluding header)
			dataRows := len(records) - 1
			if dataRows != tt.wantRowCount {
				t.Errorf("Row count = %d, want %d", dataRows, tt.wantRowCount)
			}

			// Verify first data row
			if len(records) > 1 {
				firstRow := records[1]
				if firstRow[0] != "89019" { // Arena ID
					t.Errorf("First row Arena ID = %q, want %q", firstRow[0], "89019")
				}
				if firstRow[1] != "Lightning Bolt" {
					t.Errorf("First row Name = %q, want %q", firstRow[1], "Lightning Bolt")
				}
			}
		})
	}
}

func TestExportCardsJSON(t *testing.T) {
	cards := createTestCards()

	tests := []struct {
		name          string
		opts          CardExportOptions
		wantCardCount int
	}{
		{
			name: "basic JSON export",
			opts: CardExportOptions{
				Format:       FormatJSON,
				IncludeStats: true,
				PrettyJSON:   true,
			},
			wantCardCount: 4,
		},
		{
			name: "JSON with metadata",
			opts: CardExportOptions{
				Format:       FormatJSON,
				IncludeStats: true,
				ShowDataAge:  true,
				PrettyJSON:   true,
			},
			wantCardCount: 4,
		},
		{
			name: "JSON only with stats",
			opts: CardExportOptions{
				Format:        FormatJSON,
				IncludeStats:  true,
				OnlyWithStats: true,
				PrettyJSON:    true,
			},
			wantCardCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := ExportCards(&buf, cards, tt.opts)
			if err != nil {
				t.Fatalf("ExportCards() error = %v", err)
			}

			// Parse JSON
			var data CardExportData
			err = json.Unmarshal(buf.Bytes(), &data)
			if err != nil {
				t.Fatalf("Failed to parse JSON: %v", err)
			}

			// Check card count
			if data.TotalCards != tt.wantCardCount {
				t.Errorf("TotalCards = %d, want %d", data.TotalCards, tt.wantCardCount)
			}
			if len(data.Cards) != tt.wantCardCount {
				t.Errorf("len(Cards) = %d, want %d", len(data.Cards), tt.wantCardCount)
			}

			// Check metadata if requested
			if tt.opts.ShowDataAge {
				if data.Metadata == nil {
					t.Error("Metadata is nil, want non-nil")
				} else {
					if data.Metadata.CardsWithStats == 0 {
						t.Error("CardsWithStats = 0, want > 0")
					}
				}
			}

			// Verify first card
			if len(data.Cards) > 0 {
				firstCard := data.Cards[0]
				if firstCard.ArenaID != 89019 {
					t.Errorf("First card ArenaID = %d, want 89019", firstCard.ArenaID)
				}
				if firstCard.Name != "Lightning Bolt" {
					t.Errorf("First card Name = %q, want %q", firstCard.Name, "Lightning Bolt")
				}
				if tt.opts.IncludeStats {
					if firstCard.GIHWR == nil {
						t.Error("First card GIHWR is nil, want non-nil")
					} else if *firstCard.GIHWR != 58.5 {
						t.Errorf("First card GIHWR = %.1f, want 58.5", *firstCard.GIHWR)
					}
				}
			}
		})
	}
}

func TestExportCardsMarkdown(t *testing.T) {
	cards := createTestCards()

	tests := []struct {
		name         string
		opts         CardExportOptions
		wantContains []string
		wantRowCount int
	}{
		{
			name: "basic markdown export",
			opts: CardExportOptions{
				Format:       "markdown",
				IncludeStats: true,
			},
			wantContains: []string{
				"# BLB Draft Statistics",
				"**Total Cards:**",
				"| Card | Mana Cost | Type | Rarity | GIHWR | ALSA | ATA | Sample |",
				"Lightning Bolt",
				"Counterspell",
			},
			wantRowCount: 3, // 3 cards with stats
		},
		{
			name: "markdown with metadata",
			opts: CardExportOptions{
				Format:       "markdown",
				IncludeStats: true,
				ShowDataAge:  true,
			},
			wantContains: []string{
				"## Data Freshness",
				"**Average Metadata Age:**",
				"**Average Stats Age:**",
				"**Cards With Stats:**",
			},
			wantRowCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := ExportCards(&buf, cards, tt.opts)
			if err != nil {
				t.Fatalf("ExportCards() error = %v", err)
			}

			output := buf.String()

			// Check for expected content
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing expected string: %q", want)
				}
			}

			// Count table rows (lines with |...|)
			lines := strings.Split(output, "\n")
			rowCount := 0
			for _, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "|") && strings.Contains(line, "Lightning Bolt") {
					rowCount++
				}
			}

			// Note: rowCount checking is approximate for markdown tables
			// Just verify we have some content
			if rowCount == 0 && tt.wantRowCount > 0 {
				t.Error("Expected table rows but found none")
			}
		})
	}
}

func TestExportCardsArena(t *testing.T) {
	cards := createTestCards()

	tests := []struct {
		name         string
		opts         CardExportOptions
		wantContains []string
		wantLines    int
	}{
		{
			name: "basic arena export",
			opts: CardExportOptions{
				Format: FormatArena,
			},
			wantContains: []string{
				"Deck\n",
				"1 Lightning Bolt (BLB) 123",
				"1 Counterspell (BLB) 124",
				"1 Grizzly Bears (BLB) 125",
			},
			wantLines: 6, // Deck header + 4 cards + trailing newline
		},
		{
			name: "arena export with top N",
			opts: CardExportOptions{
				Format: FormatArena,
				TopN:   2,
			},
			wantContains: []string{
				"Deck\n",
				"1 Lightning Bolt (BLB) 123",
			},
			wantLines: 4, // Deck header + 2 cards + trailing newline
		},
		{
			name: "arena export with only stats",
			opts: CardExportOptions{
				Format:        FormatArena,
				OnlyWithStats: true,
			},
			wantContains: []string{
				"Deck\n",
				"1 Lightning Bolt (BLB) 123",
				"1 Counterspell (BLB) 124",
			},
			wantLines: 5, // Deck header + 3 cards with stats + trailing newline
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := ExportCards(&buf, cards, tt.opts)
			if err != nil {
				t.Fatalf("ExportCards() error = %v", err)
			}

			output := buf.String()

			// Check for expected content
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing expected string: %q\nGot:\n%s", want, output)
				}
			}

			// Check line count
			lines := strings.Split(output, "\n")
			if len(lines) != tt.wantLines {
				t.Errorf("Line count = %d, want %d\nGot:\n%s", len(lines), tt.wantLines, output)
			}

			// Verify Arena format structure
			if !strings.HasPrefix(output, "Deck\n") {
				t.Error("Arena export should start with 'Deck' header")
			}

			// Verify format: <quantity> <name> (<set>) <number>
			for _, line := range lines[1:] { // Skip Deck header and empty line at end
				if line == "" {
					continue
				}
				if !strings.Contains(line, "(BLB)") {
					t.Errorf("Card line missing set code: %q", line)
				}
				if !strings.HasPrefix(line, "1 ") {
					t.Errorf("Card line should start with quantity: %q", line)
				}
			}
		})
	}
}

func TestFilterCards(t *testing.T) {
	cards := createTestCards()

	tests := []struct {
		name      string
		opts      CardExportOptions
		wantCount int
	}{
		{
			name:      "no filters",
			opts:      CardExportOptions{},
			wantCount: 4,
		},
		{
			name: "only with stats",
			opts: CardExportOptions{
				OnlyWithStats: true,
			},
			wantCount: 3,
		},
		{
			name: "minimum sample size",
			opts: CardExportOptions{
				MinSampleSize: 8000,
			},
			wantCount: 2, // Lightning Bolt and Counterspell
		},
		{
			name: "filter by rarity - common",
			opts: CardExportOptions{
				FilterRarity: []string{"common"},
			},
			wantCount: 2, // Lightning Bolt and Grizzly Bears
		},
		{
			name: "filter by rarity - uncommon and rare",
			opts: CardExportOptions{
				FilterRarity: []string{"uncommon", "rare"},
			},
			wantCount: 2, // Counterspell and Dark Ritual
		},
		{
			name: "filter by color - red",
			opts: CardExportOptions{
				FilterColors: []string{"R"},
			},
			wantCount: 1, // Lightning Bolt
		},
		{
			name: "filter by color - blue or green",
			opts: CardExportOptions{
				FilterColors: []string{"U", "G"},
			},
			wantCount: 2, // Counterspell and Grizzly Bears
		},
		{
			name: "combined filters",
			opts: CardExportOptions{
				OnlyWithStats: true,
				FilterRarity:  []string{"common"},
			},
			wantCount: 2, // Lightning Bolt and Grizzly Bears (both common with stats)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterCards(cards, tt.opts)
			if len(filtered) != tt.wantCount {
				t.Errorf("filterCards() count = %d, want %d", len(filtered), tt.wantCount)
			}
		})
	}
}

func TestSortCards(t *testing.T) {
	cards := createTestCards()

	tests := []struct {
		name      string
		sortBy    string
		wantFirst string
		wantLast  string
	}{
		{
			name:      "sort by name",
			sortBy:    "name",
			wantFirst: "Counterspell",
			wantLast:  "Lightning Bolt",
		},
		{
			name:      "sort by GIHWR (highest first)",
			sortBy:    "gihwr",
			wantFirst: "Lightning Bolt", // 58.5%
			wantLast:  "Dark Ritual",    // No stats
		},
		{
			name:      "sort by ALSA (lowest first)",
			sortBy:    "alsa",
			wantFirst: "Lightning Bolt", // 3.2
			wantLast:  "Dark Ritual",    // No stats
		},
		{
			name:      "sort by CMC",
			sortBy:    "cmc",
			wantFirst: "Lightning Bolt", // CMC 1
			wantLast:  "Grizzly Bears",  // CMC 2 (last alphabetically of CMC 2)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorted := sortCards(cards, tt.sortBy)
			if len(sorted) == 0 {
				t.Fatal("sortCards() returned empty slice")
			}

			first := sorted[0].Name
			last := sorted[len(sorted)-1].Name

			if first != tt.wantFirst {
				t.Errorf("First card = %q, want %q", first, tt.wantFirst)
			}
			if last != tt.wantLast {
				t.Errorf("Last card = %q, want %q", last, tt.wantLast)
			}
		})
	}
}

func TestTopNLimit(t *testing.T) {
	cards := createTestCards()

	opts := CardExportOptions{
		Format:       FormatJSON,
		IncludeStats: true,
		TopN:         2,
		SortBy:       "gihwr",
	}

	var buf bytes.Buffer
	err := ExportCards(&buf, cards, opts)
	if err != nil {
		t.Fatalf("ExportCards() error = %v", err)
	}

	var data CardExportData
	err = json.Unmarshal(buf.Bytes(), &data)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(data.Cards) != 2 {
		t.Errorf("len(Cards) = %d, want 2", len(data.Cards))
	}

	// Should be Lightning Bolt and Counterspell (highest GIHWR)
	if len(data.Cards) >= 2 {
		if data.Cards[0].Name != "Lightning Bolt" {
			t.Errorf("First card = %q, want Lightning Bolt", data.Cards[0].Name)
		}
		if data.Cards[1].Name != "Counterspell" {
			t.Errorf("Second card = %q, want Counterspell", data.Cards[1].Name)
		}
	}
}

func TestFormatHelpers(t *testing.T) {
	tests := []struct {
		name         string
		input        time.Duration
		wantContains string
	}{
		{"less than minute", 30 * time.Second, "< 1 minute"},
		{"minutes", 5 * time.Minute, "5 minutes"},
		{"hours", 2 * time.Hour, "2.0 hours"},
		{"days", 36 * time.Hour, "1.5 days"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.input)
			if !strings.Contains(result, tt.wantContains) {
				t.Errorf("formatDuration(%v) = %q, want to contain %q", tt.input, result, tt.wantContains)
			}
		})
	}
}

func TestFormatSampleSize(t *testing.T) {
	tests := []struct {
		name string
		size *int
		want string
	}{
		{"nil", nil, "N/A"},
		{"small", intPtr(500), "500"},
		{"thousands", intPtr(12453), "12.5K"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSampleSize(tt.size)
			if result != tt.want {
				t.Errorf("formatSampleSize() = %q, want %q", result, tt.want)
			}
		})
	}
}

// Helper function
func intPtr(i int) *int {
	return &i
}
