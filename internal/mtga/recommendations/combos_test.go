package recommendations

import (
	"testing"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
)

func TestGetSynergyPackages(t *testing.T) {
	packages := GetSynergyPackages()

	// Should have at least 10 packages
	if len(packages) < 10 {
		t.Errorf("GetSynergyPackages() returned %d packages, want at least 10", len(packages))
	}

	// Check that each package has required fields
	for _, pkg := range packages {
		if pkg.Name == "" {
			t.Error("Package has empty Name")
		}
		if pkg.Description == "" {
			t.Errorf("Package %q has empty Description", pkg.Name)
		}
		if len(pkg.Roles) < 2 {
			t.Errorf("Package %q has fewer than 2 roles", pkg.Name)
		}
		if pkg.MinRoles < 1 {
			t.Errorf("Package %q has invalid MinRoles: %d", pkg.Name, pkg.MinRoles)
		}
	}
}

func TestGetCardRoles_TokenGenerator(t *testing.T) {
	oracleText := "When this creature enters the battlefield, create a 1/1 white Soldier creature token."
	card := &cards.Card{
		ArenaID:    12345,
		Name:       "Attended Knight",
		TypeLine:   "Creature — Human Knight",
		OracleText: &oracleText,
	}

	roles := GetCardRoles(card)

	foundTokenGenerator := false
	for _, role := range roles {
		if role.RoleName == "token_generator" || role.RoleName == "token_maker" {
			foundTokenGenerator = true
			break
		}
	}

	if !foundTokenGenerator {
		t.Errorf("Expected card to be identified as token generator, got roles: %v", roles)
	}
}

func TestGetCardRoles_SacOutlet(t *testing.T) {
	oracleText := "Sacrifice a creature: Scry 1."
	card := &cards.Card{
		ArenaID:    12346,
		Name:       "Viscera Seer",
		TypeLine:   "Creature — Vampire Wizard",
		OracleText: &oracleText,
	}

	roles := GetCardRoles(card)

	foundSacOutlet := false
	for _, role := range roles {
		if role.RoleName == "sac_outlet" {
			foundSacOutlet = true
			break
		}
	}

	if !foundSacOutlet {
		t.Errorf("Expected card to be identified as sac outlet, got roles: %v", roles)
	}
}

func TestGetCardRoles_DeathPayoff(t *testing.T) {
	oracleText := "Whenever another creature you control dies, each opponent loses 1 life and you gain 1 life."
	card := &cards.Card{
		ArenaID:    12347,
		Name:       "Blood Artist",
		TypeLine:   "Creature — Vampire",
		OracleText: &oracleText,
	}

	roles := GetCardRoles(card)

	foundDeathPayoff := false
	for _, role := range roles {
		if role.RoleName == "death_payoff" {
			foundDeathPayoff = true
			break
		}
	}

	if !foundDeathPayoff {
		t.Errorf("Expected card to be identified as death payoff, got roles: %v", roles)
	}
}

func TestGetCardRoles_ProwessCreature(t *testing.T) {
	oracleText := "Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)"
	card := &cards.Card{
		ArenaID:    12348,
		Name:       "Monastery Swiftspear",
		TypeLine:   "Creature — Human Monk",
		OracleText: &oracleText,
	}

	roles := GetCardRoles(card)

	foundProwess := false
	for _, role := range roles {
		if role.RoleName == "prowess_creature" {
			foundProwess = true
			break
		}
	}

	if !foundProwess {
		t.Errorf("Expected card to be identified as prowess creature, got roles: %v", roles)
	}
}

func TestGetCardRoles_CheapSpell(t *testing.T) {
	oracleText := "Lightning Bolt deals 3 damage to any target."
	card := &cards.Card{
		ArenaID:    12349,
		Name:       "Lightning Bolt",
		TypeLine:   "Instant",
		CMC:        1,
		OracleText: &oracleText,
	}

	roles := GetCardRoles(card)

	foundCheapSpell := false
	for _, role := range roles {
		if role.RoleName == "cheap_spell" {
			foundCheapSpell = true
			break
		}
	}

	if !foundCheapSpell {
		t.Errorf("Expected card to be identified as cheap spell, got roles: %v", roles)
	}
}

