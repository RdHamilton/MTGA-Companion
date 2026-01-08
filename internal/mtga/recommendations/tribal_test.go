package recommendations

import (
	"testing"
)

func TestGetTribalInfo(t *testing.T) {
	tests := []struct {
		name         string
		creatureType string
		wantNil      bool
		wantSupport  TribalSupport
	}{
		{
			name:         "strong tribe - Elf",
			creatureType: "Elf",
			wantNil:      false,
			wantSupport:  TribalSupportStrong,
		},
		{
			name:         "moderate tribe - Cat",
			creatureType: "Cat",
			wantNil:      false,
			wantSupport:  TribalSupportModerate,
		},
		{
			name:         "weak tribe - Monk",
			creatureType: "Monk",
			wantNil:      false,
			wantSupport:  TribalSupportWeak,
		},
		{
			name:         "unknown tribe",
			creatureType: "Unicorn",
			wantNil:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := GetTribalInfo(tt.creatureType)
			if tt.wantNil {
				if info != nil {
					t.Errorf("GetTribalInfo(%q) = %v, want nil", tt.creatureType, info)
				}
			} else {
				if info == nil {
					t.Fatalf("GetTribalInfo(%q) = nil, want non-nil", tt.creatureType)
				}
				if info.Support != tt.wantSupport {
					t.Errorf("GetTribalInfo(%q).Support = %q, want %q", tt.creatureType, info.Support, tt.wantSupport)
				}
			}
		})
	}
}

func TestGetTribalSynergyWeight(t *testing.T) {
	tests := []struct {
		name         string
		creatureType string
		wantMin      float64
		wantMax      float64
	}{
		{
			name:         "strong tribe has high weight",
			creatureType: "Elf",
			wantMin:      1.2,
			wantMax:      1.5,
		},
		{
			name:         "moderate tribe has medium weight",
			creatureType: "Cat",
			wantMin:      1.0,
			wantMax:      1.2,
		},
		{
			name:         "weak tribe has lower weight",
			creatureType: "Beast",
			wantMin:      0.7,
			wantMax:      1.0,
		},
		{
			name:         "unknown tribe returns neutral",
			creatureType: "Unicorn",
			wantMin:      1.0,
			wantMax:      1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weight := GetTribalSynergyWeight(tt.creatureType)
			if weight < tt.wantMin || weight > tt.wantMax {
				t.Errorf("GetTribalSynergyWeight(%q) = %v, want between %v and %v",
					tt.creatureType, weight, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestGetRelatedTypes(t *testing.T) {
	tests := []struct {
		name         string
		creatureType string
		wantContains string
		wantNil      bool
	}{
		{
			name:         "Elf has Druid as related",
			creatureType: "Elf",
			wantContains: "Druid",
		},
		{
			name:         "Rogue has Faerie as related",
			creatureType: "Rogue",
			wantContains: "Faerie",
		},
		{
			name:         "unknown type returns nil",
			creatureType: "Unicorn",
			wantNil:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			related := GetRelatedTypes(tt.creatureType)
			if tt.wantNil {
				if related != nil {
					t.Errorf("GetRelatedTypes(%q) = %v, want nil", tt.creatureType, related)
				}
				return
			}
			found := false
			for _, r := range related {
				if r == tt.wantContains {
					found = true
					break
				}
			}
			if !found && tt.wantContains != "" {
				t.Errorf("GetRelatedTypes(%q) = %v, want to contain %q",
					tt.creatureType, related, tt.wantContains)
			}
		})
	}
}

func TestIsStrongTribalSupport(t *testing.T) {
	tests := []struct {
		name         string
		creatureType string
		want         bool
	}{
		{"Elf is strong", "Elf", true},
		{"Goblin is strong", "Goblin", true},
		{"Vampire is strong", "Vampire", true},
		{"Cat is not strong", "Cat", false},
		{"Monk is not strong", "Monk", false},
		{"Unknown is not strong", "Unicorn", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsStrongTribalSupport(tt.creatureType)
			if got != tt.want {
				t.Errorf("IsStrongTribalSupport(%q) = %v, want %v", tt.creatureType, got, tt.want)
			}
		})
	}
}

func TestGetTribalKeywords(t *testing.T) {
	tests := []struct {
		name         string
		creatureType string
		wantContains string
	}{
		{
			name:         "Elf has mana ramp",
			creatureType: "Elf",
			wantContains: "mana ramp",
		},
		{
			name:         "Goblin has haste",
			creatureType: "Goblin",
			wantContains: "haste",
		},
		{
			name:         "Zombie has graveyard",
			creatureType: "Zombie",
			wantContains: "graveyard",
		},
		{
			name:         "Cat has tokens",
			creatureType: "Cat",
			wantContains: "tokens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := GetTribalKeywords(tt.creatureType)
			found := false
			for _, kw := range keywords {
				if kw == tt.wantContains {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("GetTribalKeywords(%q) = %v, want to contain %q",
					tt.creatureType, keywords, tt.wantContains)
			}
		})
	}
}

func TestIsChangeling(t *testing.T) {
	tests := []struct {
		name       string
		oracleText string
		want       bool
	}{
		{
			name:       "has changeling keyword",
			oracleText: "Changeling (This card is every creature type.)",
			want:       true,
		},
		{
			name:       "is every creature type",
			oracleText: "This creature is every creature type.",
			want:       true,
		},
		{
			name:       "no changeling",
			oracleText: "Flying, vigilance",
			want:       false,
		},
		{
			name:       "empty text",
			oracleText: "",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsChangeling(tt.oracleText)
			if got != tt.want {
				t.Errorf("IsChangeling(%q) = %v, want %v", tt.oracleText, got, tt.want)
			}
		})
	}
}

func TestGetAllSupportedTribes(t *testing.T) {
	tribes := GetAllSupportedTribes()

	// Should have at least 20 tribes
	if len(tribes) < 20 {
		t.Errorf("GetAllSupportedTribes() returned %d tribes, want at least 20", len(tribes))
	}

	// Should contain key tribes
	keyTribes := []string{"Elf", "Goblin", "Vampire", "Zombie", "Cat", "Dragon"}
	for _, expected := range keyTribes {
		found := false
		for _, tribe := range tribes {
			if tribe == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetAllSupportedTribes() missing %q", expected)
		}
	}
}

func TestGetStrongTribes(t *testing.T) {
	tribes := GetStrongTribes()

	// Should have at least 6 strong tribes
	if len(tribes) < 6 {
		t.Errorf("GetStrongTribes() returned %d tribes, want at least 6", len(tribes))
	}

	// All returned tribes should be strong
	for _, tribe := range tribes {
		if !IsStrongTribalSupport(tribe) {
			t.Errorf("GetStrongTribes() returned %q which is not a strong tribe", tribe)
		}
	}
}

func TestTribalDatabaseCompleteness(t *testing.T) {
	// Verify each entry has required fields
	for tribe, info := range tribalDatabase {
		if info.Type == "" {
			t.Errorf("Tribe %q has empty Type field", tribe)
		}
		if info.Type != tribe {
			t.Errorf("Tribe %q has mismatched Type field: %q", tribe, info.Type)
		}
		if info.Support == "" {
			t.Errorf("Tribe %q has empty Support field", tribe)
		}
		if info.SynergyWeight <= 0 {
			t.Errorf("Tribe %q has invalid SynergyWeight: %v", tribe, info.SynergyWeight)
		}
		if info.Description == "" {
			t.Errorf("Tribe %q has empty Description field", tribe)
		}
	}
}
