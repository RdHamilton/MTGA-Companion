# Collection Features User Guide

This guide explains how to use the collection tracking features in MTGA-Companion to manage and analyze your Magic: The Gathering Arena card collection.

## Table of Contents

- [How Collection Tracking Works](#how-collection-tracking-works)
- [Collection Page](#collection-page)
- [Set Completion Tracking](#set-completion-tracking)
- [Missing Cards Analysis](#missing-cards-analysis)
- [Collection Change Notifications](#collection-change-notifications)
- [Tips and Best Practices](#tips-and-best-practices)

---

## How Collection Tracking Works

MTGA-Companion automatically tracks your card collection by monitoring your MTGA game logs. Here's what you need to know:

### Automatic Log-Based Tracking

- **No manual input required**: Your collection is automatically updated as you play
- **Works across all platforms**: Windows, macOS, and Linux are fully supported
- **Real-time updates**: Collection changes are detected within seconds of occurring in-game

### Data Sources

Your collection is populated from multiple in-game activities:

- **Opening booster packs**: Cards from packs are automatically added
- **Draft picks**: All cards drafted are tracked
- **Deck rewards**: Cards earned from events and rewards
- **Crafting**: Cards crafted with wildcards are included

### Card Quantity Rules

Collection tracking follows Arena's rules:

- **4-copy limit**: Most cards are capped at 4 copies (matching Arena's deck-building rules)
- **Basic lands**: Unlimited copies allowed (not tracked for set completion)
- **Wildcards**: Your wildcard inventory is also tracked

### Building Your Collection Over Time

Your collection data accumulates gradually as you play MTGA. The more you play, the more complete your collection data becomes. If you're a new user, your collection will populate as the app detects cards from your gameplay.

---

## Collection Page

The Collection page is your central hub for browsing and managing your card collection.

### Viewing Your Cards

Cards are displayed in a visual grid showing:

- **Card images**: Full card artwork for easy recognition
- **Quantity badges**: Shows how many copies you own (displayed on each card)
- **Rarity indicators**: Badge color matches card rarity

Cards are paginated (50 per page) for smooth performance with large collections.

### Collection Statistics

The page header displays at-a-glance stats:

- **Unique Cards**: The number of different cards in your collection
- **Total Cards**: The total count including all copies

### Searching and Filtering

Find specific cards quickly with powerful filters:

| Filter | Description |
|--------|-------------|
| **Search** | Type to search by card name (case-insensitive) |
| **Set** | Filter by specific MTG set from the dropdown |
| **Rarity** | Show only Common, Uncommon, Rare, or Mythic cards |
| **Colors** | Click color buttons (W/U/B/R/G) to filter by mana color |
| **Owned Only** | Toggle to show only cards you currently own |

Multiple filters can be combined. For example, search for "dragon" + filter to "Rare" + select "Red" to find all rare red dragons in your collection.

### Sorting Options

Organize your cards with flexible sorting:

| Sort Option | Description |
|-------------|-------------|
| **Name (A-Z / Z-A)** | Alphabetical ordering |
| **Quantity (High / Low)** | Cards you have the most/least copies of |
| **Rarity (High / Low)** | Mythic > Rare > Uncommon > Common |
| **CMC (Low / High)** | Converted mana cost ordering |

### Pagination

Navigate large collections easily:

- **First / Last**: Jump to beginning or end
- **Previous / Next**: Move one page at a time
- **Page indicator**: Shows current page and total pages

---

## Set Completion Tracking

Track your progress toward completing each MTG set with the Set Completion panel.

### Accessing Set Completion

Click the **"Show Set Completion"** button on the Collection page to open the Set Completion panel.

### Understanding the Display

For each set, you'll see:

- **Set icon and name**: Visual identification of the set
- **Progress bar**: Visual representation of completion percentage
- **Completion stats**: "X owned / Y total (Z%)" format

### Rarity Breakdown

Click on any set to expand and see completion by rarity:

| Rarity | Color |
|--------|-------|
| Mythic | Orange |
| Rare | Gold |
| Uncommon | Silver |
| Common | Black |

Each rarity shows its own progress bar and owned/total count.

### Sorting Sets

Use the dropdown to sort sets by:

- **Newest First**: Most recently released sets at top
- **Oldest First**: Oldest sets at top
- **Most Complete**: Sets you're closest to finishing
- **Least Complete**: Sets with the most cards to collect

This helps you decide which sets to focus on completing.

---

## Missing Cards Analysis

Understand exactly what cards you need to complete decks or sets.

### Missing Cards for Decks

When viewing a deck, you can see:

- **List of missing cards**: Every card you need but don't own
- **Quantity needed**: How many more copies of each card
- **Wildcard cost breakdown**:
  - Common wildcards needed
  - Uncommon wildcards needed
  - Rare wildcards needed
  - Mythic wildcards needed
- **Total wildcard cost**: Sum of all wildcards required

Cards are sorted by rarity (Mythic first) to help prioritize crafting decisions.

### Missing Cards for Sets

For any set, you can analyze:

- **Completion percentage**: How close you are to finishing
- **All missing cards**: Complete list of cards you don't have
- **Rarity breakdown**: Missing cards grouped by rarity
- **Wildcard cost to complete**: Total wildcards needed to finish the set

This helps you plan which sets to invest wildcards in.

---

## Collection Change Notifications

Track recent changes to your collection.

### Recent Changes View

View your most recent collection updates:

- **Card name and set**: Which card was added or removed
- **Quantity change**: +X for additions, -X for removals
- **Timestamp**: When the change occurred
- **Source**: Where the card came from (pack, draft, craft, etc.)

### Using Change History

Change history is useful for:

- Verifying pack contents after opening
- Tracking draft progress
- Reviewing what you've crafted
- Identifying unexpected collection changes

---

## Tips and Best Practices

### Getting Started

1. **Launch MTGA first**: The app reads from MTGA's log file, so start the game
2. **Play some games**: Collection data populates as you play
3. **Check the Collection page**: Your cards will appear automatically

### Maximizing Collection Data

- **Play regularly**: More gameplay means more complete data
- **Open packs in-game**: Pack openings are tracked automatically
- **Complete drafts**: All drafted cards are recorded

### Troubleshooting

**Collection not updating?**
- Ensure MTGA is running and you've played recently
- Check that the app has access to your MTGA log file location
- Try restarting both MTGA and MTGA-Companion

**Missing older cards?**
- Collection builds over time from log data
- Cards from before you started using the app may not appear
- Play games to trigger inventory updates which capture your full collection

**Set completion showing 0%?**
- You may need to fetch set card data first (Settings > Fetch Set Cards)
- This downloads the card database needed for completion calculations

### Platform-Specific Notes

| Platform | Log Location |
|----------|--------------|
| Windows | `%APPDATA%\..\LocalLow\Wizards Of The Coast\MTGA\` |
| macOS | `~/Library/Logs/Wizards Of The Coast/MTGA/` |
| Linux | Varies by installation method |

The app automatically detects the correct log location for your platform.

---

## Summary

The Collection features in MTGA-Companion provide:

1. **Automatic tracking**: No manual data entry required
2. **Visual browsing**: See your entire collection with card images
3. **Powerful search**: Find cards quickly with filters and sorting
4. **Set completion**: Track progress toward completing each set
5. **Missing cards analysis**: Know exactly what you need for decks and sets
6. **Change history**: Review recent additions to your collection

Your collection data grows more complete the more you play, giving you better insights into your card library over time.