func TestGetCardRoles_ETBCreature(t *testing.T) {
	oracleText := "When Mulldrifter enters the battlefield, draw two cards."
	card := &cards.Card{
		ArenaID:    12350,
		Name:       "Mulldrifter",
		TypeLine:   "Creature — Elemental",
		OracleText: &oracleText,
	}

	roles := GetCardRoles(card)

	foundETB := false
	for _, role := range roles {
		if role.RoleName == "etb_creature" {
			foundETB = true
			break
		}
	}

	if !foundETB {
		t.Errorf("Expected card to be identified as ETB creature, got roles: %v", roles)
	}
}

func TestGetCardRoles_BlinkEnabler(t *testing.T) {
	oracleText := "Exile target creature you control, then return that card to the battlefield under your control."
	card := &cards.Card{
		ArenaID:    12351,
		Name:       "Cloudshift",
		TypeLine:   "Instant",
		OracleText: &oracleText,
	}

	roles := GetCardRoles(card)

	foundBlink := false
	for _, role := range roles {
		if role.RoleName == "blink_enabler" {
			foundBlink = true
			break
		}
	}

	if !foundBlink {
		t.Errorf("Expected card to be identified as blink enabler, got roles: %v", roles)
	}
}

func TestGetCardRoles_LifegainPayoff(t *testing.T) {
	oracleText := "Whenever you gain life, put a +1/+1 counter on Ajani's Pridemate."
	card := &cards.Card{
		ArenaID:    12352,
		Name:       "Ajani's Pridemate",
		TypeLine:   "Creature — Cat Soldier",
		OracleText: &oracleText,
	}

	roles := GetCardRoles(card)

	foundLifegainPayoff := false
	for _, role := range roles {
		if role.RoleName == "lifegain_payoff" {
			foundLifegainPayoff = true
			break
		}
	}

	if !foundLifegainPayoff {
		t.Errorf("Expected card to be identified as lifegain payoff, got roles: %v", roles)
	}
}

func TestAnalyzeDeckPackages_Aristocrats(t *testing.T) {
	tokenText := "When this creature enters the battlefield, create a 1/1 white Soldier creature token."
	sacText := "Sacrifice a creature: Scry 1."
	payoffText := "Whenever another creature you control dies, each opponent loses 1 life."

	deckCards := []*cards.Card{
		{ArenaID: 1, Name: "Token Maker", TypeLine: "Creature", OracleText: &tokenText},
		{ArenaID: 2, Name: "Sac Outlet", TypeLine: "Creature", OracleText: &sacText},
		{ArenaID: 3, Name: "Death Payoff", TypeLine: "Creature", OracleText: &payoffText},
	}

	analyses := AnalyzeDeckPackages(deckCards)

	foundAristocrats := false
	for _, analysis := range analyses {
		if analysis.Package.Name == "Aristocrats" {
			foundAristocrats = true
			if !analysis.IsActive {
				t.Error("Expected Aristocrats package to be active")
			}
			if analysis.Completeness < 1.0 {
				t.Errorf("Expected Aristocrats completeness 1.0, got %.2f", analysis.Completeness)
			}
		}
	}

	if !foundAristocrats {
		t.Error("Expected to find Aristocrats package in analysis")
	}
}

