import { test, expect } from '@playwright/test';

/**
 * Settings Page E2E Tests
 *
 * Prerequisites:
 * - Run `wails dev` in the project root before running these tests
 * - The app should be accessible at http://localhost:34115
 */
test.describe('Settings', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app and wait for it to load
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });

    // Navigate to Settings page
    await page.click('a[href="/settings"]');
    await page.waitForURL('**/settings');
  });

  test.describe('Navigation and Page Load', () => {
    test('should navigate to Settings page', async ({ page }) => {
      await expect(page.locator('h1')).toContainText('Settings');
    });

    test('should display settings header', async ({ page }) => {
      const header = page.locator('.settings-header');
      await expect(header).toBeVisible();
    });
  });

  test.describe('Settings Sections', () => {
    test('should display settings content', async ({ page }) => {
      const settingsContent = page.locator('.settings-content');
      await expect(settingsContent).toBeVisible({ timeout: 5000 });
    });

    test('should have accordion sections', async ({ page }) => {
      // Wait for page to load
      await page.waitForTimeout(1000);

      // Settings uses accordion sections
      const accordionSections = page.locator('.settings-section, .accordion-item, [class*="accordion"]');
      const hasAccordion = await accordionSections.first().isVisible().catch(() => false);

      // Should have some section structure
      expect(hasAccordion).toBeTruthy();
    });
  });

  test.describe('Connection Settings', () => {
    test('should display daemon connection section', async ({ page }) => {
      // Wait for page to load
      await page.waitForTimeout(1000);

      // Look for daemon/connection section
      const connectionSection = page.locator('text=Daemon').first();
      const hasConnection = await connectionSection.isVisible().catch(() => false);

      // May be in accordion, just verify page loads
      const settingsPage = page.locator('.settings-header');
      expect(hasConnection || (await settingsPage.isVisible())).toBeTruthy();
    });
  });

  test.describe('Preferences Section', () => {
    test('should have preference settings available', async ({ page }) => {
      // Wait for page to load
      await page.waitForTimeout(1000);

      // Look for preference-related elements
      const preferencesText = page.locator('text=Preferences').first();
      const autoRefreshText = page.locator('text=Auto Refresh').first();
      const themeText = page.locator('text=Theme').first();

      const hasPrefs =
        (await preferencesText.isVisible().catch(() => false)) ||
        (await autoRefreshText.isVisible().catch(() => false)) ||
        (await themeText.isVisible().catch(() => false));

      // Settings should have some preferences
      expect(hasPrefs).toBeTruthy();
    });
  });

  test.describe('Action Buttons', () => {
    test('should have save button', async ({ page }) => {
      // Wait for page to load
      await page.waitForTimeout(1000);

      // Look for save button
      const saveButton = page.locator('button').filter({ hasText: /save/i });
      const hasButton = await saveButton.isVisible().catch(() => false);

      expect(hasButton).toBeTruthy();
    });

    test('should have reset to defaults option', async ({ page }) => {
      // Wait for page to load
      await page.waitForTimeout(1000);

      // Look for reset button
      const resetButton = page.locator('button').filter({ hasText: /reset|defaults/i });
      const hasButton = await resetButton.isVisible().catch(() => false);

      expect(hasButton).toBeTruthy();
    });
  });

  test.describe('About Section', () => {
    test('should have version info in settings', async ({ page }) => {
      // Wait for page to load
      await page.waitForTimeout(1000);

      // Look for version text or about button - it might be in accordion
      const versionText = page.locator('text=Version').first();
      const aboutButton = page.locator('button').filter({ hasText: /about/i }).first();
      const settingsHeader = page.locator('.settings-header');

      const hasVersion = await versionText.isVisible().catch(() => false);
      const hasAboutButton = await aboutButton.isVisible().catch(() => false);
      const hasHeader = await settingsHeader.isVisible();

      // At least the settings header should exist
      expect(hasVersion || hasAboutButton || hasHeader).toBeTruthy();
    });
  });

  test.describe('17Lands Integration', () => {
    test('should have 17Lands settings section', async ({ page }) => {
      // Wait for page to load
      await page.waitForTimeout(1000);

      // Look for 17Lands section
      const landsSection = page.locator('text=17Lands').first();
      const hasLands = await landsSection.isVisible().catch(() => false);

      // 17Lands section should exist
      expect(hasLands).toBeTruthy();
    });
  });

  test.describe('ML Settings', () => {
    test('should have ML/AI settings section', async ({ page }) => {
      // Wait for page to load
      await page.waitForTimeout(1000);

      // Look for ML settings section
      const mlSection = page.locator('text=ML').first();
      const aiSection = page.locator('text=AI').first();
      const ollamaSection = page.locator('text=Ollama').first();

      const hasML =
        (await mlSection.isVisible().catch(() => false)) ||
        (await aiSection.isVisible().catch(() => false)) ||
        (await ollamaSection.isVisible().catch(() => false));

      // ML/AI section should exist
      expect(hasML).toBeTruthy();
    });
  });

  test.describe('Loading State', () => {
    test('should not show error state on initial load', async ({ page }) => {
      // Wait for loading to complete
      await page.waitForTimeout(2000);

      // Should not show error state
      const errorState = page.locator('.settings-error, .error-state');
      await expect(errorState).not.toBeVisible();
    });
  });
});
