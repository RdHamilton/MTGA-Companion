import { test, expect } from '@playwright/test';

/**
 * Charts Pages E2E Tests
 *
 * Prerequisites:
 * - Run `wails dev` in the project root before running these tests
 * - The app should be accessible at http://localhost:34115
 *
 * Note: Charts are accessed via the main "Charts" tab which goes to win-rate-trend,
 * then sub-navigation for other chart pages appears.
 */
test.describe('Charts', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app and wait for it to load
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });

    // Navigate to Charts section (main tab goes to win-rate-trend)
    await page.click('a.tab[href="/charts/win-rate-trend"]');
    await page.waitForURL('**/charts/**');
  });

  test.describe('Win Rate Trend', () => {
    test('should navigate to Win Rate Trend page', async ({ page }) => {
      // We're already on win-rate-trend from beforeEach
      // The active sub-tab should indicate we're on the right page
      const activeSubTab = page.locator('.sub-tab-bar a.active');
      await expect(activeSubTab).toContainText(/Win Rate Trend/i);
    });

    test('should display chart content after loading', async ({ page }) => {
      // Wait for chart to render
      await page.waitForTimeout(2000);

      // Page should have filter controls for the chart
      const dateRangeFilter = page.locator('select').first();
      await expect(dateRangeFilter).toBeVisible();
    });
  });

  test.describe('Sub-navigation', () => {
    test('should display sub-navigation bar', async ({ page }) => {
      // Sub-navigation should be visible when on charts page
      const subTabBar = page.locator('.sub-tab-bar');
      await expect(subTabBar).toBeVisible();
    });

    test('should have all chart sub-tabs', async ({ page }) => {
      const subTabBar = page.locator('.sub-tab-bar');
      await expect(subTabBar).toBeVisible();

      // Check for all sub-tabs
      await expect(subTabBar.locator('a[href="/charts/win-rate-trend"]')).toBeVisible();
      await expect(subTabBar.locator('a[href="/charts/deck-performance"]')).toBeVisible();
      await expect(subTabBar.locator('a[href="/charts/rank-progression"]')).toBeVisible();
      await expect(subTabBar.locator('a[href="/charts/format-distribution"]')).toBeVisible();
      await expect(subTabBar.locator('a[href="/charts/result-breakdown"]')).toBeVisible();
    });
  });

  test.describe('Deck Performance', () => {
    test('should navigate to Deck Performance page via sub-nav', async ({ page }) => {
      // Click sub-tab
      await page.click('.sub-tab-bar a[href="/charts/deck-performance"]');
      await page.waitForURL('**/charts/deck-performance');

      // Verify active sub-tab
      const activeSubTab = page.locator('.sub-tab-bar a.active');
      await expect(activeSubTab).toContainText(/Deck Performance/i);
    });
  });

  test.describe('Rank Progression', () => {
    test('should navigate to Rank Progression page via sub-nav', async ({ page }) => {
      // Click sub-tab
      await page.click('.sub-tab-bar a[href="/charts/rank-progression"]');
      await page.waitForURL('**/charts/rank-progression');

      // Verify active sub-tab
      const activeSubTab = page.locator('.sub-tab-bar a.active');
      await expect(activeSubTab).toContainText(/Rank Progression/i);
    });
  });

  test.describe('Format Distribution', () => {
    test('should navigate to Format Distribution page via sub-nav', async ({ page }) => {
      // Click sub-tab
      await page.click('.sub-tab-bar a[href="/charts/format-distribution"]');
      await page.waitForURL('**/charts/format-distribution');

      // Verify active sub-tab
      const activeSubTab = page.locator('.sub-tab-bar a.active');
      await expect(activeSubTab).toContainText(/Format Distribution/i);
    });
  });

  test.describe('Result Breakdown', () => {
    test('should navigate to Result Breakdown page via sub-nav', async ({ page }) => {
      // Click sub-tab
      await page.click('.sub-tab-bar a[href="/charts/result-breakdown"]');
      await page.waitForURL('**/charts/result-breakdown');

      // Verify active sub-tab
      const activeSubTab = page.locator('.sub-tab-bar a.active');
      await expect(activeSubTab).toContainText(/Result Breakdown/i);
    });
  });

  test.describe('Navigation Between Charts', () => {
    test('should allow navigation between chart pages via sub-nav', async ({ page }) => {
      // Navigate to deck performance
      await page.click('.sub-tab-bar a[href="/charts/deck-performance"]');
      await page.waitForURL('**/charts/deck-performance');
      await expect(page.locator('.sub-tab-bar a.active')).toContainText(/Deck Performance/i);

      // Navigate to format distribution
      await page.click('.sub-tab-bar a[href="/charts/format-distribution"]');
      await page.waitForURL('**/charts/format-distribution');
      await expect(page.locator('.sub-tab-bar a.active')).toContainText(/Format Distribution/i);

      // Navigate to result breakdown
      await page.click('.sub-tab-bar a[href="/charts/result-breakdown"]');
      await page.waitForURL('**/charts/result-breakdown');
      await expect(page.locator('.sub-tab-bar a.active')).toContainText(/Result Breakdown/i);
    });
  });
});
