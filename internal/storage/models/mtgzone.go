package models

import "time"

// CardRole represents the role of a card in an archetype.
type CardRole string

const (
	CardRoleCore      CardRole = "core"
	CardRoleFlex      CardRole = "flex"
	CardRoleSideboard CardRole = "sideboard"
)

// CardImportance represents how important a card is to an archetype.
type CardImportance string

const (
	CardImportanceEssential CardImportance = "essential"
	CardImportanceImportant CardImportance = "important"
	CardImportanceOptional  CardImportance = "optional"
)

// ArticleType represents the type of MTGZone article.
type ArticleType string

const (
	ArticleTypeDeckGuide ArticleType = "deck_guide"
	ArticleTypeTierList  ArticleType = "tier_list"
	ArticleTypeSetReview ArticleType = "set_review"
	ArticleTypeStrategy  ArticleType = "strategy"
)

// MTGZoneArchetype represents an archetype definition from MTGZone.
type MTGZoneArchetype struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Format      string    `json:"format" db:"format"`
	Tier        string    `json:"tier,omitempty" db:"tier"`
	Description string    `json:"description,omitempty" db:"description"`
	PlayStyle   string    `json:"playStyle,omitempty" db:"play_style"`
	SourceURL   string    `json:"sourceUrl,omitempty" db:"source_url"`
	LastUpdated time.Time `json:"lastUpdated" db:"last_updated"`
}

// MTGZoneArchetypeCard represents a card that belongs to an archetype.
type MTGZoneArchetypeCard struct {
	ID          int64          `json:"id" db:"id"`
	ArchetypeID int64          `json:"archetypeId" db:"archetype_id"`
	CardName    string         `json:"cardName" db:"card_name"`
	Role        CardRole       `json:"role" db:"role"`
	Copies      int            `json:"copies" db:"copies"`
	Importance  CardImportance `json:"importance,omitempty" db:"importance"`
	Notes       string         `json:"notes,omitempty" db:"notes"`
	LastUpdated time.Time      `json:"lastUpdated" db:"last_updated"`
}

// MTGZoneSynergy represents a synergy relationship with explanation.
type MTGZoneSynergy struct {
	ID               int64     `json:"id" db:"id"`
	CardA            string    `json:"cardA" db:"card_a"`
	CardB            string    `json:"cardB" db:"card_b"`
	Reason           string    `json:"reason" db:"reason"`
	SourceURL        string    `json:"sourceUrl,omitempty" db:"source_url"`
	ArchetypeContext string    `json:"archetypeContext,omitempty" db:"archetype_context"`
	Confidence       float64   `json:"confidence" db:"confidence"`
	LastUpdated      time.Time `json:"lastUpdated" db:"last_updated"`
}

// MTGZoneArticle represents metadata about a processed article.
type MTGZoneArticle struct {
	ID             int64       `json:"id" db:"id"`
	URL            string      `json:"url" db:"url"`
	Title          string      `json:"title" db:"title"`
	ArticleType    ArticleType `json:"articleType,omitempty" db:"article_type"`
	Format         string      `json:"format,omitempty" db:"format"`
	Archetype      string      `json:"archetype,omitempty" db:"archetype"`
	PublishedAt    *time.Time  `json:"publishedAt,omitempty" db:"published_at"`
	ProcessedAt    time.Time   `json:"processedAt" db:"processed_at"`
	CardsMentioned string      `json:"cardsMentioned,omitempty" db:"cards_mentioned"` // JSON array
}

// ArchetypeWithCards combines an archetype with its card list.
type ArchetypeWithCards struct {
	Archetype MTGZoneArchetype       `json:"archetype"`
	CoreCards []MTGZoneArchetypeCard `json:"coreCards"`
	FlexCards []MTGZoneArchetypeCard `json:"flexCards"`
	Sideboard []MTGZoneArchetypeCard `json:"sideboard"`
}

// GetTierRank returns a numeric rank for tier comparison (lower is better).
func GetTierRank(tier string) int {
	ranks := map[string]int{
		"S":  1,
		"A":  2,
		"A+": 2,
		"A-": 3,
		"B":  4,
		"B+": 4,
		"B-": 5,
		"C":  6,
		"C+": 6,
		"C-": 7,
		"D":  8,
		"F":  9,
	}
	if rank, ok := ranks[tier]; ok {
		return rank
	}
	return 10 // Unknown tier
}

// IsTopTier returns true if the tier is S, A, or A+.
func IsTopTier(tier string) bool {
	return GetTierRank(tier) <= 2
}
