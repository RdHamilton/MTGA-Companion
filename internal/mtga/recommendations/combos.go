package recommendations

import (
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
)

// PackageRole defines a role within a synergy package.
type PackageRole struct {
	Name        string   // Role identifier (e.g., "token_generator", "sac_outlet")
	DisplayName string   // Human-readable name
	Patterns    []string // Oracle text patterns that identify this role
	TypeLines   []string // Type line patterns (e.g., "Instant", "Sorcery")
	Keywords    []string // Keywords that identify this role
	Required    bool     // Is this role required for the package to function?
}

// SynergyPackage defines a multi-card synergy chain.
type SynergyPackage struct {
	Name        string        // Package name (e.g., "Aristocrats")
	Description string        // Brief description of the strategy
	Roles       []PackageRole // Roles that make up this package
	MinRoles    int           // Minimum roles needed for the package to be active
}

// PackageAnalysis contains analysis of how well a deck fits a synergy package.
type PackageAnalysis struct {
	Package      *SynergyPackage
	FilledRoles  map[string]int // Role name -> count of cards filling this role
	MissingRoles []string       // Roles that are missing or underrepresented
	Completeness float64        // 0.0-1.0, how complete is the package
	IsActive     bool           // Does the deck have enough roles for the package to function?
}

// CardRoleMatch represents a card matching a role in a package.
type CardRoleMatch struct {
	PackageName string
	RoleName    string
	RoleDisplay string
	IsRequired  bool
}

