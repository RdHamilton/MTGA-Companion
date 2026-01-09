package export

import (
	"strings"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

func TestExportDraftTo17Lands_NilData(t *testing.T) {
	_, err := ExportDraftTo17Lands(nil)
	if err == nil {
		t.Error("Expected error for nil data")
	}
}

func TestExportDraftTo17Lands_NilSession(t *testing.T) {
	data := &DraftExportData{}
	_, err := ExportDraftTo17Lands(data)
	if err == nil {
		t.Error("Expected error for nil session")
	}
}

func TestExportDraftTo17Lands_BasicExport(t *testing.T) {
	startTime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)

	data := &DraftExportData{
		Session: &models.DraftSession{
			ID:        "test-session-123",
			SetCode:   "TLA",
			DraftType: "QuickDraft",
			EventName: "QuickDraft_TLA",
			StartTime: startTime,
		},
		Picks: []*models.DraftPickSession{
			{
				ID:         1,
				SessionID:  "test-session-123",
				PackNumber: 0,
				PickNumber: 1,
				CardID:     "12345",
				Timestamp:  startTime.Add(1 * time.Minute),
			},
			{
				ID:         2,
				SessionID:  "test-session-123",
				PackNumber: 0,
				PickNumber: 2,
				CardID:     "12346",
				Timestamp:  startTime.Add(2 * time.Minute),
			},
		},
		Packs: []*models.DraftPackSession{
			{
				ID:         1,
				SessionID:  "test-session-123",
				PackNumber: 0,
				PickNumber: 1,
				CardIDs:    []string{"12345", "12347", "12348"},
			},
			{
				ID:         2,
				SessionID:  "test-session-123",
				PackNumber: 0,
				PickNumber: 2,
				CardIDs:    []string{"12346", "12349"},
			},
		},
	}

	result, err := ExportDraftTo17Lands(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify basic fields
	if result.DraftID != "test-session-123" {
		t.Errorf("Expected draft ID 'test-session-123', got '%s'", result.DraftID)
	}

	if result.SetCode != "TLA" {
		t.Errorf("Expected set code 'TLA', got '%s'", result.SetCode)
	}

	if result.EventType != "QuickDraft" {
		t.Errorf("Expected event type 'QuickDraft', got '%s'", result.EventType)
	}

	// Verify picks were converted (0-based to 1-based)
	if len(result.Picks) != 2 {
		t.Fatalf("Expected 2 picks, got %d", len(result.Picks))
	}

	// First pick should have pack 1, pick 2 (1-based: 0+1=1, 1+1=2)
	if result.Picks[0].PackNumber != 1 {
		t.Errorf("Expected pack number 1, got %d", result.Picks[0].PackNumber)
	}
	if result.Picks[0].PickNumber != 2 {
		t.Errorf("Expected pick number 2 (0-based 1 + 1), got %d", result.Picks[0].PickNumber)
	}
	if result.Picks[0].Pick != 12345 {
		t.Errorf("Expected picked card 12345, got %d", result.Picks[0].Pick)
	}

	// Verify metadata
	if result.Metadata == nil {
		t.Error("Expected metadata to be set")
	} else {
		if result.Metadata.ExportedFrom != "MTGA-Companion" {
			t.Errorf("Expected exported from 'MTGA-Companion', got '%s'", result.Metadata.ExportedFrom)
		}
	}
}

func TestExportDraftTo17Lands_WithGradeInfo(t *testing.T) {
	grade := "A"
	score := 85
	winRate := 0.65

	data := &DraftExportData{
		Session: &models.DraftSession{
			ID:               "test-session-456",
			SetCode:          "DSK",
			DraftType:        "PremierDraft",
			EventName:        "PremierDraft_DSK",
			StartTime:        time.Now(),
			OverallGrade:     &grade,
			OverallScore:     &score,
			PredictedWinRate: &winRate,
		},
	}

	result, err := ExportDraftTo17Lands(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Metadata.OverallGrade == nil || *result.Metadata.OverallGrade != "A" {
		t.Errorf("Expected overall grade 'A', got %v", result.Metadata.OverallGrade)
	}

	if result.Metadata.OverallScore == nil || *result.Metadata.OverallScore != 85 {
		t.Errorf("Expected overall score 85, got %v", result.Metadata.OverallScore)
	}

	if result.Metadata.PredictedWinRate == nil || *result.Metadata.PredictedWinRate != 0.65 {
		t.Errorf("Expected predicted win rate 0.65, got %v", result.Metadata.PredictedWinRate)
	}
}

func TestNormalizeEventType(t *testing.T) {
	tests := []struct {
		draftType string
		eventName string
		expected  string
	}{
		{"quick_draft", "", "QuickDraft"},
		{"QuickDraft", "", "QuickDraft"},
		{"premier_draft", "", "PremierDraft"},
		{"PremierDraft", "", "PremierDraft"},
		{"traditional_draft", "", "TradDraft"},
		{"TraditionalDraft", "", "TradDraft"},
		{"sealed", "", "Sealed"},
		{"Sealed", "", "Sealed"},
		{"unknown", "QuickDraft_TLA", "QuickDraft"},
		{"unknown", "PremierDraft_DSK", "PremierDraft"},
		{"unknown", "TraditionalDraft_FDN", "TradDraft"},
		{"unknown", "TradDraft", "TradDraft"},
		{"unknown", "SomeOtherEvent", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.draftType+"_"+tt.eventName, func(t *testing.T) {
			result := normalizeEventType(tt.draftType, tt.eventName)
			if result != tt.expected {
				t.Errorf("normalizeEventType(%q, %q) = %q, want %q",
					tt.draftType, tt.eventName, result, tt.expected)
			}
		})
	}
}

func TestExportDraftToJSON(t *testing.T) {
	data := &DraftExportData{
		Session: &models.DraftSession{
			ID:        "json-test",
			SetCode:   "TLA",
			DraftType: "QuickDraft",
			StartTime: time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
		},
	}

	jsonStr, err := ExportDraftToJSON(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify it's valid JSON (contains expected fields)
	if jsonStr == "" {
		t.Error("Expected non-empty JSON string")
	}

	// Check for expected fields in JSON
	if !strings.Contains(jsonStr, "draft_id") {
		t.Error("JSON should contain draft_id field")
	}
	if !strings.Contains(jsonStr, "set_code") {
		t.Error("JSON should contain set_code field")
	}
	if !strings.Contains(jsonStr, "TLA") {
		t.Error("JSON should contain set code value 'TLA'")
	}
}
