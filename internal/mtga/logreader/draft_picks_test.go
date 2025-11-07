package logreader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDraftPicks(t *testing.T) {
	tests := []struct {
		name    string
		entries []*LogEntry
		want    []*DraftPicks
		wantNil bool
	}{
		{
			name: "no draft pick data",
			entries: []*LogEntry{
				{
					IsJSON: true,
					JSON: map[string]interface{}{
						"otherEvent": map[string]interface{}{},
					},
				},
			},
			wantNil: true,
		},
		{
			name: "draft pick with pack and pick numbers",
			entries: []*LogEntry{
				{
					IsJSON:    true,
					Timestamp: "[UnityCrossThreadLogger]2024-01-15 10:30:45",
					JSON: map[string]interface{}{
						"humanDraftEvent": map[string]interface{}{
							"CourseId":     "course-1",
							"SelfPack":     float64(1),
							"SelfPick":     float64(1),
							"PackCards":    []interface{}{float64(12345), float64(12346), float64(12347)},
							"SelectedCard": float64(12345),
						},
					},
				},
			},
			want: []*DraftPicks{
				{
					CourseID: "course-1",
					Picks: []DraftPick{
						{
							CourseID:       "course-1",
							PackNumber:     1,
							PickNumber:     1,
							AvailableCards: []int{12345, 12346, 12347},
							SelectedCard:   12345,
						},
					},
				},
			},
		},
		{
			name: "multiple picks for same course",
			entries: []*LogEntry{
				{
					IsJSON:    true,
					Timestamp: "[UnityCrossThreadLogger]2024-01-15 10:30:45",
					JSON: map[string]interface{}{
						"humanDraftEvent": map[string]interface{}{
							"CourseId":     "course-1",
							"SelfPack":     float64(1),
							"SelfPick":     float64(1),
							"PackCards":    []interface{}{float64(12345)},
							"SelectedCard": float64(12345),
						},
					},
				},
				{
					IsJSON:    true,
					Timestamp: "[UnityCrossThreadLogger]2024-01-15 10:31:00",
					JSON: map[string]interface{}{
						"humanDraftEvent": map[string]interface{}{
							"CourseId":     "course-1",
							"SelfPack":     float64(1),
							"SelfPick":     float64(2),
							"PackCards":    []interface{}{float64(12346)},
							"SelectedCard": float64(12346),
						},
					},
				},
			},
			want: []*DraftPicks{
				{
					CourseID: "course-1",
					Picks: []DraftPick{
						{
							CourseID:       "course-1",
							PackNumber:     1,
							PickNumber:     1,
							AvailableCards: []int{12345},
							SelectedCard:   12345,
						},
						{
							CourseID:       "course-1",
							PackNumber:     1,
							PickNumber:     2,
							AvailableCards: []int{12346},
							SelectedCard:   12346,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDraftPicks(tt.entries)
			if err != nil {
				t.Errorf("ParseDraftPicks() error = %v", err)
				return
			}

			if tt.wantNil {
				if got != nil {
					t.Errorf("ParseDraftPicks() expected nil, got %v", got)
				}
				return
			}

			if got == nil {
				t.Errorf("ParseDraftPicks() expected picks, got nil")
				return
			}

			// Check number of courses
			if len(got) != len(tt.want) {
				t.Errorf("ParseDraftPicks() course count = %d, want %d", len(got), len(tt.want))
			}

			// Check picks for each course
			for i, wantPicks := range tt.want {
				if i >= len(got) {
					t.Errorf("ParseDraftPicks() missing course %s", wantPicks.CourseID)
					continue
				}

				gotPicks := got[i]
				if gotPicks.CourseID != wantPicks.CourseID {
					t.Errorf("ParseDraftPicks() course ID = %s, want %s", gotPicks.CourseID, wantPicks.CourseID)
				}

				if len(gotPicks.Picks) != len(wantPicks.Picks) {
					t.Errorf("ParseDraftPicks() pick count = %d, want %d", len(gotPicks.Picks), len(wantPicks.Picks))
				}
			}
		})
	}
}

func TestParseDraftPicks_FromLogFile(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test_player.log")

	// Create test log data with draft picks
	testData := `[UnityCrossThreadLogger]{"humanDraftEvent":{"CourseId":"course-1","SelfPack":1,"SelfPick":1,"PackCards":[12345,12346,12347],"SelectedCard":12345}}
`
	if err := os.WriteFile(logPath, []byte(testData), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	// Read entries
	reader, err := NewReader(logPath)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			t.Errorf("Error closing reader: %v", err)
		}
	}()

	entries, err := reader.ReadAllJSON()
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}

	// Parse draft picks
	picks, err := ParseDraftPicks(entries)
	if err != nil {
		t.Fatalf("ParseDraftPicks() error = %v", err)
	}

	if picks == nil {
		t.Fatal("ParseDraftPicks() expected picks, got nil")
	}

	if len(picks) != 1 {
		t.Errorf("ParseDraftPicks() course count = %d, want 1", len(picks))
	}

	if len(picks[0].Picks) != 1 {
		t.Errorf("ParseDraftPicks() pick count = %d, want 1", len(picks[0].Picks))
	}
}
