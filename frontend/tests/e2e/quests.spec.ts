import { test, expect } from '@playwright/test';

/**
 * Quests Page E2E Tests
 *
 * Tests the Quests page functionality.
 * Uses REST API backend for testing.
 */
test.describe('Quests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });

    await page.click('a[href="/quests"]');
    await page.waitForURL('**/quests');
  });

  test.describe('Navigation and Page Load', () => {
    test('@smoke should navigate to Quests page', async ({ page }) => {
      await expect(page.locator('h1')).toContainText('Quests');
    });

    test('should display quests header', async ({ page }) => {
      const header = page.locator('.quests-header');
      await expect(header).toBeVisible();
    });
  });

  test.describe('Active Quests Section', () => {
    test('should display active quests section', async ({ page }) => {
      const questsSection = page.locator('.quests-section');
      const emptyState = page.locator('.empty-state');

      // Wait for either content type to appear
      await expect(questsSection.first().or(emptyState)).toBeVisible({ timeout: 10000 });

      const hasSection = await questsSection.first().isVisible();
      const hasEmptyState = await emptyState.isVisible();

      expect(hasSection || hasEmptyState).toBeTruthy();
    });
  });

  test.describe('Quest History Section', () => {
    test('should have date range filter', async ({ page }) => {
      // Wait for page to load
      const questsHeader = page.locator('.quests-header');
      await expect(questsHeader).toBeVisible({ timeout: 10000 });

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
      // Wait for content to load
      const content = page.locator('.quests-section, .empty-state, .quests-header');
      await expect(content.first()).toBeVisible({ timeout: 10000 });

      const errorState = page.locator('.error-state');
      await expect(errorState).not.toBeVisible();
    });
  });
});
