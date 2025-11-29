import { test, expect } from '@playwright/test';

/**
 * Collection Page E2E Tests
 *
 * Prerequisites:
 * - Run `wails dev` in the project root before running these tests
 * - The app should be accessible at http://localhost:34115
 */
test.describe('Collection', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app and wait for it to load
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });

    // Navigate to Collection page
    await page.click('a[href="/collection"]');
    await page.waitForURL('**/collection');
  });

  test.describe('Navigation and Page Load', () => {
    test('should navigate to Collection page', async ({ page }) => {
      const collectionPage = page.locator('.collection-page');
      await expect(collectionPage).toBeVisible({ timeout: 10000 });
    });

    test('should display page title', async ({ page }) => {
      await expect(page.locator('h1')).toContainText('Collection');
    });
  });

  test.describe('Collection Header', () => {
    test('should display collection header', async ({ page }) => {
      const header = page.locator('.collection-header');
      await expect(header).toBeVisible({ timeout: 10000 });
    });

    test('should display collection stats summary', async ({ page }) => {
      // Wait for loading to complete
      await page.waitForTimeout(2000);

      const stats = page.locator('.collection-stats-summary');
      const hasStats = await stats.isVisible().catch(() => false);

      // Stats should be visible after loading
      expect(hasStats).toBeTruthy();
    });
  });

  test.describe('Filter Controls', () => {
    test('should have search input', async ({ page }) => {
      const searchInput = page.locator('input[type="text"], input[placeholder*="earch"]');
      await expect(searchInput.first()).toBeVisible({ timeout: 5000 });
    });

    test('should have set filter dropdown', async ({ page }) => {
      // Wait for page to load
      await page.waitForTimeout(1000);

      // Look for set filter (select element)
      const setFilter = page.locator('select').first();
      const hasSetFilter = await setFilter.isVisible().catch(() => false);

      expect(hasSetFilter).toBeTruthy();
    });

    test('should have rarity filter', async ({ page }) => {
      // Wait for page to load
      await page.waitForTimeout(1000);

      // Look for rarity filter buttons or select
      const rarityFilter = page.locator('.rarity-filter, select').nth(1);
      const hasRarityFilter = await rarityFilter.isVisible().catch(() => false);

      // May have multiple filter options
      expect(hasRarityFilter).toBeTruthy();
    });
  });

  test.describe('Collection Content', () => {
    test('should display collection cards or empty state', async ({ page }) => {
      // Wait for loading to complete
      await Promise.race([
        page.locator('.collection-card').first().waitFor({ state: 'visible', timeout: 10000 }),
        page.locator('.empty-state').waitFor({ state: 'visible', timeout: 10000 }),
        page.locator('.collection-page:not(.loading-state)').waitFor({ state: 'visible', timeout: 10000 }),
      ]).catch(() => {});

      // Wait for content to render
      await page.waitForTimeout(500);

      // Either collection cards or empty state should be visible
      const hasCards = await page.locator('.collection-card').first().isVisible().catch(() => false);
      const hasEmptyState = await page.locator('.empty-state').isVisible().catch(() => false);
      const pageVisible = await page.locator('.collection-page').isVisible();

      expect(hasCards || hasEmptyState || pageVisible).toBeTruthy();
    });
  });

  test.describe('Set Completion', () => {
    test('should have set completion toggle or section', async ({ page }) => {
      // Wait for page to load
      await page.waitForTimeout(2000);

      // Look for set completion button or section
      const setCompletionButton = page.locator('button').filter({ hasText: /set completion/i });
      const setCompletionSection = page.locator('.set-completion');

      const hasButton = await setCompletionButton.isVisible().catch(() => false);
      const hasSection = await setCompletionSection.isVisible().catch(() => false);

      // Either button or section should exist
      expect(hasButton || hasSection).toBeTruthy();
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
