package logreader

import (
	"strings"
	"time"
)

// DeckLibrary represents all player decks.
type DeckLibrary struct {
	Decks         map[string]*PlayerDeck
	TotalDecks    int
	DecksByFormat map[string][]*PlayerDeck
}

// PlayerDeck represents a single saved deck.
type PlayerDeck struct {
	DeckID      string
	Name        string
	Format      string
	Description string
	MainDeck    []DeckCard
	Sideboard   []DeckCard
	Created     time.Time
	Modified    time.Time
	LastPlayed  *time.Time
}

// ParseDecks extracts saved player decks from log entries.
// It looks for deck data in EventGetCoursesV2 responses (Courses array).
func ParseDecks(entries []*LogEntry) (*DeckLibrary, error) {
	library := &DeckLibrary{
		Decks:         make(map[string]*PlayerDeck),
		DecksByFormat: make(map[string][]*PlayerDeck),
	}

	seenDecks := make(map[string]bool)

	// Look for EventGetCoursesV2 responses which contain deck data in the Courses array
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if !entry.IsJSON {
			continue
		}

		// Check for Courses key (from EventGetCoursesV2 responses)
		coursesData, ok := entry.JSON["Courses"]
		if !ok {
			continue
		}

		// Parse courses array
		coursesArray, ok := coursesData.([]interface{})
		if !ok {
			continue
		}

		// Process each course (which may contain a deck)
		for _, courseData := range coursesArray {
			courseMap, ok := courseData.(map[string]interface{})
			if !ok {
				continue
			}

			// Extract deck from CourseDeckSummary and CourseDeck
			deck := parseCourseWithDeck(courseMap)
			if deck == nil || deck.DeckID == "" {
				continue
			}

			// Skip if already seen
			if seenDecks[deck.DeckID] {
				continue
			}
			seenDecks[deck.DeckID] = true

			library.Decks[deck.DeckID] = deck
			library.DecksByFormat[deck.Format] = append(library.DecksByFormat[deck.Format], deck)
		}
	}

	library.TotalDecks = len(library.Decks)

	if library.TotalDecks == 0 {
		return nil, nil
	}

	return library, nil
}

// parseCourseWithDeck extracts deck information from a Course object (from EventGetCoursesV2).
// The structure is: Course -> CourseDeckSummary (metadata) + CourseDeck (cards).
func parseCourseWithDeck(courseMap map[string]interface{}) *PlayerDeck {
	deck := &PlayerDeck{
		MainDeck:  []DeckCard{},
		Sideboard: []DeckCard{},
	}

	// Extract CourseDeckSummary for metadata
	summaryData, ok := courseMap["CourseDeckSummary"]
	if !ok {
		return nil
	}

	summaryMap, ok := summaryData.(map[string]interface{})
	if !ok {
		return nil
	}

	// Extract DeckId
	if id, ok := summaryMap["DeckId"].(string); ok {
		deck.DeckID = id
	}

	if deck.DeckID == "" {
		return nil
	}

	// Extract Name (clean up localization keys if present)
	if name, ok := summaryMap["Name"].(string); ok {
		deck.Name = cleanDeckName(name)
	}

	// Extract Description
	if desc, ok := summaryMap["Description"].(string); ok {
		deck.Description = desc
	}

	// Extract Attributes array for format and timestamps
	if attrsData, ok := summaryMap["Attributes"].([]interface{}); ok {
		for _, attrData := range attrsData {
			attrMap, ok := attrData.(map[string]interface{})
			if !ok {
				continue
			}

			name, _ := attrMap["name"].(string)
			value, _ := attrMap["value"].(string)

			switch name {
			case "Format":
				deck.Format = value
			case "LastPlayed":
				// Value is quoted: "\"2024-06-21T09:35:17...\""
				// Remove outer quotes
				if len(value) > 2 && value[0] == '"' && value[len(value)-1] == '"' {
					value = value[1 : len(value)-1]
				}
				if t, err := time.Parse(time.RFC3339, value); err == nil {
					deck.LastPlayed = &t
				}
			case "LastUpdated":
				// Value is quoted: "\"2024-05-10T09:54:31...\""
				if len(value) > 2 && value[0] == '"' && value[len(value)-1] == '"' {
					value = value[1 : len(value)-1]
				}
				if t, err := time.Parse(time.RFC3339, value); err == nil {
					deck.Modified = t
				}
			}
		}
	}

	// Default format if not found
	if deck.Format == "" {
		deck.Format = "Unknown"
	}

	// Extract CourseDeck for card lists
	deckData, ok := courseMap["CourseDeck"]
	if !ok {
		return deck // Return deck with metadata even if cards missing
	}

	deckMap, ok := deckData.(map[string]interface{})
	if !ok {
		return deck
	}

	// Extract MainDeck
	if mainDeckData, ok := deckMap["MainDeck"].([]interface{}); ok {
		deck.MainDeck = parseDeckCards(mainDeckData)
	}

	// Extract Sideboard
	if sideboardData, ok := deckMap["Sideboard"].([]interface{}); ok {
		deck.Sideboard = parseDeckCards(sideboardData)
	}

	// Note: CommandZone and Companions are also available but not currently stored

	return deck
}

// parseDeckCards parses an array of card objects into DeckCard slice.
func parseDeckCards(cardsData []interface{}) []DeckCard {
	var cards []DeckCard

	for _, cardData := range cardsData {
		cardMap, ok := cardData.(map[string]interface{})
		if !ok {
			continue
		}

		card := DeckCard{}

		// Extract card ID
		if cardID, ok := cardMap["cardId"].(float64); ok {
			card.CardID = int(cardID)
		} else if cardID, ok := cardMap["CardId"].(float64); ok {
			card.CardID = int(cardID)
		} else if cardID, ok := cardMap["card_id"].(float64); ok {
			card.CardID = int(cardID)
		}

		// Extract quantity
		if quantity, ok := cardMap["quantity"].(float64); ok {
			card.Quantity = int(quantity)
		} else if quantity, ok := cardMap["Quantity"].(float64); ok {
			card.Quantity = int(quantity)
		}

		if card.CardID > 0 && card.Quantity > 0 {
			cards = append(cards, card)
		}
	}

	return cards
}

// cleanDeckName converts MTGA localization keys to readable deck names.
// Example: "?=?Loc/Decks/Precon/Precon_EPP2024_UW" -> "Precon EPP2024 UW"
func cleanDeckName(name string) string {
	// Check for localization key pattern
	if !strings.HasPrefix(name, "?=?Loc/") {
		return name
	}

	// Extract the last path segment: "?=?Loc/Decks/Precon/Precon_EPP2024_UW" -> "Precon_EPP2024_UW"
	lastSlash := strings.LastIndex(name, "/")
	if lastSlash == -1 || lastSlash >= len(name)-1 {
		return name
	}

	identifier := name[lastSlash+1:]

	// Replace underscores with spaces
	return strings.ReplaceAll(identifier, "_", " ")
}
