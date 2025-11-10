package display

import (
	"context"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

func TestDraftPicksDisplayer_DisplayPicks(t *testing.T) {
	ctx := context.Background()
	displayer := NewDraftPicksDisplayer(nil) // nil card service for testing

	picks := []*models.DraftPick{
		{
			ID:             1,
			DraftEventID:   "draft-1",
			PackNumber:     1,
			PickNumber:     1,
			AvailableCards: []int{100, 200, 300, 400, 500},
			SelectedCard:   200,
			Timestamp:      time.Now(),
		},
		{
			ID:             2,
			DraftEventID:   "draft-1",
			PackNumber:     1,
			PickNumber:     2,
			AvailableCards: []int{100, 300, 400, 500},
			SelectedCard:   300,
			Timestamp:      time.Now(),
		},
		{
			ID:             3,
			DraftEventID:   "draft-1",
			PackNumber:     2,
			PickNumber:     1,
			AvailableCards: []int{600, 700, 800, 900},
			SelectedCard:   700,
			Timestamp:      time.Now(),
		},
	}

	// Should not error
	err := displayer.DisplayPicks(ctx, picks)
	if err != nil {
		t.Errorf("DisplayPicks failed: %v", err)
	}
}

func TestDraftPicksDisplayer_DisplayPicksEmpty(t *testing.T) {
	ctx := context.Background()
	displayer := NewDraftPicksDisplayer(nil)

	// Empty picks should not error
	err := displayer.DisplayPicks(ctx, []*models.DraftPick{})
	if err != nil {
		t.Errorf("DisplayPicks with empty picks failed: %v", err)
	}
}

func TestDraftPicksDisplayer_DisplayPicksCompact(t *testing.T) {
	ctx := context.Background()
	displayer := NewDraftPicksDisplayer(nil)

	picks := []*models.DraftPick{
		{
			PackNumber:     1,
			PickNumber:     1,
			AvailableCards: []int{100, 200, 300},
			SelectedCard:   200,
			Timestamp:      time.Now(),
		},
		{
			PackNumber:     1,
			PickNumber:     2,
			AvailableCards: []int{100, 300},
			SelectedCard:   300,
			Timestamp:      time.Now(),
		},
	}

	err := displayer.DisplayPicksCompact(ctx, picks)
	if err != nil {
		t.Errorf("DisplayPicksCompact failed: %v", err)
	}
}

func TestDraftPicksDisplayer_DisplayPicksSummary(t *testing.T) {
	ctx := context.Background()
	displayer := NewDraftPicksDisplayer(nil)

	picks := []*models.DraftPick{
		{
			PackNumber:     1,
			PickNumber:     1,
			AvailableCards: []int{100, 200, 300},
			SelectedCard:   200,
			Timestamp:      time.Now(),
		},
		{
			PackNumber:     2,
			PickNumber:     1,
			AvailableCards: []int{400, 500, 600},
			SelectedCard:   500,
			Timestamp:      time.Now(),
		},
		{
			PackNumber:     3,
			PickNumber:     1,
			AvailableCards: []int{700, 800, 900},
			SelectedCard:   800,
			Timestamp:      time.Now(),
		},
	}

	err := displayer.DisplayPicksSummary(ctx, picks)
	if err != nil {
		t.Errorf("DisplayPicksSummary failed: %v", err)
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "no truncation needed",
			input:    "Short",
			maxLen:   10,
			expected: "Short",
		},
		{
			name:     "exact length",
			input:    "Exactly10!",
			maxLen:   10,
			expected: "Exactly10!",
		},
		{
			name:     "needs truncation",
			input:    "This is a very long card name",
			maxLen:   15,
			expected: "This is a ve...",
		},
		{
			name:     "very short maxLen",
			input:    "Hello",
			maxLen:   3,
			expected: "Hel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestDraftPicksDisplayer_GetCardName(t *testing.T) {
	displayer := NewDraftPicksDisplayer(nil)

	// With nil card service, should return "Card #N"
	name := displayer.getCardName(12345)
	expected := "Card #12345"
	if name != expected {
		t.Errorf("getCardName(12345) = %q, want %q", name, expected)
	}
}
