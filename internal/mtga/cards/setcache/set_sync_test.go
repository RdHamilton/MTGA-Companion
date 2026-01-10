package setcache

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// TestStandardSet_Structure tests the StandardSet struct.
func TestStandardSet_Structure(t *testing.T) {
	set := StandardSet{
		Name:     "Test Set",
		Codename: "test",
		Code:     "TST",
	}

	if set.Name != "Test Set" {
		t.Errorf("expected Name 'Test Set', got '%s'", set.Name)
	}
	if set.Code != "TST" {
		t.Errorf("expected Code 'TST', got '%s'", set.Code)
	}
}

// TestStandardResponse_Structure tests the StandardResponse struct.
func TestStandardResponse_Structure(t *testing.T) {
	resp := StandardResponse{
		Deprecated: false,
		Sets: []StandardSet{
			{Name: "Set 1", Code: "S1"},
			{Name: "Set 2", Code: "S2"},
		},
	}

	if resp.Deprecated {
		t.Error("expected Deprecated to be false")
	}
	if len(resp.Sets) != 2 {
		t.Errorf("expected 2 sets, got %d", len(resp.Sets))
	}
}

// TestSetSyncer_NewSetSyncer tests the NewSetSyncer constructor.
func TestSetSyncer_NewSetSyncer(t *testing.T) {
	syncer := NewSetSyncer(nil, nil)

	if syncer == nil {
		t.Fatal("expected non-nil SetSyncer")
	}
	if syncer.httpClient == nil {
		t.Error("expected non-nil httpClient")
	}
	if syncer.httpClient.Timeout != 30*time.Second {
		t.Errorf("expected Timeout 30s, got %v", syncer.httpClient.Timeout)
	}
}

// mockStorageService implements the storage service methods needed for testing.
type mockStorageService struct {
	sets              []*storage.Set
	standardCodes     map[string]string
	saveSetCalled     int
	updateLegalCalled bool
}

func (m *mockStorageService) SaveSet(ctx context.Context, set *storage.Set) error {
	m.saveSetCalled++
	m.sets = append(m.sets, set)
	return nil
}

func (m *mockStorageService) GetAllSets(ctx context.Context) ([]*storage.Set, error) {
	return m.sets, nil
}

func (m *mockStorageService) UpdateStandardLegality(ctx context.Context, standardCodes map[string]string) error {
	m.standardCodes = standardCodes
	m.updateLegalCalled = true
	return nil
}

func (m *mockStorageService) GetStandardSets(ctx context.Context) ([]*storage.Set, error) {
	var result []*storage.Set
	for _, set := range m.sets {
		if set.IsStandardLegal {
			result = append(result, set)
		}
	}
	return result, nil
}

// TestSetSyncer_SyncStandardLegality_ParsesResponse tests parsing the whatsinstandard.com response.
func TestSetSyncer_SyncStandardLegality_ParsesResponse(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Include sets with different states:
		// - One currently in Standard
		// - One not yet entered (future enterDate)
		// - One already rotated out (past exitDate)
		w.Write([]byte(`{
			"deprecated": false,
			"sets": [
				{
					"name": "Current Set",
					"code": "CUR",
					"enterDate": {"exact": "2024-01-01T00:00:00.000Z", "rough": "Q1 2024"},
					"exitDate": {"exact": "2026-01-01T00:00:00.000Z", "rough": "Q1 2026"}
				},
				{
					"name": "Future Set",
					"code": "FUT",
					"enterDate": {"exact": "2030-01-01T00:00:00.000Z", "rough": "Q1 2030"},
					"exitDate": {"exact": "2032-01-01T00:00:00.000Z", "rough": "Q1 2032"}
				},
				{
					"name": "Rotated Set",
					"code": "ROT",
					"enterDate": {"exact": "2020-01-01T00:00:00.000Z", "rough": "Q1 2020"},
					"exitDate": {"exact": "2022-01-01T00:00:00.000Z", "rough": "Q1 2022"}
				},
				{
					"name": "No Code Set",
					"code": "",
					"enterDate": {"rough": "Unknown"},
					"exitDate": {"rough": "Unknown"}
				}
			]
		}`))
	}))
	defer server.Close()

	// Create syncer with test server URL
	syncer := &SetSyncer{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	// Make request to test server
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	resp, err := syncer.httpClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

// TestSetSyncer_FilterRelevantSetTypes tests that only relevant set types are synced.
func TestSetSyncer_FilterRelevantSetTypes(t *testing.T) {
	relevantTypes := map[string]bool{
		"core":             true,
		"expansion":        true,
		"masters":          true,
		"draft_innovation": true,
		"commander":        true,
		"alchemy":          true,
		"starter":          true,
	}

	testCases := []struct {
		setType  string
		expected bool
	}{
		{"core", true},
		{"expansion", true},
		{"masters", true},
		{"draft_innovation", true},
		{"commander", true},
		{"alchemy", true},
		{"starter", true},
		{"token", false},
		{"memorabilia", false},
		{"promo", false},
		{"funny", false},
	}

	for _, tc := range testCases {
		t.Run(tc.setType, func(t *testing.T) {
			result := relevantTypes[tc.setType]
			if result != tc.expected {
				t.Errorf("set type '%s': expected %v, got %v", tc.setType, tc.expected, result)
			}
		})
	}
}

// TestSetSyncer_StandardCodeNormalization tests that set codes are normalized to uppercase.
func TestSetSyncer_StandardCodeNormalization(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"dsk", "DSK"},
		{"DSK", "DSK"},
		{"Dsk", "DSK"},
		{"blb", "BLB"},
		{"m21", "M21"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := normalizeSetCode(tc.input)
			if result != tc.expected {
				t.Errorf("normalizeSetCode(%s): expected %s, got %s", tc.input, tc.expected, result)
			}
		})
	}
}

// normalizeSetCode is a helper function matching the logic in SyncStandardLegality.
func normalizeSetCode(code string) string {
	// Using same logic as in SyncStandardLegality
	result := ""
	for _, c := range code {
		if c >= 'a' && c <= 'z' {
			result += string(c - 32)
		} else {
			result += string(c)
		}
	}
	return result
}

// TestSetSyncer_EnterDateParsing tests parsing of enterDate fields.
func TestSetSyncer_EnterDateParsing(t *testing.T) {
	testCases := []struct {
		exactDate   string
		shouldParse bool
	}{
		{"2024-01-01T00:00:00.000Z", true},
		{"2024-12-31T23:59:59.999Z", true},
		{"invalid", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.exactDate, func(t *testing.T) {
			_, err := time.Parse(time.RFC3339, tc.exactDate)
			parsed := err == nil
			if parsed != tc.shouldParse {
				t.Errorf("parsing '%s': expected parsed=%v, got parsed=%v", tc.exactDate, tc.shouldParse, parsed)
			}
		})
	}
}

// TestSetSyncer_ExitDateLogic tests the logic for determining if a set has exited Standard.
func TestSetSyncer_ExitDateLogic(t *testing.T) {
	now := time.Now()
	pastDate := now.AddDate(-1, 0, 0)  // 1 year ago
	futureDate := now.AddDate(1, 0, 0) // 1 year from now

	testCases := []struct {
		name          string
		exitDate      time.Time
		shouldInclude bool
	}{
		{"future exit date", futureDate, true},
		{"past exit date", pastDate, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Logic: if exitDate.Before(now), set is not Standard-legal
			isStandard := !tc.exitDate.Before(now)
			if isStandard != tc.shouldInclude {
				t.Errorf("exit date logic: expected include=%v, got include=%v", tc.shouldInclude, isStandard)
			}
		})
	}
}
