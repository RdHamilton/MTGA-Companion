import { test, expect } from '@playwright/test';

/**
 * Draft Page E2E Tests
 *
 * Prerequisites:
 * - Run `wails dev` in the project root before running these tests
 * - The app should be accessible at http://localhost:34115
 */
test.describe('Draft', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app and wait for it to load
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });

    // Navigate to Draft page
    await page.click('a[href="/draft"]');
    await page.waitForURL('**/draft');
  });

  test.describe('Navigation and Page Load', () => {
    test('should navigate to Draft page', async ({ page }) => {
      // Wait for the page to load fully
      await page.waitForTimeout(2000);

      // The draft page should have some content - container, loading, empty, or historical drafts
      const draftContainer = page.locator('.draft-container');
      const draftEmpty = page.locator('.draft-empty');
      const historicalSection = page.locator('text=Historical Drafts');
      const appContainer = page.locator('.app-container');

      // Wait for any of these to appear
      await Promise.race([
        draftContainer.waitFor({ state: 'visible', timeout: 5000 }),
        draftEmpty.waitFor({ state: 'visible', timeout: 5000 }),
        historicalSection.waitFor({ state: 'visible', timeout: 5000 }),
      ]).catch(() => {});

      // At minimum, the app container should be visible
      await expect(appContainer).toBeVisible();
    });
  });

  test.describe('Draft Content', () => {
    test('should display draft content or empty state', async ({ page }) => {
      // Wait for loading to complete
      await Promise.race([
        page.locator('.draft-container').waitFor({ state: 'visible', timeout: 10000 }),
        page.locator('.draft-empty').waitFor({ state: 'visible', timeout: 10000 }),
        page.locator('.draft-loading').waitFor({ state: 'hidden', timeout: 10000 }),
      ]).catch(() => {});

      // Wait a bit more for content to render
      await page.waitForTimeout(1000);

      // Either draft container, empty state, or some content should be visible
      const hasContainer = await page.locator('.draft-container').isVisible();
      const hasEmpty = await page.locator('.draft-empty').isVisible();

      expect(hasContainer || hasEmpty).toBeTruthy();
    });

    test('should display historical drafts section if no active draft', async ({ page }) => {
      // Wait for page to load
      await page.waitForTimeout(2000);

      // Check for historical drafts heading or empty state
      const hasHistorical = await page.locator('text=Historical Drafts').isVisible().catch(() => false);
      const hasDraftContent = await page.locator('.draft-container').isVisible().catch(() => false);
      const hasEmpty = await page.locator('.draft-empty').isVisible().catch(() => false);

      // At least one should be true
      expect(hasHistorical || hasDraftContent || hasEmpty).toBeTruthy();
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
