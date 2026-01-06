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

  test.describe('DeckBuilder Build Around', () => {
    test('should show Build Around button for non-draft decks', async ({ page }) => {
      // Wait for decks to load
      const deckCard = page.locator('.deck-card');
      const emptyState = page.locator('.empty-state');
      await expect(deckCard.first().or(emptyState)).toBeVisible({ timeout: 10000 });

      const hasCards = await deckCard.first().isVisible();

      if (hasCards) {
        // Find a non-draft deck (look for Standard, Historic, etc. format badges)
        const nonDraftDeck = page.locator('.deck-card').filter({
          has: page.locator('.deck-format:not(:has-text("Limited"))')
        }).first();

        const hasNonDraft = await nonDraftDeck.isVisible().catch(() => false);

        if (hasNonDraft) {
          // Click on the deck to go to DeckBuilder
          await nonDraftDeck.click();
          await page.waitForURL('**/decks/**');

          // Wait for DeckBuilder to load
          const deckBuilder = page.locator('.deck-builder');
          await expect(deckBuilder).toBeVisible({ timeout: 10000 });

          // Build Around button should be visible for non-draft decks
          const buildAroundButton = page.locator('button.build-around-btn');
          await expect(buildAroundButton).toBeVisible();
          await expect(buildAroundButton).toContainText('Build Around');
        }
      }
    });

    test('should open Build Around modal when button clicked', async ({ page }) => {
      // Wait for decks to load
      const deckCard = page.locator('.deck-card');
      const emptyState = page.locator('.empty-state');
      await expect(deckCard.first().or(emptyState)).toBeVisible({ timeout: 10000 });

      const hasCards = await deckCard.first().isVisible();

      if (hasCards) {
        // Find a non-draft deck
        const nonDraftDeck = page.locator('.deck-card').filter({
          has: page.locator('.deck-format:not(:has-text("Limited"))')
        }).first();

        const hasNonDraft = await nonDraftDeck.isVisible().catch(() => false);

        if (hasNonDraft) {
          await nonDraftDeck.click();
          await page.waitForURL('**/decks/**');

          const buildAroundButton = page.locator('button.build-around-btn');
          const isButtonVisible = await buildAroundButton.isVisible().catch(() => false);

          if (isButtonVisible) {
            await buildAroundButton.click();

            // Modal should open
            const modal = page.locator('.build-around-modal');
            await expect(modal).toBeVisible({ timeout: 5000 });

            // Modal header should be visible
            const modalHeader = modal.locator('h2');
            await expect(modalHeader).toContainText('Build Around Card');

            // Search input should be present
            const searchInput = modal.locator('input[placeholder*="Search"]');
            await expect(searchInput).toBeVisible();

            // Close button should work
            const closeButton = modal.locator('.close-button');
            await closeButton.click();
            await expect(modal).not.toBeVisible();
          }
        }
      }
    });

    test('should NOT show Build Around button for draft decks', async ({ page }) => {
      // Wait for decks to load
      const deckCard = page.locator('.deck-card');
      const emptyState = page.locator('.empty-state');
      await expect(deckCard.first().or(emptyState)).toBeVisible({ timeout: 10000 });

      const hasCards = await deckCard.first().isVisible();

      if (hasCards) {
        // Find a draft/limited deck
        const draftDeck = page.locator('.deck-card').filter({
          has: page.locator('.deck-format:has-text("Limited")')
        }).first();

        const hasDraft = await draftDeck.isVisible().catch(() => false);

        if (hasDraft) {
          await draftDeck.click();
          await page.waitForURL('**/decks/**');

          const deckBuilder = page.locator('.deck-builder');
          await expect(deckBuilder).toBeVisible({ timeout: 10000 });

          // Build Around button should NOT be visible for draft decks
          const buildAroundButton = page.locator('button.build-around-btn');
          await expect(buildAroundButton).not.toBeVisible();

          // But Suggest Decks button SHOULD be visible for draft decks
          const suggestDecksButton = page.locator('button.suggest-decks-btn');
          await expect(suggestDecksButton).toBeVisible();
        }
      }
    });
  });
});
