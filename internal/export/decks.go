package export

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/viewer"
)

// DeckFormat represents the deck export format.
type DeckFormat string

const (
	// DeckFormatArena represents MTG Arena deck format.
	DeckFormatArena DeckFormat = "arena"
	// DeckFormatText represents simple text format.
	DeckFormatText DeckFormat = "text"
	// DeckFormatJSON represents JSON format.
	DeckFormatJSON DeckFormat = "json"
	// DeckFormatCSV represents CSV format.
	DeckFormatCSV DeckFormat = "csv"
)

// DeckExportRow represents a deck card for CSV export.
type DeckExportRow struct {
	DeckID      string `csv:"deck_id" json:"deck_id"`
	DeckName    string `csv:"deck_name" json:"deck_name"`
	DeckFormat  string `csv:"deck_format" json:"deck_format"`
	Board       string `csv:"board" json:"board"`
	Quantity    int    `csv:"quantity" json:"quantity"`
	CardID      int    `csv:"card_id" json:"card_id"`
	CardName    string `csv:"card_name" json:"card_name"`
	ManaCost    string `csv:"mana_cost" json:"mana_cost"`
	Type        string `csv:"type" json:"type"`
	Rarity      string `csv:"rarity" json:"rarity"`
	SetCode     string `csv:"set_code" json:"set_code"`
	CollectorNo string `csv:"collector_number" json:"collector_number"`
}

// DeckJSON represents a complete deck in JSON format.
type DeckJSON struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Format        string         `json:"format"`
	Description   string         `json:"description,omitempty"`
	ColorIdentity string         `json:"color_identity,omitempty"`
	MainDeck      []DeckCardJSON `json:"main_deck"`
	Sideboard     []DeckCardJSON `json:"sideboard"`
	TotalCards    int            `json:"total_cards"`
	CreatedAt     string         `json:"created_at"`
	ModifiedAt    string         `json:"modified_at"`
	LastPlayed    string         `json:"last_played,omitempty"`
}

// DeckCardJSON represents a card in JSON deck format.
type DeckCardJSON struct {
	Quantity    int    `json:"quantity"`
	CardID      int    `json:"card_id"`
	CardName    string `json:"card_name"`
	ManaCost    string `json:"mana_cost,omitempty"`
	Type        string `json:"type,omitempty"`
	Rarity      string `json:"rarity,omitempty"`
	SetCode     string `json:"set_code,omitempty"`
	CollectorNo string `json:"collector_number,omitempty"`
}

// ExportDecks exports one or more decks to the specified format.
func ExportDecks(ctx context.Context, deckViewer *viewer.DeckViewer, deckIDs []string, deckFormat DeckFormat, outputPath string) error {
	if len(deckIDs) == 0 {
		return fmt.Errorf("no deck IDs provided")
	}

	// Retrieve deck data
	var allDecks []*models.DeckView
	for _, deckID := range deckIDs {
		deck, err := deckViewer.GetDeck(ctx, deckID)
		if err != nil {
			return fmt.Errorf("failed to get deck %s: %w", deckID, err)
		}
		if deck == nil {
			return fmt.Errorf("deck not found: %s", deckID)
		}
		allDecks = append(allDecks, deck)
	}

	switch deckFormat {
	case DeckFormatArena:
		return exportDecksArenaFormat(allDecks, outputPath)
	case DeckFormatText:
		return exportDecksTextFormat(allDecks, outputPath)
	case DeckFormatJSON:
		return exportDecksJSONFormat(allDecks, outputPath)
	case DeckFormatCSV:
		return exportDecksCSVFormat(allDecks, outputPath)
	default:
		return fmt.Errorf("unsupported deck format: %s", deckFormat)
	}
}

// exportDecksArenaFormat exports decks in MTG Arena format.
func exportDecksArenaFormat(decks []*models.DeckView, outputPath string) error {
	var content strings.Builder

	for i, deck := range decks {
		if i > 0 {
			content.WriteString("\n\n")
		}

		// Header comment with deck info
		content.WriteString("Deck\n")

		// Main deck
		for _, card := range deck.MainboardCards {
			if card.Metadata != nil && card.Metadata.Name != "" {
				content.WriteString(fmt.Sprintf("%d %s\n", card.Quantity, card.Metadata.Name))
			} else {
				content.WriteString(fmt.Sprintf("%d Card#%d\n", card.Quantity, card.CardID))
			}
		}

		// Sideboard (if any)
		if len(deck.SideboardCards) > 0 {
			content.WriteString("\n")
			for _, card := range deck.SideboardCards {
				if card.Metadata != nil && card.Metadata.Name != "" {
					content.WriteString(fmt.Sprintf("%d %s\n", card.Quantity, card.Metadata.Name))
				} else {
					content.WriteString(fmt.Sprintf("%d Card#%d\n", card.Quantity, card.CardID))
				}
			}
		}
	}

	return writeStringToFile(content.String(), outputPath)
}