func TestAnalyzeDeckPackages_PartialPackage(t *testing.T) {
	tokenText := "Create a 1/1 white Soldier creature token."
	payoffText := "Whenever a creature you control dies, draw a card."

	// Missing sac outlet
	deckCards := []*cards.Card{
		{ArenaID: 1, Name: "Token Maker", TypeLine: "Creature", OracleText: &tokenText},
		{ArenaID: 2, Name: "Death Payoff", TypeLine: "Creature", OracleText: &payoffText},
	}

	analyses := AnalyzeDeckPackages(deckCards)

	foundAristocrats := false
	for _, analysis := range analyses {
		if analysis.Package.Name == "Aristocrats" {
			foundAristocrats = true
			// Should have missing roles
			if len(analysis.MissingRoles) == 0 {
				t.Error("Expected Aristocrats to have missing roles")
			}
			// Completeness should be less than 1.0
			if analysis.Completeness >= 1.0 {
				t.Errorf("Expected Aristocrats completeness < 1.0, got %.2f", analysis.Completeness)
			}
			break
		}
	}
	if !foundAristocrats {
		t.Error("Expected to find Aristocrats package in analysis")
	}
}

func TestScoreCardForPackages_CompletesPackage(t *testing.T) {
	tokenText := "Create a 1/1 white Soldier creature token."
	payoffText := "Whenever a creature you control dies, draw a card."

	// Deck has tokens and death payoff, missing sac outlet
	deckCards := []*cards.Card{
		{ArenaID: 1, Name: "Token Maker", TypeLine: "Creature", OracleText: &tokenText},
		{ArenaID: 2, Name: "Death Payoff", TypeLine: "Creature", OracleText: &payoffText},
	}

	analyses := AnalyzeDeckPackages(deckCards)

	// Card that adds the missing sac outlet
	sacText := "Sacrifice a creature: Scry 1."
	newCard := &cards.Card{
		ArenaID:    99,
		Name:       "Viscera Seer",
		TypeLine:   "Creature",
		OracleText: &sacText,
	}

	bonus, reasons := ScoreCardForPackages(newCard, analyses)

	if bonus <= 0 {
		t.Errorf("Expected positive bonus for completing package, got %.2f", bonus)
	}

	if len(reasons) == 0 {
		t.Error("Expected reasons explaining the bonus")
	}
}

func TestScoreCardForPackages_NoBonus(t *testing.T) {
	// Empty deck
	analyses := AnalyzeDeckPackages([]*cards.Card{})

	// Random card with no package synergy
	text := "Flying"
	card := &cards.Card{
		ArenaID:    99,
		Name:       "Flying Creature",
		TypeLine:   "Creature — Bird",
		OracleText: &text,
	}

	bonus, reasons := ScoreCardForPackages(card, analyses)

	if bonus != 0 {
		t.Errorf("Expected no bonus for unrelated card, got %.2f", bonus)
	}

	if len(reasons) != 0 {
		t.Errorf("Expected no reasons for unrelated card, got %v", reasons)
	}
}

func TestGetMissingRoleSuggestion(t *testing.T) {
	tokenText := "Create a 1/1 token."
	payoffText := "Whenever a creature dies, draw a card."

	deckCards := []*cards.Card{
		{ArenaID: 1, Name: "Token Maker", TypeLine: "Creature", OracleText: &tokenText},
		{ArenaID: 2, Name: "Death Payoff", TypeLine: "Creature", OracleText: &payoffText},
	}

	analyses := AnalyzeDeckPackages(deckCards)

	foundQualifying := false
	for _, analysis := range analyses {
		if analysis.Package.Name == "Aristocrats" && analysis.Completeness >= 0.5 {
			foundQualifying = true
			suggestion := GetMissingRoleSuggestion(&analysis)
			if suggestion == "" {
				t.Error("Expected suggestion for partially complete package")
			}
			break
		}
	}
	if !foundQualifying {
		t.Error("Expected to find Aristocrats package with completeness >= 0.5")
	}
}

func TestPackageRolesHavePatterns(t *testing.T) {
	packages := GetSynergyPackages()

	for _, pkg := range packages {
		for _, role := range pkg.Roles {
			// Each role should have at least one detection method
			hasDetection := len(role.Patterns) > 0 || len(role.TypeLines) > 0 || len(role.Keywords) > 0 || role.Name == "big_target"
			if !hasDetection {
				t.Errorf("Package %q role %q has no detection patterns", pkg.Name, role.Name)
			}
		}
	}
}
