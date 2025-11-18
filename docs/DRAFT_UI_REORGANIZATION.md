# Draft UI Project Reorganization

**Date:** 2025-11-18
**Purpose:** Reorganize Draft UI project to align with new real-time draft assistant vision

---

## New Vision Summary

### Real-Time Draft Assistant UI
A window-based draft companion that displays during Quick Draft, showing:

1. **Complete Set Card Grid** (~25% of screen)
   - All cards from the set being drafted
   - Hover tooltips with card images and details
   - Dynamic highlighting based on draft state

2. **Pick Highlighting**
   - **Green highlight + pinned to top**: Cards the player has picked
   - **Orange highlight**: Cards with types in common with picked cards

3. **Individual Pick History Grid**
   - Shows each card picked in chronological order
   - Visual record of the draft progression

4. **"Cards to Look For" Synergy Panel**
   - Suggests cards that synergize with current picks
   - Based on card types, mechanics, and archetypes

### Log Parsing Requirements
Parse these MTGA log events for Quick Draft:
- `Client.SceneChange` with `toSceneName:"Draft"` → Draft started
- `BotDraftDraftStatus` → Initial pack and picked cards
  - `PackNumber` (0-indexed), `PickNumber`, `DraftPack[]`, `PickedCards[]`
- `BotDraftDraftPick` → Player made a pick
  - `CardIds[]`, `PackNumber`, `PickNumber`
- `Client.SceneChange` with `toSceneName:"DeckBuilder"` → Draft ended

---

## Implementation Phases

### Phase 1: Foundation (Backend & Data)
**Goal:** Parse logs, fetch card data, store draft sessions

**New Issues to Create:**
1. **Parse Quick Draft log events** (CRITICAL)
   - Detect draft start/end from SceneChange events
   - Parse BotDraftDraftStatus for initial state
   - Parse BotDraftDraftPick for each pick
   - Store in database

2. **Fetch and cache set card data from Scryfall**
   - Identify set from EventName (e.g., "QuickDraft_TDM_20251111" → TDM set)
   - Fetch complete card list for set via Scryfall API
   - Cache card images, types, mana costs, rules text
   - Update cache when new sets release

3. **Database schema for draft sessions**
   - `draft_sessions` table (id, event_name, set_code, start_time, end_time)
   - `draft_picks` table (session_id, pack_number, pick_number, card_id, timestamp)
   - `draft_packs` table (session_id, pack_number, pick_number, card_ids[])
   - `set_cards` cache table (set_code, card_id, name, types, mana_cost, image_url, etc.)

**Existing Issues to Keep:**
- #271 - Add archetype definitions to set guides
- #272 - Improve card categorization using Scryfall card types
- #273 - Add set mechanics parsing and display

---

### Phase 2: Core UI (Real-Time Draft Assistant)
**Goal:** Build the main draft assistant UI window

**New Issues to Create:**
4. **Real-time draft UI window**
   - Detect when draft starts (parse logs or manual trigger)
   - Open draft assistant window
   - Close when draft ends
   - Responsive layout with 3 main sections: set grid, pick history, synergies

5. **Complete set card grid component**
   - Display all cards from current set
   - Grid layout with card images
   - Filter/search functionality
   - Takes ~25% of window width
   - Responsive grid sizing

6. **Card tooltip component with images**
   - Hover over card → show tooltip
   - Display: card image, name, mana cost, types, rules text, rarity
   - Use Scryfall image URLs
   - Smooth hover animation

7. **Picked card highlighting and pinning**
   - When card is picked → highlight green
   - Pin picked cards to top of grid
   - Show pick order number on card
   - Visual distinction from unpicked cards

8. **Type synergy highlighting**
   - Analyze picked cards → extract card types (Creature, Instant, etc.)
   - Highlight cards with matching types in orange
   - Support multiple synergy types (types, mechanics, archetypes)
   - Configurable synergy rules

9. **Individual pick history grid**
   - Display each pick in chronological order
   - Show: pack number, pick number, card image/name
   - Visual timeline of draft progression
   - Scrollable if many picks

10. **"Cards to Look For" synergy panel**
    - Based on current picks, suggest synergistic cards
    - Show card name, image, reason for suggestion
    - Examples: "Tribal synergy (Elves)", "Removal for aggro deck", etc.
    - Update in real-time as picks are made

**Existing Issues to Keep:**
- #181 - Implement draft statistics and mana curve visualization (add to UI as stats panel)
- #266 - Add performance monitoring and metrics for draft overlay

---

### Phase 3: Enhancement (Advanced Features)
**Goal:** Add advanced analytics and recommendations

**Existing Issues to Keep:**
- #179 - Implement missing cards detection from draft packs (show which cards others picked)
- #119 - Enhance draft recommendations with card ratings and archetype analysis
- #275 - Add tier list export to CSV and JSON formats
- #276 - Add format comparison for card ratings (Premier vs Quick Draft)

