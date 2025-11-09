# Collection and Deck Parsing Analysis

## Problem
Collection and deck data is not showing up in the MTGA Companion application.

## Root Causes

### 1. Deck Parsing Issue

**Current Implementation** (`internal/mtga/logreader/deck.go`):
- Looking for events: `Deck.GetDeckLists`, `getDeckLists`, `DeckLists`, `decks`
- These events DO NOT exist in MTGA logs

**Actual MTGA Log Format**:
```
<== EventGetCoursesV2(guid)
{"Courses":[...]}
```

The deck data is actually in `EventGetCoursesV2` responses, specifically in:
- `Courses[].CourseDeckSummary` - Deck metadata (name, format, etc.)
- `Courses[].CourseDeck.MainDeck` - Main deck cards
- `Courses[].CourseDeck.Sideboard` - Sideboard cards
- `Courses[].CourseDeck.CommandZone` - Commander (for Brawl)

**Example Structure**:
```json
{
  "Courses": [{
    "CourseId": "...",
    "InternalEventName": "Explorer_Ladder",
    "CourseDeckSummary": {
      "DeckId": "cd6a8186-c42f-4821-b089-24244828afa2",
      "Name": "Rakdos Sacrifice",
      "Attributes": [
        {"name": "Format", "value": "Explorer"},
        {"name": "LastPlayed", "value": "\"2024-06-21T09:35:17...\""},
        {"name": "LastUpdated", "value": "\"2024-05-10T09:54:31...\""}
      ],
      "DeckTileId": 71289
    },
    "CourseDeck": {
      "MainDeck": [
        {"cardId": 75662, "quantity": 4},
        {"cardId": 81181, "quantity": 2}
      ],
      "Sideboard": [
        {"cardId": 71289, "quantity": 1}
      ],
      "CommandZone": []
    }
  }]
}
```

### 2. Collection Parsing Issue

**Current Implementation** (`internal/mtga/logreader/collection.go`):
- Looking for `InventoryInfo` with `Cards`, `CardInventory`, or `cardInventory` fields
- These fields DO NOT exist in the InventoryInfo structure

**Actual MTGA Log Format**:
```json
{
  "InventoryInfo": {
    "SeqId": 1,
    "Changes": [],
    "Gems": 1885,
    "Gold": 7850,
    "TotalVaultProgress": 428,
    "WildCardCommons": 48,
    "WildCardUnCommons": 164,
    "WildCardRares": 11,
    "WildCardMythics": 14,
    "CustomTokens": {...},
    "Boosters": [],
    "Cosmetics": {
      "ArtStyles": [...]
    }
  }
}
```

**The card collection is NOT in InventoryInfo!**

**Where to Find Collection Data**:
According to research of other MTGA parsers, card collection data may be in:
1. `PlayerInventory.GetPlayerCardsV3` responses (if it exists)
2. Separate collection update events
3. May require monitoring specific events that track card acquisition

**Note**: The actual card collection (which cards you own and how many) may not be logged in the Player.log file in a easily parseable format, or may only appear during specific events (like opening packs, crafting, etc.).

## Comparison with Python Parser

The `python-mtga-helper` parser:
- **For Decks**: Uses `EventGetCoursesV2` callback to parse `Courses` array (CORRECT)
- **For Collection**: Does NOT appear to parse full collection - focuses on limited/draft pools in the `CardPool` field of courses

## Recommendations

### Fix 1: Update Deck Parser

Modify `internal/mtga/logreader/deck.go` to:
1. Look for `"Courses"` key (from `EventGetCoursesV2` responses)
2. Parse each course's `CourseDeckSummary` and `CourseDeck`
3. Extract deck metadata from `Attributes` array
4. Handle all deck types: MainDeck, Sideboard, CommandZone, Companions

### Fix 2: Investigate Collection Data

Options:
1. **Search for other events**: Look for `PlayerInventory`, `GetPlayerCards`, or similar events in logs
2. **Track Changes**: Monitor `InventoryInfo.Changes` array which may contain card acquisition events
3. **Accept Limitation**: Document that full collection may not be available in Player.log
4. **Use Alternative Source**: Check if collection data is available via other means (API, different log file)

### Fix 3: Test with Real Data

After implementing fixes:
1. Run MTGA and play a game
2. Check for `EventGetCoursesV2` in log
3. Verify parser can extract deck data
4. Determine if collection data is actually logged anywhere

## Event Types Found in MTGA Logs

```
DeckUpsertDeckV2       - Deck creation/update
EventEnterPairing      - Match pairing
EventGetActiveMatches  - Active matches
EventGetCoursesV2      - Course/event data (CONTAINS DECKS!)
EventJoin              - Join event
EventSetDeckV2         - Set deck for event
GetFormats             - Available formats
GraphGetGraphState     - Graph state
PeriodicRewardsGetStatus - Rewards
QuestGetQuests         - Quests
RankGetCombinedRankInfo - Rank info
RankGetSeasonAndRankDetails - Rank details
StartHook              - Startup
```

## Next Steps

1. ✅ Document current findings
2. ⬜ Implement fixed deck parser using `Courses` data
3. ⬜ Research if collection data exists in logs at all
4. ⬜ Add tests with real MTGA log samples
5. ⬜ Update documentation about what data is available
