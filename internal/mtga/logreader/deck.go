package logreader

import "time"

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
// It looks for deck-related events in MTGA logs.
func ParseDecks(entries []*LogEntry) (*DeckLibrary, error) {
	library := &DeckLibrary{
		Decks:         make(map[string]*PlayerDeck),
		DecksByFormat: make(map[string][]*PlayerDeck),
	}

	seenDecks := make(map[string]bool)

	// Look for deck data in various event types
	// MTGA may store decks in different event types like "Deck.GetDeckLists" or similar
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if !entry.IsJSON {
			continue
		}

		// Try different possible deck event types
		var deckData interface{}
		var found bool

		// Check for deck list events
		if decks, ok := entry.JSON["Deck.GetDeckLists"]; ok {
			deckData = decks
			found = true
		} else if decks, ok := entry.JSON["getDeckLists"]; ok {
			deckData = decks
			found = true
		} else if decks, ok := entry.JSON["DeckLists"]; ok {
			deckData = decks
			found = true
		} else if decks, ok := entry.JSON["decks"]; ok {
			deckData = decks
			found = true
		}

		if !found {
			continue
		}

		// Parse deck list (could be array or object)
		decksArray, ok := deckData.([]interface{})
		if !ok {
			// Try as object with decks array
			if decksMap, ok := deckData.(map[string]interface{}); ok {
				if decks, ok := decksMap["decks"].([]interface{}); ok {
					decksArray = decks
				} else if decks, ok := decksMap["DeckLists"].([]interface{}); ok {
					decksArray = decks
				}
			}
		}

		if decksArray == nil {
			continue
		}

		// Process each deck
		for _, deckData := range decksArray {
			deckMap, ok := deckData.(map[string]interface{})
			if !ok {
				continue
			}

			deck := parsePlayerDeck(deckMap)
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

		// If we found decks, we can stop searching
		if len(library.Decks) > 0 {
			break
		}
	}

	library.TotalDecks = len(library.Decks)

	if library.TotalDecks == 0 {
		return nil, nil
	}

	return library, nil
}

// parsePlayerDeck extracts deck information from a deck object.
func parsePlayerDeck(deckMap map[string]interface{}) *PlayerDeck {
	deck := &PlayerDeck{
		MainDeck:  []DeckCard{},
		Sideboard: []DeckCard{},
	}

	// Extract deck ID
	if id, ok := deckMap["id"].(string); ok {
		deck.DeckID = id
	} else if id, ok := deckMap["Id"].(string); ok {
		deck.DeckID = id
	} else if id, ok := deckMap["deckId"].(string); ok {
		deck.DeckID = id
	} else if id, ok := deckMap["DeckId"].(string); ok {
		deck.DeckID = id
	}

	if deck.DeckID == "" {
		return nil
	}

	// Extract deck name
	if name, ok := deckMap["name"].(string); ok {
		deck.Name = name
	} else if name, ok := deckMap["Name"].(string); ok {
		deck.Name = name
	}

	// Extract format
	if format, ok := deckMap["format"].(string); ok {
		deck.Format = format
	} else if format, ok := deckMap["Format"].(string); ok {
		deck.Format = format
	} else {
		deck.Format = "Unknown"
	}

	// Extract description
	if desc, ok := deckMap["description"].(string); ok {
		deck.Description = desc
	} else if desc, ok := deckMap["Description"].(string); ok {
		deck.Description = desc
	}

	// Extract main deck
	if mainDeckData, ok := deckMap["mainDeck"].([]interface{}); ok {
		deck.MainDeck = parseDeckCards(mainDeckData)
	} else if mainDeckData, ok := deckMap["MainDeck"].([]interface{}); ok {
		deck.MainDeck = parseDeckCards(mainDeckData)
	}

	// Extract sideboard
	if sideboardData, ok := deckMap["sideboard"].([]interface{}); ok {
		deck.Sideboard = parseDeckCards(sideboardData)
	} else if sideboardData, ok := deckMap["Sideboard"].([]interface{}); ok {
		deck.Sideboard = parseDeckCards(sideboardData)
	}

	// Extract timestamps
	if created, ok := deckMap["created"].(string); ok {
		if t, err := time.Parse(time.RFC3339, created); err == nil {
			deck.Created = t
		}
	} else if created, ok := deckMap["Created"].(string); ok {
		if t, err := time.Parse(time.RFC3339, created); err == nil {
			deck.Created = t
		}
	}

	if modified, ok := deckMap["modified"].(string); ok {
		if t, err := time.Parse(time.RFC3339, modified); err == nil {
			deck.Modified = t
		}
	} else if modified, ok := deckMap["Modified"].(string); ok {
		if t, err := time.Parse(time.RFC3339, modified); err == nil {
			deck.Modified = t
		}
	}

	if lastPlayed, ok := deckMap["lastPlayed"].(string); ok {
		if t, err := time.Parse(time.RFC3339, lastPlayed); err == nil {
			deck.LastPlayed = &t
		}
	} else if lastPlayed, ok := deckMap["LastPlayed"].(string); ok {
		if t, err := time.Parse(time.RFC3339, lastPlayed); err == nil {
			deck.LastPlayed = &t
		}
	}

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
