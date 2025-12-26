# E2E Testing with Playwright

This directory contains end-to-end tests for MTGA Companion using Playwright.

## Architecture

E2E tests use the **REST API mode** which enables testing without the Wails runtime:
- **Go REST API server** runs on port 8080
- **Vite dev server** runs on port 5173 with `VITE_USE_REST_API=true`
- Playwright starts both servers automatically

## Quick Start

From the `frontend/` directory:

```bash
# Run all E2E tests (servers start automatically)
npm run test:e2e

# Run tests with UI (useful for debugging)
npm run test:e2e:ui

# Run tests in debug mode
npm run test:e2e:debug
```

**Run specific test files:**
```bash
npm run test:e2e -- match-history.spec.ts
npm run test:e2e -- quests.spec.ts
npm run test:e2e -- collection.spec.ts
```

**Debug mode (see browser):**
```bash
npm run test:e2e -- --headed
npm run test:e2e -- --debug
```

## Manual Server Start (Optional)

If you prefer to start servers manually:

**Terminal 1 - Start Go API Server:**
```bash
cd /path/to/MTGA-Companion
go run ./cmd/apiserver
```
Wait for: `API server running at http://localhost:8080`

**Terminal 2 - Start Vite Dev Server:**
```bash
cd /path/to/MTGA-Companion/frontend
VITE_USE_REST_API=true npm run dev
```
Wait for: `Local: http://localhost:5173`

**Terminal 3 - Run E2E Tests:**
```bash
cd /path/to/MTGA-Companion/frontend
npm run test:e2e
```

## Test Structure

### Match History Tests (`match-history.spec.ts`)
- Navigation and page load
- Filter controls (date range, format, queue, result)
- Match table display and sorting
- Loading and empty states

### Collection Tests (`collection.spec.ts`)
- Collection overview
- Set completion
- Card filtering and search

### Draft Tests (`draft.spec.ts`)
- Draft session list
- Pick analysis
- Deck building from draft

### Deck Tests (`decks.spec.ts`)
- Deck library
- Create/edit/delete decks
- Import/export functionality

### Quest Tests (`quests.spec.ts`)
- Active quests display
- Quest history
- Daily/weekly win tracking

### Charts Tests (`charts.spec.ts`)
- Win rate over time
- Format distribution
- Performance metrics

### Settings Tests (`settings.spec.ts`)
- Application settings
- Database configuration
- Theme preferences

### Meta Tests (`meta.spec.ts`)
- Meta analysis
- Archetype information

## Writing Tests

E2E tests should:
- Test complete user workflows from start to finish
- Use data-testid attributes for reliable element selection when possible
- Be independent and not rely on the state from other tests
- Clean up any test data they create
- Use the REST API adapter (tests run in REST API mode)

## Configuration

See `playwright.config.ts` in the frontend root for configuration details.

The config automatically starts:
1. Go REST API server (`go run ./cmd/apiserver`) on port 8080
2. Vite dev server (`VITE_USE_REST_API=true npm run dev`) on port 5173

## Troubleshooting

**Test fails with "page.goto: net::ERR_CONNECTION_REFUSED"**
- The API server or Vite dev server may have failed to start
- Check the terminal output for error messages
- Verify ports 8080 and 5173 are available

**API server fails to start**
- Ensure Go is installed and in your PATH
- Run `go build ./cmd/apiserver` to check for compilation errors
- Check if port 8080 is already in use

**Tests are flaky**
- Check the trace files in `test-results/` for detailed execution info
- Use `await page.pause()` to debug interactively
- Increase timeout values if needed for slower operations
- The REST API mode is more stable than Wails mode

**View test report:**
```bash
npx playwright show-report
```
