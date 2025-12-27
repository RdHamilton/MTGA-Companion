import { test, expect } from '@playwright/test';

/**
 * Draft Page E2E Tests
 *
 * Tests the Draft page functionality including navigation and content display.
 * Uses REST API backend for testing.
 */
test.describe('Draft', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });

    await page.click('a[href="/draft"]');
    await page.waitForURL('**/draft');
  });

  test.describe('Navigation and Page Load', () => {
    test('@smoke should navigate to Draft page', async ({ page }) => {
      // Wait for page content to load
      const appContainer = page.locator('.app-container');
      await expect(appContainer).toBeVisible();

      // Verify we're on the draft page
      const url = page.url();
      expect(url).toContain('/draft');
    });
  });

  test.describe('Draft Content', () => {
    test('should display draft content or empty state', async ({ page }) => {
      const draftContainer = page.locator('.draft-container');
      const draftEmpty = page.locator('.draft-empty');

      // Wait for either content type to appear
      await expect(draftContainer.or(draftEmpty)).toBeVisible({ timeout: 10000 });

      const hasContainer = await draftContainer.isVisible();
      const hasEmpty = await draftEmpty.isVisible();

      expect(hasContainer || hasEmpty).toBeTruthy();
    });

    test('should display historical drafts section if no active draft', async ({ page }) => {
      const historicalSection = page.locator('text=Historical Drafts');
      const draftContainer = page.locator('.draft-container');
      const draftEmpty = page.locator('.draft-empty');

      // Wait for content to load
      await expect(draftContainer.or(draftEmpty).or(historicalSection)).toBeVisible({ timeout: 10000 });

      const hasHistorical = await historicalSection.isVisible();
      const hasDraftContent = await draftContainer.isVisible();
      const hasEmpty = await draftEmpty.isVisible();

      expect(hasHistorical || hasDraftContent || hasEmpty).toBeTruthy();
    });
  });

  test.describe('Loading State', () => {
    test('should not show error state on initial load', async ({ page }) => {
      // Wait for content to load
      const content = page.locator('.draft-container, .draft-empty');
      await expect(content.first()).toBeVisible({ timeout: 10000 });

      const errorState = page.locator('.error-state');
      await expect(errorState).not.toBeVisible();
    });
  });
});
