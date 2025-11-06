package logreader

import (
	"runtime"
	"strings"
	"testing"
)

func TestDefaultLogPath(t *testing.T) {
	path, err := DefaultLogPath()
	if err != nil {
		t.Fatalf("DefaultLogPath() returned error: %v", err)
	}

	if path == "" {
		t.Fatal("DefaultLogPath() returned empty path")
	}

	// Verify path contains platform-specific components
	switch runtime.GOOS {
	case "darwin":
		if !strings.Contains(path, "Library/Application Support/com.wizards.mtga") {
			t.Errorf("macOS path does not contain expected directory: %s", path)
		}
		// Should end with either UTC_Log*.log or Player.log
		if !strings.HasSuffix(path, ".log") {
			t.Errorf("path does not end with .log: %s", path)
		}
		if !strings.Contains(path, "UTC_Log") && !strings.HasSuffix(path, "Player.log") {
			t.Errorf("path should be UTC_Log or Player.log: %s", path)
		}

	case "windows":
		if !strings.Contains(path, "AppData") || !strings.Contains(path, "Wizards Of The Coast") {
			t.Errorf("Windows path does not contain expected directory: %s", path)
		}
		// Should end with either UTC_Log*.log or Player.log
		if !strings.HasSuffix(path, ".log") {
			t.Errorf("path does not end with .log: %s", path)
		}
		if !strings.Contains(path, "UTC_Log") && !strings.HasSuffix(path, "Player.log") {
			t.Errorf("path should be UTC_Log or Player.log: %s", path)
		}
	}
}

func TestLogExists(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() string
		cleanup func(string)
		want    bool
		wantErr bool
	}{
		{
			name: "nonexistent file",
			setup: func() string {
				return "/nonexistent/path/to/Player.log"
			},
			cleanup: func(s string) {},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			defer tt.cleanup(path)

			exists, err := LogExists(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LogExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if exists != tt.want {
				t.Errorf("LogExists() = %v, want %v", exists, tt.want)
			}
		})
	}
}
