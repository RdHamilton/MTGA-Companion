package cards

import "time"

// Card represents comprehensive metadata about a Magic card.
type Card struct {
	// MTGA Arena ID (primary identifier in MTGA)
	ArenaID int `json:"arena_id"`

	// Scryfall identifiers
	ScryfallID   string  `json:"id"`
	OracleID     *string `json:"oracle_id"`
	MultiverseID *int    `json:"multiverse_ids,omitempty"`

	// Basic card information
	Name     string `json:"name"`
	TypeLine string `json:"type_line"`
	SetCode  string `json:"set"`
	SetName  string `json:"set_name"`

	// Mana information
	ManaCost *string `json:"mana_cost"`
	CMC      float64 `json:"cmc"` // Converted mana cost

	// Colors and identity
	Colors        []string `json:"colors"`
	ColorIdentity []string `json:"color_identity"`

	// Rarity and legality
	Rarity string `json:"rarity"` // "common", "uncommon", "rare", "mythic"

	// Power/Toughness (for creatures)
	Power     *string `json:"power,omitempty"`
	Toughness *string `json:"toughness,omitempty"`

	// Loyalty (for planeswalkers)
	Loyalty *string `json:"loyalty,omitempty"`

	// Text and imagery
	OracleText *string `json:"oracle_text,omitempty"`
	FlavorText *string `json:"flavor_text,omitempty"`
	ImageURI   *string `json:"image_uri,omitempty"`

	// Layout information
	Layout string `json:"layout"` // "normal", "split", "transform", etc.

	// Metadata
	CollectorNumber string    `json:"collector_number"`
	ReleasedAt      time.Time `json:"released_at"`
}

// CardFace represents one face of a multi-faced card.
type CardFace struct {
	Name      string   `json:"name"`
	TypeLine  string   `json:"type_line"`
	ManaCost  *string  `json:"mana_cost"`
	OracleText *string `json:"oracle_text"`
	Colors    []string `json:"colors"`
	Power     *string  `json:"power,omitempty"`
	Toughness *string  `json:"toughness,omitempty"`
	Loyalty   *string  `json:"loyalty,omitempty"`
	ImageURI  *string  `json:"image_uri,omitempty"`
}

// ScryfallCard represents the Scryfall API response format.
// This matches Scryfall's card object schema.
type ScryfallCard struct {
	ID              string        `json:"id"`
	OracleID        string        `json:"oracle_id"`
	MultiverseIDs   []int         `json:"multiverse_ids,omitempty"`
	ArenaID         int           `json:"arena_id"`
	Name            string        `json:"name"`
	Lang            string        `json:"lang"`
	ReleasedAt      string        `json:"released_at"`
	URI             string        `json:"uri"`
	ScryfallURI     string        `json:"scryfall_uri"`
	Layout          string        `json:"layout"`
	ImageURIs       *ImageURIs    `json:"image_uris,omitempty"`
	ManaCost        string        `json:"mana_cost"`
	CMC             float64       `json:"cmc"`
	TypeLine        string        `json:"type_line"`
	OracleText      string        `json:"oracle_text,omitempty"`
	Power           string        `json:"power,omitempty"`
	Toughness       string        `json:"toughness,omitempty"`
	Loyalty         string        `json:"loyalty,omitempty"`
	Colors          []string      `json:"colors"`
	ColorIdentity   []string      `json:"color_identity"`
	Set             string        `json:"set"`
	SetName         string        `json:"set_name"`
	CollectorNumber string        `json:"collector_number"`
	Rarity          string        `json:"rarity"`
	FlavorText      string        `json:"flavor_text,omitempty"`
	CardFaces       []ScryfallCardFace `json:"card_faces,omitempty"`
}

// ScryfallCardFace represents a face of a multi-faced card in Scryfall format.
type ScryfallCardFace struct {
	Name       string     `json:"name"`
	TypeLine   string     `json:"type_line"`
	ManaCost   string     `json:"mana_cost"`
	OracleText string     `json:"oracle_text"`
	Colors     []string   `json:"colors"`
	Power      string     `json:"power,omitempty"`
	Toughness  string     `json:"toughness,omitempty"`
	Loyalty    string     `json:"loyalty,omitempty"`
	ImageURIs  *ImageURIs `json:"image_uris,omitempty"`
}

// ImageURIs contains URLs for card images in various sizes.
type ImageURIs struct {
	Small      string `json:"small"`
	Normal     string `json:"normal"`
	Large      string `json:"large"`
	PNG        string `json:"png"`
	ArtCrop    string `json:"art_crop"`
	BorderCrop string `json:"border_crop"`
}

// ToCard converts a ScryfallCard to our internal Card representation.
func (sc *ScryfallCard) ToCard() *Card {
	releasedAt, _ := time.Parse("2006-01-02", sc.ReleasedAt)

	card := &Card{
		ArenaID:         sc.ArenaID,
		ScryfallID:      sc.ID,
		OracleID:        &sc.OracleID,
		Name:            sc.Name,
		TypeLine:        sc.TypeLine,
		SetCode:         sc.Set,
		SetName:         sc.SetName,
		CMC:             sc.CMC,
		Colors:          sc.Colors,
		ColorIdentity:   sc.ColorIdentity,
		Rarity:          sc.Rarity,
		Layout:          sc.Layout,
		CollectorNumber: sc.CollectorNumber,
		ReleasedAt:      releasedAt,
	}

	// Handle mana cost
	if sc.ManaCost != "" {
		card.ManaCost = &sc.ManaCost
	}

	// Handle multiverse IDs
	if len(sc.MultiverseIDs) > 0 {
		card.MultiverseID = &sc.MultiverseIDs[0]
	}

	// Handle power/toughness
	if sc.Power != "" {
		card.Power = &sc.Power
	}
	if sc.Toughness != "" {
		card.Toughness = &sc.Toughness
	}

	// Handle loyalty
	if sc.Loyalty != "" {
		card.Loyalty = &sc.Loyalty
	}

	// Handle oracle text
	if sc.OracleText != "" {
		card.OracleText = &sc.OracleText
	}

	// Handle flavor text
	if sc.FlavorText != "" {
		card.FlavorText = &sc.FlavorText
	}

	// Handle image URIs
	if sc.ImageURIs != nil && sc.ImageURIs.Normal != "" {
		card.ImageURI = &sc.ImageURIs.Normal
	}

	return card
}