// exportDecksTextFormat exports decks in simple text format.
func exportDecksTextFormat(decks []*models.DeckView, outputPath string) error {
	var content strings.Builder

	for i, deck := range decks {
		if i > 0 {
			content.WriteString("\n\n")
		}

		// Deck header
		content.WriteString(fmt.Sprintf("=== %s ===\n", deck.Deck.Name))
		content.WriteString(fmt.Sprintf("Format: %s\n", deck.Deck.Format))
		if deck.Deck.Description != nil && *deck.Deck.Description != "" {
			content.WriteString(fmt.Sprintf("Description: %s\n", *deck.Deck.Description))
		}
		if len(deck.ColorIdentity) > 0 {
			content.WriteString(fmt.Sprintf("Colors: %s\n", strings.Join(deck.ColorIdentity, "")))
		}
		totalCards := deck.TotalMainboard + deck.TotalSideboard
		content.WriteString(fmt.Sprintf("Total Cards: %d\n", totalCards))
		content.WriteString("\n")

		// Main deck
		content.WriteString("Main Deck:\n")
		for _, card := range deck.MainboardCards {
			cardInfo := formatCardInfo(card)
			content.WriteString(fmt.Sprintf("  %d %s\n", card.Quantity, cardInfo))
		}

		// Sideboard (if any)
		if len(deck.SideboardCards) > 0 {
			content.WriteString("\nSideboard:\n")
			for _, card := range deck.SideboardCards {
				cardInfo := formatCardInfo(card)
				content.WriteString(fmt.Sprintf("  %d %s\n", card.Quantity, cardInfo))
			}
		}

		// Deck statistics (if available)
		if totalCards > 0 {
			content.WriteString(fmt.Sprintf("\nMain Deck: %d cards\n", deck.TotalMainboard))
			content.WriteString(fmt.Sprintf("Sideboard: %d cards\n", deck.TotalSideboard))
		}
	}

	return writeStringToFile(content.String(), outputPath)
}

