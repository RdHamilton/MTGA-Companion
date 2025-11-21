# MTGA Log Event Analysis - Cross-Referenced with Actual Logs

**Analysis Date**: November 14, 2025
**Log File**: UTC_Log - 11-14-2025 15.02.33.log (79MB, 9,596 JSON events)

## Executive Summary

✅ **Confirmed**: Many events from older community parsers **NO LONGER EXIST** in current MTGA logs
❌ **Not Found**: `PlayerInventoryGetPlayerCards`, `CrackBooster`, `EventClaimPrize`, `RankUpdated`
✅ **Actually Exist**: 61 unique top-level JSON event types found in current logs

---

## Currently Parsed Events (Cross-Referenced)

| Event Type | Status | Data Available | Currently Used |
|------------|--------|----------------|----------------|
| `authenticateResponse` | ✅ EXISTS | screenName, clientId | ✅ Player profile |
| `InventoryInfo` | ✅ EXISTS | Gems, Gold, Wildcards, Boosters, Vault, Cosmetics | ✅ Economy tracking |
| `Courses` | ✅ EXISTS | Deck lists, event info, wins, card pool | ✅ Deck parsing |
| `matchGameRoomStateChangedEvent` | ✅ EXISTS | Match outcomes, format, wins/losses | ✅ Match history |
| `RankUpdated` | ❌ NOT FOUND | - | ❌ Removed by Wizards |
| Draft events | ⚠️ PARTIAL | See below | ⚠️ Limited |

### Rank Progression Status

