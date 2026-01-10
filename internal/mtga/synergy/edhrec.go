package synergy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// EDHRECClient fetches synergy data from EDHREC's JSON API.
type EDHRECClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewEDHRECClient creates a new EDHREC API client.
func NewEDHRECClient() *EDHRECClient {
	return &EDHRECClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://json.edhrec.com/pages",
	}
}

// EDHRECCardPage represents the response from the card endpoint.
type EDHRECCardPage struct {
	Container   *EDHRECContainer `json:"container"`
	Similar     []*EDHRECSimilar `json:"similar"`
	Description string           `json:"description"`
}

// EDHRECThemePage represents the response from the theme endpoint.
type EDHRECThemePage struct {
	Container   *EDHRECContainer `json:"container"`
	Header      string           `json:"header"`
	Description string           `json:"description"`
}

// EDHRECContainer holds the main data structure.
type EDHRECContainer struct {
	JSONDict    *EDHRECJSONDict `json:"json_dict"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
}

// EDHRECJSONDict contains card lists and main card info.
type EDHRECJSONDict struct {
	CardLists []*EDHRECCardList `json:"cardlists"`
	Card      *EDHRECCardInfo   `json:"card"`
}

// EDHRECCardList represents a categorized list of cards.
type EDHRECCardList struct {
	Tag       string            `json:"tag"`
	Header    string            `json:"header"`
	CardViews []*EDHRECCardView `json:"cardviews"`
}

// EDHRECCardView represents a card with synergy information.
type EDHRECCardView struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Sanitized      string  `json:"sanitized"`
	URL            string  `json:"url"`
	Synergy        float64 `json:"synergy"`
	Lift           float64 `json:"lift"`
	Inclusion      int     `json:"inclusion"`
	NumDecks       int     `json:"num_decks"`
	PotentialDecks int     `json:"potential_decks"`
}

// EDHRECCardInfo represents the main card being queried.
type EDHRECCardInfo struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Sanitized     string   `json:"sanitized"`
	CMC           float64  `json:"cmc"`
	ColorIdentity []string `json:"color_identity"`
	Salt          float64  `json:"salt"`
	NumDecks      int      `json:"num_decks"`
}

// EDHRECSimilar represents similar cards.
type EDHRECSimilar struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Sanitized     string   `json:"sanitized"`
	CMC           float64  `json:"cmc"`
	ColorIdentity []string `json:"color_identity"`
	PrimaryType   string   `json:"primary_type"`
	Rarity        string   `json:"rarity"`
	Salt          float64  `json:"salt"`
}

// EDHRECSynergyData contains synergy information for a card.
type EDHRECSynergyData struct {
	CardName      string
	HighSynergy   []*EDHRECCardView
	TopCards      []*EDHRECCardView
	NewCards      []*EDHRECCardView
	SimilarCards  []*EDHRECSimilar
	NumDecks      int
	Salt          float64
	ColorIdentity []string
}

// EDHRECThemeData contains theme information.
type EDHRECThemeData struct {
	ThemeName    string
	Description  string
	HighSynergy  []*EDHRECCardView
	TopCards     []*EDHRECCardView
	Creatures    []*EDHRECCardView
	Enchantments []*EDHRECCardView
	Artifacts    []*EDHRECCardView
}

// SanitizeCardName converts a card name to EDHREC's URL format.
func SanitizeCardName(name string) string {
	// Convert to lowercase and replace spaces with hyphens
	sanitized := strings.ToLower(name)
	sanitized = strings.ReplaceAll(sanitized, " ", "-")
	sanitized = strings.ReplaceAll(sanitized, "'", "")
	sanitized = strings.ReplaceAll(sanitized, ",", "")
	sanitized = strings.ReplaceAll(sanitized, ":", "")
	return sanitized
}

// GetCardSynergy fetches synergy data for a specific card.
func (c *EDHRECClient) GetCardSynergy(ctx context.Context, cardName string) (*EDHRECSynergyData, error) {
	sanitized := SanitizeCardName(cardName)
	url := fmt.Sprintf("%s/cards/%s.json", c.baseURL, sanitized)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "MTGA-Companion/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch card data: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("card not found: %s", cardName)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var page EDHRECCardPage
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return c.extractSynergyData(cardName, &page), nil
}

// extractSynergyData extracts synergy information from the card page.
func (c *EDHRECClient) extractSynergyData(cardName string, page *EDHRECCardPage) *EDHRECSynergyData {
	data := &EDHRECSynergyData{
		CardName:     cardName,
		SimilarCards: page.Similar,
	}

	if page.Container == nil || page.Container.JSONDict == nil {
		return data
	}

	// Extract card info
	if page.Container.JSONDict.Card != nil {
		card := page.Container.JSONDict.Card
		data.NumDecks = card.NumDecks
		data.Salt = card.Salt
		data.ColorIdentity = card.ColorIdentity
	}

	// Extract card lists by tag
	for _, cardList := range page.Container.JSONDict.CardLists {
		switch cardList.Tag {
		case "highsynergycards":
			data.HighSynergy = cardList.CardViews
		case "topcards":
			data.TopCards = cardList.CardViews
		case "newcards":
			data.NewCards = cardList.CardViews
		}
	}

	return data
}

// GetThemeSynergy fetches synergy data for a theme (e.g., "tokens", "aristocrats").
func (c *EDHRECClient) GetThemeSynergy(ctx context.Context, themeName string) (*EDHRECThemeData, error) {
	sanitized := SanitizeCardName(themeName)
	url := fmt.Sprintf("%s/themes/%s.json", c.baseURL, sanitized)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "MTGA-Companion/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch theme data: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("theme not found: %s", themeName)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var page EDHRECThemePage
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return c.extractThemeData(themeName, &page), nil
}

// extractThemeData extracts theme information from the theme page.
func (c *EDHRECClient) extractThemeData(themeName string, page *EDHRECThemePage) *EDHRECThemeData {
	data := &EDHRECThemeData{
		ThemeName:   themeName,
		Description: page.Description,
	}

	if page.Container == nil || page.Container.JSONDict == nil {
		return data
	}

	// Extract card lists by tag
	for _, cardList := range page.Container.JSONDict.CardLists {
		switch cardList.Tag {
		case "highsynergycards":
			data.HighSynergy = cardList.CardViews
		case "topcards":
			data.TopCards = cardList.CardViews
		case "creatures":
			data.Creatures = cardList.CardViews
		case "enchantments":
			data.Enchantments = cardList.CardViews
		case "utilityartifacts":
			data.Artifacts = cardList.CardViews
		}
	}

	return data
}

// GetHighSynergyCards returns the top synergy cards for a given card.
// Limit controls how many cards to return (0 = all).
func (c *EDHRECClient) GetHighSynergyCards(ctx context.Context, cardName string, limit int) ([]*EDHRECCardView, error) {
	data, err := c.GetCardSynergy(ctx, cardName)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > len(data.HighSynergy) {
		return data.HighSynergy, nil
	}

	return data.HighSynergy[:limit], nil
}

// KnownThemes returns a list of known EDHREC themes that can be queried.
var KnownThemes = []string{
	"tokens",
	"aristocrats",
	"counters",
	"blink",
	"reanimator",
	"spellslinger",
	"artifacts",
	"enchantments",
	"lifegain",
	"mill",
	"voltron",
	"landfall",
	"graveyard",
	"sacrifice",
	"clones",
	"equipment",
	"wheels",
	"extra-turns",
	"extra-combats",
	"big-mana",
}
