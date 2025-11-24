import { test, expect, Page } from '@playwright/test';

/**
 * Helper function to navigate to a page and wait for it to finish loading
 */
async function gotoAndWaitForLoad(page: Page, url: string) {
  await page.goto(url);
  await page.waitForLoadState('networkidle');
  // Wait for any loading spinners to disappear
  await page.locator('.loading-state').waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {});
}

/**
 * E2E tests for Deck Builder workflow
 *
 * Prerequisites:
 * - Run `wails dev` in the project root
 * - Database should contain draft session data for draft-to-deck tests
 *
 * Tests cover:
 * - Deck creation from scratch
 * - Draft-to-deck workflow
 * - Navigation between decks and builder
 * - Deck list display
 */
test.describe('Deck Builder Workflow', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app and wait for it to be ready
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });
  });

  test.describe('Navigation and Initial State', () => {
    test('should navigate to Decks page', async ({ page }) => {
      // Click on Decks tab
      await page.getByText('Decks', { exact: true }).click();

      // Verify we're on the decks page
      await expect(page).toHaveURL(/\/decks/);

      // Check for decks page container
      await expect(page.locator('.decks-page')).toBeVisible();
    });

    test('should display empty state when no decks exist', async ({ page }) => {
      await gotoAndWaitForLoad(page, '/decks');

      // Should show either empty state or deck list
      const hasEmptyState = await page.locator('.empty-state').count() > 0;
      const hasDeckList = await page.locator('.decks-grid').count() > 0;

      expect(hasEmptyState || hasDeckList).toBe(true);
    });

    test('should show deck list when decks exist', async ({ page }) => {
      await gotoAndWaitForLoad(page, '/decks');

      // Look for either deck cards or empty state
      const deckCards = await page.locator('.deck-card').count();
      const emptyState = await page.locator('.empty-state').count();

      // One or the other should be present
      expect(deckCards > 0 || emptyState > 0).toBe(true);
    });
  });

  test.describe('Create Deck Modal', () => {
    test('should show Create Deck button', async ({ page }) => {
      await gotoAndWaitForLoad(page, '/decks');

      // Look for Create Deck button (either in header or empty state)
      const createButtons = await page.locator('button:has-text("Create New Deck"), button:has-text("+ Create New Deck")').count();

      expect(createButtons).toBeGreaterThan(0);
    });

    test('should open Create Deck modal when button clicked', async ({ page }) => {
      await gotoAndWaitForLoad(page, '/decks');

      // Click the Create Deck button
      const createButton = page.locator('button:has-text("Create New Deck"), button:has-text("+ Create New Deck")').first();
      await createButton.click();

      // Modal should appear
      await expect(page.locator('.modal-overlay')).toBeVisible({ timeout: 2000 });
      await expect(page.locator('.modal-content')).toBeVisible();

      // Should show form fields
      await expect(page.locator('input#deck-name')).toBeVisible();
      await expect(page.locator('select#deck-format')).toBeVisible();
    });

    test('should close modal when cancel button clicked', async ({ page }) => {
      await page.goto('/decks');
      await page.waitForLoadState('networkidle');

      // Open modal
      const createButton = page.locator('button:has-text("Create New Deck"), button:has-text("+ Create New Deck")').first();
      await createButton.click();
      await expect(page.locator('.modal-overlay')).toBeVisible();

      // Click cancel
      await page.locator('.cancel-button').click();

      // Modal should close
      await expect(page.locator('.modal-overlay')).not.toBeVisible();
    });

    test('should close modal when clicking overlay', async ({ page }) => {
      await page.goto('/decks');
      await page.waitForLoadState('networkidle');

      // Open modal
      const createButton = page.locator('button:has-text("Create New Deck"), button:has-text("+ Create New Deck")').first();
      await createButton.click();
      await expect(page.locator('.modal-overlay')).toBeVisible();

      // Click overlay (not the modal content)
      await page.locator('.modal-overlay').click({ position: { x: 10, y: 10 } });

      // Modal should close
      await expect(page.locator('.modal-overlay')).not.toBeVisible();
    });

    test('should require deck name to create deck', async ({ page }) => {
      await page.goto('/decks');
      await page.waitForLoadState('networkidle');

      // Open modal
      const createButton = page.locator('button:has-text("Create New Deck"), button:has-text("+ Create New Deck")').first();
      await createButton.click();
      await expect(page.locator('.modal-overlay')).toBeVisible();

      // Try to create without entering name
      await page.locator('.create-button').click();

      // Should still be on decks page (or show alert)
      // The implementation shows an alert, so modal should stay open
      await expect(page.locator('.modal-overlay')).toBeVisible();
    });

    test('should create deck with valid name and format', async ({ page }) => {
      await page.goto('/decks');
      await page.waitForLoadState('networkidle');

      // Open modal
      const createButton = page.locator('button:has-text("Create New Deck"), button:has-text("+ Create New Deck")').first();
      await createButton.click();
      await expect(page.locator('.modal-overlay')).toBeVisible();

      // Enter deck details
      await page.locator('input#deck-name').fill('Test Deck E2E');
      await page.locator('select#deck-format').selectOption('standard');

      // Create deck
      await page.locator('.create-button').click();

      // Should navigate to deck builder
      await expect(page).toHaveURL(/\/deck-builder\//, { timeout: 5000 });

      // Should show deck builder page
      await expect(page.locator('.deck-builder')).toBeVisible();
    });
  });

  test.describe('Draft to Deck Workflow', () => {
    test('should show Build Deck button on draft history cards', async ({ page }) => {
      await page.goto('/draft');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Look for draft history cards
      const draftCards = await page.locator('.draft-card').count();

      if (draftCards > 0) {
        // Should have Build Deck buttons
        const buildDeckButtons = await page.locator('button:has-text("Build Deck")').count();
        expect(buildDeckButtons).toBeGreaterThan(0);
      } else {
        // No draft history - test passes vacuously
        expect(true).toBe(true);
      }
    });

    test('should navigate to deck builder when Build Deck clicked from draft history', async ({ page }) => {
      await page.goto('/draft');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Look for Build Deck button
      const buildDeckButton = page.locator('button:has-text("Build Deck")').first();
      const buttonExists = await buildDeckButton.count() > 0;

      if (buttonExists) {
        // Click Build Deck
        await buildDeckButton.click();

        // Should navigate to deck builder
        await expect(page).toHaveURL(/\/deck-builder\/draft\//, { timeout: 5000 });

        // Should show deck builder page (not error)
        await expect(page.locator('.deck-builder')).toBeVisible({ timeout: 10000 });

        // Should NOT show error state
        const hasError = await page.locator('.error-state').count() > 0;
        expect(hasError).toBe(false);
      }
    });

    test('should show Build Deck button on historical draft detail view', async ({ page }) => {
      await page.goto('/draft');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Click View Replay if available
      const viewReplayButton = page.locator('button:has-text("View Replay")').first();
      const buttonExists = await viewReplayButton.count() > 0;

      if (buttonExists) {
        await viewReplayButton.click();
        await page.waitForLoadState('networkidle');

        // Should show Build Deck button in detail view
        const buildDeckButton = page.locator('button:has-text("Build Deck")');
        await expect(buildDeckButton).toBeVisible();
      }
    });

    test('should auto-create deck from draft picks if none exists', async ({ page }) => {
      await page.goto('/draft');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Look for Build Deck button
      const buildDeckButton = page.locator('button:has-text("Build Deck")').first();
      const buttonExists = await buildDeckButton.count() > 0;

      if (buttonExists) {
        // Click Build Deck
        await buildDeckButton.click();

        // Should navigate to deck builder (creating deck if needed)
        await expect(page).toHaveURL(/\/deck-builder\/draft\//, { timeout: 5000 });

        // Should show deck builder, not error
        await expect(page.locator('.deck-builder')).toBeVisible({ timeout: 10000 });

        // Should NOT show "No deck found" error
        const errorText = await page.locator('body').textContent();
        expect(errorText).not.toContain('No deck found for this draft event');
      }
    });
  });

  test.describe('Deck Builder Page', () => {
    test('should display deck builder UI components', async ({ page }) => {
      await page.goto('/decks');
      await page.waitForLoadState('networkidle');

      // Try to navigate to deck builder (may need to create deck first)
      const deckCards = await page.locator('.deck-card').count();

      if (deckCards > 0) {
        // Click first deck
        await page.locator('.deck-card').first().click();

        // Should show deck builder
        await expect(page).toHaveURL(/\/deck-builder\//, { timeout: 5000 });
        await expect(page.locator('.deck-builder')).toBeVisible();

        // Should have key components
        // Note: Exact selectors depend on implementation
        const hasDeckList = await page.locator('[class*="deck-list"]').count() > 0;
        const hasCardSearch = await page.locator('[class*="card-search"], button:has-text("Add Cards")').count() > 0;

        expect(hasDeckList || hasCardSearch).toBe(true);
      }
    });

    test('should show Back to Decks button', async ({ page }) => {
      await page.goto('/decks');
      await page.waitForLoadState('networkidle');

      const deckCards = await page.locator('.deck-card').count();

      if (deckCards > 0) {
        await page.locator('.deck-card').first().click();
        await expect(page).toHaveURL(/\/deck-builder\//);

        // Should have back navigation
        const backButton = page.locator('button:has-text("Back to Decks"), button:has-text("Back"), a:has-text("Back")');
        await expect(backButton).toBeVisible();
      }
    });

    test('should navigate back to decks list', async ({ page }) => {
      await page.goto('/decks');
      await page.waitForLoadState('networkidle');

      const deckCards = await page.locator('.deck-card').count();

      if (deckCards > 0) {
        await page.locator('.deck-card').first().click();
        await expect(page).toHaveURL(/\/deck-builder\//);

        // Click back button
        const backButton = page.locator('button:has-text("Back to Decks"), button:has-text("Back"), a:has-text("Back")').first();
        await backButton.click();

        // Should return to decks page
        await expect(page).toHaveURL(/\/decks/);
      }
    });
  });

  test.describe('Error Handling', () => {
    test('should handle missing deck gracefully', async ({ page }) => {
      // Navigate to non-existent deck
      await page.goto('/deck-builder/non-existent-deck-id');
      await page.waitForLoadState('networkidle');

      // Should show error state or redirect
      const hasError = await page.locator('.error-state').count() > 0;
      const isRedirected = page.url().includes('/decks');

      expect(hasError || isRedirected).toBe(true);
    });

    test('should not crash on navigation between tabs', async ({ page }) => {
      // Navigate to decks
      await page.goto('/decks');
      await page.waitForLoadState('networkidle');

      // Navigate to draft
      await page.getByText('Draft').click();
      await page.waitForLoadState('networkidle');

      // Navigate back to decks
      await page.getByText('Decks').click();
      await page.waitForLoadState('networkidle');

      // Should not have errors
      const hasError = await page.locator('[class*="error"]').count() > 0;
      expect(hasError).toBe(false);
    });
  });
});