// synergyPackages defines the known synergy chains/combos.
var synergyPackages = []SynergyPackage{
	{
		Name:        "Aristocrats",
		Description: "Sacrifice creatures for value with death triggers",
		MinRoles:    2,
		Roles: []PackageRole{
			{
				Name:        "token_generator",
				DisplayName: "Token Generator",
				Patterns:    []string{"create a.*token", "create.*creature token", "creates a.*token"},
				Required:    true,
			},
			{
				Name:        "sac_outlet",
				DisplayName: "Sacrifice Outlet",
				Patterns:    []string{"sacrifice a creature:", "sacrifice another creature:", "sacrifice a creature,"},
				Required:    true,
			},
			{
				Name:        "death_payoff",
				DisplayName: "Death Payoff",
				Patterns:    []string{"whenever.*creature.*dies", "whenever another creature you control dies", "whenever a creature you control dies"},
				Required:    true,
			},
		},
	},
	{
		Name:        "Spellslinger",
		Description: "Cast many spells to trigger prowess and spell payoffs",
		MinRoles:    2,
		Roles: []PackageRole{
			{
				Name:        "prowess_creature",
				DisplayName: "Prowess Creature",
				Patterns:    []string{"prowess", "whenever you cast a noncreature spell", "magecraft"},
				Required:    true,
			},
			{
				Name:        "cheap_spell",
				DisplayName: "Cheap Spell",
				TypeLines:   []string{"Instant", "Sorcery"},
				Required:    true,
			},
			{
				Name:        "spell_payoff",
				DisplayName: "Spell Payoff",
				Patterns:    []string{"whenever you cast an instant or sorcery", "instant or sorcery.*from your graveyard"},
				Required:    false,
			},
		},
	},
	{
		Name:        "Blink",
		Description: "Exile and return creatures to reuse ETB effects",
		MinRoles:    2,
		Roles: []PackageRole{
			{
				Name:        "etb_creature",
				DisplayName: "ETB Creature",
				Patterns:    []string{"when.*enters the battlefield", "when.*enters,"},
				Required:    true,
			},
			{
				Name:        "blink_enabler",
				DisplayName: "Blink Enabler",
				Patterns:    []string{"exile.*then return", "exile target creature you control.*return", "flicker"},
				Required:    true,
			},
			{
				Name:        "blink_payoff",
				DisplayName: "Blink Payoff",
				Patterns:    []string{"whenever.*enters the battlefield under your control", "whenever another creature enters"},
				Required:    false,
			},
		},
	},
	{
		Name:        "Reanimator",
		Description: "Put creatures in graveyard and bring them back",
		MinRoles:    2,
		Roles: []PackageRole{
			{
				Name:        "discard_enabler",
				DisplayName: "Discard Enabler",
				Patterns:    []string{"discard a card", "discard.*cards", "mill", "put.*into your graveyard"},
				Required:    true,
			},
			{
				Name:        "reanimation",
				DisplayName: "Reanimation Spell",
				Patterns:    []string{"return.*creature.*from.*graveyard to the battlefield", "put.*creature.*from.*graveyard onto the battlefield"},
				Required:    true,
			},
			{
				Name:        "big_target",
				DisplayName: "Reanimation Target",
				Patterns:    []string{}, // Detected by CMC >= 5 and creature type
				Required:    false,
			},
		},
	},
	{
		Name:        "Tokens",
		Description: "Create many tokens and buff them with anthems",
		MinRoles:    2,
		Roles: []PackageRole{
			{
				Name:        "token_maker",
				DisplayName: "Token Maker",
				Patterns:    []string{"create a.*token", "create.*tokens", "create two", "create three"},
				Required:    true,
			},
			{
				Name:        "anthem",
				DisplayName: "Anthem Effect",
				Patterns:    []string{"creatures you control get +", "other creatures you control get +", "creatures you control have"},
				Required:    true,
			},
			{
				Name:        "token_payoff",
				DisplayName: "Token Payoff",
				Patterns:    []string{"for each creature you control", "equal to the number of creatures"},
				Required:    false,
			},
		},
	},
	{
		Name:        "Counters",
		Description: "Accumulate +1/+1 counters and proliferate",
		MinRoles:    2,
		Roles: []PackageRole{
			{
				Name:        "counter_source",
				DisplayName: "Counter Source",
				Patterns:    []string{"put a +1/+1 counter", "put.*+1/+1 counters", "enters.*with.*+1/+1 counter"},
				Required:    true,
			},
			{
				Name:        "counter_synergy",
				DisplayName: "Counter Synergy",
				Patterns:    []string{"whenever.*+1/+1 counter", "for each +1/+1 counter", "proliferate", "double the number of +1/+1 counters"},
				Required:    true,
			},
			{
				Name:        "counter_payoff",
				DisplayName: "Counter Payoff",
				Patterns:    []string{"with.*or more +1/+1 counters", "has a +1/+1 counter"},
				Required:    false,
			},
		},
	},
	{
		Name:        "Lifegain",
		Description: "Gain life and trigger lifegain payoffs",
		MinRoles:    2,
		Roles: []PackageRole{
			{
				Name:        "lifegain_source",
				DisplayName: "Lifegain Source",
				Patterns:    []string{"gain.*life", "lifelink"},
				Keywords:    []string{"lifelink"},
				Required:    true,
			},
			{
				Name:        "lifegain_payoff",
				DisplayName: "Lifegain Payoff",
				Patterns:    []string{"whenever you gain life", "whenever.*gains life"},
				Required:    true,
			},
			{
				Name:        "lifegain_finisher",
				DisplayName: "Lifegain Finisher",
				Patterns:    []string{"equal to your life total", "life total becomes"},
				Required:    false,
			},
		},
	},
	{
		Name:        "Graveyard",
		Description: "Fill graveyard and use it as a resource",
		MinRoles:    2,
		Roles: []PackageRole{
			{
				Name:        "self_mill",
				DisplayName: "Self Mill",
				Patterns:    []string{"mill", "put.*from.*library into.*graveyard", "surveil"},
				Required:    true,
			},
			{
				Name:        "graveyard_payoff",
				DisplayName: "Graveyard Payoff",
				Patterns:    []string{"for each card in your graveyard", "in your graveyard", "from your graveyard", "escape", "flashback", "disturb"},
				Required:    true,
			},
			{
				Name:        "recursion",
				DisplayName: "Recursion",
				Patterns:    []string{"return.*from your graveyard", "cast.*from your graveyard"},
				Required:    false,
			},
		},
	},
	{
		Name:        "Artifacts",
		Description: "Artifacts matter synergies",
		MinRoles:    2,
		Roles: []PackageRole{
			{
				Name:        "artifact_producer",
				DisplayName: "Artifact Producer",
				Patterns:    []string{"create.*artifact token", "create a treasure", "create a clue", "create a food", "create a blood"},
				Required:    true,
			},
			{
				Name:        "artifact_payoff",
				DisplayName: "Artifact Payoff",
				Patterns:    []string{"whenever an artifact enters", "for each artifact you control", "artifact you control"},
				Required:    true,
			},
			{
				Name:        "artifact_sacrifice",
				DisplayName: "Artifact Sacrifice",
				Patterns:    []string{"sacrifice an artifact", "sacrifice a treasure", "sacrifice a food"},
				Required:    false,
			},
		},
	},
	{
		Name:        "Energy",
		Description: "Generate and spend energy counters",
		MinRoles:    2,
		Roles: []PackageRole{
			{
				Name:        "energy_producer",
				DisplayName: "Energy Producer",
				Patterns:    []string{"you get {e}", "energy counter"},
				Required:    true,
			},
			{
				Name:        "energy_spender",
				DisplayName: "Energy Spender",
				Patterns:    []string{"pay {e}", "spend.*energy"},
				Required:    true,
			},
		},
	},
}

// GetSynergyPackages returns all defined synergy packages.
func GetSynergyPackages() []SynergyPackage {
	return synergyPackages
}

// GetCardRoles identifies which roles a card can fill across all packages.
func GetCardRoles(card *cards.Card) []CardRoleMatch {
	var roles []CardRoleMatch

	oracleText := ""
	if card.OracleText != nil {
		oracleText = strings.ToLower(*card.OracleText)
	}
	typeLine := strings.ToLower(card.TypeLine)

	for _, pkg := range synergyPackages {
		for _, role := range pkg.Roles {
			if matchesRole(oracleText, typeLine, card, &role) {
				roles = append(roles, CardRoleMatch{
					PackageName: pkg.Name,
					RoleName:    role.Name,
					RoleDisplay: role.DisplayName,
					IsRequired:  role.Required,
				})
			}
		}
	}

	return roles
}

