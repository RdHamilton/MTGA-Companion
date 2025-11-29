import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for E2E testing of MTGA Companion
 *
 * Note: This app uses Wails. Tests can either:
 * 1. Use an already running `wails dev` server (start manually)
 * 2. Use the webServer config below (uncomment to auto-start server)
 *
 * The webServer config will start the server before tests and stop it after.
 * If a server is already running on port 34115, it will reuse that server.
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
    ['list'],
  ],

  // Shared settings for all the projects below
  use: {
    // Base URL for the Wails dev server
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

  // Uncomment to automatically start Wails dev server before tests
  // Note: Requires `wails` to be in PATH (or use full path like ~/go/bin/wails)
  // webServer: {
  //   command: 'cd .. && wails dev',
  //   url: 'http://localhost:34115',
  //   timeout: 60 * 1000,
  //   reuseExistingServer: true,
  //   stdout: 'pipe',
  //   stderr: 'pipe',
  // },
});
