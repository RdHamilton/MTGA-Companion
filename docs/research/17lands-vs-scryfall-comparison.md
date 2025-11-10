# Card Data Source Comparison: 17Lands vs Scryfall

**Research Date:** 2025-11-10
**Issue:** #177
**Status:** ✅ Complete

## Executive Summary

**Recommendation: Hybrid Approach** - Use 17Lands for draft statistics and Scryfall for comprehensive card metadata.

This combines the strengths of both APIs:
- 17Lands provides unique draft performance data essential for draft features
- Scryfall offers comprehensive card metadata with better documentation and reliability
- Each service covers gaps in the other

---

## Detailed Comparison

### 17Lands API

**Primary Focus:** Community-sourced draft and limited format statistics

#### Available Data

**Win Rate Metrics:**
- `GIHWR` (Games in Hand Win Rate) - Win rate when card is in hand
- `OHWR` (Opening Hand Win Rate) - Win rate in opening hand
- `GPWR` (Game Present Win Rate) - Win rate when card is in deck
- `GNDWR` (Game Not Drawn Win Rate) - Win rate when in deck but not drawn
- `GDWR` (Game Drawn Win Rate) - Win rate when card is drawn
- `IWD` (Improvement When Drawn) - GIHWR - GNDWR

**Draft Metrics:**
- `ALSA` (Average Last Seen At) - Average pick position card disappears
- `ATA` (Average Taken At) - Average pick position when taken
- Sample sizes: `GIH`, `OH`, `GP`, `GD` counts

**Advanced Features:**
- Color-specific ratings (W, U, B, R, G, WU, UB, BR, etc.)
- Format-specific data (PremierDraft, QuickDraft, TradDraft, Sealed)
- Date range filtering for historical analysis
- Bayesian averaging for small sample sizes
- Deck archetype win rates

#### API Endpoints

```
# Card ratings
https://www.17lands.com/card_ratings/data?expansion={set}&format={type}&start_date={start}&end_date={end}&colors={colors}

# Color ratings
https://www.17lands.com/color_ratings/data?expansion={set}&event_type={type}&start_date={start}&end_date={end}&combine_splash=true
```

#### Data Format

- **Format:** JSON array
- **Fields per card:**
  - `name`, `mtga_id`, `color`, `rarity`, `card_types`, `layout`
  - `url` (front image), `url_back` (back image for DFCs)
  - `seen_count`, `avg_seen`, `pick_count`, `avg_pick`
  - `game_count`, `play_rate`, `win_rate`
  - `opening_hand_game_count`, `opening_hand_win_rate`
  - `drawn_game_count`, `drawn_win_rate`, `ever_drawn_win_rate`
  - `never_drawn_win_rate`, `drawn_improvement_win_rate`

#### Testing Results

✅ **API Status:** Working (tested with Bloomburrow BLB set)
- Endpoint accessible
- Returns valid JSON
- Card data structure confirmed

#### Limitations

**Documentation:**
- ❌ No public API documentation
- ❌ No published rate limits
- ❌ No official terms of service for API usage
- ❌ No authentication requirements documented

**Coverage:**
- ⚠️ Limited to sets with active draft data
- ⚠️ Data sparse for newly released sets (needs draft play accumulation)
- ⚠️ Historical sets may have limited or no data
- ⚠️ No constructed format data

**Reliability:**
- ⚠️ Unofficial API (no service guarantees)
- ⚠️ May change without notice
- ⚠️ No uptime SLA

### Scryfall API

**Primary Focus:** Comprehensive Magic card database and metadata

#### Available Data

**Card Metadata:**
- Names (in all languages)
- Mana cost, CMC, color identity
- Type line, oracle text, flavor text
- Power/toughness, loyalty
- Set information, collector number
- Rarity, legality by format
- Artist information
- Card variations and printings

**Images:**
- Multiple sizes: `small`, `normal`, `large`, `png`, `art_crop`, `border_crop`
- High-quality images for all printings
- Card faces for DFCs, MDFCs, split cards

**Additional Features:**
- Full-text search with advanced syntax
- Rulings and oracle text
- Price information (multiple markets)
- Set symbols and icons
- Related cards and tokens
- MTGO/MTGA IDs

#### API Endpoints

```
# Search cards
https://api.scryfall.com/cards/search?q={query}

# Get specific card
https://api.scryfall.com/cards/{id}

# Get set information
https://api.scryfall.com/sets/{code}

# Bulk data downloads
https://api.scryfall.com/bulk-data
```

