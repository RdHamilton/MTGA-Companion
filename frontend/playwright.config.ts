import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for E2E testing of MTGA Companion
 *
 * This configuration uses the REST API backend for testing, which allows
 * E2E tests to run without the Wails runtime.
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
  ],

  // Shared settings for all the projects below
  use: {
    // Base URL for the Vite dev server
    baseURL: 'http://localhost:5173',

    // Collect trace on failure for debugging
    trace: 'on-first-retry',

    // Take screenshot on failure
    screenshot: 'only-on-failure',

    // Record video on failure
    video: 'retain-on-failure',
  },

  // Configure projects for major browsers
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  // Start the Vite dev server with REST API mode before tests
  webServer: {
    command: 'VITE_USE_REST_API=true npm run dev',
    url: 'http://localhost:5173',
    timeout: 60 * 1000,
    reuseExistingServer: !process.env.CI,
    stdout: 'pipe',
    stderr: 'pipe',
  },
});
