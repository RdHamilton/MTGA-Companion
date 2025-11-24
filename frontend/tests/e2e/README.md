# E2E Testing with Playwright

This directory contains end-to-end tests for MTGA Companion using Playwright.

## ⚠️ CI Status

**E2E tests are disabled in CI** and must be run locally before submitting PRs.

**Why?**
- Wails builds native desktop applications (not web apps)
- Desktop apps require native GUI components (GTK, WebKit, native windows)
- Running desktop apps in headless CI environments with Xvfb is unreliable
- GitHub Actions runners have difficulty with virtual displays and native graphics

**CI Coverage:** Unit tests, component tests, linting, and security scans all run in CI.

## Prerequisites

Before running E2E tests, you need to have the Wails application running:

1. From the project root directory, start the development server:
   ```bash
   wails dev
   ```

2. Wait for the app to be fully loaded (accessible at http://localhost:34115)

## Running Tests

From the `frontend/` directory:

```bash
# Run all E2E tests (headless mode)
npm run test:e2e

# Run tests with UI (useful for debugging)
npm run test:e2e:ui

# Run tests in debug mode with step-by-step execution
npm run test:e2e:debug

# View the last test report
npx playwright show-report
```

## Test Structure

### Smoke Tests (`smoke.spec.ts`)
Basic smoke tests to verify app loads and navigation works.

### Deck Builder Tests (`deck-builder.spec.ts`) - 18 tests
- Navigation and initial state (3 tests)
- Create deck modal (6 tests)
- Draft-to-deck workflow (4 tests)
- Deck builder page (3 tests)
- Error handling (2 tests)

### Quest Tests (`quests.spec.ts`) - 30+ tests
- Navigation and initial state (3 tests)
- Quest list display (4 tests)
- Quest statistics (3 tests)
- Quest status and completion (3 tests)
- Quest filtering and sorting (2 tests)
- Quest details (3 tests)
- Empty state (2 tests)
- Error handling (3 tests)
- Performance and loading (2 tests)
- Visual layout (2 tests)

## Writing Tests

E2E tests should:
- Test complete user workflows from start to finish
- Use data-testid attributes for reliable element selection when possible
- Be independent and not rely on the state from other tests
- Clean up any test data they create

## Configuration

See `playwright.config.ts` in the frontend root for configuration details.

## Troubleshooting

**Test fails with "page.goto: net::ERR_CONNECTION_REFUSED"**
- Make sure `wails dev` is running before starting tests
- Verify the app is accessible at http://localhost:34115

**Tests are flaky**
- Check the trace files in `test-results/` for detailed execution info
- Use `await page.pause()` to debug interactively
- Increase timeout values if needed for slower operations