// exportDecksJSONFormat exports decks in JSON format.
func exportDecksJSONFormat(decks []*models.DeckView, outputPath string) error {
	jsonDecks := make([]DeckJSON, len(decks))

	for i, deck := range decks {
		totalCards := deck.TotalMainboard + deck.TotalSideboard
		jsonDeck := DeckJSON{
			ID:         deck.Deck.ID,
			Name:       deck.Deck.Name,
			Format:     deck.Deck.Format,
			TotalCards: totalCards,
			CreatedAt:  deck.Deck.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			ModifiedAt: deck.Deck.ModifiedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if deck.Deck.Description != nil {
			jsonDeck.Description = *deck.Deck.Description
		}
		if len(deck.ColorIdentity) > 0 {
			jsonDeck.ColorIdentity = strings.Join(deck.ColorIdentity, "")
		}
		if deck.Deck.LastPlayed != nil {
			jsonDeck.LastPlayed = deck.Deck.LastPlayed.Format("2006-01-02T15:04:05Z07:00")
		}

		// Main deck cards
		jsonDeck.MainDeck = make([]DeckCardJSON, len(deck.MainboardCards))
		for j, card := range deck.MainboardCards {
			cardJSON := DeckCardJSON{
				Quantity: card.Quantity,
				CardID:   card.CardID,
			}
			if card.Metadata != nil {
				cardJSON.CardName = card.Metadata.Name
				if card.Metadata.ManaCost != nil {
					cardJSON.ManaCost = *card.Metadata.ManaCost
				}
				cardJSON.Type = card.Metadata.TypeLine
				cardJSON.Rarity = card.Metadata.Rarity
				cardJSON.SetCode = card.Metadata.SetCode
				cardJSON.CollectorNo = card.Metadata.CollectorNumber
			}
			jsonDeck.MainDeck[j] = cardJSON
		}

		// Sideboard cards
		jsonDeck.Sideboard = make([]DeckCardJSON, len(deck.SideboardCards))
		for j, card := range deck.SideboardCards {
			cardJSON := DeckCardJSON{
				Quantity: card.Quantity,
				CardID:   card.CardID,
			}
			if card.Metadata != nil {
				cardJSON.CardName = card.Metadata.Name
				if card.Metadata.ManaCost != nil {
					cardJSON.ManaCost = *card.Metadata.ManaCost
				}
				cardJSON.Type = card.Metadata.TypeLine
				cardJSON.Rarity = card.Metadata.Rarity
				cardJSON.SetCode = card.Metadata.SetCode
				cardJSON.CollectorNo = card.Metadata.CollectorNumber
			}
			jsonDeck.Sideboard[j] = cardJSON
		}

		jsonDecks[i] = jsonDeck
	}

	opts := Options{
		Format:     FormatJSON,
		FilePath:   outputPath,
		PrettyJSON: true,
		Overwrite:  true,
	}

	exporter := NewExporter(opts)
	return exporter.Export(jsonDecks)
}

// exportDecksCSVFormat exports decks in CSV format (one row per card).
func exportDecksCSVFormat(decks []*models.DeckView, outputPath string) error {
	var rows []DeckExportRow

	for _, deck := range decks {
		// Main deck cards
		for _, card := range deck.MainboardCards {
			row := DeckExportRow{
				DeckID:     deck.Deck.ID,
				DeckName:   deck.Deck.Name,
				DeckFormat: deck.Deck.Format,
				Board:      "main",
				Quantity:   card.Quantity,
				CardID:     card.CardID,
			}
			if card.Metadata != nil {
				row.CardName = card.Metadata.Name
				if card.Metadata.ManaCost != nil {
					row.ManaCost = *card.Metadata.ManaCost
				}
				row.Type = card.Metadata.TypeLine
				row.Rarity = card.Metadata.Rarity
				row.SetCode = card.Metadata.SetCode
				row.CollectorNo = card.Metadata.CollectorNumber
			}
			rows = append(rows, row)
		}

		// Sideboard cards
		for _, card := range deck.SideboardCards {
			row := DeckExportRow{
				DeckID:     deck.Deck.ID,
				DeckName:   deck.Deck.Name,
				DeckFormat: deck.Deck.Format,
				Board:      "sideboard",
				Quantity:   card.Quantity,
				CardID:     card.CardID,
			}
			if card.Metadata != nil {
				row.CardName = card.Metadata.Name
				if card.Metadata.ManaCost != nil {
					row.ManaCost = *card.Metadata.ManaCost
				}
				row.Type = card.Metadata.TypeLine
				row.Rarity = card.Metadata.Rarity
				row.SetCode = card.Metadata.SetCode
				row.CollectorNo = card.Metadata.CollectorNumber
			}
			rows = append(rows, row)
		}
	}

	if len(rows) == 0 {
		return fmt.Errorf("no cards to export")
	}

	opts := Options{
		Format:    FormatCSV,
		FilePath:  outputPath,
		Overwrite: true,
	}

	exporter := NewExporter(opts)
	return exporter.Export(rows)
}

// ExportAllDecks exports all decks in the database for the specified account.
func ExportAllDecks(ctx context.Context, deckViewer *viewer.DeckViewer, accountID int, deckFormat DeckFormat, outputPath string) error {
	decks, err := deckViewer.ListDecks(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to list decks: %w", err)
	}

	if len(decks) == 0 {
		return fmt.Errorf("no decks found")
	}

	deckIDs := make([]string, len(decks))
	for i, deck := range decks {
		deckIDs[i] = deck.ID
	}

	return ExportDecks(ctx, deckViewer, deckIDs, deckFormat, outputPath)
}

// ExportDecksByFormat exports all decks of a specific format for the specified account.
func ExportDecksByFormat(ctx context.Context, deckViewer *viewer.DeckViewer, accountID int, format string, deckFormat DeckFormat, outputPath string) error {
	decks, err := deckViewer.GetDecksByFormat(ctx, accountID, format)
	if err != nil {
		return fmt.Errorf("failed to get decks by format: %w", err)
	}

	if len(decks) == 0 {
		return fmt.Errorf("no decks found for format: %s", format)
	}

	deckIDs := make([]string, len(decks))
	for i, deck := range decks {
		deckIDs[i] = deck.ID
	}

	return ExportDecks(ctx, deckViewer, deckIDs, deckFormat, outputPath)
}

// Helper functions

func formatCardInfo(card *models.DeckCardView) string {
	if card.Metadata == nil || card.Metadata.Name == "" {
		return fmt.Sprintf("Card#%d", card.CardID)
	}

	var parts []string
	parts = append(parts, card.Metadata.Name)

	if card.Metadata.ManaCost != nil && *card.Metadata.ManaCost != "" {
		parts = append(parts, fmt.Sprintf("(%s)", *card.Metadata.ManaCost))
	}

	if card.Metadata.TypeLine != "" {
		parts = append(parts, fmt.Sprintf("[%s]", card.Metadata.TypeLine))
	}

	return strings.Join(parts, " ")
}

func writeStringToFile(content string, outputPath string) error {
	opts := Options{
		Format:    Format("text"), // Custom format for plain text
		FilePath:  outputPath,
		Overwrite: true,
	}

	return (&Exporter{opts: opts}).writeToFile([]byte(content))
}

// GenerateDeckFilename generates a default filename for deck export.
func GenerateDeckFilename(deckName string, format DeckFormat) string {
	// Sanitize deck name for filename
	safeName := strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
			return '_'
		}
		return r
	}, deckName)

	extension := string(format)
	if format == DeckFormatArena || format == DeckFormatText {
		extension = "txt"
	}

	return fmt.Sprintf("%s.%s", safeName, extension)
}

