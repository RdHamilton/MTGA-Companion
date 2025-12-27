import { test, expect } from '@playwright/test';

/**
 * Collection Page E2E Tests
 *
 * Tests the Collection page functionality including navigation and filters.
 * Uses REST API backend for testing.
 */
test.describe('Collection', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to home first, then to collection
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 15000 });
    await page.goto('/collection');
  });

  test.describe('Navigation and Page Load', () => {
    // TODO: Fix this test - collection page doesn't load properly in CI
    // See: https://github.com/RdHamilton/MTGA-Companion/issues/725
    test.skip('@smoke should navigate to Collection page', async ({ page }) => {
      // Verify URL is collection
      await expect(page).toHaveURL(/.*\/collection/);

      // Wait for page content - check for h1 with Collection text instead of specific class
      await expect(page.locator('h1')).toContainText('Collection', { timeout: 20000 });
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
      // Wait for stats to load
      const stats = page.locator('.collection-stats-summary');
      await expect(stats).toBeVisible({ timeout: 10000 });
    });
  });

  test.describe('Filter Controls', () => {
    test('should have search input', async ({ page }) => {
      const searchInput = page.locator('input[type="text"], input[placeholder*="earch"]');
      await expect(searchInput.first()).toBeVisible({ timeout: 5000 });
    });

    test('should have set filter dropdown', async ({ page }) => {
      const setFilter = page.locator('select').first();
      await expect(setFilter).toBeVisible({ timeout: 5000 });
    });

    test('should have rarity filter', async ({ page }) => {
      const rarityFilter = page.locator('.rarity-filter, select').nth(1);
      await expect(rarityFilter).toBeVisible({ timeout: 5000 });
    });
  });

  test.describe('Collection Content', () => {
    test('should display collection cards or empty state', async ({ page }) => {
      const collectionCard = page.locator('.collection-card');
      const emptyState = page.locator('.empty-state');
      const collectionPage = page.locator('.collection-page');

      // Wait for content to load
      await expect(collectionCard.first().or(emptyState).or(collectionPage)).toBeVisible({ timeout: 10000 });

      const hasCards = await collectionCard.first().isVisible();
      const hasEmptyState = await emptyState.isVisible();
      const pageVisible = await collectionPage.isVisible();

      expect(hasCards || hasEmptyState || pageVisible).toBeTruthy();
    });
  });

  test.describe('Set Completion', () => {
    test('should have set completion toggle or section', async ({ page }) => {
      // Wait for page to load
      const collectionPage = page.locator('.collection-page');
      await expect(collectionPage).toBeVisible({ timeout: 10000 });

      const setCompletionButton = page.locator('button').filter({ hasText: /set completion/i });
      const setCompletionSection = page.locator('.set-completion');

      const hasButton = await setCompletionButton.isVisible().catch(() => false);
      const hasSection = await setCompletionSection.isVisible().catch(() => false);

      expect(hasButton || hasSection).toBeTruthy();
    });
  });

  test.describe('Loading State', () => {
    test('should not show error state on initial load', async ({ page }) => {
      // Wait for content to load
      const content = page.locator('.collection-page');
      await expect(content).toBeVisible({ timeout: 10000 });

      const errorState = page.locator('.error-state');
      await expect(errorState).not.toBeVisible();
    });
  });
});
