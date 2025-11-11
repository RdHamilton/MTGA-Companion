package scryfall

import (
	"fmt"
	"time"
)

// Card represents a Magic card from Scryfall.
type Card struct {
	// Core fields
	ID       string `json:"id"`
	OracleID string `json:"oracle_id"`
	ArenaID  *int   `json:"arena_id,omitempty"`

	// Card details
	Name          string     `json:"name"`
	Lang          string     `json:"lang"`
	ReleasedAt    string     `json:"released_at"`
	URI           string     `json:"uri"`
	ScryfallURI   string     `json:"scryfall_uri"`
	Layout        string     `json:"layout"`
	ImageURIs     *ImageURIs `json:"image_uris,omitempty"`
	ManaCost      string     `json:"mana_cost,omitempty"`
	CMC           float64    `json:"cmc"`
	TypeLine      string     `json:"type_line"`
	OracleText    string     `json:"oracle_text,omitempty"`
	Colors        []string   `json:"colors,omitempty"`
	ColorIdentity []string   `json:"color_identity"`
	Keywords      []string   `json:"keywords,omitempty"`

	// Gameplay
	Power        string `json:"power,omitempty"`
	Toughness    string `json:"toughness,omitempty"`
	Loyalty      string `json:"loyalty,omitempty"`
	LifeModifier string `json:"life_modifier,omitempty"`
	HandModifier string `json:"hand_modifier,omitempty"`

	// Print details
	SetCode         string `json:"set"`
	SetName         string `json:"set_name"`
	CollectorNumber string `json:"collector_number"`
	Rarity          string `json:"rarity"`
	Artist          string `json:"artist,omitempty"`
	FlavorText      string `json:"flavor_text,omitempty"`

	// Card faces (for DFCs, MDFCs, split cards)
	CardFaces []CardFace `json:"card_faces,omitempty"`

	// Legality
	Legalities Legalities `json:"legalities"`

	// Prices
	Prices Prices `json:"prices"`

	// Related
	RelatedURIs  map[string]string `json:"related_uris,omitempty"`
	PurchaseURIs map[string]string `json:"purchase_uris,omitempty"`
}

// CardFace represents one face of a multi-faced card.
type CardFace struct {
	Name       string     `json:"name"`
	ManaCost   string     `json:"mana_cost,omitempty"`
	TypeLine   string     `json:"type_line"`
	OracleText string     `json:"oracle_text,omitempty"`
	Colors     []string   `json:"colors,omitempty"`
	Power      string     `json:"power,omitempty"`
	Toughness  string     `json:"toughness,omitempty"`
	Loyalty    string     `json:"loyalty,omitempty"`
	Artist     string     `json:"artist,omitempty"`
	ImageURIs  *ImageURIs `json:"image_uris,omitempty"`
	FlavorText string     `json:"flavor_text,omitempty"`
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

// Legalities represents the legality of a card in various formats.
type Legalities struct {
	Standard        string `json:"standard"`
	Future          string `json:"future"`
	Historic        string `json:"historic"`
	Gladiator       string `json:"gladiator"`
	Pioneer         string `json:"pioneer"`
	Explorer        string `json:"explorer"`
	Modern          string `json:"modern"`
	Legacy          string `json:"legacy"`
	Pauper          string `json:"pauper"`
	Vintage         string `json:"vintage"`
	Penny           string `json:"penny"`
	Commander       string `json:"commander"`
	Oathbreaker     string `json:"oathbreaker"`
	Brawl           string `json:"brawl"`
	HistoricBrawl   string `json:"historicbrawl"`
	Alchemy         string `json:"alchemy"`
	PauperCommander string `json:"paupercommander"`
	Duel            string `json:"duel"`
	OldSchool       string `json:"oldschool"`
	Premodern       string `json:"premodern"`
	Predh           string `json:"predh"`
}

// Prices represents the prices of a card in various currencies.
type Prices struct {
	USD       *string `json:"usd,omitempty"`
	USDFoil   *string `json:"usd_foil,omitempty"`
	USDEtched *string `json:"usd_etched,omitempty"`
	EUR       *string `json:"eur,omitempty"`
	EURFoil   *string `json:"eur_foil,omitempty"`
	TIX       *string `json:"tix,omitempty"`
}

// Set represents a Magic set from Scryfall.
type Set struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	URI         string `json:"uri"`
	ScryfallURI string `json:"scryfall_uri"`
	SearchURI   string `json:"search_uri"`
	ReleasedAt  string `json:"released_at,omitempty"`
	SetType     string `json:"set_type"`
	CardCount   int    `json:"card_count"`
	Digital     bool   `json:"digital"`
	FoilOnly    bool   `json:"foil_only"`
	IconSVGURI  string `json:"icon_svg_uri"`
}

// SetList represents a list of sets from Scryfall.
type SetList struct {
	Object  string `json:"object"`
	HasMore bool   `json:"has_more"`
	Data    []Set  `json:"data"`
}

// SearchResult represents search results from Scryfall.
type SearchResult struct {
	Object     string `json:"object"`
	TotalCards int    `json:"total_cards"`
	HasMore    bool   `json:"has_more"`
	NextPage   string `json:"next_page,omitempty"`
	Data       []Card `json:"data"`
}

// BulkDataList represents the list of bulk data files.
type BulkDataList struct {
	Object  string     `json:"object"`
	HasMore bool       `json:"has_more"`
	Data    []BulkData `json:"data"`
}

// BulkData represents a bulk data file download.
type BulkData struct {
	ID              string    `json:"id"`
	Object          string    `json:"object"`
	Type            string    `json:"type"`
	UpdatedAt       time.Time `json:"updated_at"`
	URI             string    `json:"uri"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	CompressedSize  int       `json:"compressed_size"`
	DownloadURI     string    `json:"download_uri"`
	ContentType     string    `json:"content_type"`
	ContentEncoding string    `json:"content_encoding"`
}

// APIError represents an error response from the Scryfall API.
type APIError struct {
	Object   string   `json:"object"`
	Code     string   `json:"code"`
	Status   int      `json:"status"`
	Details  string   `json:"details"`
	Type     string   `json:"type,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// Error implements the error interface for APIError.
func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("Scryfall API error (HTTP %d): %s", e.Status, e.Details)
	}
	return fmt.Sprintf("Scryfall API error (HTTP %d): %s", e.Status, e.Code)
}

// NotFoundError represents a 404 error from the API.
type NotFoundError struct {
	URL string
}

// Error implements the error interface for NotFoundError.
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("resource not found: %s", e.URL)
}

// IsNotFound returns true if the error is a NotFoundError.
func IsNotFound(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// Migration represents a card migration from Scryfall.
// Migrations occur when Scryfall merges duplicate cards or removes invalid entries.
type Migration struct {
	ID                string    `json:"id"`
	URI               string    `json:"uri"`
	PerformedAt       time.Time `json:"performed_at"`
	MigrationStrategy string    `json:"migration_strategy"` // "merge" or "delete"
	OldScryfallID     string    `json:"old_scryfall_id"`
	NewScryfallID     *string   `json:"new_scryfall_id,omitempty"` // Only present for merge strategy
	Note              string    `json:"note"`
}

// MigrationList represents a list of migrations from Scryfall.
type MigrationList struct {
	Object   string      `json:"object"`
	HasMore  bool        `json:"has_more"`
	NextPage string      `json:"next_page,omitempty"`
	Data     []Migration `json:"data"`
}