// ExportToMTGAFormat exports a deck to MTG Arena format string.
func ExportToMTGAFormat(deck *models.Deck) string {
	// For Deck without cards, return minimal format
	var content strings.Builder
	content.WriteString("Deck\n")
	content.WriteString(fmt.Sprintf("// Name: %s\n", deck.Name))
	content.WriteString(fmt.Sprintf("// Format: %s\n", deck.Format))
	return content.String()
}

// ExportToTextFormat exports a deck to simple text format string.
func ExportToTextFormat(deck *models.Deck) string {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("=== %s ===\n", deck.Name))
	content.WriteString(fmt.Sprintf("Format: %s\n", deck.Format))
	if deck.Description != nil && *deck.Description != "" {
		content.WriteString(fmt.Sprintf("Description: %s\n", *deck.Description))
	}
	if deck.ColorIdentity != nil && *deck.ColorIdentity != "" {
		content.WriteString(fmt.Sprintf("Colors: %s\n", *deck.ColorIdentity))
	}
	return content.String()
}

// ExportDeckViewToMTGAFormat exports a DeckView to MTG Arena format string.
func ExportDeckViewToMTGAFormat(deck *models.DeckView) string {
	var content strings.Builder
	content.WriteString("Deck\n")

	// Main deck
	for _, card := range deck.MainboardCards {
		if card.Metadata != nil && card.Metadata.Name != "" {
			content.WriteString(fmt.Sprintf("%d %s\n", card.Quantity, card.Metadata.Name))
		} else {
			content.WriteString(fmt.Sprintf("%d Card#%d\n", card.Quantity, card.CardID))
		}
	}

	// Sideboard (if any)
	if len(deck.SideboardCards) > 0 {
		content.WriteString("\nSideboard\n")
		for _, card := range deck.SideboardCards {
			if card.Metadata != nil && card.Metadata.Name != "" {
				content.WriteString(fmt.Sprintf("%d %s\n", card.Quantity, card.Metadata.Name))
			} else {
				content.WriteString(fmt.Sprintf("%d Card#%d\n", card.Quantity, card.CardID))
			}
		}
	}

	return content.String()
}

