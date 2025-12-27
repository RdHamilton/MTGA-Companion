import { test, expect } from '@playwright/test';

/**
 * Decks Page E2E Tests
 *
 * Tests the Decks page functionality including navigation and deck management.
 * Uses REST API backend for testing.
 */
test.describe('Decks', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });

    await page.click('a[href="/decks"]');
    await page.waitForURL('**/decks');
  });

  test.describe('Navigation and Page Load', () => {
    test('@smoke should navigate to Decks page', async ({ page }) => {
      await expect(page.locator('h1')).toContainText('Decks');
    });

    test('should display page title', async ({ page }) => {
      const header = page.locator('h1');
      await expect(header).toBeVisible();
      await expect(header).toContainText('Decks');
    });
  });

  test.describe('Deck List', () => {
    test('should display deck cards or empty state', async ({ page }) => {
      const deckCard = page.locator('.deck-card');
      const emptyState = page.locator('.empty-state');

      // Wait for either content type to appear
      await expect(deckCard.first().or(emptyState)).toBeVisible({ timeout: 10000 });

      const hasCards = await deckCard.first().isVisible();
      const hasEmptyState = await emptyState.isVisible();

      expect(hasCards || hasEmptyState).toBeTruthy();
    });
  });

  test.describe('Create Deck', () => {
    test('should have create deck button', async ({ page }) => {
      // Wait for page to fully load
      const pageContent = page.locator('.deck-card, .empty-state, .decks-header');
      await expect(pageContent.first()).toBeVisible({ timeout: 10000 });

      const createButton = page.locator('button').filter({ hasText: /create|new/i });
      const hasButton = await createButton.isVisible().catch(() => false);

      expect(hasButton).toBeTruthy();
    });
  });

  test.describe('Loading State', () => {
    test('should not show error state on initial load', async ({ page }) => {
      // Wait for content to load
      const content = page.locator('.deck-card, .empty-state');
      await expect(content.first()).toBeVisible({ timeout: 10000 });

      const errorState = page.locator('.error-state');
      await expect(errorState).not.toBeVisible();
    });
  });
});
