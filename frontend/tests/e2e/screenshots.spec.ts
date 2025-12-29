import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

/**
 * Documentation Screenshots
 *
 * This test file generates screenshots for documentation purposes.
 * Screenshots are saved to docs/images/ for use in README and documentation.
 *
 * Run with: npm run screenshots
 *
 * Note: Screenshots are generated using test fixtures for realistic data.
 * Dark theme is used by default to match the application's default appearance.
 */

const SCREENSHOT_DIR = path.join(__dirname, '../../../../docs/images');

// Ensure the screenshots directory exists
test.beforeAll(async () => {
  if (!fs.existsSync(SCREENSHOT_DIR)) {
    fs.mkdirSync(SCREENSHOT_DIR, { recursive: true });
  }
});

test.describe('Documentation Screenshots', () => {
  test.beforeEach(async ({ page }) => {
    // Set viewport for consistent screenshots
    await page.setViewportSize({ width: 1280, height: 800 });
  });

  test('capture Match History page', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 15000 });

    // Wait for data to load
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(500); // Allow animations to complete

    await page.screenshot({
      path: path.join(SCREENSHOT_DIR, 'match-history.png'),
      fullPage: false,
    });
  });

  test('capture Draft History page', async ({ page }) => {
    await page.goto('/draft');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 15000 });

    // Wait for draft list to load
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(500);

    await page.screenshot({
      path: path.join(SCREENSHOT_DIR, 'draft-history.png'),
      fullPage: false,
    });
  });

  test('capture Decks page', async ({ page }) => {
    await page.goto('/decks');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 15000 });

    // Wait for decks to load
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(500);

    await page.screenshot({
      path: path.join(SCREENSHOT_DIR, 'decks.png'),
      fullPage: false,
    });
  });

  test('capture Quests page', async ({ page }) => {
    await page.goto('/quests');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 15000 });

    // Wait for quest data to load
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(500);

    await page.screenshot({
      path: path.join(SCREENSHOT_DIR, 'quests.png'),
      fullPage: false,
    });
  });

  test('capture Meta Dashboard', async ({ page }) => {
    await page.goto('/meta');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 15000 });

    // Wait for meta data to load
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000); // Meta data may take longer

    await page.screenshot({
      path: path.join(SCREENSHOT_DIR, 'meta-dashboard.png'),
      fullPage: false,
    });
  });

  test('capture Charts - Deck Performance', async ({ page }) => {
    await page.goto('/charts/deck-performance');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 15000 });

    // Wait for chart to render
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(500);

    await page.screenshot({
      path: path.join(SCREENSHOT_DIR, 'charts-deck-performance.png'),
      fullPage: false,
    });
  });

  test('capture Charts - Format Distribution', async ({ page }) => {
    await page.goto('/charts/format-distribution');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 15000 });

    // Wait for chart to render
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(500);

    await page.screenshot({
      path: path.join(SCREENSHOT_DIR, 'charts-format-distribution.png'),
      fullPage: false,
    });
  });

  test('capture Charts - Result Breakdown', async ({ page }) => {
    await page.goto('/charts/result-breakdown');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 15000 });

    // Wait for chart to render
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(500);

    await page.screenshot({
      path: path.join(SCREENSHOT_DIR, 'charts-result-breakdown.png'),
      fullPage: false,
    });
  });

  test('capture Settings page', async ({ page }) => {
    await page.goto('/settings');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 15000 });

    // Wait for settings to load
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(500);

    await page.screenshot({
      path: path.join(SCREENSHOT_DIR, 'settings.png'),
      fullPage: false,
    });
  });
});