// ExportDeckViewToTextFormat exports a DeckView to simple text format string.
func ExportDeckViewToTextFormat(deck *models.DeckView) string {
	var content strings.Builder

	// Deck header
	content.WriteString(fmt.Sprintf("=== %s ===\n", deck.Deck.Name))
	content.WriteString(fmt.Sprintf("Format: %s\n", deck.Deck.Format))
	if deck.Deck.Description != nil && *deck.Deck.Description != "" {
		content.WriteString(fmt.Sprintf("Description: %s\n", *deck.Deck.Description))
	}
	content.WriteString(fmt.Sprintf("Cards: %d mainboard", len(deck.MainboardCards)))
	if len(deck.SideboardCards) > 0 {
		content.WriteString(fmt.Sprintf(", %d sideboard", len(deck.SideboardCards)))
	}
	content.WriteString("\n\n")

	// Main deck
	content.WriteString("Main Deck:\n")
	for _, card := range deck.MainboardCards {
		content.WriteString(fmt.Sprintf("  %d %s\n", card.Quantity, formatCardInfo(card)))
	}

	// Sideboard (if any)
	if len(deck.SideboardCards) > 0 {
		content.WriteString("\nSideboard:\n")
		for _, card := range deck.SideboardCards {
			content.WriteString(fmt.Sprintf("  %d %s\n", card.Quantity, formatCardInfo(card)))
		}
	}

	return content.String()
}

// ExportDeckViewToArenaFormatWithSetCodes exports a DeckView to MTG Arena format with set codes.
// This format is more portable as it includes set information for card identification.
func ExportDeckViewToArenaFormatWithSetCodes(deck *models.DeckView) string {
	var content strings.Builder
	content.WriteString("Deck\n")

	// Main deck
	for _, card := range deck.MainboardCards {
		content.WriteString(formatArenaCardLine(card))
	}

	// Sideboard (if any)
	if len(deck.SideboardCards) > 0 {
		content.WriteString("\nSideboard\n")
		for _, card := range deck.SideboardCards {
			content.WriteString(formatArenaCardLine(card))
		}
	}

	return content.String()
}

// formatArenaCardLine formats a card line in Arena format: "4 Card Name (SET) 123"
func formatArenaCardLine(card *models.DeckCardView) string {
	if card.Metadata == nil || card.Metadata.Name == "" {
		return fmt.Sprintf("%d Card#%d\n", card.Quantity, card.CardID)
	}

	// Full Arena format: "4 Card Name (SET) 123"
	if card.Metadata.SetCode != "" && card.Metadata.CollectorNumber != "" {
		return fmt.Sprintf("%d %s (%s) %s\n",
			card.Quantity,
			card.Metadata.Name,
			strings.ToUpper(card.Metadata.SetCode),
			card.Metadata.CollectorNumber)
	}

	// Fallback to just name
	return fmt.Sprintf("%d %s\n", card.Quantity, card.Metadata.Name)
}

// MoxfieldDeckExport represents the structure for Moxfield import.
type MoxfieldDeckExport struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Format      string                 `json:"format,omitempty"`
	MainBoard   map[string]interface{} `json:"mainboard"`
	SideBoard   map[string]interface{} `json:"sideboard,omitempty"`
}

// GenerateMoxfieldURL generates a Moxfield import URL for a deck.
// Note: Moxfield doesn't have a direct URL import - we generate a deck string for clipboard import.
func GenerateMoxfieldURL(deck *models.DeckView) string {
	// Moxfield uses the same text format as Arena for import
	// Users can paste this on moxfield.com/decks/new
	return ExportDeckViewToArenaFormatWithSetCodes(deck)
}

// GenerateMoxfieldImportData generates JSON data for Moxfield API import.
func GenerateMoxfieldImportData(deck *models.DeckView) (string, error) {
	export := MoxfieldDeckExport{
		Name:      deck.Deck.Name,
		Format:    mapFormatToMoxfield(deck.Deck.Format),
		MainBoard: make(map[string]interface{}),
		SideBoard: make(map[string]interface{}),
	}

	if deck.Deck.Description != nil {
		export.Description = *deck.Deck.Description
	}

	// Main deck cards
	for _, card := range deck.MainboardCards {
		if card.Metadata != nil && card.Metadata.Name != "" {
			export.MainBoard[card.Metadata.Name] = card.Quantity
		}
	}

	// Sideboard cards
	for _, card := range deck.SideboardCards {
		if card.Metadata != nil && card.Metadata.Name != "" {
			export.SideBoard[card.Metadata.Name] = card.Quantity
		}
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Moxfield data: %w", err)
	}

	return string(data), nil
}

// mapFormatToMoxfield maps internal format names to Moxfield format names.
func mapFormatToMoxfield(format string) string {
	formatMap := map[string]string{
		"standard":      "standard",
		"historic":      "historic",
		"explorer":      "explorer",
		"pioneer":       "pioneer",
		"modern":        "modern",
		"legacy":        "legacy",
		"vintage":       "vintage",
		"pauper":        "pauper",
		"commander":     "commander",
		"brawl":         "brawl",
		"historicbrawl": "historicBrawl",
		"alchemy":       "alchemy",
		"timeless":      "timeless",
	}

	if mapped, ok := formatMap[strings.ToLower(format)]; ok {
		return mapped
	}
	return strings.ToLower(format)
}

