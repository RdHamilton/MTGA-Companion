package recommendations

// TribalSupport indicates the level of tribal support for a creature type.
type TribalSupport string

const (
	// TribalSupportStrong indicates 20+ cards with tribal synergies.
	TribalSupportStrong TribalSupport = "strong"
	// TribalSupportModerate indicates 10-20 cards with tribal synergies.
	TribalSupportModerate TribalSupport = "moderate"
	// TribalSupportWeak indicates fewer than 10 cards with tribal synergies.
	TribalSupportWeak TribalSupport = "weak"
)

// TribalInfo contains information about a creature type's tribal support.
type TribalInfo struct {
	Type           string        // The creature type (e.g., "Cat", "Elf")
	Support        TribalSupport // Level of tribal support
	SynergyWeight  float64       // Weight multiplier for synergy scoring (0.5-1.5)
	RelatedTypes   []string      // Related creature types that synergize
	CommonKeywords []string      // Keywords commonly found in this tribe
	Description    string        // Brief description of the tribe's strategy
}

// tribalDatabase maps creature types to their tribal information.
// This enables better synergy detection for tribal deck building.
var tribalDatabase = map[string]TribalInfo{
	// Strong Support Tribes (20+ cards in Standard/Historic)
	"Elf": {
		Type:           "Elf",
		Support:        TribalSupportStrong,
		SynergyWeight:  1.3,
		RelatedTypes:   []string{"Druid", "Ranger", "Shaman"},
		CommonKeywords: []string{"mana ramp", "tokens", "+1/+1 counters"},
		Description:    "Mana production and overwhelming board presence",
	},
	"Goblin": {
		Type:           "Goblin",
		Support:        TribalSupportStrong,
		SynergyWeight:  1.3,
		RelatedTypes:   []string{"Shaman", "Warrior"},
		CommonKeywords: []string{"haste", "sacrifice", "tokens", "damage"},
		Description:    "Aggressive swarm tactics with sacrifice synergies",
	},
	"Vampire": {
		Type:           "Vampire",
		Support:        TribalSupportStrong,
		SynergyWeight:  1.3,
		RelatedTypes:   []string{"Noble", "Knight"},
		CommonKeywords: []string{"lifelink", "+1/+1 counters", "blood tokens", "drain"},
		Description:    "Life manipulation and incremental advantage",
	},
	"Zombie": {
		Type:           "Zombie",
		Support:        TribalSupportStrong,
		SynergyWeight:  1.3,
		RelatedTypes:   []string{"Skeleton"},
		CommonKeywords: []string{"graveyard", "death triggers", "tokens", "sacrifice"},
		Description:    "Graveyard recursion and death synergies",
	},
	"Merfolk": {
		Type:           "Merfolk",
		Support:        TribalSupportStrong,
		SynergyWeight:  1.2,
		RelatedTypes:   []string{"Wizard"},
		CommonKeywords: []string{"+1/+1 counters", "unblockable", "card draw"},
		Description:    "Evasion and counter-based growth",
	},
	"Human": {
		Type:           "Human",
		Support:        TribalSupportStrong,
		SynergyWeight:  1.2,
		RelatedTypes:   []string{"Soldier", "Knight", "Cleric", "Wizard"},
		CommonKeywords: []string{"tokens", "anthem", "+1/+1 counters"},
		Description:    "Versatile tribe with wide support across colors",
	},
	"Wizard": {
		Type:           "Wizard",
		Support:        TribalSupportStrong,
		SynergyWeight:  1.2,
		RelatedTypes:   []string{"Human", "Merfolk"},
		CommonKeywords: []string{"spells matter", "card draw", "counters"},
		Description:    "Spell synergies and card advantage",
	},
	"Soldier": {
		Type:           "Soldier",
		Support:        TribalSupportStrong,
		SynergyWeight:  1.2,
		RelatedTypes:   []string{"Human", "Knight", "Warrior"},
		CommonKeywords: []string{"tokens", "anthem", "go wide"},
		Description:    "Token generation and army building",
	},

	// Moderate Support Tribes (10-20 cards)
	"Cat": {
		Type:           "Cat",
		Support:        TribalSupportModerate,
		SynergyWeight:  1.1,
		RelatedTypes:   []string{"Beast"},
		CommonKeywords: []string{"tokens", "lifegain", "+1/+1 counters"},
		Description:    "Token generation and lifegain synergies",
	},
	"Dog": {
		Type:           "Dog",
		Support:        TribalSupportModerate,
		SynergyWeight:  1.0,
		RelatedTypes:   []string{"Beast"},
		CommonKeywords: []string{"tokens", "anthem", "fetch lands"},
		Description:    "Loyal companions with pack tactics",
	},
	"Dinosaur": {
		Type:           "Dinosaur",
		Support:        TribalSupportModerate,
		SynergyWeight:  1.1,
		RelatedTypes:   []string{"Beast"},
		CommonKeywords: []string{"trample", "enrage", "damage"},
		Description:    "Big creatures with damage-based triggers",
	},
	"Dragon": {
		Type:           "Dragon",
		Support:        TribalSupportModerate,
		SynergyWeight:  1.1,
		RelatedTypes:   []string{},
		CommonKeywords: []string{"flying", "treasure tokens", "damage"},
		Description:    "Powerful flyers with treasure generation",
	},
	"Angel": {
		Type:           "Angel",
		Support:        TribalSupportModerate,
		SynergyWeight:  1.1,
		RelatedTypes:   []string{},
		CommonKeywords: []string{"flying", "lifelink", "vigilance"},
		Description:    "Evasive threats with lifegain",
	},
	"Demon": {
		Type:           "Demon",
		Support:        TribalSupportModerate,
		SynergyWeight:  1.0,
		RelatedTypes:   []string{"Devil"},
		CommonKeywords: []string{"flying", "sacrifice", "pay life"},
		Description:    "Powerful creatures with sacrifice costs",
	},
	"Spirit": {
		Type:           "Spirit",
		Support:        TribalSupportModerate,
		SynergyWeight:  1.1,
		RelatedTypes:   []string{},
		CommonKeywords: []string{"flying", "disturb", "enchantments matter"},
		Description:    "Evasive creatures with graveyard synergies",
	},
	"Rogue": {
		Type:           "Rogue",
		Support:        TribalSupportModerate,
		SynergyWeight:  1.1,
		RelatedTypes:   []string{"Assassin", "Faerie"},
		CommonKeywords: []string{"mill", "flash", "evasion"},
		Description:    "Mill strategies and sneaky tactics",
	},
	"Warrior": {
		Type:           "Warrior",
		Support:        TribalSupportModerate,
		SynergyWeight:  1.0,
		RelatedTypes:   []string{"Berserker", "Barbarian"},
		CommonKeywords: []string{"haste", "attack triggers", "+1/+1 counters"},
		Description:    "Aggressive combat-focused creatures",
	},
	"Knight": {
		Type:           "Knight",
		Support:        TribalSupportModerate,
		SynergyWeight:  1.1,
		RelatedTypes:   []string{"Human", "Soldier"},
		CommonKeywords: []string{"first strike", "vigilance", "equipment"},
		Description:    "Combat specialists with equipment synergies",
	},
	"Faerie": {
		Type:           "Faerie",
		Support:        TribalSupportModerate,
		SynergyWeight:  1.1,
		RelatedTypes:   []string{"Rogue"},
		CommonKeywords: []string{"flying", "flash", "spells matter"},
		Description:    "Tricky flyers with instant-speed plays",
	},

	// Weak/Niche Support Tribes
	"Monk": {
		Type:           "Monk",
		Support:        TribalSupportWeak,
		SynergyWeight:  0.9,
		RelatedTypes:   []string{"Human"},
		CommonKeywords: []string{"prowess", "spells matter", "+1/+1 counters"},
		Description:    "Spell-focused combat creatures",
	},
	"Ninja": {
		Type:           "Ninja",
		Support:        TribalSupportWeak,
		SynergyWeight:  1.0,
		RelatedTypes:   []string{"Rogue"},
		CommonKeywords: []string{"ninjutsu", "evasion", "card draw"},
		Description:    "Combat damage triggers with ninjutsu",
	},
	"Samurai": {
		Type:           "Samurai",
		Support:        TribalSupportWeak,
		SynergyWeight:  0.9,
		RelatedTypes:   []string{"Warrior", "Human"},
		CommonKeywords: []string{"attack triggers", "first strike", "vigilance"},
		Description:    "Solo attackers with combat bonuses",
	},
	"Pirate": {
		Type:           "Pirate",
		Support:        TribalSupportWeak,
		SynergyWeight:  1.0,
		RelatedTypes:   []string{"Rogue"},
		CommonKeywords: []string{"treasure tokens", "evasion", "raid"},
		Description:    "Treasure generation and raid mechanics",
	},
	"Cleric": {
		Type:           "Cleric",
		Support:        TribalSupportWeak,
		SynergyWeight:  0.9,
		RelatedTypes:   []string{"Human"},
		CommonKeywords: []string{"lifegain", "party", "removal"},
		Description:    "Life manipulation and party synergies",
	},
	"Shaman": {
		Type:           "Shaman",
		Support:        TribalSupportWeak,
		SynergyWeight:  0.9,
		RelatedTypes:   []string{"Elf", "Goblin"},
		CommonKeywords: []string{"mana ramp", "sacrifice", "ETB"},
		Description:    "Mana and sacrifice abilities",
	},
	"Elemental": {
		Type:           "Elemental",
		Support:        TribalSupportModerate,
		SynergyWeight:  1.0,
		RelatedTypes:   []string{},
		CommonKeywords: []string{"landfall", "evoke", "ETB"},
		Description:    "Land synergies and evoke abilities",
	},
	"Beast": {
		Type:           "Beast",
		Support:        TribalSupportWeak,
		SynergyWeight:  0.8,
		RelatedTypes:   []string{"Cat", "Dog", "Dinosaur"},
		CommonKeywords: []string{"trample", "+1/+1 counters"},
		Description:    "Generic green creatures with size",
	},
	"Artifact Creature": {
		Type:           "Artifact Creature",
		Support:        TribalSupportModerate,
		SynergyWeight:  1.1,
		RelatedTypes:   []string{"Construct", "Golem"},
		CommonKeywords: []string{"artifacts matter", "modular", "+1/+1 counters"},
		Description:    "Artifact synergies and modular mechanics",
	},
	"Construct": {
		Type:           "Construct",
		Support:        TribalSupportWeak,
		SynergyWeight:  0.9,
		RelatedTypes:   []string{"Artifact Creature", "Golem"},
		CommonKeywords: []string{"artifacts matter", "tokens"},
		Description:    "Artifact creature tokens and synergies",
	},
	"Rat": {
		Type:           "Rat",
		Support:        TribalSupportWeak,
		SynergyWeight:  1.0,
		RelatedTypes:   []string{},
		CommonKeywords: []string{"tokens", "discard", "sacrifice"},
		Description:    "Swarm tactics and discard synergies",
	},
	"Bird": {
		Type:           "Bird",
		Support:        TribalSupportWeak,
		SynergyWeight:  0.8,
		RelatedTypes:   []string{},
		CommonKeywords: []string{"flying", "tokens"},
		Description:    "Evasive flyers",
	},
	"Phyrexian": {
		Type:           "Phyrexian",
		Support:        TribalSupportModerate,
		SynergyWeight:  1.0,
		RelatedTypes:   []string{},
		CommonKeywords: []string{"toxic", "proliferate", "oil counters"},
		Description:    "Poison and counter synergies",
	},
}

