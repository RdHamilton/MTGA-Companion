import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for E2E testing of MTGA Companion
 *
 * This configuration uses the REST API backend for testing, which allows
 * E2E tests to run without the Wails runtime.
 *
 * Test Projects:
 * - smoke: Quick critical path tests (@smoke tagged) - for post-merge validation
 * - full: All E2E tests - for release validation
 * - firefox: Smoke tests on Firefox - for cross-browser release testing
 * - webkit: Smoke tests on WebKit - for cross-browser release testing
 *
 * The webServer config starts:
 * 1. Go REST API server on port 8080
 * 2. Vite dev server on port 5173 (with REST API mode enabled)
 */
export default defineConfig({
  testDir: './tests/e2e',

  // Maximum time one test can run for
  timeout: 30 * 1000,

  // Run tests sequentially for consistent state management
  fullyParallel: false,
  workers: 1,

  // Fail the build on CI if you accidentally left test.only in the source code
  forbidOnly: !!process.env.CI,

  // Retry on CI only
  retries: process.env.CI ? 2 : 0,

  // Reporter to use
  reporter: [
    ['html', { open: 'never' }],
    ['list'],
    ...(process.env.CI ? [['github'] as const] : []),
  ],

  // Shared settings for all the projects below
  use: {
    // Base URL for the Vite dev server
    baseURL: 'http://localhost:3000',

    // Collect trace on failure for debugging
    trace: 'on-first-retry',

    // Take screenshot on failure
    screenshot: 'only-on-failure',

    // Record video on failure
    video: 'retain-on-failure',
  },

  // Configure projects for different test scenarios
  projects: [
    // Smoke tests - quick critical path validation
    {
      name: 'smoke',
      grep: /@smoke/,
      use: { ...devices['Desktop Chrome'] },
    },
    // Full test suite - all E2E tests
    {
      name: 'full',
      use: { ...devices['Desktop Chrome'] },
    },
    // Cross-browser smoke tests for releases
    {
      name: 'firefox',
      grep: /@smoke/,
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'webkit',
      grep: /@smoke/,
      use: { ...devices['Desktop Safari'] },
    },
  ],

  // Start both servers for E2E testing
  webServer: [
    // Go REST API server on port 8080
    // In CI, use a temp database with fixtures; locally, reuse existing server
    {
      command: process.env.CI
        ? 'cd .. && go run ./cmd/apiserver --db-path=/tmp/e2e-test.db --load-fixtures=frontend/tests/e2e/fixtures/test-data.sql'
        : 'cd .. && go run ./cmd/apiserver',
      url: 'http://localhost:8080/health',
      timeout: 120 * 1000,
      reuseExistingServer: !process.env.CI,
      stdout: 'pipe',
      stderr: 'pipe',
    },
    // Vite dev server on port 3000 with REST API mode
    {
      command: 'VITE_USE_REST_API=true npm run dev',
      url: 'http://localhost:3000',
      timeout: 60 * 1000,
      reuseExistingServer: !process.env.CI,
      stdout: 'pipe',
      stderr: 'pipe',
    },
  ],
});