#### Data Format

- **Format:** JSON
- **Standardized:** Well-documented schema
- **Versioned:** API changes are documented

#### Testing Results

✅ **API Status:** Working (tested with Bloomburrow BLB set)
- Returns complete set information
- 398 cards in BLB set
- Release date: August 2, 2024

#### Terms of Service

**Rate Limits:**
- ✅ Documented: 50-100ms delay between requests (~10 req/sec)
- `HTTP 429` for exceeding limits
- Temporary or permanent IP ban for abuse

**Usage Restrictions:**
- ✅ **Free to use** for Magic-related software and research
- ❌ **No paywalling** Scryfall data
- ❌ **No republishing** or proxying Scryfall data
- ❌ **No game creation** using Scryfall data
- ✅ **Caching encouraged** (minimum 24 hours)

**Image Usage:**
- Cannot distort or watermark
- Require artist attribution for art crops
- Cannot misrepresent source

**Enforcement:**
- May restrict/block API access for violations

#### Strengths

- ✅ Comprehensive documentation
- ✅ Clear terms of service
- ✅ Documented rate limits
- ✅ All Magic sets (historical to current)
- ✅ Multiple image qualities
- ✅ Reliable uptime
- ✅ Active development and support
- ✅ Bulk data downloads available

---

## Feature Comparison Matrix

| Feature | 17Lands | Scryfall | Winner |
|---------|----------|----------|--------|
| **Card Metadata** | Basic | Comprehensive | Scryfall |
| **Win Rate Data** | ✅ Extensive | ❌ None | 17Lands |
| **Draft Statistics** | ✅ Yes | ❌ No | 17Lands |
| **Pick/Play Data** | ✅ Yes | ❌ No | 17Lands |
| **Color-Specific Ratings** | ✅ Yes | ❌ No | 17Lands |
| **Card Images** | ✅ Yes | ✅ Higher quality | Scryfall |
| **Free API** | ✅ Yes | ✅ Yes | Tie |
| **Rate Limits** | ❌ Unknown | ✅ Documented | Scryfall |
| **Documentation** | ❌ None | ✅ Excellent | Scryfall |
| **Terms of Service** | ❌ Unclear | ✅ Clear | Scryfall |
| **Historical Data** | ✅ Date ranges | ✅ All sets | Scryfall |
| **Real-time Updates** | ✅ Yes | ✅ Yes | Tie |
| **Set Coverage** | ⚠️ Limited | ✅ Complete | Scryfall |
| **Reliability** | ⚠️ Unofficial | ✅ Official | Scryfall |
| **Bayesian Averaging** | ✅ Yes | ❌ N/A | 17Lands |
| **Bulk Downloads** | ❌ No | ✅ Yes | Scryfall |

---

## Recommended Approach: Hybrid

### Architecture

```
┌─────────────────────┐
│   Card Metadata     │
│     (Scryfall)      │
│  - Names, types     │
│  - Mana costs       │
│  - Images (HQ)      │
│  - Oracle text      │
└──────────┬──────────┘
           │
           ├─── Combined ──→ Local Database
           │
┌──────────┴──────────┐
│  Draft Statistics   │
│    (17Lands)        │
│  - Win rates        │
│  - Pick statistics  │
│  - Color ratings    │
│  - ALSA/ATA         │
└─────────────────────┘
```

### Implementation Strategy

#### Phase 1: Scryfall Integration (Foundation)
1. **Set up Scryfall client** with rate limiting (100ms delays)
2. **Implement card cache** (SQLite database)
3. **Bulk data import** for initial seed
4. **Update mechanism** for new sets
5. **Card lookup by MTGA ID** (primary key for MTGA data)

#### Phase 2: 17Lands Integration (Enhanced Features)
1. **17Lands client** with conservative rate limiting
2. **Draft statistics cache** (separate table)
3. **Periodic updates** for active sets (daily/weekly)
4. **Graceful fallback** when 17Lands unavailable
5. **Historical data retention** for trend analysis

#### Phase 3: Combined Data Layer
1. **Unified card model** merging both sources
2. **Priority system**: Scryfall for metadata, 17Lands for stats
3. **Data staleness tracking** for refresh decisions
4. **Export combined data** for offline use

### Benefits of Hybrid Approach

**1. Best of Both Worlds**
- Comprehensive card data (Scryfall)
- Unique draft insights (17Lands)

