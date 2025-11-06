package logreader

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// LogEntry represents a parsed line from the Player.log file.
// The log contains JSON data mixed with plain text timestamps and metadata.
type LogEntry struct {
	Raw       string                 // The original raw line from the log
	Timestamp string                 // Extracted timestamp if present
	JSON      map[string]interface{} // Parsed JSON data if the line contains valid JSON
	IsJSON    bool                   // Whether this line contains JSON data
}

// Reader reads and parses MTGA Player.log files.
type Reader struct {
	file    *os.File
	scanner *bufio.Scanner
}

// NewReader creates a new Reader for the given log file path.
func NewReader(path string) (*Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	return &Reader{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

// Close closes the underlying log file.
func (r *Reader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// ReadEntry reads the next log entry from the file.
// It returns io.EOF when there are no more entries to read.
func (r *Reader) ReadEntry() (*LogEntry, error) {
	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return nil, fmt.Errorf("scan log file: %w", err)
		}
		return nil, io.EOF
	}

	line := r.scanner.Text()
	entry := &LogEntry{
		Raw: line,
	}

	// Try to parse as JSON
	// MTGA logs often have format: [UnityCrossThreadLogger]timestamp JSON
	// We need to extract the JSON portion
	entry.parseJSON()

	return entry, nil
}

// parseJSON attempts to parse JSON data from the log line.
func (e *LogEntry) parseJSON() {
	line := e.Raw

	// Look for JSON content - typically starts with { or [
	jsonStart := strings.Index(line, "{")
	if jsonStart == -1 {
		jsonStart = strings.Index(line, "[")
	}

	if jsonStart == -1 {
		// No JSON in this line
		e.IsJSON = false
		return
	}

	// Extract potential timestamp/prefix before JSON
	if jsonStart > 0 {
		e.Timestamp = strings.TrimSpace(line[:jsonStart])
	}

	// Try to parse the JSON portion
	jsonData := line[jsonStart:]
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err == nil {
		e.JSON = data
		e.IsJSON = true
	}
}

// ReadAll reads all entries from the log file.
func (r *Reader) ReadAll() ([]*LogEntry, error) {
	var entries []*LogEntry

	for {
		entry, err := r.ReadEntry()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// ReadAllJSON reads all JSON entries from the log file, skipping non-JSON lines.
func (r *Reader) ReadAllJSON() ([]*LogEntry, error) {
	var entries []*LogEntry

	for {
		entry, err := r.ReadEntry()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if entry.IsJSON {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}
