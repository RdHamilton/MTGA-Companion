import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for E2E testing of MTGA Companion
 *
 * Note: This app uses Wails, which means tests need to run against
 * a running Wails development server. Before running tests, ensure
 * the app is running via `wails dev` in the project root.
 */
export default defineConfig({
  testDir: './tests/e2e',

  // Maximum time one test can run for
  timeout: 30 * 1000,

  // Run tests sequentially (important for Wails app state management)
  fullyParallel: false,
  workers: 1,

  // Fail the build on CI if you accidentally left test.only in the source code
  forbidOnly: !!process.env.CI,

  // Retry on CI only
  retries: process.env.CI ? 2 : 0,

  // Reporter to use
  reporter: [
    ['html', { open: 'never' }],
    ['list']
  ],

  // Shared settings for all the projects below
  use: {
    // Base URL for the Wails dev server
    // Note: You need to start `wails dev` before running tests
    baseURL: 'http://localhost:34115',

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
});
