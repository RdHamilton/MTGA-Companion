# Phase 2 Progress - Card Data Integration (Hybrid Approach)

## Overview
Phase 2 integrates 17Lands draft statistics with Phase 1's Scryfall card metadata to provide comprehensive card data.

## Completed

### Phase 1 (100% Complete)
- ✅ #209 - Scryfall client library (PR #224)
- ✅ #210 - Database schema for card metadata (PR #225)
- ✅ #211 - Bulk data import (PR #226)
- ✅ #212 - Update mechanism for new sets (PR #227)
- ✅ #213 - Card lookup integration (PR #228)

### Phase 2 (22% Complete)
- ✅ #214 - 17Lands client library (PR #229)
- ✅ #215 - Database schema for draft statistics (PR #230) ⭐ JUST COMPLETED

## Phase 2 Remaining Tasks (7 tasks)

### Next Up: #216 - Periodic Updates for Active Draft Sets
**Dependencies**: #214 ✅, #215 ✅
**Status**: Ready to implement
**What's needed**:
- Scheduler for active draft sets (weekly updates)
- Identify "active" sets (current Standard rotation)
- Background updater service
- Configuration for update frequency

### #217 - Graceful Fallback When 17Lands Unavailable
**Dependencies**: #214, #215
**What's needed**:
- Circuit breaker pattern
- Fallback to cached data (even if stale)
- Health check endpoint
- Retry logic with backoff

### #218 - Historical Draft Data Retention
**Dependencies**: #215
**What's needed**:
- Retention policy (keep N months)
- Cleanup job for old data
- Archival strategy
- Configuration for retention period

### #219 - Unified Card Model
**Dependencies**: #213 (Phase 1), #214, #215
**What's needed**:
- Combined model merging Scryfall + 17Lands data
- Single interface for both data sources
- Smart merging (Scryfall for metadata, 17Lands for draft stats)

### #220 - Query Interface with Data Priority
**Dependencies**: #219
**What's needed**:
- Unified query interface
- Data source priority (cache > 17Lands > Scryfall)
- Smart fallbacks
- Response combining

### #221 - Staleness Tracking and Refresh Scheduler
**Dependencies**: #215, #216
**What's needed**:
- Track last update per set/format
- Auto-refresh scheduler for stale data
- Configurable staleness thresholds
- Background refresh jobs

### #222 - Export Combined Card Data
**Dependencies**: #220
**What's needed**:
- Export commands for combined data
- CSV/JSON format support
- Include both metadata + draft stats
- CLI integration

## Current Branch State
- **Main branch**: Up to date with all Phase 1
- **Open PRs**:
  - #229 (17Lands client library - #214)
  - #230 (Database schema for draft statistics - #215) ⭐ JUST CREATED
- **Last completed**: #215 - Database schema for draft statistics

## Architecture Notes

### Data Flow
1. **Scryfall** (Phase 1) → Card metadata (name, type, colors, etc.)
2. **17Lands** (Phase 2) → Draft statistics (win rates, pick data)
3. **Unified Model** → Combined view of both sources

### Storage Strategy
- Separate tables for different data sources
- Join at query time for unified view
- Independent staleness tracking per source
- Batch updates for efficiency

## Next Steps for Continuation

1. **Complete #216** (Periodic Updates):
   - Create updater service
   - Add scheduler for active draft sets
   - CLI commands for manual updates
   - Configuration for update frequency

2. **Complete #217** (Graceful Fallback):
   - Circuit breaker implementation
   - Health monitoring
   - Fallback logic

3. **Complete #218** (Historical Data Retention):
   - Retention policy
   - Cleanup job for old data

4. Continue through remaining tasks (#219-#222) in order

## Key Files to Know

### Phase 1 (Scryfall)
- `internal/mtga/cards/scryfall/` - Scryfall client
- `internal/mtga/cardlookup/` - Card lookup service
- `internal/storage/cards.go` - Card storage
- `internal/storage/migrations/000006_*` - Card metadata schema

### Phase 2 (17Lands)
- `internal/mtga/cards/seventeenlands/` - 17Lands client ✅
- `internal/storage/migrations/000007_*` - Draft stats schema ✅
- `internal/storage/draft_statistics.go` - Data access layer ✅
- `internal/storage/draft_statistics_test.go` - Comprehensive tests ✅
- (More to be added as Phase 2 progresses)

## Testing Strategy
- Unit tests for all new storage methods
- Integration tests with real 17Lands data
- Migration tests (up/down)
- End-to-end tests for unified queries

## Performance Targets
- Query combined data: <50ms
- Batch insert 1000 ratings: <5s
- Cache staleness check: <10ms
- Background update: Once per week per active set
