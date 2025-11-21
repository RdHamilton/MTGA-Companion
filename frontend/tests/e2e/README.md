# E2E Testing with Playwright

This directory contains end-to-end tests for MTGA Companion using Playwright.

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

- `smoke.spec.ts` - Basic smoke tests to verify app loads and navigation works
- More workflow-specific tests will be added in subsequent tickets

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