**Existing Issues to DEPRIORITIZE (move to v2.0 or later):**
- #114 - Replay draft decisions (post-draft analytics, not real-time)
- #115 - Analyze drafting patterns (post-draft analytics)
- #117 - Add visual draft pick timeline (post-draft visualization)
- #120 - Add archetype performance analysis for drafts (post-draft analytics)
- #121 - Add temporal trend analysis for drafts (post-draft analytics)
- #122 - Compare draft stats with community averages (post-draft analytics)
- #180 - Implement draft deck suggester (complex feature, defer)
- #265 - Export draft data to 17Lands JSON format (export feature, defer)
- #326 - Feature: Limited Card Pool Tracker (different feature scope)
- #256 - Research platform-specific overlay implementations (not needed for window UI)

---

## Issue Actions Summary

### CREATE (10 new issues)
1. Parse Quick Draft log events
2. Fetch and cache set card data from Scryfall
3. Database schema for draft sessions
4. Real-time draft UI window
5. Complete set card grid component
6. Card tooltip component with images
7. Picked card highlighting and pinning
8. Type synergy highlighting
9. Individual pick history grid
10. "Cards to Look For" synergy panel

### KEEP in v1.1 (8 issues)
- #119 - Card ratings integration
- #179 - Missing cards detection
- #181 - Draft statistics and mana curve
- #266 - Performance monitoring
- #271 - Archetype definitions
- #272 - Card categorization (Scryfall)
- #273 - Set mechanics parsing
- #275 - Tier list export
- #276 - Format comparison

### MOVE to v2.0 (10 issues)
- #114 - Replay draft decisions
- #115 - Analyze drafting patterns
- #117 - Visual draft pick timeline
- #120 - Archetype performance analysis
- #121 - Temporal trend analysis
- #122 - Community comparison
- #180 - Draft deck suggester
- #265 - 17Lands export
- #326 - Limited card pool tracker
- #256 - Platform-specific overlay research

---

## Reorganization Complete ✅

### Issues Created (9 new issues)

**Phase 1: Foundation (Backend & Data)**
- #378 - Parse Quick Draft log events from Player.log
- #379 - Fetch and cache set card data from Scryfall API
- #380 - Database schema for draft sessions, picks, and packs

**Phase 2: Core UI (Real-Time Draft Assistant)**
- #381 - Real-time draft UI window - main container and layout
- #382 - Complete set card grid component with filtering
- #383 - Card tooltip component with images and details
- #384 - Type synergy highlighting algorithm
- #385 - Individual pick history grid component
- #386 - "Cards to Look For" synergy suggestions panel

### Issues Moved to v2.0 (10 issues)
- #114 - Replay draft decisions
- #115 - Analyze drafting patterns
- #117 - Visual draft pick timeline
- #120 - Archetype performance analysis
- #121 - Temporal trend analysis
- #122 - Community comparison
- #180 - Draft deck suggester
- #256 - Platform-specific overlay research
- #265 - 17Lands export
- #326 - Limited card pool tracker

### Issues Kept in v1.1 (9 issues)
- #119 - Card ratings integration
- #179 - Missing cards detection
- #181 - Draft statistics and mana curve
- #266 - Performance monitoring
- #271 - Archetype definitions
- #272 - Card categorization (Scryfall)
- #273 - Set mechanics parsing
- #275 - Tier list export
- #276 - Format comparison

### Implementation Order (Organized by Milestone)

**📦 PHASE 1: Draft Infrastructure (Milestone 27)**
Foundation - Blockers (must complete first)
1. #380 - Database schema for draft sessions, picks, and packs
2. #378 - Parse Quick Draft log events from Player.log
3. #379 - Fetch and cache set card data from Scryfall API

**🎨 PHASE 2: Draft UI/UX Improvements (Milestone 23)**
Core UI Components (depends on Phase 1)
4. #381 - Real-time draft UI window - main container and layout
5. #383 - Card tooltip component with images and details
6. #382 - Complete set card grid component with filtering
7. #385 - Individual pick history grid component

**⚡ PHASE 3: Draft Recommendations Enhancement (Milestone 25)**
Advanced Features (depends on Phase 2)
8. #384 - Type synergy highlighting algorithm
9. #386 - "Cards to Look For" synergy suggestions panel

**🔧 PHASE 4: Enhancements (Various Milestones)**
Additional features (can be done in parallel with Phase 3)
10. #181 - Draft statistics and mana curve (Milestone: Draft Analytics & Statistics)
11. #179 - Missing cards detection (Milestone: Draft Analytics & Statistics)
12. #119 - Card ratings integration (Milestone: Draft Recommendations Enhancement)

---

**Project Board:** https://github.com/users/RdHamilton/projects/20
**Date Completed:** 2025-11-18
