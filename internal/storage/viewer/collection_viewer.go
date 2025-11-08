package viewer

import (
	"context"
	"database/sql"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// CollectionViewer provides card collection viewing with metadata.
type CollectionViewer struct {
	collectionRepo repository.CollectionRepository
	cardService    *cards.Service
}

// NewCollectionViewer creates a new collection viewer.
func NewCollectionViewer(db *sql.DB, cardService *cards.Service) *CollectionViewer {
	return &CollectionViewer{
		collectionRepo: repository.NewCollectionRepository(db),
		cardService:    cardService,
	}
}

// GetCollection retrieves the entire collection with card metadata.
func (cv *CollectionViewer) GetCollection(ctx context.Context) ([]*models.CollectionCardView, error) {
	// Get collection from repository
	collection, err := cv.collectionRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Extract card IDs
	cardIDs := make([]int, 0, len(collection))
	for cardID := range collection {
		cardIDs = append(cardIDs, cardID)
	}

	// Fetch metadata for all cards
	cardMetadata, err := cv.cardService.GetCards(cardIDs)
	if err != nil {
		// Continue even if some cards fail to load metadata
		cardMetadata = make(map[int]*cards.Card)
	}

	// Build collection card views
	views := make([]*models.CollectionCardView, 0, len(collection))
	for cardID, quantity := range collection {
		view := &models.CollectionCardView{
			CardID:   cardID,
			Quantity: quantity,
			Metadata: cardMetadata[cardID], // May be nil if not found
		}
		views = append(views, view)
	}

	return views, nil
}

// GetCardWithMetadata retrieves a specific card from the collection with metadata.
func (cv *CollectionViewer) GetCardWithMetadata(ctx context.Context, cardID int) (*models.CollectionCardView, error) {
	// Get quantity from collection
	quantity, err := cv.collectionRepo.GetCard(ctx, cardID)
	if err != nil {
		return nil, err
	}

	// Get metadata
	metadata, err := cv.cardService.GetCard(cardID)
	if err != nil {
		// Return card without metadata if lookup fails
		metadata = nil
	}

	return &models.CollectionCardView{
		CardID:   cardID,
		Quantity: quantity,
		Metadata: metadata,
	}, nil
}

// SearchByName searches the collection for cards by name.
// This requires card metadata to be available in the database.
func (cv *CollectionViewer) SearchByName(ctx context.Context, nameQuery string) ([]*models.CollectionCardView, error) {
	// Get entire collection
	views, err := cv.GetCollection(ctx)
	if err != nil {
		return nil, err
	}

	// Filter by name (case-insensitive substring match)
	var results []*models.CollectionCardView
	for _, view := range views {
		if view.Metadata != nil {
			// Simple substring match - could be enhanced with fuzzy matching
			if contains(view.Metadata.Name, nameQuery) {
				results = append(results, view)
			}
		}
	}

	return results, nil
}

// FilterByRarity filters the collection by rarity.
func (cv *CollectionViewer) FilterByRarity(ctx context.Context, rarity string) ([]*models.CollectionCardView, error) {
	views, err := cv.GetCollection(ctx)
	if err != nil {
		return nil, err
	}

	var results []*models.CollectionCardView
	for _, view := range views {
		if view.Metadata != nil && view.Metadata.Rarity == rarity {
			results = append(results, view)
		}
	}

	return results, nil
}

// FilterBySet filters the collection by set code.
func (cv *CollectionViewer) FilterBySet(ctx context.Context, setCode string) ([]*models.CollectionCardView, error) {
	views, err := cv.GetCollection(ctx)
	if err != nil {
		return nil, err
	}

	var results []*models.CollectionCardView
	for _, view := range views {
		if view.Metadata != nil && view.Metadata.SetCode == setCode {
			results = append(results, view)
		}
	}

	return results, nil
}

// FilterByColors filters the collection by color identity.
func (cv *CollectionViewer) FilterByColors(ctx context.Context, colors []string) ([]*models.CollectionCardView, error) {
	views, err := cv.GetCollection(ctx)
	if err != nil {
		return nil, err
	}

	var results []*models.CollectionCardView
	for _, view := range views {
		if view.Metadata != nil && matchesColors(view.Metadata.ColorIdentity, colors) {
			results = append(results, view)
		}
	}

	return results, nil
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	// Simple implementation - could use strings.Contains with ToLower
	return len(substr) > 0 && len(s) >= len(substr) &&
		containsIgnoreCase(s, substr)
}

func containsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive substring search
	sLower := toLower(s)
	substrLower := toLower(substr)

	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}

// matchesColors checks if a card's color identity matches the specified colors.
func matchesColors(cardColors, filterColors []string) bool {
	if len(filterColors) == 0 {
		return true
	}

	colorSet := make(map[string]bool)
	for _, color := range cardColors {
		colorSet[color] = true
	}

	for _, color := range filterColors {
		if !colorSet[color] {
			return false
		}
	}

	return true
}
