package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
)

type EventSample struct {
	EventType  string
	Count      int
	Sample     map[string]interface{}
	Properties []string
}

func main() {
	// Find MTGA log file
	logPath, err := logreader.DefaultLogPath()
	if err != nil {
		log.Fatalf("Failed to find MTGA log file: %v", err)
	}

	log.Printf("Reading log file: %s", logPath)

	// Read log entries
	reader, err := logreader.NewReader(logPath)
	if err != nil {
		log.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	entries, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read log: %v", err)
	}

	log.Printf("Found %d log entries", len(entries))

	// Analyze events
	eventSamples := make(map[string]*EventSample)

	for _, entry := range entries {
		if !entry.IsJSON {
			continue
		}

		// Extract top-level keys
		for key, value := range entry.JSON {
			if _, exists := eventSamples[key]; !exists {
				eventSamples[key] = &EventSample{
					EventType:  key,
					Count:      0,
					Sample:     make(map[string]interface{}),
					Properties: []string{},
				}
			}

			eventSamples[key].Count++

			// Store first sample
			if len(eventSamples[key].Sample) == 0 {
				eventSamples[key].Sample = entry.JSON
			}

			// Collect all properties if this is an object
			if obj, ok := value.(map[string]interface{}); ok {
				for prop := range obj {
					found := false
					for _, p := range eventSamples[key].Properties {
						if p == prop {
							found = true
							break
						}
					}
					if !found {
						eventSamples[key].Properties = append(eventSamples[key].Properties, prop)
					}
				}
			}
		}
	}

	// Sort event types
	var eventTypes []string
	for eventType := range eventSamples {
		eventTypes = append(eventTypes, eventType)
	}
	sort.Strings(eventTypes)

	// Generate markdown
	var md strings.Builder
	md.WriteString("# MTGA Log Events Reference\n\n")
	md.WriteString("This document contains all top-level events found in MTGA Player.log files with sample data.\n\n")
	md.WriteString(fmt.Sprintf("**Total unique event types**: %d\n\n", len(eventTypes)))
	md.WriteString("---\n\n")

	for _, eventType := range eventTypes {
		sample := eventSamples[eventType]
		md.WriteString(fmt.Sprintf("## %s\n\n", eventType))
		md.WriteString(fmt.Sprintf("**Occurrences**: %d\n\n", sample.Count))

		if len(sample.Properties) > 0 {
			md.WriteString("**Properties**:\n")
			sort.Strings(sample.Properties)
			for _, prop := range sample.Properties {
				md.WriteString(fmt.Sprintf("- `%s`\n", prop))
			}
			md.WriteString("\n")
		}

		md.WriteString("**Sample Data**:\n```json\n")
		sampleJSON, err := json.MarshalIndent(sample.Sample, "", "  ")
		if err == nil {
			md.WriteString(string(sampleJSON))
		}
		md.WriteString("\n```\n\n")
		md.WriteString("---\n\n")
	}

	// Write to docs folder
	docsDir := filepath.Join("docs")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		log.Fatalf("Failed to create docs directory: %v", err)
	}

	outputPath := filepath.Join(docsDir, "MTGA_LOG_EVENTS.md")
	if err := os.WriteFile(outputPath, []byte(md.String()), 0o644); err != nil {
		log.Fatalf("Failed to write documentation: %v", err)
	}

	log.Printf("Documentation written to: %s", outputPath)
	log.Printf("Total event types documented: %d", len(eventTypes))
}
