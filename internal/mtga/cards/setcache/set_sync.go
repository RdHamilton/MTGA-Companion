package setcache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// SetSyncer handles synchronization of set metadata from Scryfall and Standard legality from whatsinstandard.com.
type SetSyncer struct {
	scryfallClient *scryfall.Client
	storage        *storage.Service
	httpClient     *http.Client
}

// NewSetSyncer creates a new SetSyncer.
func NewSetSyncer(scryfallClient *scryfall.Client, storage *storage.Service) *SetSyncer {
	return &SetSyncer{
		scryfallClient: scryfallClient,
		storage:        storage,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// StandardSet represents a set from whatsinstandard.com API.
type StandardSet struct {
	Name     string `json:"name"`
	Codename string `json:"codename"`
	Code     string `json:"code"`
	Symbol   struct {
		Common     string `json:"common"`
		Uncommon   string `json:"uncommon"`
		Rare       string `json:"rare"`
		MythicRare string `json:"mythicRare"`
	} `json:"symbol"`
	EnterDate struct {
		Exact string `json:"exact"`
		Rough string `json:"rough"`
	} `json:"enterDate"`
	ExitDate struct {
		Exact string `json:"exact"`
		Rough string `json:"rough"`
	} `json:"exitDate"`
}

// StandardResponse represents the response from whatsinstandard.com API.
type StandardResponse struct {
	Deprecated bool          `json:"deprecated"`
	Sets       []StandardSet `json:"sets"`
}

// SyncSets fetches all sets from Scryfall and saves them to the database.
// Returns an error if any sets fail to save.
func (s *SetSyncer) SyncSets(ctx context.Context) error {
	log.Println("[SetSyncer] Fetching sets from Scryfall...")

	sets, err := s.scryfallClient.GetSets(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch sets from Scryfall: %w", err)
	}

	log.Printf("[SetSyncer] Fetched %d sets from Scryfall", len(sets.Data))

	// Filter to relevant set types (exclude tokens, memorabilia, etc.)
	relevantTypes := map[string]bool{
		"core":             true,
		"expansion":        true,
		"masters":          true,
		"draft_innovation": true,
		"commander":        true,
		"alchemy":          true,
		"starter":          true,
	}

	savedCount := 0
	failedCount := 0
	var lastErr error

	for _, scryfallSet := range sets.Data {
		// Skip digital-only sets except Alchemy
		if scryfallSet.Digital && scryfallSet.SetType != "alchemy" {
			continue
		}

		// Skip irrelevant set types
		if !relevantTypes[scryfallSet.SetType] {
			continue
		}

		set := &storage.Set{
			Code:       strings.ToUpper(scryfallSet.Code),
			Name:       scryfallSet.Name,
			SetType:    &scryfallSet.SetType,
			CardCount:  &scryfallSet.CardCount,
			IconSVGURI: &scryfallSet.IconSVGURI,
		}

		if scryfallSet.ReleasedAt != "" {
			set.ReleasedAt = &scryfallSet.ReleasedAt
		}

		if err := s.storage.SaveSet(ctx, set); err != nil {
			log.Printf("[SetSyncer] Failed to save set %s: %v", scryfallSet.Code, err)
			failedCount++
			lastErr = err
			continue
		}
		savedCount++
	}

	log.Printf("[SetSyncer] Saved %d sets to database (%d failed)", savedCount, failedCount)

	if failedCount > 0 {
		return fmt.Errorf("failed to save %d sets, last error: %w", failedCount, lastErr)
	}

	return nil
}

// SyncStandardLegality fetches Standard-legal sets from whatsinstandard.com and updates the database.
func (s *SetSyncer) SyncStandardLegality(ctx context.Context) error {
	log.Println("[SetSyncer] Fetching Standard legality from whatsinstandard.com...")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://whatsinstandard.com/api/v6/standard.json", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch Standard data: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var standardResp StandardResponse
	if err := json.NewDecoder(resp.Body).Decode(&standardResp); err != nil {
		return fmt.Errorf("failed to decode Standard response: %w", err)
	}

	// Check if API version is deprecated
	if standardResp.Deprecated {
		log.Println("[SetSyncer] Warning: whatsinstandard.com API v6 is deprecated, skipping Standard legality update")
		return nil
	}

	log.Printf("[SetSyncer] Fetched %d sets from whatsinstandard.com", len(standardResp.Sets))

	now := time.Now()
	standardCodes := make(map[string]string) // code -> rotation date

	for _, stdSet := range standardResp.Sets {
		if stdSet.Code == "" {
			continue // Skip sets without codes (unreleased)
		}

		code := strings.ToUpper(stdSet.Code)

		// Check if set has entered Standard
		if stdSet.EnterDate.Exact != "" {
			enterDate, err := time.Parse(time.RFC3339, stdSet.EnterDate.Exact)
			if err == nil && enterDate.After(now) {
				continue // Not yet in Standard
			}
		}

		// Check if set has exited Standard
		if stdSet.ExitDate.Exact != "" {
			exitDate, err := time.Parse(time.RFC3339, stdSet.ExitDate.Exact)
			if err == nil && exitDate.Before(now) {
				continue // Already rotated out
			}
			standardCodes[code] = stdSet.ExitDate.Exact
		} else {
			// No exact exit date, use rough date
			standardCodes[code] = stdSet.ExitDate.Rough
		}
	}

	log.Printf("[SetSyncer] Found %d Standard-legal sets", len(standardCodes))

	// Update Standard legality in database
	if err := s.storage.UpdateStandardLegality(ctx, standardCodes); err != nil {
		return fmt.Errorf("failed to update Standard legality: %w", err)
	}

	log.Println("[SetSyncer] Standard legality updated successfully")
	return nil
}

// SyncAll performs a full sync of sets and Standard legality.
func (s *SetSyncer) SyncAll(ctx context.Context) error {
	if err := s.SyncSets(ctx); err != nil {
		return fmt.Errorf("failed to sync sets: %w", err)
	}

	if err := s.SyncStandardLegality(ctx); err != nil {
		return fmt.Errorf("failed to sync Standard legality: %w", err)
	}

	return nil
}

// SyncIfEmpty syncs sets only if the sets table is empty.
func (s *SetSyncer) SyncIfEmpty(ctx context.Context) error {
	sets, err := s.storage.GetAllSets(ctx)
	if err != nil {
		return fmt.Errorf("failed to check existing sets: %w", err)
	}

	if len(sets) > 0 {
		log.Printf("[SetSyncer] Sets table has %d entries, skipping sync", len(sets))
		return nil
	}

	log.Println("[SetSyncer] Sets table is empty, performing initial sync...")
	return s.SyncAll(ctx)
}

// GetStandardSets returns all Standard-legal sets.
func (s *SetSyncer) GetStandardSets(ctx context.Context) ([]*storage.Set, error) {
	return s.storage.GetStandardSets(ctx)
}
