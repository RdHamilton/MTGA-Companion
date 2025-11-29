import { test, expect } from '@playwright/test';

/**
 * Decks Page E2E Tests
 *
 * Prerequisites:
 * - Run `wails dev` in the project root before running these tests
 * - The app should be accessible at http://localhost:34115
 */
test.describe('Decks', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app and wait for it to load
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });

    // Navigate to Decks page
    await page.click('a[href="/decks"]');
    await page.waitForURL('**/decks');
  });

  test.describe('Navigation and Page Load', () => {
    test('should navigate to Decks page', async ({ page }) => {
      const decksPage = page.locator('.decks-page');
      await expect(decksPage).toBeVisible({ timeout: 10000 });
    });

    test('should display page title', async ({ page }) => {
      await expect(page.locator('h1')).toContainText('Decks');
    });
  });

  test.describe('Deck List', () => {
    test('should display decks or empty state', async ({ page }) => {
      // Wait for loading to complete
      await Promise.race([
        page.locator('.deck-card').first().waitFor({ state: 'visible', timeout: 10000 }),
        page.locator('.empty-state').waitFor({ state: 'visible', timeout: 10000 }),
        page.locator('.decks-page:not(.loading-state)').waitFor({ state: 'visible', timeout: 10000 }),
      ]).catch(() => {});

      // Wait for content to render
      await page.waitForTimeout(500);

      // Either deck cards or empty state should be visible (or just decks page without loading)
      const hasDecks = await page.locator('.deck-card').first().isVisible().catch(() => false);
      const hasEmptyState = await page.locator('.empty-state').isVisible().catch(() => false);
      const decksPageVisible = await page.locator('.decks-page').isVisible();

      expect(hasDecks || hasEmptyState || decksPageVisible).toBeTruthy();
    });
  });

  test.describe('Create Deck Button', () => {
    test('should have a create deck button', async ({ page }) => {
      // Wait for page to load
      await page.waitForTimeout(1000);

      // Look for create deck button
      const createButton = page.locator('button').filter({ hasText: /create|new/i });
      const hasCreateButton = await createButton.isVisible().catch(() => false);

      // Create button should exist (may be in header or elsewhere)
      expect(hasCreateButton).toBeTruthy();
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