// ArchidektDeckExport represents the structure for Archidekt import.
type ArchidektDeckExport struct {
	Name   string              `json:"name"`
	Format int                 `json:"format"`
	Cards  []ArchidektCardData `json:"cards"`
}

// ArchidektCardData represents a card for Archidekt import.
type ArchidektCardData struct {
	Quantity   int    `json:"quantity"`
	CardName   string `json:"name"`
	Categories string `json:"categories,omitempty"`
}

// GenerateArchidektURL generates an Archidekt import URL for a deck.
// Archidekt supports URL-based import via /decks/import with base64 deck string.
func GenerateArchidektURL(deck *models.DeckView) string {
	// Build deck string in simple format for URL encoding
	var deckStr strings.Builder

	for _, card := range deck.MainboardCards {
		if card.Metadata != nil && card.Metadata.Name != "" {
			deckStr.WriteString(fmt.Sprintf("%d %s\n", card.Quantity, card.Metadata.Name))
		}
	}

	if len(deck.SideboardCards) > 0 {
		deckStr.WriteString("\n// Sideboard\n")
		for _, card := range deck.SideboardCards {
			if card.Metadata != nil && card.Metadata.Name != "" {
				deckStr.WriteString(fmt.Sprintf("%d %s\n", card.Quantity, card.Metadata.Name))
			}
		}
	}

	// URL encode the deck string
	encoded := url.QueryEscape(deckStr.String())

	return fmt.Sprintf("https://archidekt.com/decks/new?deck=%s", encoded)
}

// GenerateArchidektImportData generates JSON data for Archidekt API import.
func GenerateArchidektImportData(deck *models.DeckView) (string, error) {
	export := ArchidektDeckExport{
		Name:   deck.Deck.Name,
		Format: mapFormatToArchidekt(deck.Deck.Format),
		Cards:  make([]ArchidektCardData, 0),
	}

	// Main deck cards
	for _, card := range deck.MainboardCards {
		if card.Metadata != nil && card.Metadata.Name != "" {
			export.Cards = append(export.Cards, ArchidektCardData{
				Quantity:   card.Quantity,
				CardName:   card.Metadata.Name,
				Categories: "Mainboard",
			})
		}
	}

	// Sideboard cards
	for _, card := range deck.SideboardCards {
		if card.Metadata != nil && card.Metadata.Name != "" {
			export.Cards = append(export.Cards, ArchidektCardData{
				Quantity:   card.Quantity,
				CardName:   card.Metadata.Name,
				Categories: "Sideboard",
			})
		}
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Archidekt data: %w", err)
	}

	return string(data), nil
}

// mapFormatToArchidekt maps internal format names to Archidekt format IDs.
func mapFormatToArchidekt(format string) int {
	// Archidekt format IDs
	formatMap := map[string]int{
		"standard":  1,
		"modern":    2,
		"legacy":    3,
		"vintage":   4,
		"commander": 5,
		"pauper":    6,
		"pioneer":   7,
		"historic":  8,
		"brawl":     9,
		"explorer":  10,
		"alchemy":   11,
		"timeless":  12,
	}

	if id, ok := formatMap[strings.ToLower(format)]; ok {
		return id
	}
	return 1 // Default to Standard
}

// DeckExportResponse represents the response from deck export API.
type DeckExportResponse struct {
	DeckID      string `json:"deck_id"`
	DeckName    string `json:"deck_name"`
	Format      string `json:"format"`
	ExportType  string `json:"export_type"`
	Content     string `json:"content"`
	URL         string `json:"url,omitempty"`
	FileName    string `json:"file_name,omitempty"`
	ContentType string `json:"content_type"`
}

