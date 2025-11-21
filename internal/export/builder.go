package export

import (
	"fmt"
	"io"
	"time"
)

// ExportBuilder provides a fluent API for configuring and executing export operations.
// It implements the Builder pattern to make export configuration more readable and maintainable.
//
// Example usage:
//
//	builder := NewExportBuilder().
//	    WithFormat(FormatJSON).
//	    WithFilePath("/path/to/output.json").
//	    WithPrettyJSON(true).
//	    WithOverwrite(true)
//
//	err := builder.Export(data)
type ExportBuilder struct {
	format     Format
	filePath   string
	prettyJSON bool
	overwrite  bool
	writer     io.Writer
	useWriter  bool
}

// NewExportBuilder creates a new ExportBuilder with default settings.
// Default settings:
//   - Format: FormatJSON
//   - PrettyJSON: false
//   - Overwrite: false
func NewExportBuilder() *ExportBuilder {
	return &ExportBuilder{
		format:     FormatJSON,
		prettyJSON: false,
		overwrite:  false,
		useWriter:  false,
	}
}

// WithFormat sets the export format (CSV, JSON, Markdown, Arena).
// Returns the builder for method chaining.
func (b *ExportBuilder) WithFormat(format Format) *ExportBuilder {
	b.format = format
	return b
}

// WithFilePath sets the output file path for the export.
// The directory will be created if it doesn't exist.
// Returns the builder for method chaining.
func (b *ExportBuilder) WithFilePath(filePath string) *ExportBuilder {
	b.filePath = filePath
	b.useWriter = false
	return b
}

// WithWriter sets an io.Writer as the output destination instead of a file.
// This is useful for writing to stdout, buffers, or other streams.
// Returns the builder for method chaining.
func (b *ExportBuilder) WithWriter(w io.Writer) *ExportBuilder {
	b.writer = w
	b.useWriter = true
	return b
}

// WithPrettyJSON enables pretty-printing for JSON exports (indentation and newlines).
// Only affects JSON format exports. Default is false.
// Returns the builder for method chaining.
func (b *ExportBuilder) WithPrettyJSON(pretty bool) *ExportBuilder {
	b.prettyJSON = pretty
	return b
}

// WithOverwrite enables overwriting existing files.
// If false and the file exists, Export will return an error.
// Default is false.
// Returns the builder for method chaining.
func (b *ExportBuilder) WithOverwrite(overwrite bool) *ExportBuilder {
	b.overwrite = overwrite
	return b
}

// WithDefaultFilename generates a timestamped filename based on the export type and format.
// For example: "matches_20240101_120000.json"
// Returns the builder for method chaining.
func (b *ExportBuilder) WithDefaultFilename(exportType string) *ExportBuilder {
	filename := GenerateFilename(exportType, b.format)
	b.filePath = filename
	b.useWriter = false
	return b
}

// WithTimestampedFilename generates a filename with a custom prefix and timestamp.
// Returns the builder for method chaining.
func (b *ExportBuilder) WithTimestampedFilename(prefix string) *ExportBuilder {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.%s", prefix, timestamp, b.format)
	b.filePath = filename
	b.useWriter = false
	return b
}

// Build creates an Options struct from the builder's configuration.
// This allows the builder to interoperate with existing code that uses Options directly.
func (b *ExportBuilder) Build() Options {
	return Options{
		Format:     b.format,
		FilePath:   b.filePath,
		PrettyJSON: b.prettyJSON,
		Overwrite:  b.overwrite,
	}
}

// Export executes the export operation with the configured settings.
// data can be a slice of structs or a single struct.
// Returns an error if the export fails.
func (b *ExportBuilder) Export(data interface{}) error {
	// Validate configuration
	if err := b.validate(); err != nil {
		return err
	}

	// If using writer, use ExportToWriter
	if b.useWriter {
		return ExportToWriter(b.writer, b.format, data, b.prettyJSON)
	}

	// Otherwise use file-based export
	opts := b.Build()
	exporter := NewExporter(opts)
	return exporter.Export(data)
}

// validate checks that the builder configuration is valid.
func (b *ExportBuilder) validate() error {
	// Check that either file path or writer is set
	if !b.useWriter && b.filePath == "" {
		return fmt.Errorf("either file path or writer must be set")
	}

	// Check that format is supported
	switch b.format {
	case FormatCSV, FormatJSON, FormatMarkdown, FormatArena:
		// Valid format
	default:
		return fmt.Errorf("unsupported export format: %s", b.format)
	}

	return nil
}

// Clone creates a deep copy of the builder.
// This is useful for creating variations of a base configuration.
func (b *ExportBuilder) Clone() *ExportBuilder {
	return &ExportBuilder{
		format:     b.format,
		filePath:   b.filePath,
		prettyJSON: b.prettyJSON,
		overwrite:  b.overwrite,
		writer:     b.writer,
		useWriter:  b.useWriter,
	}
}

// Reset resets the builder to default settings.
// Returns the builder for method chaining.
func (b *ExportBuilder) Reset() *ExportBuilder {
	b.format = FormatJSON
	b.filePath = ""
	b.prettyJSON = false
	b.overwrite = false
	b.writer = nil
	b.useWriter = false
	return b
}
