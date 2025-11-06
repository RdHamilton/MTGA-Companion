package logreader

import (
	"testing"
)

func TestParseDraftHistory(t *testing.T) {
	tests := []struct {
		name      string
		entries   []*LogEntry
		wantNil   bool
		wantCount int
	}{
		{
			name:    "no draft events",
			entries: []*LogEntry{},
			wantNil: true,
		},
		{
			name: "no JSON entries",
			entries: []*LogEntry{
				{Raw: "Plain text line", IsJSON: false},
			},
			wantNil: true,
		},
		{
			name: "constructed events only",
			entries: []*LogEntry{
				{
					Raw:    `{"Courses": [{"CourseId": "test-1", "InternalEventName": "Ladder"}]}`,
					IsJSON: true,
					JSON: map[string]interface{}{
						"Courses": []interface{}{
							map[string]interface{}{
								"CourseId":          "test-1",
								"InternalEventName": "Ladder",
							},
						},
					},
				},
			},
			wantNil: true,
		},
		{
			name: "single draft event",
			entries: []*LogEntry{
				{
					Raw:    `{"Courses": [{"CourseId": "draft-1", "InternalEventName": "PremierDraft_BLB"}]}`,
					IsJSON: true,
					JSON: map[string]interface{}{
						"Courses": []interface{}{
							map[string]interface{}{
								"CourseId":          "draft-1",
								"InternalEventName": "PremierDraft_BLB",
								"CurrentModule":     "DeckBuild",
								"CurrentWins":       float64(3),
								"CurrentLosses":     float64(1),
								"CourseDeck": map[string]interface{}{
									"MainDeck": []interface{}{
										map[string]interface{}{
											"cardId":   float64(12345),
											"quantity": float64(2),
										},
									},
								},
								"CourseDeckSummary": map[string]interface{}{
									"Name": "BLB Draft Deck",
								},
							},
						},
					},
				},
			},
			wantNil:   false,
			wantCount: 1,
		},
		{
			name: "multiple draft events",
			entries: []*LogEntry{
				{
					Raw:    `{"Courses": [...]}`,
					IsJSON: true,
					JSON: map[string]interface{}{
						"Courses": []interface{}{
							map[string]interface{}{
								"CourseId":          "draft-1",
								"InternalEventName": "PremierDraft_BLB",
								"CurrentModule":     "CreateMatch",
								"CurrentWins":       float64(7),
							},
							map[string]interface{}{
								"CourseId":          "draft-2",
								"InternalEventName": "QuickDraft_FDN",
								"CurrentModule":     "DeckBuild",
								"CurrentWins":       float64(2),
							},
							map[string]interface{}{
								"CourseId":          "constructed-1",
								"InternalEventName": "Ladder",
							},
						},
					},
				},
			},
			wantNil:   false,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history, err := ParseDraftHistory(tt.entries)
			if err != nil {
				t.Errorf("ParseDraftHistory() unexpected error: %v", err)
				return
			}

			if tt.wantNil {
				if history != nil {
					t.Errorf("ParseDraftHistory() expected nil, got %v", history)
				}
				return
			}

			if history == nil {
				t.Error("ParseDraftHistory() expected non-nil result")
				return
			}

			if len(history.Drafts) != tt.wantCount {
				t.Errorf("ParseDraftHistory() got %d drafts, want %d", len(history.Drafts), tt.wantCount)
			}

			// Additional validation for single draft test
			if tt.wantCount == 1 && len(history.Drafts) > 0 {
				draft := history.Drafts[0]
				if draft.EventID != "draft-1" {
					t.Errorf("Draft EventID = %s, want draft-1", draft.EventID)
				}
				if draft.EventName != "PremierDraft_BLB" {
					t.Errorf("Draft EventName = %s, want PremierDraft_BLB", draft.EventName)
				}
				if draft.Wins != 3 {
					t.Errorf("Draft Wins = %d, want 3", draft.Wins)
				}
				if draft.Losses != 1 {
					t.Errorf("Draft Losses = %d, want 1", draft.Losses)
				}
				if draft.Deck.Name != "BLB Draft Deck" {
					t.Errorf("Deck Name = %s, want BLB Draft Deck", draft.Deck.Name)
				}
				if len(draft.Deck.MainDeck) != 1 {
					t.Errorf("MainDeck length = %d, want 1", len(draft.Deck.MainDeck))
				}
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"PremierDraft_BLB", "Draft", true},
		{"QuickDraft_FDN", "Draft", true},
		{"Sealed_BLB", "Sealed", true},
		{"Ladder", "Draft", false},
		{"Play", "Draft", false},
		{"", "Draft", false},
		{"Draft", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_contains_"+tt.substr, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

func TestParseArenaStats(t *testing.T) {
	tests := []struct {
		name              string
		entries           []*LogEntry
		wantNil           bool
		wantTotalMatches  int
		wantMatchWins     int
		wantMatchLosses   int
		wantTotalGames    int
		wantGameWins      int
		wantGameLosses    int
		wantFormatCount   int
	}{
		{
			name:    "no match events",
			entries: []*LogEntry{},
			wantNil: true,
		},
		{
			name: "no JSON entries",
			entries: []*LogEntry{
				{Raw: "Plain text line", IsJSON: false},
			},
			wantNil: true,
		},
		{
			name: "match without final result",
			entries: []*LogEntry{
				{
					Raw:    `{"matchGameRoomStateChangedEvent": {"gameRoomInfo": {}}}`,
					IsJSON: true,
					JSON: map[string]interface{}{
						"matchGameRoomStateChangedEvent": map[string]interface{}{
							"gameRoomInfo": map[string]interface{}{
								"stateType": "MatchGameRoomStateType_Playing",
							},
						},
					},
				},
			},
			wantNil: true,
		},
		{
			name: "single match win (player team 1)",
			entries: []*LogEntry{
				{
					Raw:    `{"matchGameRoomStateChangedEvent": {...}}`,
					IsJSON: true,
					JSON: map[string]interface{}{
						"matchGameRoomStateChangedEvent": map[string]interface{}{
							"gameRoomInfo": map[string]interface{}{
								"finalMatchResult": map[string]interface{}{
									"matchId": "match-1",
									"resultList": []interface{}{
										map[string]interface{}{
											"scope":          "MatchScope_Match",
											"winningTeamId":  float64(1),
											"result":         "ResultType_WinLoss",
										},
										map[string]interface{}{
											"scope":          "MatchScope_Game",
											"winningTeamId":  float64(1),
											"result":         "ResultType_WinLoss",
										},
									},
								},
								"gameRoomConfig": map[string]interface{}{
									"matchId": "match-1",
									"reservedPlayers": []interface{}{
										map[string]interface{}{
											"userId":   "player1",
											"teamId":   float64(1),
											"eventId":  "Play",
										},
									},
								},
							},
						},
					},
				},
			},
			wantNil:          false,
			wantTotalMatches: 1,
			wantMatchWins:    1,
			wantMatchLosses:  0,
			wantTotalGames:   1,
			wantGameWins:     1,
			wantGameLosses:   0,
			wantFormatCount:  1,
		},
		{
			name: "single match loss (player team 2 loses)",
			entries: []*LogEntry{
				{
					Raw:    `{"matchGameRoomStateChangedEvent": {...}}`,
					IsJSON: true,
					JSON: map[string]interface{}{
						"matchGameRoomStateChangedEvent": map[string]interface{}{
							"gameRoomInfo": map[string]interface{}{
								"finalMatchResult": map[string]interface{}{
									"matchId": "match-2",
									"resultList": []interface{}{
										map[string]interface{}{
											"scope":          "MatchScope_Match",
											"winningTeamId":  float64(1),
											"result":         "ResultType_WinLoss",
										},
										map[string]interface{}{
											"scope":          "MatchScope_Game",
											"winningTeamId":  float64(1),
											"result":         "ResultType_WinLoss",
										},
									},
								},
								"gameRoomConfig": map[string]interface{}{
									"matchId": "match-2",
									"reservedPlayers": []interface{}{
										map[string]interface{}{
											"userId":   "player1",
											"teamId":   float64(2),
											"eventId":  "Ladder",
										},
									},
								},
							},
						},
					},
				},
			},
			wantNil:          false,
			wantTotalMatches: 1,
			wantMatchWins:    0,
			wantMatchLosses:  1,
			wantTotalGames:   1,
			wantGameWins:     0,
			wantGameLosses:   1,
			wantFormatCount:  1,
		},
		{
			name: "multiple matches different formats",
			entries: []*LogEntry{
				{
					Raw:    `{"matchGameRoomStateChangedEvent": {...}}`,
					IsJSON: true,
					JSON: map[string]interface{}{
						"matchGameRoomStateChangedEvent": map[string]interface{}{
							"gameRoomInfo": map[string]interface{}{
								"finalMatchResult": map[string]interface{}{
									"matchId": "match-3",
									"resultList": []interface{}{
										map[string]interface{}{
											"scope":          "MatchScope_Match",
											"winningTeamId":  float64(1),
										},
										map[string]interface{}{
											"scope":          "MatchScope_Game",
											"winningTeamId":  float64(1),
										},
									},
								},
								"gameRoomConfig": map[string]interface{}{
									"reservedPlayers": []interface{}{
										map[string]interface{}{
											"teamId":   float64(1),
											"eventId":  "Play",
										},
									},
								},
							},
						},
					},
				},
				{
					Raw:    `{"matchGameRoomStateChangedEvent": {...}}`,
					IsJSON: true,
					JSON: map[string]interface{}{
						"matchGameRoomStateChangedEvent": map[string]interface{}{
							"gameRoomInfo": map[string]interface{}{
								"finalMatchResult": map[string]interface{}{
									"matchId": "match-4",
									"resultList": []interface{}{
										map[string]interface{}{
											"scope":          "MatchScope_Match",
											"winningTeamId":  float64(2),
										},
										map[string]interface{}{
											"scope":          "MatchScope_Game",
											"winningTeamId":  float64(2),
										},
									},
								},
								"gameRoomConfig": map[string]interface{}{
									"reservedPlayers": []interface{}{
										map[string]interface{}{
											"teamId":   float64(1),
											"eventId":  "Ladder",
										},
									},
								},
							},
						},
					},
				},
			},
			wantNil:          false,
			wantTotalMatches: 2,
			wantMatchWins:    1,
			wantMatchLosses:  1,
			wantTotalGames:   2,
			wantGameWins:     1,
			wantGameLosses:   1,
			wantFormatCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats, err := ParseArenaStats(tt.entries)
			if err != nil {
				t.Errorf("ParseArenaStats() unexpected error: %v", err)
				return
			}

			if tt.wantNil {
				if stats != nil {
					t.Errorf("ParseArenaStats() expected nil, got %v", stats)
				}
				return
			}

			if stats == nil {
				t.Error("ParseArenaStats() expected non-nil result")
				return
			}

			if stats.TotalMatches != tt.wantTotalMatches {
				t.Errorf("TotalMatches = %d, want %d", stats.TotalMatches, tt.wantTotalMatches)
			}
			if stats.MatchWins != tt.wantMatchWins {
				t.Errorf("MatchWins = %d, want %d", stats.MatchWins, tt.wantMatchWins)
			}
			if stats.MatchLosses != tt.wantMatchLosses {
				t.Errorf("MatchLosses = %d, want %d", stats.MatchLosses, tt.wantMatchLosses)
			}
			if stats.TotalGames != tt.wantTotalGames {
				t.Errorf("TotalGames = %d, want %d", stats.TotalGames, tt.wantTotalGames)
			}
			if stats.GameWins != tt.wantGameWins {
				t.Errorf("GameWins = %d, want %d", stats.GameWins, tt.wantGameWins)
			}
			if stats.GameLosses != tt.wantGameLosses {
				t.Errorf("GameLosses = %d, want %d", stats.GameLosses, tt.wantGameLosses)
			}
			if len(stats.FormatStats) != tt.wantFormatCount {
				t.Errorf("FormatStats count = %d, want %d", len(stats.FormatStats), tt.wantFormatCount)
			}
		})
	}
}
