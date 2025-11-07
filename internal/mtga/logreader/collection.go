package logreader

import "fmt"

// CardCollection represents the player's card collection.
type CardCollection struct {
	Cards           map[int]*Card // cardId -> Card
	TotalCards      int
	UniqueCards     int
	SetCompletion   map[string]*SetProgress
	RarityBreakdown map[string]int
}

// Card represents a single card in the collection.
type Card struct {
	CardID      int
	Name        string // Will be empty until we have metadata source
	Set         string // Will be empty until we have metadata source
	Rarity      string // Will be empty until we have metadata source
	Colors      []string
	Type        string
	Quantity    int
	MaxQuantity int // Usually 4, but could be different
}

// SetProgress tracks completion for a card set.
type SetProgress struct {
	SetCode          string
	SetName          string
	TotalCards       int
	OwnedCards       int
	CompletionPct    float64
	MissingCommons   int
	MissingUncommons int
	MissingRares     int
	MissingMythics   int
}

// ParseCollection extracts card collection information from log entries.
// It looks for InventoryInfo events that contain card collection data.
func ParseCollection(entries []*LogEntry) (*CardCollection, error) {
	// Look for the most recent InventoryInfo
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if !entry.IsJSON {
			continue
		}

		// Check if this is an InventoryInfo event
		if invInfo, ok := entry.JSON["InventoryInfo"]; ok {
			invMap, ok := invInfo.(map[string]interface{})
			if !ok {
				continue
			}

			// Check for card collection data
			// MTGA logs may have "Cards" or "CardInventory" field
			var cardsData interface{}
			var found bool

			// Try different possible field names and types
			if cards, ok := invMap["Cards"]; ok {
				cardsData = cards
				found = true
			} else if cards, ok := invMap["CardInventory"]; ok {
				cardsData = cards
				found = true
			} else if cards, ok := invMap["cardInventory"]; ok {
				cardsData = cards
				found = true
			}

			if !found {
				continue
			}

			collection := &CardCollection{
				Cards:           make(map[int]*Card),
				SetCompletion:   make(map[string]*SetProgress),
				RarityBreakdown: make(map[string]int),
			}

			// Parse cards from the collection data
			// The structure might be: map[string]interface{} where keys are card IDs
			// or it might be an array of card objects
			cardsMap, ok := cardsData.(map[string]interface{})
			if ok {
				// Cards are stored as a map: cardID -> quantity
				for cardIDStr, quantityData := range cardsMap {
					// Parse card ID from string key
					var cardID int
					if _, err := fmt.Sscanf(cardIDStr, "%d", &cardID); err != nil {
						continue
					}

					// Parse quantity
					var quantity int
					if qtyFloat, ok := quantityData.(float64); ok {
						quantity = int(qtyFloat)
					} else if qtyInt, ok := quantityData.(int); ok {
						quantity = qtyInt
					} else {
						continue
					}

					if cardID <= 0 || quantity <= 0 {
						continue
					}

					// Create card entry
					card := &Card{
						CardID:      cardID,
						Quantity:    quantity,
						MaxQuantity: 4, // Default to 4, can be updated with metadata
					}

					collection.Cards[cardID] = card
					collection.TotalCards += quantity
				}
			} else if cardsArray, ok := cardsData.([]interface{}); ok {
				// Cards might be stored as an array of objects
				for _, cardData := range cardsArray {
					cardMap, ok := cardData.(map[string]interface{})
					if !ok {
						continue
					}

					var cardID, quantity int
					if id, ok := cardMap["cardId"].(float64); ok {
						cardID = int(id)
					} else if id, ok := cardMap["CardId"].(float64); ok {
						cardID = int(id)
					} else {
						continue
					}

					if qty, ok := cardMap["quantity"].(float64); ok {
						quantity = int(qty)
					} else if qty, ok := cardMap["Quantity"].(float64); ok {
						quantity = int(qty)
					} else {
						continue
					}

					if cardID <= 0 || quantity <= 0 {
						continue
					}

					card := &Card{
						CardID:      cardID,
						Quantity:    quantity,
						MaxQuantity: 4,
					}

					collection.Cards[cardID] = card
					collection.TotalCards += quantity
				}
			}

			collection.UniqueCards = len(collection.Cards)

			// Calculate rarity breakdown (will be empty until we have metadata)
			// For now, we'll just count cards
			for _, card := range collection.Cards {
				if card.Rarity != "" {
					collection.RarityBreakdown[card.Rarity] += card.Quantity
				}
			}

			if len(collection.Cards) > 0 {
				return collection, nil
			}
		}
	}

	return nil, nil
}
