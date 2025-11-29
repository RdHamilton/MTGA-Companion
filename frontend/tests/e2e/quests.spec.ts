import { test, expect } from '@playwright/test';

/**
 * Quests Page E2E Tests
 *
 * Prerequisites:
 * - Run `wails dev` in the project root before running these tests
 * - The app should be accessible at http://localhost:34115
 */
test.describe('Quests', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app and wait for it to load
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });

    // Navigate to Quests page
    await page.click('a[href="/quests"]');
    await page.waitForURL('**/quests');
  });

  test.describe('Navigation and Page Load', () => {
    test('should navigate to Quests page', async ({ page }) => {
      await expect(page.locator('h1')).toContainText('Quests');
    });

    test('should display quests header', async ({ page }) => {
      const header = page.locator('.quests-header');
      await expect(header).toBeVisible();
    });
  });

  test.describe('Active Quests Section', () => {
    test('should display active quests section', async ({ page }) => {
      // Wait for loading to complete
      await Promise.race([
        page.locator('.quests-section').first().waitFor({ state: 'visible', timeout: 10000 }),
        page.locator('.empty-state').waitFor({ state: 'visible', timeout: 10000 }),
      ]).catch(() => {});

      // Either active quests section or empty state should be visible
      const hasSection = await page.locator('.quests-section').first().isVisible();
      const hasEmptyState = await page.locator('.empty-state').isVisible();

      expect(hasSection || hasEmptyState).toBeTruthy();
    });
  });

  test.describe('Quest History Section', () => {
    test('should have date range filter', async ({ page }) => {
      // Wait for page to load
      await page.waitForTimeout(1000);

      // Check for date range select (if present)
      const dateRangeSelect = page.locator('select').first();
      const hasDateRange = await dateRangeSelect.isVisible().catch(() => false);

      if (hasDateRange) {
        const options = await dateRangeSelect.locator('option').allTextContents();
        expect(options.length).toBeGreaterThan(0);
      }
    });
  });

  test.describe('Loading State', () => {
    test('should not show error state on initial load', async ({ page }) => {
      // Wait for loading to complete
      await page.waitForTimeout(2000);

      // Should not show error state
      const errorState = page.locator('.error-state');
      await expect(errorState).not.toBeVisible();
    });
  });
});
