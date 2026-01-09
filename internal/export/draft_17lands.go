package export

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// SeventeenLandsDraftExport represents the 17Lands draft data format.
type SeventeenLandsDraftExport struct {
	DraftID   string                   `json:"draft_id"`
	EventType string                   `json:"event_type"`
	SetCode   string                   `json:"set_code"`
	DraftTime string                   `json:"draft_time"`
	Picks     []SeventeenLandsPickData `json:"picks"`
	FinalDeck []int                    `json:"final_deck,omitempty"`
	Sideboard []int                    `json:"sideboard,omitempty"`
	Metadata  *SeventeenLandsMetadata  `json:"metadata,omitempty"`
}

// SeventeenLandsPickData represents a single pick in 17Lands format.
type SeventeenLandsPickData struct {
	PackNumber int    `json:"pack_number"`
	PickNumber int    `json:"pick_number"`
	Pack       []int  `json:"pack"`
	Pick       int    `json:"pick"`
	PickTime   string `json:"pick_time"`
}

// SeventeenLandsMetadata contains additional metadata about the draft.
type SeventeenLandsMetadata struct {
	ExportedAt       string   `json:"exported_at"`
	ExportedFrom     string   `json:"exported_from"`
	OverallGrade     *string  `json:"overall_grade,omitempty"`
	OverallScore     *int     `json:"overall_score,omitempty"`
	PredictedWinRate *float64 `json:"predicted_win_rate,omitempty"`
}

// DraftExportData contains all the data needed to export a draft.
type DraftExportData struct {
	Session   *models.DraftSession
	Picks     []*models.DraftPickSession
	Packs     []*models.DraftPackSession
	FinalDeck []int // Arena card IDs
	Sideboard []int // Arena card IDs
}

// ExportDraftTo17Lands converts draft data to 17Lands JSON format.
func ExportDraftTo17Lands(data *DraftExportData) (*SeventeenLandsDraftExport, error) {
	if data == nil || data.Session == nil {
		return nil, fmt.Errorf("draft data is required")
	}

	session := data.Session

	// Generate draft ID from session ID
	draftID := session.ID

	// Normalize event type for 17Lands
	eventType := normalizeEventType(session.DraftType, session.EventName)

	export := &SeventeenLandsDraftExport{
		DraftID:   draftID,
		EventType: eventType,
		SetCode:   session.SetCode,
		DraftTime: session.StartTime.UTC().Format(time.RFC3339),
		Picks:     make([]SeventeenLandsPickData, 0),
		FinalDeck: data.FinalDeck,
		Sideboard: data.Sideboard,
		Metadata: &SeventeenLandsMetadata{
			ExportedAt:   time.Now().UTC().Format(time.RFC3339),
			ExportedFrom: "MTGA-Companion",
		},
	}

	// Add grade info if available
	if session.OverallGrade != nil {
		export.Metadata.OverallGrade = session.OverallGrade
	}
	if session.OverallScore != nil {
		export.Metadata.OverallScore = session.OverallScore
	}
	if session.PredictedWinRate != nil {
		export.Metadata.PredictedWinRate = session.PredictedWinRate
	}

	// Build a map of packs by (pack_number, pick_number) for quick lookup
	packMap := make(map[string][]int)
	for _, pack := range data.Packs {
		key := fmt.Sprintf("%d-%d", pack.PackNumber, pack.PickNumber)
		// Convert string card IDs to ints
		cardIDs := make([]int, 0, len(pack.CardIDs))
		for _, cardID := range pack.CardIDs {
			if id, err := strconv.Atoi(cardID); err == nil {
				cardIDs = append(cardIDs, id)
			}
		}
		packMap[key] = cardIDs
	}

	// Convert picks to 17Lands format
	for _, pick := range data.Picks {
		// 17Lands uses 1-based pack numbers, our DB uses 0-based
		packNum := pick.PackNumber + 1
		pickNum := pick.PickNumber + 1

		// Get the pack contents for this pick
		key := fmt.Sprintf("%d-%d", pick.PackNumber, pick.PickNumber)
		packContents := packMap[key]

		// Convert picked card ID to int
		pickedCardID, err := strconv.Atoi(pick.CardID)
		if err != nil {
			return nil, fmt.Errorf("invalid card ID '%s' for pick %d in pack %d: %w", pick.CardID, pickNum, packNum, err)
		}

		pickData := SeventeenLandsPickData{
			PackNumber: packNum,
			PickNumber: pickNum,
			Pack:       packContents,
			Pick:       pickedCardID,
			PickTime:   pick.Timestamp.UTC().Format(time.RFC3339),
		}

		export.Picks = append(export.Picks, pickData)
	}

	return export, nil
}

// ExportDraftToFile writes draft data to a JSON file.
func ExportDraftToFile(data *DraftExportData, outputPath string) error {
	export, err := ExportDraftTo17Lands(data)
	if err != nil {
		return fmt.Errorf("failed to convert draft data: %w", err)
	}

	jsonData, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, jsonData, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ExportDraftToJSON returns the draft data as a JSON string.
func ExportDraftToJSON(data *DraftExportData) (string, error) {
	export, err := ExportDraftTo17Lands(data)
	if err != nil {
		return "", fmt.Errorf("failed to convert draft data: %w", err)
	}

	jsonData, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonData), nil
}

// normalizeEventType converts internal event types to 17Lands format.
// 17Lands expects lowercase event types: quick, premier, traditional, sealed.
func normalizeEventType(draftType, eventName string) string {
	// Map common draft types to 17Lands conventions (lowercase)
	switch strings.ToLower(draftType) {
	case "quick_draft", "quickdraft", "quick":
		return "quick"
	case "premier_draft", "premierdraft", "premier":
		return "premier"
	case "traditional_draft", "traditionaldraft", "traddraft", "traditional":
		return "traditional"
	case "sealed":
		return "sealed"
	default:
		// Try to extract from event name
		eventLower := strings.ToLower(eventName)
		if strings.Contains(eventLower, "quick") {
			return "quick"
		}
		if strings.Contains(eventLower, "premier") {
			return "premier"
		}
		if strings.Contains(eventLower, "traditional") || strings.Contains(eventLower, "trad") {
			return "traditional"
		}
		if strings.Contains(eventLower, "sealed") {
			return "sealed"
		}
		// Default to lowercase version of the original value
		return strings.ToLower(draftType)
	}
}
