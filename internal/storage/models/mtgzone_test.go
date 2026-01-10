package models

import (
	"testing"
)

func TestGetTierRank(t *testing.T) {
	tests := []struct {
		tier     string
		expected int
	}{
		{"S", 1},
		{"A", 2},
		{"A+", 2},
		{"A-", 3},
		{"B", 4},
		{"B+", 4},
		{"B-", 5},
		{"C", 6},
		{"C+", 6},
		{"C-", 7},
		{"D", 8},
		{"F", 9},
		{"unknown", 10},
		{"", 10},
	}

	for _, tt := range tests {
		t.Run(tt.tier, func(t *testing.T) {
			result := GetTierRank(tt.tier)
			if result != tt.expected {
				t.Errorf("GetTierRank(%q) = %d, want %d", tt.tier, result, tt.expected)
			}
		})
	}
}

func TestIsTopTier(t *testing.T) {
	tests := []struct {
		tier     string
		expected bool
	}{
		{"S", true},
		{"A", true},
		{"A+", true},
		{"A-", false},
		{"B", false},
		{"B+", false},
		{"C", false},
		{"D", false},
		{"F", false},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.tier, func(t *testing.T) {
			result := IsTopTier(tt.tier)
			if result != tt.expected {
				t.Errorf("IsTopTier(%q) = %v, want %v", tt.tier, result, tt.expected)
			}
		})
	}
}

func TestCardRoleConstants(t *testing.T) {
	if CardRoleCore != "core" {
		t.Errorf("CardRoleCore = %q, want %q", CardRoleCore, "core")
	}
	if CardRoleFlex != "flex" {
		t.Errorf("CardRoleFlex = %q, want %q", CardRoleFlex, "flex")
	}
	if CardRoleSideboard != "sideboard" {
		t.Errorf("CardRoleSideboard = %q, want %q", CardRoleSideboard, "sideboard")
	}
}

func TestCardImportanceConstants(t *testing.T) {
	if CardImportanceEssential != "essential" {
		t.Errorf("CardImportanceEssential = %q, want %q", CardImportanceEssential, "essential")
	}
	if CardImportanceImportant != "important" {
		t.Errorf("CardImportanceImportant = %q, want %q", CardImportanceImportant, "important")
	}
	if CardImportanceOptional != "optional" {
		t.Errorf("CardImportanceOptional = %q, want %q", CardImportanceOptional, "optional")
	}
}

func TestArticleTypeConstants(t *testing.T) {
	if ArticleTypeDeckGuide != "deck_guide" {
		t.Errorf("ArticleTypeDeckGuide = %q, want %q", ArticleTypeDeckGuide, "deck_guide")
	}
	if ArticleTypeTierList != "tier_list" {
		t.Errorf("ArticleTypeTierList = %q, want %q", ArticleTypeTierList, "tier_list")
	}
	if ArticleTypeSetReview != "set_review" {
		t.Errorf("ArticleTypeSetReview = %q, want %q", ArticleTypeSetReview, "set_review")
	}
	if ArticleTypeStrategy != "strategy" {
		t.Errorf("ArticleTypeStrategy = %q, want %q", ArticleTypeStrategy, "strategy")
	}
}

func TestMTGZoneArchetype_Struct(t *testing.T) {
	archetype := MTGZoneArchetype{
		ID:          1,
		Name:        "Mono-Red Aggro",
		Format:      "Standard",
		Tier:        "S",
		Description: "Fast aggro deck",
		PlayStyle:   "aggro",
		SourceURL:   "https://mtgazone.com/deck/mono-red",
	}

	if archetype.Name != "Mono-Red Aggro" {
		t.Errorf("Name = %q, want %q", archetype.Name, "Mono-Red Aggro")
	}
	if archetype.Format != "Standard" {
		t.Errorf("Format = %q, want %q", archetype.Format, "Standard")
	}
	if archetype.Tier != "S" {
		t.Errorf("Tier = %q, want %q", archetype.Tier, "S")
	}
	if archetype.PlayStyle != "aggro" {
		t.Errorf("PlayStyle = %q, want %q", archetype.PlayStyle, "aggro")
	}
}

func TestMTGZoneArchetypeCard_Struct(t *testing.T) {
	card := MTGZoneArchetypeCard{
		ID:          1,
		ArchetypeID: 10,
		CardName:    "Lightning Bolt",
		Role:        CardRoleCore,
		Copies:      4,
		Importance:  CardImportanceEssential,
		Notes:       "Key removal spell",
	}

	if card.CardName != "Lightning Bolt" {
		t.Errorf("CardName = %q, want %q", card.CardName, "Lightning Bolt")
	}
	if card.Role != CardRoleCore {
		t.Errorf("Role = %q, want %q", card.Role, CardRoleCore)
	}
	if card.Copies != 4 {
		t.Errorf("Copies = %d, want %d", card.Copies, 4)
	}
	if card.Importance != CardImportanceEssential {
		t.Errorf("Importance = %q, want %q", card.Importance, CardImportanceEssential)
	}
}

func TestMTGZoneSynergy_Struct(t *testing.T) {
	synergy := MTGZoneSynergy{
		ID:               1,
		CardA:            "Fatal Push",
		CardB:            "Thoughtseize",
		Reason:           "Both are efficient black removal/disruption",
		SourceURL:        "https://mtgazone.com/article",
		ArchetypeContext: "Rakdos Midrange",
		Confidence:       0.85,
	}

	if synergy.CardA != "Fatal Push" {
		t.Errorf("CardA = %q, want %q", synergy.CardA, "Fatal Push")
	}
	if synergy.CardB != "Thoughtseize" {
		t.Errorf("CardB = %q, want %q", synergy.CardB, "Thoughtseize")
	}
	if synergy.Confidence != 0.85 {
		t.Errorf("Confidence = %v, want %v", synergy.Confidence, 0.85)
	}
}

func TestMTGZoneArticle_Struct(t *testing.T) {
	article := MTGZoneArticle{
		ID:             1,
		URL:            "https://mtgazone.com/article/1",
		Title:          "Top Decks in Standard",
		ArticleType:    ArticleTypeTierList,
		Format:         "Standard",
		Archetype:      "Meta Overview",
		CardsMentioned: `["Lightning Bolt", "Fatal Push"]`,
	}

	if article.Title != "Top Decks in Standard" {
		t.Errorf("Title = %q, want %q", article.Title, "Top Decks in Standard")
	}
	if article.ArticleType != ArticleTypeTierList {
		t.Errorf("ArticleType = %q, want %q", article.ArticleType, ArticleTypeTierList)
	}
}

func TestTierComparison(t *testing.T) {
	// Verify tier ordering
	if GetTierRank("S") >= GetTierRank("A") {
		t.Error("S tier should rank higher (lower number) than A tier")
	}
	if GetTierRank("A") >= GetTierRank("B") {
		t.Error("A tier should rank higher than B tier")
	}
	if GetTierRank("B") >= GetTierRank("C") {
		t.Error("B tier should rank higher than C tier")
	}
	if GetTierRank("C") >= GetTierRank("D") {
		t.Error("C tier should rank higher than D tier")
	}
	if GetTierRank("D") >= GetTierRank("F") {
		t.Error("D tier should rank higher than F tier")
	}
}
