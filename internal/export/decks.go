package export

import (
	"context"
	"fmt"
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

// ExportAllDecks exports all decks in the database.
func ExportAllDecks(ctx context.Context, deckViewer *viewer.DeckViewer, deckFormat DeckFormat, outputPath string) error {
	decks, err := deckViewer.ListDecks(ctx)
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

// ExportDecksByFormat exports all decks of a specific format.
func ExportDecksByFormat(ctx context.Context, deckViewer *viewer.DeckViewer, format string, deckFormat DeckFormat, outputPath string) error {
	decks, err := deckViewer.GetDecksByFormat(ctx, format)
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