// ExportDeckToFormat exports a deck to the specified format and returns the response.
func ExportDeckToFormat(deck *models.DeckView, exportFormat string) (*DeckExportResponse, error) {
	if deck == nil {
		return nil, fmt.Errorf("deck is required")
	}

	response := &DeckExportResponse{
		DeckID:     deck.Deck.ID,
		DeckName:   deck.Deck.Name,
		Format:     deck.Deck.Format,
		ExportType: exportFormat,
	}

	switch strings.ToLower(exportFormat) {
	case "arena":
		response.Content = ExportDeckViewToArenaFormatWithSetCodes(deck)
		response.ContentType = "text/plain"
		response.FileName = GenerateDeckFilename(deck.Deck.Name, DeckFormatArena)

	case "moxfield":
		response.Content = GenerateMoxfieldURL(deck)
		response.ContentType = "text/plain"
		response.FileName = fmt.Sprintf("%s_moxfield.txt", sanitizeFileName(deck.Deck.Name))

	case "archidekt":
		response.URL = GenerateArchidektURL(deck)
		response.Content = ExportDeckViewToArenaFormatWithSetCodes(deck)
		response.ContentType = "text/plain"
		response.FileName = fmt.Sprintf("%s_archidekt.txt", sanitizeFileName(deck.Deck.Name))

	case "json":
		jsonData, err := exportSingleDeckJSON(deck)
		if err != nil {
			return nil, err
		}
		response.Content = jsonData
		response.ContentType = "application/json"
		response.FileName = GenerateDeckFilename(deck.Deck.Name, DeckFormatJSON)

	case "text":
		response.Content = ExportDeckViewToTextFormat(deck)
		response.ContentType = "text/plain"
		response.FileName = GenerateDeckFilename(deck.Deck.Name, DeckFormatText)

	default:
		return nil, fmt.Errorf("unsupported export format: %s", exportFormat)
	}

	return response, nil
}

// exportSingleDeckJSON exports a single deck to JSON format.
func exportSingleDeckJSON(deck *models.DeckView) (string, error) {
	totalCards := deck.TotalMainboard + deck.TotalSideboard
	jsonDeck := DeckJSON{
		ID:         deck.Deck.ID,
		Name:       deck.Deck.Name,
		Format:     deck.Deck.Format,
		TotalCards: totalCards,
		CreatedAt:  deck.Deck.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		ModifiedAt: deck.Deck.ModifiedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if deck.Deck.Description != nil {
		jsonDeck.Description = *deck.Deck.Description
	}
	if len(deck.ColorIdentity) > 0 {
		jsonDeck.ColorIdentity = strings.Join(deck.ColorIdentity, "")
	}
	if deck.Deck.LastPlayed != nil {
		jsonDeck.LastPlayed = deck.Deck.LastPlayed.Format("2006-01-02T15:04:05Z07:00")
	}

	// Main deck cards
	jsonDeck.MainDeck = make([]DeckCardJSON, len(deck.MainboardCards))
	for j, card := range deck.MainboardCards {
		cardJSON := DeckCardJSON{
			Quantity: card.Quantity,
			CardID:   card.CardID,
		}
		if card.Metadata != nil {
			cardJSON.CardName = card.Metadata.Name
			if card.Metadata.ManaCost != nil {
				cardJSON.ManaCost = *card.Metadata.ManaCost
			}
			cardJSON.Type = card.Metadata.TypeLine
			cardJSON.Rarity = card.Metadata.Rarity
			cardJSON.SetCode = card.Metadata.SetCode
			cardJSON.CollectorNo = card.Metadata.CollectorNumber
		}
		jsonDeck.MainDeck[j] = cardJSON
	}

	// Sideboard cards
	jsonDeck.Sideboard = make([]DeckCardJSON, len(deck.SideboardCards))
	for j, card := range deck.SideboardCards {
		cardJSON := DeckCardJSON{
			Quantity: card.Quantity,
			CardID:   card.CardID,
		}
		if card.Metadata != nil {
			cardJSON.CardName = card.Metadata.Name
			if card.Metadata.ManaCost != nil {
				cardJSON.ManaCost = *card.Metadata.ManaCost
			}
			cardJSON.Type = card.Metadata.TypeLine
			cardJSON.Rarity = card.Metadata.Rarity
			cardJSON.SetCode = card.Metadata.SetCode
			cardJSON.CollectorNo = card.Metadata.CollectorNumber
		}
		jsonDeck.Sideboard[j] = cardJSON
	}

	data, err := json.MarshalIndent(jsonDeck, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal deck JSON: %w", err)
	}

	return string(data), nil
}

// sanitizeFileName removes invalid characters from a filename.
func sanitizeFileName(name string) string {
	return strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' || r == ' ' {
			return '_'
		}
		return r
	}, name)
}
