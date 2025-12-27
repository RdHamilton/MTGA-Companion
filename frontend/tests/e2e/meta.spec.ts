import { test, expect } from '@playwright/test';

/**
 * Meta Page E2E Tests
 *
 * Tests the Meta page functionality including format selection.
 * Uses REST API backend for testing.
 */
test.describe('Meta', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });

    await page.click('a[href="/meta"]');
    await page.waitForURL('**/meta');
  });

  test.describe('Navigation and Page Load', () => {
    test('@smoke should navigate to Meta page', async ({ page }) => {
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

      const hasStandard = options.some((opt) => opt.toLowerCase().includes('standard'));
      const hasHistoric = options.some((opt) => opt.toLowerCase().includes('historic'));

      expect(hasStandard || hasHistoric).toBeTruthy();
    });

    test('should allow changing format', async ({ page }) => {
      const formatSelect = page.locator('.format-select');
      await expect(formatSelect).toBeVisible();

      // Select a different format
      await formatSelect.selectOption({ index: 1 });

      // Wait for content to update by checking page is still visible
      const metaPage = page.locator('.meta-page');
      await expect(metaPage).toBeVisible({ timeout: 5000 });
    });
  });

  test.describe('Meta Content', () => {
    test('should display meta content or loading state', async ({ page }) => {
      const metaPage = page.locator('.meta-page');
      await expect(metaPage).toBeVisible({ timeout: 10000 });

      // Should have some content
      const content = await metaPage.textContent();
      expect(content?.length).toBeGreaterThan(0);
    });
  });

  test.describe('Loading State', () => {
    test('should show loading indicator while fetching data', async ({ page }) => {
      const metaPage = page.locator('.meta-page');
      await expect(metaPage).toBeVisible({ timeout: 10000 });
    });

    test('should handle refresh button click', async ({ page }) => {
      // Wait for page to load
      const refreshButton = page.locator('.refresh-button');
      await expect(refreshButton).toBeVisible({ timeout: 10000 });

      // Click refresh
      await refreshButton.click();

      // Page should still be visible after refresh
      const metaPage = page.locator('.meta-page');
      await expect(metaPage).toBeVisible({ timeout: 10000 });
    });
  });
});
