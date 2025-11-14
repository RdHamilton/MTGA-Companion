package logreader

import (
	"testing"
	"time"
)

func TestParseRankUpdates(t *testing.T) {
	tests := []struct {
		name     string
		entries  []*LogEntry
		expected int
		validate func(t *testing.T, updates []*RankUpdate)
	}{
		{
			name: "parse constructed rank update",
			entries: []*LogEntry{
				{
					IsJSON:    true,
					Timestamp: "2025-11-14 10:30:00",
					JSON: map[string]interface{}{
						"RankUpdated": map[string]interface{}{
							"playerId":         "test-player-123",
							"seasonOrdinal":    float64(202511),
							"newClass":         "Gold",
							"oldClass":         "Silver",
							"newLevel":         float64(4),
							"oldLevel":         float64(1),
							"newStep":          float64(1),
							"oldStep":          float64(6),
							"wasLossProtected": false,
							"rankUpdateType":   "Constructed",
						},
					},
				},
			},
			expected: 1,
			validate: func(t *testing.T, updates []*RankUpdate) {
				if len(updates) != 1 {
					t.Fatalf("expected 1 update, got %d", len(updates))
				}
				u := updates[0]
				if u.PlayerID != "test-player-123" {
					t.Errorf("expected PlayerID test-player-123, got %s", u.PlayerID)
				}
				if u.SeasonOrdinal != 202511 {
					t.Errorf("expected SeasonOrdinal 202511, got %d", u.SeasonOrdinal)
				}
				if u.NewClass != "Gold" {
					t.Errorf("expected NewClass Gold, got %s", u.NewClass)
				}
				if u.OldClass != "Silver" {
					t.Errorf("expected OldClass Silver, got %s", u.OldClass)
				}
				if u.NewLevel != 4 {
					t.Errorf("expected NewLevel 4, got %d", u.NewLevel)
				}
				if u.OldLevel != 1 {
					t.Errorf("expected OldLevel 1, got %d", u.OldLevel)
				}
				if u.NewStep != 1 {
					t.Errorf("expected NewStep 1, got %d", u.NewStep)
				}
				if u.OldStep != 6 {
					t.Errorf("expected OldStep 6, got %d", u.OldStep)
				}
				if u.WasLossProtected {
					t.Error("expected WasLossProtected false")
				}
				if u.RankUpdateType != "Constructed" {
					t.Errorf("expected RankUpdateType Constructed, got %s", u.RankUpdateType)
				}
			},
		},
		{
			name: "parse limited rank update",
			entries: []*LogEntry{
				{
					IsJSON:    true,
					Timestamp: "2025-11-14 14:15:00",
					JSON: map[string]interface{}{
						"RankUpdated": map[string]interface{}{
							"playerId":         "test-player-456",
							"seasonOrdinal":    float64(202511),
							"newClass":         "Platinum",
							"oldClass":         "Gold",
							"newLevel":         float64(4),
							"oldLevel":         float64(1),
							"newStep":          float64(2),
							"oldStep":          float64(5),
							"wasLossProtected": true,
							"rankUpdateType":   "Limited",
						},
					},
				},
			},
			expected: 1,
			validate: func(t *testing.T, updates []*RankUpdate) {
				if len(updates) != 1 {
					t.Fatalf("expected 1 update, got %d", len(updates))
				}
				u := updates[0]
				if u.RankUpdateType != "Limited" {
					t.Errorf("expected RankUpdateType Limited, got %s", u.RankUpdateType)
				}
				if !u.WasLossProtected {
					t.Error("expected WasLossProtected true")
				}
			},
		},
		{
			name: "parse multiple rank updates",
			entries: []*LogEntry{
				{
					IsJSON:    true,
					Timestamp: "2025-11-14 10:00:00",
					JSON: map[string]interface{}{
						"RankUpdated": map[string]interface{}{
							"playerId":         "player-1",
							"seasonOrdinal":    float64(202511),
							"newClass":         "Bronze",
							"oldClass":         "Bronze",
							"newLevel":         float64(2),
							"oldLevel":         float64(3),
							"newStep":          float64(1),
							"oldStep":          float64(6),
							"wasLossProtected": false,
							"rankUpdateType":   "Constructed",
						},
					},
				},
				{
					IsJSON:    true,
					Timestamp: "2025-11-14 11:00:00",
					JSON: map[string]interface{}{
						"RankUpdated": map[string]interface{}{
							"playerId":         "player-1",
							"seasonOrdinal":    float64(202511),
							"newClass":         "Bronze",
							"oldClass":         "Bronze",
							"newLevel":         float64(1),
							"oldLevel":         float64(2),
							"newStep":          float64(1),
							"oldStep":          float64(6),
							"wasLossProtected": false,
							"rankUpdateType":   "Constructed",
						},
					},
				},
			},
			expected: 2,
			validate: func(t *testing.T, updates []*RankUpdate) {
				if len(updates) != 2 {
					t.Fatalf("expected 2 updates, got %d", len(updates))
				}
			},
		},
		{
			name: "skip non-JSON entries",
			entries: []*LogEntry{
				{
					IsJSON: false,
					Raw:    "Some text log entry",
				},
				{
					IsJSON:    true,
					Timestamp: "2025-11-14 10:00:00",
					JSON: map[string]interface{}{
						"RankUpdated": map[string]interface{}{
							"playerId":         "player-1",
							"seasonOrdinal":    float64(202511),
							"newClass":         "Bronze",
							"oldClass":         "Bronze",
							"newLevel":         float64(1),
							"oldLevel":         float64(1),
							"newStep":          float64(2),
							"oldStep":          float64(1),
							"wasLossProtected": false,
							"rankUpdateType":   "Constructed",
						},
					},
				},
			},
			expected: 1,
			validate: func(t *testing.T, updates []*RankUpdate) {
				if len(updates) != 1 {
					t.Fatalf("expected 1 update, got %d", len(updates))
				}
			},
		},
		{
			name: "skip incomplete rank updates",
			entries: []*LogEntry{
				{
					IsJSON:    true,
					Timestamp: "2025-11-14 10:00:00",
					JSON: map[string]interface{}{
						"RankUpdated": map[string]interface{}{
							"playerId":      "player-1",
							"seasonOrdinal": float64(202511),
							// Missing newClass and rankUpdateType
						},
					},
				},
			},
			expected: 0,
			validate: func(t *testing.T, updates []*RankUpdate) {
				if len(updates) != 0 {
					t.Fatalf("expected 0 updates, got %d", len(updates))
				}
			},
		},
		{
			name: "skip non-RankUpdated events",
			entries: []*LogEntry{
				{
					IsJSON:    true,
					Timestamp: "2025-11-14 10:00:00",
					JSON: map[string]interface{}{
						"SomeOtherEvent": map[string]interface{}{
							"data": "value",
						},
					},
				},
			},
			expected: 0,
			validate: func(t *testing.T, updates []*RankUpdate) {
				if len(updates) != 0 {
					t.Fatalf("expected 0 updates, got %d", len(updates))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updates, err := ParseRankUpdates(tt.entries)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(updates) != tt.expected {
				t.Errorf("expected %d updates, got %d", tt.expected, len(updates))
			}
			if tt.validate != nil {
				tt.validate(t, updates)
			}
		})
	}
}

func TestRankUpdate_FormatToDBFormat(t *testing.T) {
	tests := []struct {
		name           string
		rankUpdateType string
		expected       string
	}{
		{
			name:           "constructed to lowercase",
			rankUpdateType: "Constructed",
			expected:       "constructed",
		},
		{
			name:           "limited to lowercase",
			rankUpdateType: "Limited",
			expected:       "limited",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			update := &RankUpdate{
				RankUpdateType: tt.rankUpdateType,
			}
			result := update.FormatToDBFormat()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestParseRankUpdates_Timestamp(t *testing.T) {
	entries := []*LogEntry{
		{
			IsJSON:    true,
			Timestamp: "2025-11-14 15:30:45",
			JSON: map[string]interface{}{
				"RankUpdated": map[string]interface{}{
					"playerId":         "test-player",
					"seasonOrdinal":    float64(202511),
					"newClass":         "Gold",
					"oldClass":         "Silver",
					"newLevel":         float64(4),
					"oldLevel":         float64(1),
					"newStep":          float64(1),
					"oldStep":          float64(6),
					"wasLossProtected": false,
					"rankUpdateType":   "Constructed",
				},
			},
		},
	}

	updates, err := ParseRankUpdates(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(updates))
	}

	// Timestamp should be parsed
	if updates[0].Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}

	// Verify timestamp is reasonable (not in the distant past or future)
	now := time.Now()
	diff := now.Sub(updates[0].Timestamp)
	if diff < 0 {
		diff = -diff
	}
	// Should be within a reasonable time window (e.g., not more than 10 years off)
	if diff > 10*365*24*time.Hour {
		t.Errorf("timestamp seems incorrect: %v (diff from now: %v)", updates[0].Timestamp, diff)
	}
}