**2. Reliability**
- Primary data from stable API (Scryfall)
- Enhanced features from 17Lands when available

**3. Legal Clarity**
- Clear terms of service (Scryfall)
- Compliant usage patterns

**4. Flexibility**
- Can function without 17Lands
- Can add future data sources easily

**5. User Value**
- Rich card information for all formats
- Advanced draft recommendations

### Fallback Strategy

```
1. Try 17Lands for draft stats
   ├─ Success → Use data
   └─ Failure → Continue with Scryfall only

2. Always use Scryfall for card metadata
   ├─ Cache available → Use cache
   └─ Cache miss → Fetch from API

3. Graceful degradation
   └─ Draft features work without stats
       (just missing recommendations)
```

---

## Implementation Effort Estimates

### Option 1: Hybrid (Recommended)

**Effort:** Medium-High (3-4 weeks)

**Tasks:**
1. **Scryfall Integration:** 1.5 weeks
   - Client library with rate limiting
   - Database schema for cards
   - Bulk data import
   - Update mechanism
   - Unit tests

2. **17Lands Integration:** 1 week
   - Client library
   - Database schema for statistics
   - Update mechanism
   - Caching layer
   - Unit tests

3. **Data Merging Layer:** 1 week
   - Combined data model
   - Query interface
   - Fallback logic
   - Integration tests

4. **Documentation:** 2-3 days
   - API usage guide
   - Data refresh schedules
   - Troubleshooting

**Complexity:** Moderate
- Two API integrations
- Data synchronization
- Error handling

**Risk:** Low
- Well-tested APIs
- Clear migration path

### Option 2: 17Lands Only

**Effort:** Low-Medium (2 weeks)

**Tasks:**
1. Client library: 1 week
2. Data caching: 3 days
3. Update mechanism: 2 days
4. Testing: 2 days

**Limitations:**
- Missing card metadata for non-draft features
- Sparse data for some sets
- Unofficial API risks

**Risk:** Medium
- Undocumented API may change
- No terms of service protection

### Option 3: Scryfall Only (Current State)

**Effort:** None (already implemented)

**Limitations:**
- ❌ No draft statistics
- ❌ No pick recommendations
- ❌ No win rate data
- ❌ Limits draft overlay features

**Risk:** None (stable)

---

## Legal & Terms of Service Considerations

### Scryfall ✅

**Clear Terms:**
- Free for Magic-related software
- Cannot paywall or republish data
- Must respect rate limits
- Caching encouraged

**Compliance Requirements:**
- Implement 100ms delay between requests
- Cache data for at least 24 hours
- Provide free access to Scryfall data
- Attribute images to artists

**Risk Level:** Low - Clear rules, easy to comply

### 17Lands ⚠️

**Unclear Terms:**
- No published API documentation
- No stated terms of service
- No official rate limits
- Unofficial API usage

**Mitigation Strategies:**
- Conservative rate limiting (1 req/sec or slower)
- Cache aggressively
- Implement fallback to Scryfall
- Monitor for API changes
- Consider reaching out to 17Lands team

**Risk Level:** Medium - Unofficial use, may change

---

## Conclusion

### ✅ Recommended: Hybrid Approach

The hybrid approach provides the best value:
1. **Scryfall** as primary data source (stable, documented, complete)
2. **17Lands** for enhanced draft features (unique value proposition)
3. **Graceful degradation** when 17Lands unavailable
4. **Legal compliance** through Scryfall ToS adherence
5. **Future-proof** architecture allowing additional sources

### Implementation Priorities

**High Priority:**
1. Scryfall integration (foundation for all features)
2. Card metadata caching

**Medium Priority:**
3. 17Lands integration (enhanced draft features)
4. Statistics caching

**Low Priority:**
5. Historical data analysis
6. Additional data sources

### Next Steps

1. Create issue for Scryfall integration
2. Design database schema for card metadata
3. Implement Scryfall client with rate limiting
4. Add 17Lands support after Scryfall foundation
5. Build unified query layer

---

## References

- **17Lands:** https://www.17lands.com/
- **17Lands API (undocumented):** https://www.17lands.com/card_ratings/data
- **Scryfall API:** https://scryfall.com/docs/api
- **Scryfall ToS:** https://scryfall.com/docs/api
- **MTGA_Draft_17Lands Reference:** External project using 17Lands exclusively

---

**Research Completed By:** Claude Code
**Review Status:** Ready for team review
**Decision Required:** Approve hybrid approach and prioritize implementation
