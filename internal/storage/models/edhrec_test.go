package models

import "testing"

func TestNormalizeSynergyScore(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{
			name:     "positive synergy",
			input:    0.5,
			expected: 0.75,
		},
		{
			name:     "negative synergy",
			input:    -0.5,
			expected: 0.25,
		},
		{
			name:     "zero synergy",
			input:    0.0,
			expected: 0.5,
		},
		{
			name:     "max positive synergy",
			input:    1.0,
			expected: 1.0,
		},
		{
			name:     "max negative synergy",
			input:    -1.0,
			expected: 0.0,
		},
		{
			name:     "over max positive",
			input:    1.5,
			expected: 1.0,
		},
		{
			name:     "under min negative",
			input:    -1.5,
			expected: 0.0,
		},
		{
			name:     "slight positive",
			input:    0.2,
			expected: 0.6,
		},
		{
			name:     "slight negative",
			input:    -0.2,
			expected: 0.4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeSynergyScore(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeSynergyScore(%f) = %f, want %f", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEDHRECSynergy_Fields(t *testing.T) {
	synergy := EDHRECSynergy{
		ID:              1,
		CardName:        "Sol Ring",
		SynergyCardName: "Arcane Signet",
		SynergyScore:    0.8,
		InclusionCount:  90,
		NumDecks:        50000,
		Lift:            1.5,
	}

	if synergy.CardName != "Sol Ring" {
		t.Errorf("CardName = %q, want %q", synergy.CardName, "Sol Ring")
	}
	if synergy.SynergyCardName != "Arcane Signet" {
		t.Errorf("SynergyCardName = %q, want %q", synergy.SynergyCardName, "Arcane Signet")
	}
	if synergy.SynergyScore != 0.8 {
		t.Errorf("SynergyScore = %f, want %f", synergy.SynergyScore, 0.8)
	}
}

func TestEDHRECCardMetadata_Fields(t *testing.T) {
	metadata := EDHRECCardMetadata{
		ID:            1,
		CardName:      "Sol Ring",
		SanitizedName: "sol-ring",
		NumDecks:      100000,
		SaltScore:     0.5,
		ColorIdentity: "",
	}

	if metadata.CardName != "Sol Ring" {
		t.Errorf("CardName = %q, want %q", metadata.CardName, "Sol Ring")
	}
	if metadata.SanitizedName != "sol-ring" {
		t.Errorf("SanitizedName = %q, want %q", metadata.SanitizedName, "sol-ring")
	}
	if metadata.NumDecks != 100000 {
		t.Errorf("NumDecks = %d, want %d", metadata.NumDecks, 100000)
	}
}

func TestEDHRECThemeCard_Fields(t *testing.T) {
	themeCard := EDHRECThemeCard{
		ID:            1,
		ThemeName:     "tokens",
		CardName:      "Smothering Tithe",
		SynergyScore:  0.9,
		IsTopCard:     true,
		IsHighSynergy: true,
	}

	if themeCard.ThemeName != "tokens" {
		t.Errorf("ThemeName = %q, want %q", themeCard.ThemeName, "tokens")
	}
	if themeCard.CardName != "Smothering Tithe" {
		t.Errorf("CardName = %q, want %q", themeCard.CardName, "Smothering Tithe")
	}
	if !themeCard.IsTopCard {
		t.Error("IsTopCard should be true")
	}
	if !themeCard.IsHighSynergy {
		t.Error("IsHighSynergy should be true")
	}
}