**Original Issue**: We were looking for `RankUpdated` events for rank progression (issues #320, #174)

**Reality**: `RankUpdated` does **NOT** exist in current logs.

**What Actually Exists**:
- `constructedRankInfo` - Nested in combined rank data
- `limitedRankInfo` - Nested in combined rank data
- `constructedSeasonOrdinal`, `constructedLevel`, `constructedStep`, etc. - Individual rank fields

**Current rank data fields found**:
```
constructedRankInfo
constructedClass
constructedLevel
constructedStep
constructedMatchesWon
constructedMatchesLost
constructedSeasonOrdinal
limitedRankInfo
limitedClass
limitedLevel
limitedStep
limitedMatchesWon
limitedMatchesLost
limitedSeasonOrdinal
currentSeason
```

**Implication**: Rank data exists but in a different structure than expected. We need to parse rank info from combined rank responses, not standalone `RankUpdated` events.

---

## All Events Found in Current Logs (61 Total)

### ✅ Core Game Events (High Value)

| Event | Data Fields | Potential Use | Priority |
|-------|-------------|---------------|----------|
| `InventoryInfo` | Gems, Gold, Wildcards (C/U/R/M), Boosters, Vault, Cosmetics, Vouchers | Economy tracking (already using) | ✅ CURRENT |
| `Courses` | Deck lists, event info, wins, card pool, card styles | Deck & event tracking (already using) | ✅ CURRENT |
| `MatchesV3` | Match history with detailed results | Match tracking (already using) | ✅ CURRENT |
| `authenticateResponse` | Player name, client ID | Profile (already using) | ✅ CURRENT |

### 🟡 Quest & Rewards Events (Medium-High Value)

| Event | Data Fields | Potential Use | Priority |
|-------|-------------|---------------|----------|
| `quests` | questId, goal, progress (starting/ending), rewards, canSwap | Daily quest tracking | 🔴 HIGH |
| `newQuests` | New quests assigned | Quest assignment tracking | 🟡 MEDIUM |
| `ClientPeriodicRewards` | Daily/weekly reward timers, chest descriptions | Daily win rewards, weekly rewards | 🟡 MEDIUM |

**Quest fields available**:
- `questId` - Unique identifier
- `goal` - Number of objectives needed
- `endingProgress` / `startingProgress` - Current progress
- `canSwap` - Whether quest can be rerolled
- `chestDescription` - Reward details

**New Feature Potential**: **Daily Quest Tracker**
- Track quest completion rate
- Monitor quest progress over time
- Calculate average quest gold earned
- Track quest rerolls

### 🟡 Deck & Collection Events (Medium Value)

| Event | Data Fields | Potential Use | Priority |
|-------|-------------|---------------|----------|
| `Decks` | All player decks | Alternative deck source | 🟢 LOW (redundant) |
| `DeckSummariesV2` | Deck summaries | Quick deck overview | 🟢 LOW |
| `CardPool` | Cards available for limited events | Limited card tracking | 🟡 MEDIUM |
| `CardStyles` | Card style cosmetics owned | Cosmetic collection | 🟢 LOW |
| `CardMetadataInfo` | Card metadata and legalities | Card info lookup | 🟡 MEDIUM |

**Note**: `UnownedCards` field exists in some events but appears to be empty/unused

### 🟡 Rank & Progression Events (Medium-High Value)

| Event | Data Fields | Potential Use | Priority |
|-------|-------------|---------------|----------|
| `constructedRankInfo` | Full constructed rank data | Constructed rank tracking | 🔴 HIGH |
| `limitedRankInfo` | Full limited rank data | Limited rank tracking | 🔴 HIGH |
| Individual rank fields | Class, level, step, matches won/lost, season | Granular rank tracking | 🔴 HIGH |

**Implementation Note**: Need to update rank progression parser to use these events instead of the non-existent `RankUpdated`.

### 🟡 Event Participation (Medium Value)

| Event | Data Fields | Potential Use | Priority |
|-------|-------------|---------------|----------|
| `Course` | Single event/course info | Current event status | 🟡 MEDIUM |
| `CurrentModule` | Current event module/stage | Event progress tracking | 🟡 MEDIUM |
| `InternalEventName` | Event identifier | Event tracking | 🟡 MEDIUM |

### 🟢 Cosmetic & UI Events (Low Value)

| Event | Data Fields | Potential Use | Priority |
|-------|-------------|---------------|----------|
| `PreferredCosmetics` | Sleeve, Pet, Avatar, Emotes, Title | Cosmetic preferences | 🟢 LOW |
| `SystemMessages` | System notifications | Message history | 🟢 LOW |
| `TokenDefinitions` | Custom token definitions | Token tracking | 🟢 LOW |

### 🟢 Mastery & Achievement Events (Low-Medium Value)

| Event | Data Fields | Potential Use | Priority |
|-------|-------------|---------------|----------|
| `NodeStates` | Mastery track node states | Mastery progress | 🟡 MEDIUM |
| `MilestoneStates` | Milestone completion | Achievement tracking | 🟡 MEDIUM |
| `HomePageAchievements` | Featured achievements | Achievement display | 🟢 LOW |

---

## Events from Community Parsers That NO LONGER EXIST

These events were referenced in older MTGA parser libraries but **DO NOT** appear in current logs:

❌ **NOT FOUND**:
- `PlayerInventoryGetPlayerCards` - Full card collection
- `CrackBooster` - Pack opening
- `EventClaimPrize` - Prize claiming
- `RankUpdated` - Rank updates
- `EventJoin` / `EventPayEntry` - Event entry
- `MatchStart` / `MatchEnd` - Match timing
- `DeckSubmit` / `EventDeckSubmit` - Deck submission
- `IncomingInventoryUpdate` - Inventory changes
- `MythicRatingUpdated` - Mythic rank number
- `Quest_Completed` - Quest completion
- `Track_Progress` - Mastery progress
- `DirectGameChallenge` - Friend challenges

**Conclusion**: These events were either:
1. Removed by Wizards in recent years
2. Never documented correctly by community
3. Renamed or merged into other events

---

## Recommended Features Based on ACTUAL Events

### 🔴 **HIGH PRIORITY** - Should Implement

#### 1. Fix Rank Progression Parser
**Status**: Currently broken (looking for non-existent `RankUpdated`)
**Solution**: Parse `constructedRankInfo` and `limitedRankInfo` from combined rank responses
**Effort**: ~2-4 hours
**Issues**: Addresses #320 (rank progression not working)

**Implementation**:
```go
// Instead of looking for "RankUpdated", parse from top-level rank info
if rankInfo, ok := entry.JSON["constructedRankInfo"]; ok {
    // Parse rank class, level, step from nested structure
}
if rankInfo, ok := entry.JSON["limitedRankInfo"]; ok {
    // Parse limited rank info
}
```

#### 2. Daily Quest Tracker
**Events**: `quests`, `newQuests`
**Data Available**: Quest progress, goals, rewards, rerolls
**Effort**: ~6-8 hours
**Value**: High - users want to track quest completion

**Features**:
- Quest completion history
- Average quest completion time
- Total quest gold earned
- Quest reroll tracking
- Quest types played

#### 3. Event Win Tracker
**Events**: `Courses` (already parsed) + `CurrentModule`
**Additional Data**: Current wins in event, modules completed
**Effort**: ~4-6 hours (enhancement to existing parsing)
**Value**: High - track event performance in real-time

**Features**:
- Current event wins/losses
- Event progress (which module)
- Historical event results
- Best events for win rate

### 🟡 **MEDIUM PRIORITY** - Consider Implementing

#### 4. Daily/Weekly Reward Tracker
**Events**: `ClientPeriodicRewards`
**Data**: Daily win rewards, weekly reward timers
**Effort**: ~4-6 hours
**Value**: Medium - nice supplemental feature

#### 5. Limited Card Pool Tracker
**Events**: `CardPool`, `CardPoolByCollation`
**Data**: Cards available in limited events
**Effort**: ~4-6 hours
**Value**: Medium - useful for draft/sealed players

#### 6. Mastery Pass Progress
**Events**: `NodeStates`, `MilestoneStates`
**Data**: Mastery track completion
**Effort**: ~6-8 hours
**Value**: Medium - seasonal feature

### 🟢 **LOW PRIORITY** - Nice to Have

7. **Cosmetic Collection Tracker** - Card styles, avatars, pets, sleeves
8. **Achievement Tracker** - Milestone completion
9. **Event Module Progression** - Detailed event stage tracking

---

## What We CANNOT Do (Events Don't Exist)

❌ **Collection Tracking**: No full card collection data available
- Closed issues #73-77, #83-85 correctly - this data isn't in logs
- `UnownedCards` exists but appears empty

❌ **Pack Opening Analytics**: No `CrackBooster` events
- Cannot track which cards opened from packs
- Cannot calculate pull rates

❌ **Event ROI / Prize Tracking**: No `EventClaimPrize` events
- Cannot track prizes earned from events
- Cannot calculate event profitability

❌ **Match Duration**: No `MatchStart`/`MatchEnd` timing
- Cannot calculate how long matches take
- `MatchesV3` doesn't include timestamps

❌ **Mythic Rank Number**: No `MythicRatingUpdated`
- Cannot track Mythic leaderboard position
- Only have class/tier data

❌ **Friend Matches**: No `DirectGameChallenge`
- Cannot distinguish friend games from ladder

---

## Technical Implementation Notes

### Priority 1: Fix Rank Progression (CRITICAL)

**Current Issue**: Parser looks for `RankUpdated` which doesn't exist
**File**: `internal/mtga/logreader/rank_progression.go`

**Solution**: Update to parse from actual rank events:
1. Look for `constructedRankInfo` / `limitedRankInfo` in combined responses
2. Parse individual fields: `constructedClass`, `constructedLevel`, `constructedStep`, etc.
3. Store rank snapshots when these fields change

**Example Event Structure**:
```json
{
  "currentSeason": "2025-11",
  "constructedRankInfo": [...],
  "limitedRankInfo": [...],
  "constructedSeasonOrdinal": 43,
  "constructedLevel": "gold",
  "constructedStep": 2,
  "constructedMatchesWon": 15,
  "constructedMatchesLost": 12
}
```

### Priority 2: Quest Tracker

**Database Schema**:
```sql
CREATE TABLE quests (
    id INTEGER PRIMARY KEY,
    quest_id TEXT NOT NULL,
    goal INTEGER NOT NULL,
    starting_progress INTEGER,
    ending_progress INTEGER,
    completed BOOLEAN,
    can_swap BOOLEAN,
    rewards TEXT, -- JSON
    first_seen_at TIMESTAMP,
    completed_at TIMESTAMP
);
```

**Parser**:
- Listen for `quests` events
- Track progress changes
- Mark complete when `endingProgress >= goal`
- Store reward details from `chestDescription`

### Priority 3: Enhanced Event Tracking

**Additional Fields to Capture from `Courses`**:
- `CurrentWins` - Current wins in event
- `CurrentModule` - Event stage/module
- `ModulePayload` - Module-specific data
- `CardPool` - Cards available (for limited)

---

## Summary & Recommendations

### What We Know

- ✅ **61 unique event types** exist in current MTGA logs
- ✅ **Core events** we parse (InventoryInfo, Courses, MatchesV3) still exist and work
- ❌ **Rank progression is broken** - `RankUpdated` doesn't exist anymore
- ❌ **Many old events are gone** - PlayerInventoryGetPlayerCards, CrackBooster, etc.
- ✅ **Quest data is available** - Can build quest tracker
- ✅ **Rank data exists** - Just in different format than expected

### Immediate Actions

1. **FIX RANK PROGRESSION** (2-4 hours) - Update parser for actual rank events
2. **Implement Quest Tracker** (6-8 hours) - High value, data available
3. **Enhance Event Tracking** (4-6 hours) - Use additional fields from `Courses`

### Cannot Implement (Data Not Available)

- Full collection tracking
- Pack opening analytics
- Event prize/ROI tracking
- Match duration tracking
- Mythic leaderboard position

### Est Total Effort for High-Priority Items

- Fix rank progression: ~2-4 hours
- Quest tracker: ~6-8 hours
- Enhanced event tracking: ~4-6 hours

**Total**: ~12-18 hours for all high-priority features based on actual data

---

## Files to Update

1. `internal/mtga/logreader/rank_progression.go` - Fix to use actual rank events
2. Create `internal/mtga/logreader/quests.go` - New quest parser
3. Update `internal/mtga/logreader/parse.go` - Enhanced Courses parsing
4. Create migrations for quest tracking table
5. Update GUI to display quest progress

---

**Next Steps**: Recommend starting with fixing rank progression since it's currently broken and affects existing users.