// matchesRole checks if a card matches a specific role.
func matchesRole(oracleText, typeLine string, card *cards.Card, role *PackageRole) bool {
	// Check oracle text patterns
	for _, pattern := range role.Patterns {
		if containsPattern(oracleText, pattern) {
			return true
		}
	}

	// Check type line patterns
	for _, typePattern := range role.TypeLines {
		if strings.Contains(typeLine, strings.ToLower(typePattern)) {
			// For "cheap_spell" role, also check CMC
			if role.Name == "cheap_spell" && card.CMC <= 2 {
				return true
			} else if role.Name != "cheap_spell" {
				return true
			}
		}
	}

	// Check keywords
	for _, keyword := range role.Keywords {
		if strings.Contains(oracleText, strings.ToLower(keyword)) {
			return true
		}
	}

	// Special case: big reanimation targets (CMC >= 5 creatures)
	if role.Name == "big_target" && card.CMC >= 5 && strings.Contains(typeLine, "creature") {
		return true
	}

	return false
}

// AnalyzeDeckPackages analyzes which synergy packages a deck supports.
func AnalyzeDeckPackages(deckCards []*cards.Card) []PackageAnalysis {
	var analyses []PackageAnalysis

	for i := range synergyPackages {
		pkg := &synergyPackages[i]
		analysis := analyzePackage(pkg, deckCards)
		if analysis.Completeness > 0 {
			analyses = append(analyses, analysis)
		}
	}

	return analyses
}

// analyzePackage analyzes how well a deck fits a specific package.
func analyzePackage(pkg *SynergyPackage, deckCards []*cards.Card) PackageAnalysis {
	filledRoles := make(map[string]int)
	var missingRoles []string

	// Count cards filling each role
	for _, card := range deckCards {
		oracleText := ""
		if card.OracleText != nil {
			oracleText = strings.ToLower(*card.OracleText)
		}
		typeLine := strings.ToLower(card.TypeLine)

		for _, role := range pkg.Roles {
			if matchesRole(oracleText, typeLine, card, &role) {
				filledRoles[role.Name]++
			}
		}
	}

	// Identify missing roles
	requiredFilled := 0
	totalRequired := 0
	for _, role := range pkg.Roles {
		if role.Required {
			totalRequired++
			if filledRoles[role.Name] > 0 {
				requiredFilled++
			} else {
				missingRoles = append(missingRoles, role.DisplayName)
			}
		}
	}

	// Calculate completeness
	completeness := 0.0
	if totalRequired > 0 {
		completeness = float64(requiredFilled) / float64(totalRequired)
	}

	// Package is active if minimum roles are filled
	rolesWithCards := 0
	for _, count := range filledRoles {
		if count > 0 {
			rolesWithCards++
		}
	}
	isActive := rolesWithCards >= pkg.MinRoles

	return PackageAnalysis{
		Package:      pkg,
		FilledRoles:  filledRoles,
		MissingRoles: missingRoles,
		Completeness: completeness,
		IsActive:     isActive,
	}
}

// GetMissingRoleSuggestion returns a suggestion for a missing role in a package.
func GetMissingRoleSuggestion(analysis *PackageAnalysis) string {
	if len(analysis.MissingRoles) == 0 {
		return ""
	}

	if analysis.Completeness >= 0.5 {
		return "Consider adding: " + strings.Join(analysis.MissingRoles, ", ")
	}
	return ""
}

// ScoreCardForPackages calculates bonus synergy for cards that complete packages.
func ScoreCardForPackages(card *cards.Card, deckAnalyses []PackageAnalysis) (float64, []string) {
	bonus := 0.0
	var reasons []string

	cardRoles := GetCardRoles(card)

	for _, analysis := range deckAnalyses {
		if !analysis.IsActive && analysis.Completeness < 0.5 {
			continue // Package not relevant enough
		}

		for _, cardRole := range cardRoles {
			if cardRole.PackageName != analysis.Package.Name {
				continue
			}

			// Check if this card fills a missing or underrepresented role
			currentCount := analysis.FilledRoles[cardRole.RoleName]

			if currentCount == 0 && cardRole.IsRequired {
				// Card completes a missing required role - big bonus!
				bonus += 0.3
				reasons = append(reasons, "Completes "+analysis.Package.Name+" package (adds "+cardRole.RoleDisplay+")")
			} else if currentCount < 3 {
				// Card adds to an underrepresented role
				bonus += 0.15
				reasons = append(reasons, "Strengthens "+analysis.Package.Name+" (more "+cardRole.RoleDisplay+")")
			}
		}
	}

	// Cap the bonus
	if bonus > 0.5 {
		bonus = 0.5
	}

	return bonus, reasons
}
