package logreader

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestLogEntry_parseJSON(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantJSON  bool
		wantField string // A field we expect to find in the JSON
	}{
		{
			name:     "plain text line",
			raw:      "[UnityCrossThreadLogger]2024-01-15 10:30:45",
			wantJSON: false,
		},
		{
			name:      "line with JSON object",
			raw:       `[UnityCrossThreadLogger]2024-01-15 10:30:45 {"foo":"bar","count":42}`,
			wantJSON:  true,
			wantField: "foo",
		},
		{
			name:      "JSON object only",
			raw:       `{"type":"Event","data":"test"}`,
			wantJSON:  true,
			wantField: "type",
		},
		{
			name:     "empty line",
			raw:      "",
			wantJSON: false,
		},
		{
			name:     "line with JSON array",
			raw:      `[UnityCrossThreadLogger][{"id":1},{"id":2}]`,
			wantJSON: false, // Our current implementation expects objects, not arrays
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &LogEntry{Raw: tt.raw}
			entry.parseJSON()

			if entry.IsJSON != tt.wantJSON {
				t.Errorf("parseJSON() IsJSON = %v, want %v", entry.IsJSON, tt.wantJSON)
			}

			if tt.wantJSON && tt.wantField != "" {
				if _, ok := entry.JSON[tt.wantField]; !ok {
					t.Errorf("parseJSON() missing expected field %q in JSON %v", tt.wantField, entry.JSON)
				}
			}
		})
	}
}

func TestReader(t *testing.T) {
	// Create a temporary test log file
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test_player.log")

	// Write test log data
	testData := `[UnityCrossThreadLogger]Test log start
[UnityCrossThreadLogger]2024-01-15 10:30:45 {"type":"GameStart","eventId":1}
Plain text line without JSON
[UnityCrossThreadLogger]{"type":"GameEnd","eventId":2,"result":"win"}
`
	if err := os.WriteFile(logPath, []byte(testData), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	t.Run("NewReader", func(t *testing.T) {
		r, err := NewReader(logPath)
		if err != nil {
			t.Fatalf("NewReader() error = %v", err)
		}
		defer r.Close()

		if r == nil {
			t.Fatal("NewReader() returned nil reader")
		}
	})

	t.Run("ReadEntry", func(t *testing.T) {
		r, err := NewReader(logPath)
		if err != nil {
			t.Fatalf("NewReader() error = %v", err)
		}
		defer r.Close()

		// Read first entry
		entry, err := r.ReadEntry()
		if err != nil {
			t.Fatalf("ReadEntry() error = %v", err)
		}
		if entry == nil {
			t.Fatal("ReadEntry() returned nil entry")
		}
	})

	t.Run("ReadAll", func(t *testing.T) {
		r, err := NewReader(logPath)
		if err != nil {
			t.Fatalf("NewReader() error = %v", err)
		}
		defer r.Close()

		entries, err := r.ReadAll()
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}

		if len(entries) != 4 {
			t.Errorf("ReadAll() returned %d entries, want 4", len(entries))
		}
	})

	t.Run("ReadAllJSON", func(t *testing.T) {
		r, err := NewReader(logPath)
		if err != nil {
			t.Fatalf("NewReader() error = %v", err)
		}
		defer r.Close()

		entries, err := r.ReadAllJSON()
		if err != nil {
			t.Fatalf("ReadAllJSON() error = %v", err)
		}

		// Should only get the 2 lines with valid JSON
		if len(entries) != 2 {
			t.Errorf("ReadAllJSON() returned %d entries, want 2", len(entries))
		}

		// Verify all returned entries are JSON
		for i, entry := range entries {
			if !entry.IsJSON {
				t.Errorf("ReadAllJSON() entry[%d] IsJSON = false, want true", i)
			}
		}
	})

	t.Run("EOF handling", func(t *testing.T) {
		r, err := NewReader(logPath)
		if err != nil {
			t.Fatalf("NewReader() error = %v", err)
		}
		defer r.Close()

		// Read all entries
		for {
			_, err := r.ReadEntry()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		}

		// Next read should also return EOF
		_, err = r.ReadEntry()
		if err != io.EOF {
			t.Errorf("Expected io.EOF, got %v", err)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := NewReader("/nonexistent/file.log")
		if err == nil {
			t.Fatal("NewReader() should return error for nonexistent file")
		}
	})
}
