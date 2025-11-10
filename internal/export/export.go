package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

// Format represents the export format.
type Format string

const (
	// FormatCSV represents CSV export format.
	FormatCSV Format = "csv"
	// FormatJSON represents JSON export format.
	FormatJSON Format = "json"
)

// Options holds configuration for export operations.
type Options struct {
	Format     Format
	FilePath   string
	PrettyJSON bool
	Overwrite  bool
}

// Exporter handles exporting data to various formats.
type Exporter struct {
	opts Options
}

// NewExporter creates a new Exporter with the given options.
func NewExporter(opts Options) *Exporter {
	return &Exporter{opts: opts}
}

// Export exports the given data to the specified format.
// data can be a slice of structs or a single struct.
func (e *Exporter) Export(data interface{}) error {
	switch e.opts.Format {
	case FormatCSV:
		return e.exportCSV(data)
	case FormatJSON:
		return e.exportJSON(data)
	default:
		return fmt.Errorf("unsupported export format: %s", e.opts.Format)
	}
}

// exportJSON exports data to JSON format.
func (e *Exporter) exportJSON(data interface{}) error {
	var output []byte
	var err error

	if e.opts.PrettyJSON {
		output, err = json.MarshalIndent(data, "", "  ")
	} else {
		output, err = json.Marshal(data)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return e.writeToFile(output)
}

// exportCSV exports data to CSV format.
// data must be a slice of structs.
func (e *Exporter) exportCSV(data interface{}) (err error) {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("CSV export requires a slice, got %s", v.Kind())
	}

	if v.Len() == 0 {
		return fmt.Errorf("no data to export")
	}

	// Create or open file
	file, fileErr := e.createFile()
	if fileErr != nil {
		return fileErr
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Get the first element to determine structure
	firstElem := v.Index(0)
	if firstElem.Kind() == reflect.Ptr {
		firstElem = firstElem.Elem()
	}

	if firstElem.Kind() != reflect.Struct {
		return fmt.Errorf("CSV export requires a slice of structs")
	}

	// Write header
	header := e.getCSVHeaders(firstElem.Type())
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rows
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}

		row := e.structToCSVRow(elem)
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row %d: %w", i, err)
		}
	}

	return nil
}

// getCSVHeaders extracts field names from a struct type for CSV headers.
func (e *Exporter) getCSVHeaders(t reflect.Type) []string {
	var headers []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Use csv tag if available, otherwise use field name
		if csvTag := field.Tag.Get("csv"); csvTag != "" && csvTag != "-" {
			headers = append(headers, csvTag)
		} else if field.IsExported() {
			headers = append(headers, field.Name)
		}
	}

	return headers
}

// structToCSVRow converts a struct to a CSV row (slice of strings).
func (e *Exporter) structToCSVRow(v reflect.Value) []string {
	var row []string

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields and fields tagged with csv:"-"
		if !field.IsExported() || field.Tag.Get("csv") == "-" {
			continue
		}

		fieldValue := v.Field(i)
		row = append(row, e.valueToString(fieldValue))
	}

	return row
}

// valueToString converts a reflect.Value to its string representation for CSV.
func (e *Exporter) valueToString(v reflect.Value) string {
	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.2f", v.Float())
	case reflect.Bool:
		return fmt.Sprintf("%t", v.Bool())
	case reflect.Struct:
		// Special handling for time.Time
		if v.Type() == reflect.TypeOf(time.Time{}) {
			t := v.Interface().(time.Time)
			return t.Format(time.RFC3339)
		}
		return fmt.Sprintf("%v", v.Interface())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

// writeToFile writes data to the configured file path.
func (e *Exporter) writeToFile(data []byte) (err error) {
	file, fileErr := e.createFile()
	if fileErr != nil {
		return fileErr
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

// createFile creates the output file, handling overwrite settings.
func (e *Exporter) createFile() (*os.File, error) {
	// Ensure directory exists
	dir := filepath.Dir(e.opts.FilePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(e.opts.FilePath); err == nil && !e.opts.Overwrite {
		return nil, fmt.Errorf("file already exists: %s (use overwrite option to replace)", e.opts.FilePath)
	}

	// Create/truncate file
	file, err := os.Create(e.opts.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	return file, nil
}

// ExportToWriter exports data to an io.Writer instead of a file.
// Useful for writing to stdout or other streams.
func ExportToWriter(w io.Writer, format Format, data interface{}, prettyJSON bool) error {
	switch format {
	case FormatJSON:
		encoder := json.NewEncoder(w)
		if prettyJSON {
			encoder.SetIndent("", "  ")
		}
		return encoder.Encode(data)
	case FormatCSV:
		return writeCSVToWriter(w, data)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// writeCSVToWriter writes CSV data to an io.Writer.
func writeCSVToWriter(w io.Writer, data interface{}) error {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("CSV export requires a slice, got %s", v.Kind())
	}

	if v.Len() == 0 {
		return fmt.Errorf("no data to export")
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Get the first element to determine structure
	firstElem := v.Index(0)
	if firstElem.Kind() == reflect.Ptr {
		firstElem = firstElem.Elem()
	}

	if firstElem.Kind() != reflect.Struct {
		return fmt.Errorf("CSV export requires a slice of structs")
	}

	exporter := &Exporter{}

	// Write header
	header := exporter.getCSVHeaders(firstElem.Type())
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rows
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}

		row := exporter.structToCSVRow(elem)
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row %d: %w", i, err)
		}
	}

	return nil
}

// GenerateFilename generates a default filename based on the export type and format.
func GenerateFilename(exportType string, format Format) string {
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s.%s", exportType, timestamp, format)
}
