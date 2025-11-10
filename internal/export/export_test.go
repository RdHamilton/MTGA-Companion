package export

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type TestStruct struct {
	ID        int       `csv:"id"`
	Name      string    `csv:"name"`
	Value     float64   `csv:"value"`
	Active    bool      `csv:"active"`
	CreatedAt time.Time `csv:"created_at"`
	Pointer   *string   `csv:"pointer"`
}

func TestExportJSON(t *testing.T) {
	data := []TestStruct{
		{
			ID:        1,
			Name:      "Test1",
			Value:     10.5,
			Active:    true,
			CreatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			Pointer:   stringPtr("test"),
		},
		{
			ID:        2,
			Name:      "Test2",
			Value:     20.3,
			Active:    false,
			CreatedAt: time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
			Pointer:   nil,
		},
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.json")

	exporter := NewExporter(Options{
		Format:     FormatJSON,
		FilePath:   filePath,
		PrettyJSON: true,
		Overwrite:  true,
	})

	err := exporter.Export(data)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("Export file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	var result []TestStruct
	if err := json.Unmarshal(content, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 records, got %d", len(result))
	}

	if result[0].Name != "Test1" {
		t.Errorf("Expected Name 'Test1', got '%s'", result[0].Name)
	}
}

func TestExportCSV(t *testing.T) {
	data := []TestStruct{
		{
			ID:        1,
			Name:      "Test1",
			Value:     10.5,
			Active:    true,
			CreatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			Pointer:   stringPtr("test"),
		},
		{
			ID:        2,
			Name:      "Test2",
			Value:     20.3,
			Active:    false,
			CreatedAt: time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
			Pointer:   nil,
		},
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.csv")

	exporter := NewExporter(Options{
		Format:    FormatCSV,
		FilePath:  filePath,
		Overwrite: true,
	})

	err := exporter.Export(data)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("Export file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) < 3 { // header + 2 data rows + trailing newline
		t.Fatalf("Expected at least 3 lines, got %d", len(lines))
	}

	// Verify header
	if !strings.Contains(lines[0], "id") || !strings.Contains(lines[0], "name") {
		t.Errorf("CSV header missing expected fields: %s", lines[0])
	}

	// Verify first data row contains expected values
	if !strings.Contains(lines[1], "1") || !strings.Contains(lines[1], "Test1") {
		t.Errorf("CSV first row doesn't contain expected data: %s", lines[1])
	}
}

func TestExportToWriter_JSON(t *testing.T) {
	data := []TestStruct{
		{
			ID:    1,
			Name:  "Test1",
			Value: 10.5,
		},
	}

	var buf bytes.Buffer
	err := ExportToWriter(&buf, FormatJSON, data, true)
	if err != nil {
		t.Fatalf("ExportToWriter failed: %v", err)
	}

	var result []TestStruct
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 record, got %d", len(result))
	}
}

func TestExportToWriter_CSV(t *testing.T) {
	data := []TestStruct{
		{
			ID:    1,
			Name:  "Test1",
			Value: 10.5,
		},
		{
			ID:    2,
			Name:  "Test2",
			Value: 20.3,
		},
	}

	var buf bytes.Buffer
	err := ExportToWriter(&buf, FormatCSV, data, false)
	if err != nil {
		t.Fatalf("ExportToWriter failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(output, "\n")

	if len(lines) < 3 {
		t.Fatalf("Expected at least 3 lines, got %d", len(lines))
	}

	// Verify header
	if !strings.Contains(lines[0], "id") {
		t.Errorf("CSV header missing 'id': %s", lines[0])
	}
}

func TestExportOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.json")

	data := []TestStruct{{ID: 1, Name: "Test1"}}

	// First export
	exporter1 := NewExporter(Options{
		Format:    FormatJSON,
		FilePath:  filePath,
		Overwrite: true,
	})

	err := exporter1.Export(data)
	if err != nil {
		t.Fatalf("First export failed: %v", err)
	}

	// Second export without overwrite should fail
	exporter2 := NewExporter(Options{
		Format:    FormatJSON,
		FilePath:  filePath,
		Overwrite: false,
	})

	err = exporter2.Export(data)
	if err == nil {
		t.Fatal("Expected error when overwrite is false, got nil")
	}

	// Third export with overwrite should succeed
	exporter3 := NewExporter(Options{
		Format:    FormatJSON,
		FilePath:  filePath,
		Overwrite: true,
	})

	err = exporter3.Export(data)
	if err != nil {
		t.Fatalf("Third export with overwrite failed: %v", err)
	}
}

func TestGenerateFilename(t *testing.T) {
	filename := GenerateFilename("statistics", FormatCSV)
	if !strings.HasPrefix(filename, "statistics_") {
		t.Errorf("Expected filename to start with 'statistics_', got: %s", filename)
	}
	if !strings.HasSuffix(filename, ".csv") {
		t.Errorf("Expected filename to end with '.csv', got: %s", filename)
	}

	filename2 := GenerateFilename("decks", FormatJSON)
	if !strings.HasSuffix(filename2, ".json") {
		t.Errorf("Expected filename to end with '.json', got: %s", filename2)
	}
}

func TestExportEmptySlice(t *testing.T) {
	data := []TestStruct{}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.csv")

	exporter := NewExporter(Options{
		Format:   FormatCSV,
		FilePath: filePath,
	})

	err := exporter.Export(data)
	if err == nil {
		t.Fatal("Expected error for empty slice, got nil")
	}
}

func stringPtr(s string) *string {
	return &s
}
