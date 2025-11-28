package ipc

import (
	"testing"
)

// TestEventData is a sample event struct for testing.
type TestEventData struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestDecodeEventData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected TestEventData
		wantErr  bool
	}{
		{
			name: "valid data",
			input: map[string]interface{}{
				"name":  "test",
				"count": float64(42), // JSON numbers are float64
			},
			expected: TestEventData{Name: "test", Count: 42},
			wantErr:  false,
		},
		{
			name:     "empty data",
			input:    map[string]interface{}{},
			expected: TestEventData{},
			wantErr:  false,
		},
		{
			name: "partial data",
			input: map[string]interface{}{
				"name": "partial",
			},
			expected: TestEventData{Name: "partial", Count: 0},
			wantErr:  false,
		},
		{
			name: "extra fields ignored",
			input: map[string]interface{}{
				"name":  "test",
				"count": float64(10),
				"extra": "ignored",
			},
			expected: TestEventData{Name: "test", Count: 10},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result TestEventData
			err := DecodeEventData(tt.input, &result)

			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeEventData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != tt.expected {
				t.Errorf("DecodeEventData() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTypedEventHandler(t *testing.T) {
	// Test that TypedEventHandler type can be used
	var handler TypedEventHandler[TestEventData] = func(data TestEventData) {
		if data.Name != "test" {
			t.Errorf("Expected name 'test', got '%s'", data.Name)
		}
	}

	// Call the handler
	handler(TestEventData{Name: "test", Count: 1})
}

func TestDecodeEventData_NestedStruct(t *testing.T) {
	type NestedData struct {
		Inner struct {
			Value string `json:"value"`
		} `json:"inner"`
	}

	input := map[string]interface{}{
		"inner": map[string]interface{}{
			"value": "nested",
		},
	}

	var result NestedData
	err := DecodeEventData(input, &result)
	if err != nil {
		t.Errorf("DecodeEventData() error = %v", err)
		return
	}

	if result.Inner.Value != "nested" {
		t.Errorf("Expected nested value 'nested', got '%s'", result.Inner.Value)
	}
}

func TestDecodeEventData_ArrayField(t *testing.T) {
	type ArrayData struct {
		Items []string `json:"items"`
	}

	input := map[string]interface{}{
		"items": []interface{}{"a", "b", "c"},
	}

	var result ArrayData
	err := DecodeEventData(input, &result)
	if err != nil {
		t.Errorf("DecodeEventData() error = %v", err)
		return
	}

	if len(result.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(result.Items))
	}
}
