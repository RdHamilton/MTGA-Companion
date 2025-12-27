import { test, expect } from '@playwright/test';

/**
 * Charts Pages E2E Tests
 *
 * Tests the Charts pages including Win Rate Trend and sub-navigation.
 * Uses REST API backend for testing.
 */
test.describe('Charts', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });

    await page.click('a.tab[href="/charts/win-rate-trend"]');
    await page.waitForURL('**/charts/**');
  });

  test.describe('Win Rate Trend', () => {
    test('@smoke should navigate to Win Rate Trend page', async ({ page }) => {
      const activeSubTab = page.locator('.sub-tab-bar a.active');
      await expect(activeSubTab).toContainText(/Win Rate Trend/i);
    });

    test('should display chart content after loading', async ({ page }) => {
      // Wait for chart filter to appear
      const dateRangeFilter = page.locator('select').first();
      await expect(dateRangeFilter).toBeVisible({ timeout: 10000 });
    });
  });

  test.describe('Sub-navigation', () => {
    test('should display sub-navigation bar', async ({ page }) => {
      const subTabBar = page.locator('.sub-tab-bar');
      await expect(subTabBar).toBeVisible();
    });

    test('should have all chart sub-tabs', async ({ page }) => {
      const subTabBar = page.locator('.sub-tab-bar');
      await expect(subTabBar).toBeVisible();

      await expect(subTabBar.locator('a[href="/charts/win-rate-trend"]')).toBeVisible();
      await expect(subTabBar.locator('a[href="/charts/deck-performance"]')).toBeVisible();
      await expect(subTabBar.locator('a[href="/charts/rank-progression"]')).toBeVisible();
      await expect(subTabBar.locator('a[href="/charts/format-distribution"]')).toBeVisible();
      await expect(subTabBar.locator('a[href="/charts/result-breakdown"]')).toBeVisible();
    });
  });

  test.describe('Deck Performance', () => {
    test('should navigate to Deck Performance page via sub-nav', async ({ page }) => {
      await page.click('.sub-tab-bar a[href="/charts/deck-performance"]');
      await page.waitForURL('**/charts/deck-performance');

      const activeSubTab = page.locator('.sub-tab-bar a.active');
      await expect(activeSubTab).toContainText(/Deck Performance/i);
    });
  });

  test.describe('Rank Progression', () => {
    test('should navigate to Rank Progression page via sub-nav', async ({ page }) => {
      await page.click('.sub-tab-bar a[href="/charts/rank-progression"]');
      await page.waitForURL('**/charts/rank-progression');

      const activeSubTab = page.locator('.sub-tab-bar a.active');
      await expect(activeSubTab).toContainText(/Rank Progression/i);
    });
  });

  test.describe('Format Distribution', () => {
    test('should navigate to Format Distribution page via sub-nav', async ({ page }) => {
      await page.click('.sub-tab-bar a[href="/charts/format-distribution"]');
      await page.waitForURL('**/charts/format-distribution');

      const activeSubTab = page.locator('.sub-tab-bar a.active');
      await expect(activeSubTab).toContainText(/Format Distribution/i);
    });
  });

  test.describe('Result Breakdown', () => {
    test('should navigate to Result Breakdown page via sub-nav', async ({ page }) => {
      await page.click('.sub-tab-bar a[href="/charts/result-breakdown"]');
      await page.waitForURL('**/charts/result-breakdown');

      const activeSubTab = page.locator('.sub-tab-bar a.active');
      await expect(activeSubTab).toContainText(/Result Breakdown/i);
    });
  });

  test.describe('Navigation Between Charts', () => {
    test('should allow navigation between chart pages via sub-nav', async ({ page }) => {
      await page.click('.sub-tab-bar a[href="/charts/deck-performance"]');
      await page.waitForURL('**/charts/deck-performance');
      await expect(page.locator('.sub-tab-bar a.active')).toContainText(/Deck Performance/i);

      await page.click('.sub-tab-bar a[href="/charts/format-distribution"]');
      await page.waitForURL('**/charts/format-distribution');
      await expect(page.locator('.sub-tab-bar a.active')).toContainText(/Format Distribution/i);

      await page.click('.sub-tab-bar a[href="/charts/result-breakdown"]');
      await page.waitForURL('**/charts/result-breakdown');
      await expect(page.locator('.sub-tab-bar a.active')).toContainText(/Result Breakdown/i);
    });
  });
});
