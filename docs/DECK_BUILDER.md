# Deck Builder Documentation

The MTGA Companion deck builder is a comprehensive system for creating, managing, and analyzing Magic: The Gathering Arena decks. It supports both constructed and draft-based deck building with AI-powered card recommendations, detailed statistics, and multiple import/export formats.

## Table of Contents

- [Overview](#overview)
- [User Guide](#user-guide)
  - [Building Constructed Decks](#building-constructed-decks)
  - [Building Draft Decks](#building-draft-decks)
  - [Managing Your Deck Library](#managing-your-deck-library)
- [AI-Powered Recommendations](#ai-powered-recommendations)
  - [How It Works](#how-it-works)
  - [Understanding Recommendation Scores](#understanding-recommendation-scores)
  - [Using Recommendations](#using-recommendations)
- [Import & Export Formats](#import--export-formats)
  - [Supported Import Formats](#supported-import-formats)
  - [Supported Export Formats](#supported-export-formats)
  - [Format Examples](#format-examples)
- [Deck Statistics](#deck-statistics)
  - [Understanding Your Stats](#understanding-your-stats)
  - [Format Legality](#format-legality)
- [Developer API](#developer-api)
  - [Deck Management API](#deck-management-api)
  - [Card Recommendation API](#card-recommendation-api)
  - [Statistics API](#statistics-api)
- [FAQ](#faq)

---

## Overview

The deck builder provides:

- **Deck Creation**: Build decks from scratch or import existing lists
- **Draft Integration**: Create decks directly from draft pools with validation
- **AI Recommendations**: Get intelligent card suggestions based on your deck composition
- **Comprehensive Statistics**: Analyze mana curve, color distribution, type breakdown, and more
- **Performance Tracking**: Track win rates and performance metrics for each deck
- **Multiple Formats**: Import/export in Arena, MTGO, MTGGoldfish, and plain text formats
- **Tagging & Organization**: Categorize decks with custom tags
- **Format Validation**: Check deck legality across Standard, Historic, Explorer, Alchemy, Brawl, and Commander

---

## User Guide

### Building Constructed Decks

#### Creating a New Deck

To create a constructed deck:

1. **Via API**: Call `CreateDeck` with:
   - `name`: Your deck name (e.g., "Mono Red Aggro")
   - `format`: Format (e.g., "Standard", "Historic", "Explorer")
   - `source`: Set to `"constructed"` for manually built decks

2. **Add Cards**: Use `AddCard` to add cards to your deck:
   - Specify the card ID, quantity, and board ("main" or "sideboard")
   - No restrictions on card pool for constructed decks

3. **Get Recommendations**: Request AI recommendations to fill out your deck

4. **View Statistics**: Check mana curve, color distribution, and format legality

#### Example Workflow

```
1. Create deck: "Azorius Control" (Standard, constructed)
2. Add core cards: 4x Counterspell, 2x Teferi, Hero of Dominaria
3. Get recommendations (filter: blue/white, CMC 2-4)
4. Add recommended cards
5. View statistics to check mana curve
6. Adjust land count based on recommendations (typically 24-26 for Standard)
7. Export to Arena format for use in MTGA
```

### Building Draft Decks

#### Creating a Draft Deck

Draft decks are special because they're restricted to cards from your draft pool:

1. **Automatic Creation**: After completing a draft, create a deck linked to that draft event:
   - `source`: Must be `"draft"`
   - `draftEventID`: Required - links to your draft session

2. **Card Pool Restriction**: You can only add cards that you drafted
   - The system validates each card against your draft pool
   - Attempting to add non-drafted cards will fail with an error

3. **Draft Validation**: Use `ValidateDraftDeck` to ensure all cards are from your draft

#### Draft Deck Best Practices

- **Use Recommendations**: AI recommendations are filtered to your draft pool
- **Check Statistics**: Draft decks typically need 17 lands (40 cards total)
- **Track Performance**: Link deck to matches to see how your draft choices performed

#### Example: Building from a Draft

```
1. Complete draft for "Bloomburrow Premier Draft"
2. Create deck linked to draft event ID
3. Import your draft deck list (cards will be validated against draft pool)
4. Or manually add cards one by one
5. Get recommendations filtered to draft pool only
6. Verify with ValidateDraftDeck
7. Play matches and track performance
```

### Managing Your Deck Library

#### Listing & Filtering Decks

The deck library supports advanced filtering:

- **By Source**: Filter by "draft", "constructed", or "imported"
- **By Format**: Filter by "Standard", "Historic", "Limited", etc.
- **By Tags**: Find decks with specific tags (e.g., "aggro", "competitive")
- **Sorting**: Sort by modified date, created date, name, or performance

#### Organizing with Tags

Tags help categorize your decks:

```
Examples:
- "aggro", "midrange", "control"
- "competitive", "casual", "testing"
- "best-of-1", "best-of-3"
- "budget", "optimized"
```

#### Cloning Decks

Create variations by cloning:
- Clone a deck to test modifications
- Useful for A/B testing card choices
- Maintains original deck's performance history separately

---

## AI-Powered Recommendations

### How It Works

The recommendation engine uses a multi-factor scoring system:

1. **Color Fit Analysis** (0.0-1.0)
   - Analyzes mana costs in existing cards
   - Prioritizes cards matching your color identity
   - Penalizes cards with difficult color requirements

2. **Mana Curve Optimization** (0.0-1.0)
   - Identifies gaps in your mana curve
   - Recommends cards that fill missing CMC slots
   - Balances curve for optimal gameplay

3. **Synergy Detection** (0.0-1.0)
   - Analyzes card types in your deck
   - Identifies tribal synergies, card type interactions
   - Rewards cards that work well together

4. **Card Quality Rating** (0.0-1.0)
   - Uses 17Lands data (when available)
   - Considers win rates, pick rates, and play rates
   - Adapts to format-specific power levels

5. **Format Playability** (0.0-1.0)
   - Checks format legality
   - Considers meta relevance
   - Adapts to draft vs. constructed context

### Understanding Recommendation Scores

Each recommendation includes:

- **Overall Score** (0.0-1.0): Combined weighted score
  - 0.7+: Highly recommended (strong fit)
  - 0.5-0.7: Good option (solid choice)
  - 0.3-0.5: Playable (consider if needed)
  - Below 0.3: Not recommended (filtered by default)

- **Individual Factor Scores**: See breakdown of color fit, curve, synergy, quality, playability

- **Reasoning**: Human-readable explanation of why the card is recommended

- **Confidence** (0.0-1.0): How confident the system is
  - Higher confidence = more data available for scoring
  - Lower confidence = limited data, rely more on heuristics

### Using Recommendations

#### Getting Recommendations

Request recommendations with filters:

```javascript
{
  "deckID": "abc-123",
  "maxResults": 10,          // How many suggestions
  "minScore": 0.3,           // Minimum score threshold
  "colors": ["U", "W"],      // Filter to blue/white
  "cardTypes": ["Creature"], // Only creatures
  "cmcMin": 2,               // CMC range
  "cmcMax": 4,
  "includeLands": true,      // Include land suggestions
  "onlyDraftPool": true      // For draft decks: only show drafted cards
}
```

#### Explaining Recommendations

Get detailed explanation for any card:

```javascript
ExplainRecommendation(deckID, cardID)
// Returns: Detailed breakdown of why this card fits your deck
```

### How the AI Learns

The system currently uses rule-based heuristics. Future enhancements will include:

- **Feedback Learning**: Recording which recommendations you accept/reject
- **Performance Correlation**: Learning from deck win rates
- **Meta Adaptation**: Adjusting to format metagame shifts
- **Personalization**: Learning your deckbuilding preferences

---

## Import & Export Formats

### Supported Import Formats

The deck builder can parse multiple formats automatically:

#### 1. MTGA Arena Format

The native Arena export format:

```
Deck
4 Lightning Bolt (M21) 123
20 Mountain (BLB) 280
4 Monastery Swiftspear (KTK) 118

Sideboard
2 Roast (DTK) 151
3 Abrade (AKH) 136
```

**Features**:
- Includes set codes and collector numbers
- Empty line separates mainboard from sideboard
- Most common format for Arena imports

#### 2. Plain Text Format

Simple quantity and card name:

```
4 Lightning Bolt
20 Mountain
4 Monastery Swiftspear

Sideboard:
2 Roast
3 Abrade
```

**Features**:
- Minimal format, easy to type
- "Sideboard:" header marks sideboard section
- Set codes optional

### Supported Export Formats

Export your decks in multiple formats:

#### 1. Arena Format (`"arena"`)

Native MTGA format with set codes and collector numbers.

#### 2. Plain Text Format (`"plaintext"`)

Simple list format:
```
4x Lightning Bolt
20x Mountain
```

#### 3. MTGO Format (`"mtgo"`)

Format compatible with Magic Online (.dek files).

#### 4. MTGGoldfish Format (`"mtggoldfish"`)

Format for importing to MTGGoldfish.com.

### Format Examples

#### Importing a Deck

```javascript
ImportDeck({
  "name": "Mono Red Aggro",
  "format": "Standard",
  "source": "imported",
  "importText": `
    Deck
    4 Monastery Swiftspear (KTK) 118
    4 Lightning Bolt (M21) 123
    20 Mountain (BLB) 280
  `
})

// Returns:
{
  "success": true,
  "deckID": "abc-123",
  "cardsImported": 28,
  "cardsSkipped": 0,
  "errors": [],
  "warnings": []
}
```

#### Draft Import with Validation

For draft decks, cards are validated against your draft pool:

```javascript
ImportDeck({
  "name": "My BLB Draft Deck",
  "format": "Limited",
  "source": "draft",
  "draftEventID": "draft-xyz-789",
  "importText": "..." // Your deck list
})

// If any cards aren't in your draft pool:
{
  "success": false,
  "errors": [
    "Card not in draft pool: Lightning Bolt"
  ]
}
```

#### Exporting a Deck

```javascript
ExportDeck({
  "deckID": "abc-123",
  "format": "arena",           // or "plaintext", "mtgo", "mtggoldfish"
  "includeHeaders": true,      // Include "Deck" and "Sideboard" headers
  "includeStats": true         // Include mana curve, colors as comments
})

// Returns:
{
  "content": "Deck\n4 Lightning Bolt (M21) 123\n...",
  "filename": "Mono Red Aggro.txt",
  "format": "arena"
}
```

---

## Deck Statistics

The statistics system provides comprehensive analysis of your deck composition.

### Understanding Your Stats

#### Basic Counts

- **Total Cards**: Total cards in deck (mainboard + sideboard)
- **Total Mainboard**: Cards in mainboard only
- **Total Sideboard**: Sideboard cards
- **Average CMC**: Average converted mana cost (excluding lands)

#### Mana Curve

Distribution of cards by converted mana cost:

```javascript
{
  "manaCurve": {
    "1": 8,   // 8 one-drops
    "2": 12,  // 12 two-drops
    "3": 6,   // etc.
    "4": 4,
    "5": 2
  },
  "maxCMC": 5,
  "averageCMC": 2.3
}
```

**Interpreting the Curve**:
- **Aggro**: Curve peaks at 1-2 CMC
- **Midrange**: Curve peaks at 2-4 CMC
- **Control**: Flatter curve, more high-CMC cards

#### Color Distribution

Breakdown by color identity:

```javascript
{
  "colors": {
    "white": 4,       // Mono-white cards
    "blue": 12,
    "black": 0,
    "red": 0,
    "green": 0,
    "colorless": 8,   // Artifacts, colorless cards
    "multicolor": 6   // Cards with 2+ colors
  }
}
```

#### Type Breakdown

Cards by type:

```javascript
{
  "types": {
    "creatures": 20,
    "instants": 8,
    "sorceries": 4,
    "enchantments": 2,
    "artifacts": 2,
    "planeswalkers": 1,
    "lands": 24,
    "other": 0
  }
}
```

#### Land Analysis

Detailed land statistics with recommendations:

```javascript
{
  "lands": {
    "total": 24,
    "basic": 20,
    "nonBasic": 4,
    "ratio": 40.0,              // Percentage of deck
    "recommended": 25,          // Recommended count for your curve
    "status": "optimal",        // or "too_few", "too_many"
    "statusMessage": "Land count is optimal for your deck"
  }
}
```

**Land Recommendations**:
- **60-card constructed**: 20-28 lands (typically 24-26)
  - Base of 24 for average CMC ~2.5
  - +2 lands per +0.5 average CMC
- **40-card limited**: 15-19 lands (typically 17)
  - Base of 17 for average CMC ~2.5
  - +1.5 lands per +0.5 average CMC
- **100-card Commander**: 33-42 lands (typically 37)

#### Creature Statistics

For decks with creatures:

```javascript
{
  "creatures": {
    "total": 20,
    "averagePower": 2.5,
    "averageToughness": 2.8,
    "totalPower": 50,
    "totalToughness": 56
  }
}
```

**Note**: Cards with variable power/toughness (\*/\*) are counted as 0.

### Format Legality

Checks your deck against format rules:

```javascript
{
  "legality": {
    "standard": {
      "legal": true,
      "reasons": []
    },
    "historic": {
      "legal": false,
      "reasons": [
        "Deck has only 58 cards (minimum 60 for constructed)"
      ]
    },
    "brawl": {
      "legal": false,
      "reasons": [
        "Brawl decks must have exactly 60 cards (currently 58)",
        "Card 'Lightning Bolt' has 4 copies (singleton format allows only 1)"
      ]
    }
  }
}
```

**Validation Rules**:
- **Constructed** (Standard, Historic, Explorer, Alchemy):
  - Minimum 60 cards
  - Maximum 4 copies per card (except basic lands)

- **Brawl**:
  - Exactly 60 cards
  - Singleton (max 1 copy except basic lands)

- **Commander**:
  - Exactly 99 cards (plus commander)
  - Singleton (max 1 copy except basic lands)

---

## Developer API

### Deck Management API

#### CreateDeck

Create a new deck.

```go
func (d *DeckFacade) CreateDeck(
    ctx context.Context,
    name string,
    format string,
    source string,        // "draft", "constructed", or "imported"
    draftEventID *string  // Required if source is "draft"
) (*models.Deck, error)
```

**Example**:
```go
deck, err := deckFacade.CreateDeck(
    ctx,
    "Azorius Control",
    "Standard",
    "constructed",
    nil,
)
```

#### GetDeck

Retrieve a deck with all its cards and tags.

```go
func (d *DeckFacade) GetDeck(
    ctx context.Context,
    deckID string
) (*DeckWithCards, error)
```

**Returns**:
```go
type DeckWithCards struct {
    Deck  *models.Deck
    Cards []*models.DeckCard
    Tags  []*models.DeckTag
}
```

#### AddCard

Add a card to a deck.

```go
func (d *DeckFacade) AddCard(
    ctx context.Context,
    deckID string,
    cardID int,
    quantity int,
    board string,      // "main" or "sideboard"
    fromDraft bool     // True if picked in draft
) error
```

**Validation**:
- For draft decks: Validates card is in draft pool
- For constructed: No restrictions

#### RemoveCard

Remove a card from a deck.

```go
func (d *DeckFacade) RemoveCard(
    ctx context.Context,
    deckID string,
    cardID int,
    board string  // "main" or "sideboard"
) error
```

#### UpdateDeck

Update deck metadata.

```go
func (d *DeckFacade) UpdateDeck(
    ctx context.Context,
    deck *models.Deck
) error
```

**Auto-updates**:
- `ModifiedAt` timestamp is automatically set

#### DeleteDeck

Delete a deck and all its cards.

```go
func (d *DeckFacade) DeleteDeck(
    ctx context.Context,
    deckID string
) error
```

**Warning**: This permanently deletes the deck and all associated cards. Performance history is preserved in match records.

#### CloneDeck

Create a copy of an existing deck.

```go
func (d *DeckFacade) CloneDeck(
    ctx context.Context,
    deckID string,
    newName string
) (*models.Deck, error)
```

**Use cases**:
- Testing variations
- Creating sideboard strategies
- Archiving deck versions

#### ListDecks

List all decks for the current account.

```go
func (d *DeckFacade) ListDecks(
    ctx context.Context
) ([]*DeckListItem, error)
```

**Returns**:
```go
type DeckListItem struct {
    ID            string
    Name          string
    Format        string
    Source        string
    ColorIdentity *string
    CardCount     int
    MatchesPlayed int
    MatchWinRate  float64
    ModifiedAt    time.Time
    LastPlayed    *time.Time
    Tags          []string
}
```

#### GetDeckLibrary

Advanced filtering and sorting for deck library.

```go
func (d *DeckFacade) GetDeckLibrary(
    ctx context.Context,
    filter *DeckLibraryFilter
) ([]*DeckListItem, error)
```

**Filters**:
```go
type DeckLibraryFilter struct {
    Format   *string  // Filter by format
    Source   *string  // Filter by source
    Tags     []string // Must have ALL these tags
    SortBy   string   // "modified", "created", "name", "performance"
    SortDesc bool     // Sort descending
}
```

#### Tag Management

Add and remove tags:

```go
func (d *DeckFacade) AddTag(ctx context.Context, deckID, tag string) error
func (d *DeckFacade) RemoveTag(ctx context.Context, deckID, tag string) error
```

### Card Recommendation API

#### GetRecommendations

Get AI-powered card recommendations.

```go
func (d *DeckFacade) GetRecommendations(
    ctx context.Context,
    req *GetRecommendationsRequest
) (*GetRecommendationsResponse, error)
```

**Request**:
```go
type GetRecommendationsRequest struct {
    DeckID        string
    MaxResults    int      // Default: 10
    MinScore      float64  // Default: 0.3
    Colors        []string // e.g., ["U", "W"]
    CardTypes     []string // e.g., ["Creature", "Instant"]
    CMCMin        *int
    CMCMax        *int
    IncludeLands  bool
    OnlyDraftPool bool     // For draft decks
}
```

**Response**:
```go
type GetRecommendationsResponse struct {
    Recommendations []*CardRecommendation
    Error           string
}

type CardRecommendation struct {
    CardID     int
    Name       string
    TypeLine   string
    ManaCost   string
    ImageURI   string
    Score      float64
    Reasoning  string
    Source     string
    Confidence float64
    Factors    *ScoreFactors
}
```

#### ExplainRecommendation

Get detailed explanation for a specific card recommendation.

```go
func (d *DeckFacade) ExplainRecommendation(
    ctx context.Context,
    req *ExplainRecommendationRequest
) (*ExplainRecommendationResponse, error)
```

### Statistics API

#### GetDeckStatistics

Calculate comprehensive deck statistics.

```go
func (d *DeckFacade) GetDeckStatistics(
    ctx context.Context,
    deckID string
) (*DeckStatistics, error)
```

**Returns**: Full `DeckStatistics` object with:
- Basic counts (total cards, mainboard, sideboard)
- Mana curve distribution
- Color distribution
- Type breakdown
- Land analysis with recommendations
- Creature statistics
- Format legality checks

#### GetDeckPerformance

Get performance metrics from match history.

```go
func (d *DeckFacade) GetDeckPerformance(
    ctx context.Context,
    deckID string
) (*models.DeckPerformance, error)
```

**Returns**:
```go
type DeckPerformance struct {
    DeckID            string
    MatchesPlayed     int
    MatchesWon        int
    MatchesLost       int
    GamesPlayed       int
    GamesWon          int
    GamesLost         int
    MatchWinRate      float64
    GameWinRate       float64
    LastPlayed        *time.Time
    AverageDuration   *float64
    CurrentWinStreak  int
    LongestWinStreak  int
    LongestLossStreak int
}
```

### Import/Export API

#### ImportDeck

Import a deck from text.

```go
func (d *DeckFacade) ImportDeck(
    ctx context.Context,
    req *ImportDeckRequest
) (*ImportDeckResponse, error)
```

**Request**:
```go
type ImportDeckRequest struct {
    Name         string
    Format       string
    ImportText   string
    Source       string  // "constructed", "imported", or "draft"
    DraftEventID *string // Required if source is "draft"
}
```

**Response**:
```go
type ImportDeckResponse struct {
    Success       bool
    DeckID        string
    Errors        []string
    Warnings      []string
    CardsImported int
    CardsSkipped  int
}
```

#### ExportDeck

Export a deck to various formats.

```go
func (d *DeckFacade) ExportDeck(
    ctx context.Context,
    req *ExportDeckRequest
) (*ExportDeckResponse, error)
```

**Request**:
```go
type ExportDeckRequest struct {
    DeckID         string
    Format         string // "arena", "plaintext", "mtgo", "mtggoldfish"
    IncludeHeaders bool
    IncludeStats   bool
}
```

**Response**:
```go
type ExportDeckResponse struct {
    Content  string  // The exported deck text
    Filename string  // Suggested filename
    Format   string
    Error    string
}
```

---

## FAQ

### General Questions

**Q: Can I import decks from other sources?**

A: Yes! The deck builder supports multiple import formats:
- MTGA Arena export format (most common)
- Plain text lists
- The parser automatically detects the format

**Q: How do draft decks differ from constructed decks?**

A: Draft decks are linked to a specific draft event and can only contain cards from that draft pool. The system validates all card additions against your draft picks. Constructed decks have no such restrictions.

**Q: Can I track my deck's performance?**

A: Yes! The system automatically tracks matches played with each deck, including:
- Match and game win rates
- Win/loss streaks
- Last played date
- Average match duration

**Q: How accurate are the land recommendations?**

A: Land recommendations use industry-standard heuristics:
- 60-card decks: ~24 lands baseline, adjusted for your average CMC
- 40-card limited: ~17 lands baseline
- Recommendations account for your deck's mana curve

### Import/Export Questions

**Q: What if my import has cards not in the database?**

A: The import will continue, but skipped cards are reported in the response:
```javascript
{
  "warnings": ["Skipping 'Unknown Card': card not found in database"],
  "cardsSkipped": 1
}
```

**Q: Can I import a draft deck with non-drafted cards?**

A: No. For draft decks, all cards must be from your draft pool. The import will fail with an error listing the invalid cards.

**Q: Which export format should I use?**

A:
- **Arena**: For importing directly into MTGA
- **MTGO**: For Magic Online
- **MTGGoldfish**: For uploading to MTGGoldfish.com
- **Plain Text**: For simple sharing or personal records

### Recommendation System Questions

**Q: How does the AI choose recommendations?**

A: The system analyzes five factors:
1. Color fit with your existing cards
2. Mana curve optimization
3. Synergy with deck strategy
4. Card quality (via 17Lands data when available)
5. Format playability

Each factor gets a score (0.0-1.0), which are combined into an overall recommendation score.

**Q: Why are some recommendations rated higher than others?**

A: Higher scores indicate better fit across multiple factors. A 0.8 score means the card:
- Matches your colors well
- Fills a gap in your mana curve
- Synergizes with existing cards
- Has strong performance data
- Is legal and playable in your format

**Q: Can I filter recommendations?**

A: Yes! You can filter by:
- Colors
- Card types
- CMC range
- Include/exclude lands
- Draft pool only (for draft decks)

**Q: How do I improve recommendation quality?**

A: Future versions will learn from your choices. For now:
- Build a focused color identity (fewer colors = better recommendations)
- Add enough cards to establish a strategy (10+ cards recommended)
- Use filters to narrow recommendations to what you need

### Statistics Questions

**Q: What does "optimal" land count mean?**

A: The system calculates recommended lands based on:
- Your deck size (60 vs. 40 vs. 100 cards)
- Your average CMC (excluding lands)
- Format conventions

"Optimal" means you're within ±1 land of the recommendation.

**Q: Why is my deck marked as not legal in a format?**

A: Common reasons:
- Too few cards (minimum 60 for constructed)
- Too many copies of a card (max 4, except basic lands)
- Wrong deck size for format (Brawl needs exactly 60)
- Non-singleton in Commander/Brawl formats

Check the `legality.reasons` array for specific issues.

**Q: How is average CMC calculated?**

A: Average CMC only includes non-land cards:
```
Average CMC = Total CMC of non-lands / Count of non-land cards
```

This gives you a better picture of your spell curve.

### Performance Questions

**Q: How is win rate calculated?**

A:
- **Match Win Rate**: Matches won ÷ matches played
- **Game Win Rate**: Games won ÷ games played

The deck automatically tracks this as you play matches in MTGA.

**Q: Can I see my deck's match history?**

A: Yes! Use the match history features filtered by deck ID. You can see:
- All matches played with the deck
- Win/loss record
- Opponents faced
- When the deck was last played

**Q: What if I modify a deck?**

A: Performance tracking continues with the same deck ID. If you want to track a variant separately, use `CloneDeck` to create a new deck with separate performance history.

### Tagging & Organization Questions

**Q: How many tags can I add to a deck?**

A: There's no hard limit, but we recommend 3-5 tags for effective organization.

**Q: Are tags case-sensitive?**

A: Tags are stored as-is, so "Aggro" and "aggro" are different tags. We recommend using consistent casing (e.g., lowercase).

**Q: Can I search for decks with multiple tags?**

A: Yes! Use `GetDecksByTags` or `GetDeckLibrary` with a tags array. Decks must have ALL specified tags to match.

### Draft Deck Questions

**Q: What happens if I try to add a non-drafted card?**

A: The system will return an error: "Card not in draft pool - draft decks can only contain cards from the associated draft"

**Q: Can I validate my draft deck?**

A: Yes, use `ValidateDraftDeck` to check that all cards are from your draft pool.

**Q: Can I convert a draft deck to constructed?**

A: Not directly. You can:
1. Export the draft deck
2. Create a new constructed deck
3. Import the exported list

This creates a new deck without draft pool restrictions.

**Q: What if my draft deck recommendations are limited?**

A: Draft recommendations are filtered to your draft pool only. If you have a small pool of playable cards in your colors, you'll get fewer recommendations. This is expected behavior - draft decks work with limited card pools.

---

## Troubleshooting

### Common Issues

**Import fails with "unable to parse deck format"**

- Check that your deck list has quantity numbers before card names
- Verify card names match exactly (including punctuation)
- Try adding set codes: `4 Lightning Bolt (M21) 123`

**"Card not found in database" warnings**

- The card name might be misspelled
- The card might not be in MTGA (paper-only cards)
- Try using the exact name from Scryfall or MTGA

**Land recommendations seem off**

- The system uses your non-land average CMC
- Check your mana curve - many high-CMC cards need more lands
- Limited decks naturally need fewer lands (17 vs. 24)

**No recommendations appearing**

- Ensure your deck has at least a few cards to establish identity
- Try lowering `minScore` threshold (default 0.3)
- Check your filters - they might be too restrictive
- For draft decks, ensure `onlyDraftPool` is enabled and you have draft cards available

---

## Additional Resources

- **[Main Documentation](../README.md)**: Overview of MTGA Companion
- **[Architecture Guide](ARCHITECTURE.md)**: System design and technical architecture
- **[Development Guide](DEVELOPMENT.md)**: Contributing and development setup
- **[Database Schema](https://github.com/RdHamilton/MTGA-Companion/wiki/Database-Schema)**: Database structure

---

## Changelog

### v1.3 - Deck Builder Release

**Phase 1**: Core Infrastructure
- Deck, DeckCard, and DeckTag models
- Full CRUD operations for decks
- Draft deck validation
- Tag-based organization

**Phase 2**: Import System
- Arena format parser
- Plain text format parser
- Automatic format detection
- Draft pool validation on import

**Phase 3**: AI Recommendations
- Rule-based recommendation engine
- Multi-factor scoring system
- Detailed explanations
- Draft pool filtering

**Phase 3**: Statistics & Analysis
- Comprehensive deck statistics
- Mana curve analysis
- Color and type breakdown
- Land recommendations
- Format legality checking

**Phase 4**: Export System
- Arena format export
- Plain text export
- MTGO format export
- MTGGoldfish format export
- Statistics as comments

**Phase 4**: Deck Management
- Deck library with advanced filtering
- Deck cloning
- Tag management
- Performance tracking

**Phase 5**: Testing & Documentation
- 95 passing unit tests
- Full API test coverage
- Comprehensive documentation

---

## Future Enhancements

Planned features for future releases:

- **Machine Learning**: Train recommendation engine on user feedback and deck performance
- **Sideboard Suggestions**: AI-powered sideboard recommendations based on meta
- **Deck Comparison**: Compare multiple decks side-by-side
- **Price Integration**: Show deck cost (wildcards, real money)
- **Meta Analysis**: Compare your deck to top meta decks
- **Collection Integration**: Filter recommendations by cards you own
- **Deck Versioning**: Track deck changes over time
- **Sharing**: Share decks with other users
- **Archetypes**: Auto-detect and classify deck archetypes
