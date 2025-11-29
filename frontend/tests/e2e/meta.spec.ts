import { test, expect } from '@playwright/test';

/**
 * Meta Page E2E Tests
 *
 * Prerequisites:
 * - Run `wails dev` in the project root before running these tests
 * - The app should be accessible at http://localhost:34115
 */
test.describe('Meta', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app and wait for it to load
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });

    // Navigate to Meta page
    await page.click('a[href="/meta"]');
    await page.waitForURL('**/meta');
  });

  test.describe('Navigation and Page Load', () => {
    test('should navigate to Meta page', async ({ page }) => {
      const metaPage = page.locator('.meta-page');
      await expect(metaPage).toBeVisible({ timeout: 10000 });
    });

    test('should display page title', async ({ page }) => {
      await expect(page.locator('.meta-title h1')).toContainText('Meta');
    });
  });

  test.describe('Meta Header', () => {
    test('should display meta header', async ({ page }) => {
      const header = page.locator('.meta-header');
      await expect(header).toBeVisible({ timeout: 10000 });
    });

    test('should display format selector', async ({ page }) => {
      const formatSelect = page.locator('.format-select');
      await expect(formatSelect).toBeVisible({ timeout: 5000 });
    });

    test('should have refresh button', async ({ page }) => {
      const refreshButton = page.locator('.refresh-button');
      await expect(refreshButton).toBeVisible({ timeout: 5000 });
    });
  });

  test.describe('Format Selection', () => {
    test('should have format options', async ({ page }) => {
      const formatSelect = page.locator('.format-select');
      await expect(formatSelect).toBeVisible();

      const options = await formatSelect.locator('option').allTextContents();
      expect(options.length).toBeGreaterThan(0);

      // Common formats should be available
      const hasStandard = options.some((opt) => opt.toLowerCase().includes('standard'));
      const hasHistoric = options.some((opt) => opt.toLowerCase().includes('historic'));

      expect(hasStandard || hasHistoric).toBeTruthy();
    });

    test('should allow changing format', async ({ page }) => {
      const formatSelect = page.locator('.format-select');
      await expect(formatSelect).toBeVisible();

      // Select a different format
      await formatSelect.selectOption({ index: 1 });

      // Wait for content to update
      await page.waitForTimeout(1000);

      // Page should still be visible
      const metaPage = page.locator('.meta-page');
      await expect(metaPage).toBeVisible();
    });
  });

  test.describe('Meta Content', () => {
    test('should display meta content or loading state', async ({ page }) => {
      // Wait for loading to complete
      await page.waitForTimeout(3000);

      const metaPage = page.locator('.meta-page');
      await expect(metaPage).toBeVisible();

      // Should have some content (decks, archetypes, or a message)
      const hasContent = await page.locator('.meta-page').textContent();
      expect(hasContent?.length).toBeGreaterThan(0);
    });
  });

  test.describe('Loading State', () => {
    test('should show loading indicator while fetching data', async ({ page }) => {
      // Initial page load might show loading
      const metaPage = page.locator('.meta-page');
      await expect(metaPage).toBeVisible({ timeout: 10000 });
    });

    test('should handle refresh button click', async ({ page }) => {
      // Wait for page to load
      await page.waitForTimeout(2000);

      const refreshButton = page.locator('.refresh-button');
      await expect(refreshButton).toBeVisible();

      // Click refresh
      await refreshButton.click();

      // Wait for refresh to complete
      await page.waitForTimeout(2000);

      // Page should still be visible
      const metaPage = page.locator('.meta-page');
      await expect(metaPage).toBeVisible();
    });
  });
});
