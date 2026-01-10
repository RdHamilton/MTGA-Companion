package models

import "time"

// CFB Limited Rating grades (A+ to F).
const (
	CFBLimitedGradeAPlus  = "A+"
	CFBLimitedGradeA      = "A"
	CFBLimitedGradeAMinus = "A-"
	CFBLimitedGradeBPlus  = "B+"
	CFBLimitedGradeB      = "B"
	CFBLimitedGradeBMinus = "B-"
	CFBLimitedGradeCPlus  = "C+"
	CFBLimitedGradeC      = "C"
	CFBLimitedGradeCMinus = "C-"
	CFBLimitedGradeD      = "D"
	CFBLimitedGradeF      = "F"
)

// CFB Constructed playability ratings.
const (
	CFBConstructedStaple     = "Staple"
	CFBConstructedPlayable   = "Playable"
	CFBConstructedFringe     = "Fringe"
	CFBConstructedUnplayable = "Unplayable"
)

// CFBRating represents a ChannelFireball card rating from a set review.
type CFBRating struct {
	ID       int64  `json:"id" db:"id"`
	CardName string `json:"cardName" db:"card_name"`
	SetCode  string `json:"setCode" db:"set_code"`
	ArenaID  *int   `json:"arenaId,omitempty" db:"arena_id"`

	// Limited rating (A+, A, A-, B+, B, B-, C+, C, C-, D, F)
	LimitedRating string  `json:"limitedRating" db:"limited_rating"`
	LimitedScore  float64 `json:"limitedScore" db:"limited_score"`

	// Constructed rating (Staple, Playable, Fringe, Unplayable)
	ConstructedRating string  `json:"constructedRating" db:"constructed_rating"`
	ConstructedScore  float64 `json:"constructedScore" db:"constructed_score"`

	// Archetype fit (e.g., "Best in Aggro", "Flexible", "Control only")
	ArchetypeFit string `json:"archetypeFit,omitempty" db:"archetype_fit"`

	// Commentary from the CFB review
	Commentary string `json:"commentary,omitempty" db:"commentary"`

	// Source information
	SourceURL string `json:"sourceUrl,omitempty" db:"source_url"`
	Author    string `json:"author,omitempty" db:"author"`

	// Metadata
	ImportedAt time.Time `json:"importedAt" db:"imported_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
}

// LimitedGradeToScore converts a letter grade to a numeric score (0.0-1.0).
func LimitedGradeToScore(grade string) float64 {
	scores := map[string]float64{
		CFBLimitedGradeAPlus:  1.00,
		CFBLimitedGradeA:      0.92,
		CFBLimitedGradeAMinus: 0.85,
		CFBLimitedGradeBPlus:  0.78,
		CFBLimitedGradeB:      0.70,
		CFBLimitedGradeBMinus: 0.62,
		CFBLimitedGradeCPlus:  0.55,
		CFBLimitedGradeC:      0.48,
		CFBLimitedGradeCMinus: 0.40,
		CFBLimitedGradeD:      0.30,
		CFBLimitedGradeF:      0.15,
	}
	if score, ok := scores[grade]; ok {
		return score
	}
	return 0.5 // Default for unknown grades
}

// ConstructedRatingToScore converts a constructed playability rating to a numeric score (0.0-1.0).
func ConstructedRatingToScore(rating string) float64 {
	scores := map[string]float64{
		CFBConstructedStaple:     1.00,
		CFBConstructedPlayable:   0.70,
		CFBConstructedFringe:     0.40,
		CFBConstructedUnplayable: 0.10,
	}
	if score, ok := scores[rating]; ok {
		return score
	}
	return 0.5 // Default for unknown ratings
}

// CFBRatingImport represents the structure for importing CFB ratings from JSON.
type CFBRatingImport struct {
	CardName          string `json:"card_name"`
	SetCode           string `json:"set_code"`
	LimitedRating     string `json:"limited_rating"`
	ConstructedRating string `json:"constructed_rating,omitempty"`
	ArchetypeFit      string `json:"archetype_fit,omitempty"`
	Commentary        string `json:"commentary,omitempty"`
	SourceURL         string `json:"source_url,omitempty"`
	Author            string `json:"author,omitempty"`
}
