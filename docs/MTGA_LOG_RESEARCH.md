# MTGA Log Analysis Research

This document contains manual analysis of MTGA log files to better understand game events, their timing, and how they correlate with in-game actions.

**Purpose:** Document findings from manual log analysis to improve parsing accuracy and feature implementation.

**Format:** UTC_Log files contain mixed plain text and JSON data. JSON events are what we parse.

---

## Table of Contents

- [Daily/Weekly Win Tracking](#dailyweekly-win-tracking)
- [Match Completion Events](#match-completion-events)
- [Quest Events](#quest-events)
- [Rank Progression](#rank-progression)
- [Inventory & Currency](#inventory--currency)
- [Draft/Limited Events](#draftlimited-events)
- [Achievement Events](#achievement-events)
- [Unknown/Undocumented Events](#unknownundocumented-events)
- [Implementation Status](#implementation-status)
- [Research Notes by Date](#research-notes-by-date)

---

## Daily/Weekly Win Tracking

### ClientPeriodicRewards Event

**Status:** ✅ IMPLEMENTED (PR #XXX)

**When it appears:**
- [ ] After every match win?
- [ ] Only when daily/weekly win count changes?
- [ ] At login/session start?

**JSON Structure:**
```json
{
  "ClientPeriodicRewards": {
    "_dailyRewardSequenceId": 5,
    "_weeklyRewardSequenceId": 10
  }
}
```

**Field Details:**
- `_dailyRewardSequenceId`: Integer 0-15 (current daily wins)
- `_weeklyRewardSequenceId`: Integer 0-15 (current weekly wins)

**Research Notes:**
- [ ] Does this event appear for EVERY match, or only wins that count?
- [ ] What happens at 15/15 (max daily wins)?
- [ ] When does the counter reset (midnight? server time?)
- [ ] Does it appear for bot/practice matches?

**Correlation with Match Wins:**
- [ ] How long after `matchGameRoomStateChangedEvent` does this appear?
- [ ] Can we reliably correlate the two events?
- [ ] What if two matches end within seconds of each other?

**Edge Cases Discovered:**
- [ ] Conceded matches
- [ ] Practice/bot matches
- [ ] Losses
- [ ] Direct challenge matches

---

## Match Completion Events

### matchGameRoomStateChangedEvent

**Status:** ✅ IMPLEMENTED (match history tracking)

**When it appears:**
- [ ] At end of every match?
- [ ] During match state changes?
- [ ] Only ranked/competitive matches?

**JSON Structure:**
```json
{
  "matchGameRoomStateChangedEvent": {
    "gameRoomInfo": {
      "finalMatchResult": {
        "matchId": "...",
        "resultList": [
          {
            "scope": "MatchScope_Match",
            "result": "ResultType_Won",
            "winningTeamId": 1
          },
          {
            "scope": "MatchScope_Game",
            "result": "ResultType_Won",
            "winningTeamId": 1
          }
        ]
      },
      "gameRoomConfig": {
        "reservedPlayers": [
          {
            "playerName": "...",
            "teamId": 1,
            "eventId": "..."
          }
        ]
      }
    }
  }
}
```

**Field Details:**
- `scope`: "MatchScope_Match" (overall) or "MatchScope_Game" (individual games)
- `result`: Win/loss indicator
- `winningTeamId`: Which player won (1 or 2)
- `eventId`: Format/event type (e.g., "Constructed_Ranked_Standard")

**Research Notes:**
- [ ] Best-of-3 vs Best-of-1 differences?
- [ ] How to identify player vs opponent?
- [ ] Concede vs timeout vs normal completion?

**Correlation Questions:**
- [ ] How long after match completion does this event appear?
- [ ] Is there a delay between game end and log entry?
- [ ] Can we use timestamp to correlate with daily wins?

---

## Quest Events

### QuestGetQuests Response

**Status:** ✅ IMPLEMENTED

**When it appears:**
- Session login
- After completing a match
- When claiming quest rewards
- Periodic refreshes during gameplay

**JSON Structure (Before Quest Completion):**
```json
{
  "canSwap": true,
  "quests": [
    {
      "questId": "39892322-3ea8-4131-911d-fa6386c7b2d4",
      "goal": 25,
      "locKey": "Quests/Quest_Nissas_Journey",
      "endingProgress": 17,
      "startingProgress": 17,
      "canSwap": true,
      "chestDescription": {
        "quantity": "500"
      }
    }
  ]
}
```

**JSON Structure (After Quest Completion):**
```json
{
  "canSwap": true,
  "quests": []
}
```

**Key Logic:**
- **Quest Assignment**: Quest appears in the `quests` array
- **Quest Progress**: `endingProgress` value updates as quest progresses
- **Quest Completion**: Quest **disappears from the array** when completed

**Completion Detection Strategy:**
1. Track which quest IDs are present in each `QuestGetQuests` response
2. When a quest ID that was previously in the array is no longer present in a later response, the quest was completed
3. Mark completion timestamp as the time of the response where the quest disappeared

**Important Notes:**
- Quest completion is NOT detected by `endingProgress >= goal`
- MTGA removes quests from the response array when they are completed
- A quest disappearing is the definitive signal of completion
- Multiple quests can disappear in a single response if completed together

### Quest Assignment (newQuests Event)

**Status:** ⏳ NEEDS RESEARCH

**Potential Event Names:**
- `newQuests`
- `Quest_Assigned`?
- `QuestUpdate`?

**Research Needed:**
- [ ] What event fires when a new daily quest is assigned?
- [ ] Daily quest rotation timing (midnight? server time?)
- [x] Can we detect quest rerolls? (Yes, through canSwap field)

---

## Rank Progression

### RankUpdated Event

**Status:** ✅ IMPLEMENTED

**JSON Structure:**
```json
{
  "RankUpdated": {
    "format": "Constructed",
    "seasonOrdinal": 123,
    "newClass": "Diamond",
    "newLevel": 3,
    "newStep": 4,
    "oldClass": "Diamond",
    "oldLevel": 3,
    "oldStep": 3
  }
}
```

**Research Notes:**
- [ ] When exactly does this appear relative to match end?
- [ ] Can we correlate this with match results to track rank changes per match?
- [ ] How to detect double rank-ups?
- [ ] Rank floors (can't drop below)?

---

## Inventory & Currency

### InventoryInfo Event

**Status:** ✅ IMPLEMENTED (basic parsing)

**JSON Structure:**
```json
{
  "InventoryInfo": {
    "Gems": 1885,
    "Gold": 7850,
    "WildCardCommons": 48,
    "WildCardRares": 11,
    "TotalVaultProgress": 500,
    "Boosters": [
      {
        "SetCode": "BRO",
        "Count": 5
      }
    ]
  }
}
```

**Research Needed:**
- [ ] When does this event fire? (Login? After purchases?)
- [ ] Can we track gold/gem changes per match?
- [ ] How to detect pack opening vs rewards?

### Currency Deltas

**Questions:**
- [ ] Is there a separate event for currency changes?
- [ ] How to distinguish: match rewards vs store purchases vs event prizes?

---

## Draft/Limited Events

### Draft Entry

**Status:** ⚠️ PARTIALLY IMPLEMENTED

**Needs Research:**
- [ ] Event name for entering draft?
- [ ] How to detect draft start vs deck building?

### Draft Picks

**Status:** ⏳ NEEDS RESEARCH

**Questions:**
- [ ] Can we track individual pack picks?
- [ ] What data is available per pick?
- [ ] Timestamp of each pick decision?

### Draft Completion

**Needs Research:**
- [ ] How to detect when draft deck is finalized?
- [ ] Can we capture final deck list?

---

## Achievement Events

### Event Structure

**Status:** ✅ IMPLEMENTED

**Needs Correlation Research:**
- [ ] How do achievements relate to match events?
- [ ] Progress tracking between matches?
- [ ] Completion detection timing?

---

## Mastery Pass Tracking

### NodeStates Event (Mastery Pass Graph State)

**Status:** ✅ IMPLEMENTED

**When it appears:**
- Session login/start
- After level progression
- When claiming rewards
- Related to GraphGetGraphState requests

**JSON Structure:**
```json
{
  "NodeStates": {
      "LevelTrack_Level_1": {
        "Status": "Completed"
      },
      "LevelTrack_Level_1_Reward": {
        "Status": "Completed",
        "TierRewardNodeState": {
          "CurrentTiers": ["basic"]
        }
      },
      "LevelTrack_Level_2": {
        "Status": "Completed"
      },
      "LevelTrack_Level_2_Reward": {
        "Status": "Completed",
        "TierRewardNodeState": {
          "CurrentTiers": ["basic"]
        }
      },
      "LevelTrack_Level_55": {
        "Status": "Completed"
      },
      "LevelTrack_Level_55_Reward": {
        "Status": "Completed",
        "TierRewardNodeState": {
          "CurrentTiers": ["basic"]
        }
      },
      "LevelTrack_Level_56": {
        "Status": "Available",
        "ProgressNodeState": {
          "CurrentProgress": 900
        }
      }
    },
    "MilestoneStates": {}
  }
}
```

**Field Details:**
- `NodeStates`: Object containing all mastery pass level nodes
- `LevelTrack_Level_X`: Progress node for level X
- `LevelTrack_Level_X_Reward`: Reward node for level X
- `Status`: "Completed" | "Available" | "Locked"
- `TierRewardNodeState.CurrentTiers`: Array containing pass type (e.g., ["basic"] or ["advanced"])
- `ProgressNodeState.CurrentProgress`: XP progress toward next level

**Key Logic:**
- **Current Level**: Highest level number X where `LevelTrack_Level_X_Reward` has `Status: "Completed"`
- **Pass Type**: Extract from `CurrentTiers[0]` in any `TierRewardNodeState` (e.g., "basic" -> "Basic")
- **Max Level**: Highest level number X found in any `LevelTrack_Level_X` node
- **Current XP**: `CurrentProgress` from the next available level's `ProgressNodeState`

**Parsing Strategy:**
1. Look for events with `NodeStates` as a top-level key
2. Iterate through all keys in the `NodeStates` object
3. Filter keys matching pattern `LevelTrack_Level_(\d+)_Reward`
4. Extract level number from key name (e.g., "LevelTrack_Level_55_Reward" → 55)
5. Check if `Status === "Completed"`
6. Track highest completed level
7. Extract `CurrentTiers[0]` from `TierRewardNodeState` for pass type

**Research Notes:**
- Mastery Pass typically has ~80 levels
- Requires ~5 daily wins to reach max level in season
- Free pass ("basic") available to all players
- Advanced pass requires gems/purchase
- Each level has two nodes: progress node and reward node

**Daily Win Goal Context:**
- 5 wins/day × ~50 days = max mastery level
- Tracking daily win streaks helps ensure mastery completion
- Visual indicators should highlight 5-win milestone

**Implementation Priority:** HIGH - Core progression metric

---

## Unknown/Undocumented Events

### Events to Investigate

List any interesting events you discover here:

**Event Name:** `[Event name from JSON]`
**When Observed:** [Describe what triggered it]
**Structure:**
```json
{
  // Paste JSON here
}
```
**Hypothesis:** [What you think it does]
**Questions:**
- [ ] Question 1
- [ ] Question 2

---

## Implementation Status

Track which research findings have been implemented:

- [x] **Daily/Weekly Win Tracking** - Basic implementation (reads counts)
  - [ ] Correlation with match wins - TODO
  - [ ] Toast notifications on increment - TODO
  - [ ] Detection of which match caused increment - TODO

- [x] **Match History** - Screen name-based opponent detection
  - [ ] Correlation with daily/weekly wins - TODO
  - [ ] Match-to-quest completion correlation - TODO

- [x] **Quest Tracking** - Basic quest detection
  - [ ] Quest progress tracking - TODO
  - [ ] Quest completion correlation with matches - TODO

- [x] **Rank Tracking** - RankUpdated event parsing
  - [ ] Rank change per match correlation - TODO
  - [ ] Double rank-up detection - TODO

- [ ] **Currency Tracking** - Not implemented
- [ ] **Draft Pick Tracking** - Not implemented

---

## Research Notes by Date

### Template for New Entries

```markdown
### YYYY-MM-DD

**Matches Played:** X matches
**Log File:** UTC_Log - MM-DD-YYYY HH.MM.SS.log

**Findings:**
- [Discovery 1]
- [Discovery 2]

**Questions:**
- [ ] Question raised
- [ ] Another question

**Next Steps:**
- [ ] Action item 1
- [ ] Action item 2
```

---

### 2024-11-16

**Matches Played:** [Add your findings]
**Log File:** [Add log file name]

**Findings:**
- Initial template created
- [Add your manual parsing discoveries]

**Questions:**
- [ ] [Add questions that arose during research]

**Next Steps:**
- [ ] [Add action items]

---

## Research Methodology

**Tools Used:**
- Text editor for manual log review
- JSON formatters/validators
- Timestamp correlation analysis

**Process:**
1. Play specific match scenarios (win, loss, concede, etc.)
2. Note exact time and circumstances
3. Search log for events around that timestamp
4. Document JSON structure and timing
5. Test correlation theories with multiple examples

**Tips:**
- Search for your screen name in logs to find player-related events
- Look for timestamps within ±30 seconds of in-game actions
- Compare multiple similar events to find patterns
- Test edge cases (concedes, timeouts, etc.)

---

## Contributing Research Findings

When you discover something:

1. **Document the finding** in the relevant section above
2. **Add timestamp** and log file reference
3. **Include JSON examples** (sanitize personal info)
4. **Note correlations** with other events
5. **Raise questions** for further investigation
6. **Update implementation status** if code changes are needed

**Format for Implementation Requests:**

When ready to implement a finding, note it like this:

```
**Ready for Implementation: [Feature Name]**
- Research documented in: [Section]
- Event name: [JSON event key]
- Correlation established: [Yes/No/Partial]
- Edge cases tested: [List]
- Priority: [High/Medium/Low]
- Related GitHub Issue: #XXX
```

---

## Glossary

**Match:** A complete game session (may be Bo1 or Bo3)
**Game:** Individual game within a match
**Event:** In-game format/queue (e.g., "Constructed_Ranked_Standard")
**Scope:** Whether result is for overall match or individual game
**UTC_Log:** MTGA's log file format (timestamped, mixed text/JSON)
**ClientPeriodicRewards:** Daily/weekly win progress event
**matchGameRoomStateChangedEvent:** Match completion event

---

*Last Updated: 2024-11-16*
*Primary Researcher: [Your Name]*