// GetTribalInfo returns tribal information for a creature type.
// Returns nil if the creature type is not in the database.
func GetTribalInfo(creatureType string) *TribalInfo {
	if info, ok := tribalDatabase[creatureType]; ok {
		return &info
	}
	return nil
}

// GetTribalSynergyWeight returns the synergy weight for a creature type.
// Returns 1.0 (neutral) if the creature type is not in the database.
func GetTribalSynergyWeight(creatureType string) float64 {
	if info, ok := tribalDatabase[creatureType]; ok {
		return info.SynergyWeight
	}
	return 1.0
}

// GetRelatedTypes returns creature types that synergize with the given type.
func GetRelatedTypes(creatureType string) []string {
	if info, ok := tribalDatabase[creatureType]; ok {
		return info.RelatedTypes
	}
	return nil
}

// IsStrongTribalSupport returns true if the creature type has strong tribal support.
func IsStrongTribalSupport(creatureType string) bool {
	if info, ok := tribalDatabase[creatureType]; ok {
		return info.Support == TribalSupportStrong
	}
	return false
}

// GetTribalKeywords returns common keywords associated with a creature type.
func GetTribalKeywords(creatureType string) []string {
	if info, ok := tribalDatabase[creatureType]; ok {
		return info.CommonKeywords
	}
	return nil
}

// IsChangeling checks if a card has changeling (counts as all creature types).
func IsChangeling(oracleText string) bool {
	if oracleText == "" {
		return false
	}
	return containsPattern(oracleText, "changeling") ||
		containsPattern(oracleText, "is every creature type")
}

// GetAllSupportedTribes returns a list of all creature types in the tribal database.
func GetAllSupportedTribes() []string {
	tribes := make([]string, 0, len(tribalDatabase))
	for tribe := range tribalDatabase {
		tribes = append(tribes, tribe)
	}
	return tribes
}

// GetStrongTribes returns creature types with strong tribal support.
func GetStrongTribes() []string {
	var tribes []string
	for tribe, info := range tribalDatabase {
		if info.Support == TribalSupportStrong {
			tribes = append(tribes, tribe)
		}
	}
	return tribes
}
