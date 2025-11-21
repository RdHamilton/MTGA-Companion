import { test, expect } from '@playwright/test';

/**
 * Smoke test to verify Playwright E2E setup is working correctly
 *
 * Prerequisites:
 * - Run `wails dev` in the project root before running these tests
 * - The app should be accessible at http://localhost:34115
 */
test.describe('Smoke Tests', () => {
  test('should load the application', async ({ page }) => {
    // Navigate to the app
    await page.goto('/');

    // Wait for the app to be ready - check for the main layout
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });

    // Verify the navigation tabs are present
    await expect(page.getByText('Match History')).toBeVisible();
    await expect(page.getByText('Draft')).toBeVisible();
    await expect(page.getByText('Charts')).toBeVisible();
    await expect(page.getByText('Settings')).toBeVisible();
  });

  test('should have a working navigation', async ({ page }) => {
    await page.goto('/');

    // Wait for app to be ready
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });

    // Click on Draft tab
    await page.getByText('Draft').click();

    // Verify navigation worked by checking URL or content
    await expect(page).toHaveURL(/\/draft/);
  });

  test('should display footer statistics', async ({ page }) => {
    await page.goto('/');

    // Wait for app to be ready
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });

    // Check for footer presence
    const footer = page.locator('.app-footer');
    await expect(footer).toBeVisible();
  });
});
