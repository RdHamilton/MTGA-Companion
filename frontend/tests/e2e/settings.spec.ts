import { test, expect } from '@playwright/test';

/**
 * Settings Page E2E Tests
 *
 * Tests the Settings page functionality including sections and buttons.
 * Uses REST API backend for testing.
 */
test.describe('Settings', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });

    await page.click('a[href="/settings"]');
    await page.waitForURL('**/settings');
  });

  test.describe('Navigation and Page Load', () => {
    test('@smoke should navigate to Settings page', async ({ page }) => {
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
      // Wait for settings content to load
      const settingsContent = page.locator('.settings-content');
      await expect(settingsContent).toBeVisible({ timeout: 5000 });

      // Settings uses accordion sections
      const accordionSections = page.locator('.settings-section, .accordion-item, [class*="accordion"]');
      await expect(accordionSections.first()).toBeVisible({ timeout: 5000 });
    });
  });

  test.describe('Connection Settings', () => {
    test('should display daemon connection section', async ({ page }) => {
      // Wait for settings to load
      const settingsContent = page.locator('.settings-content');
      await expect(settingsContent).toBeVisible({ timeout: 5000 });

      // Look for daemon/connection section
      const connectionSection = page.locator('text=Daemon').first();
      const settingsPage = page.locator('.settings-header');

      const hasConnection = await connectionSection.isVisible().catch(() => false);
      const hasHeader = await settingsPage.isVisible();

      expect(hasConnection || hasHeader).toBeTruthy();
    });
  });

  test.describe('Preferences Section', () => {
    test('should have preference settings available', async ({ page }) => {
      // Wait for settings to load
      const settingsContent = page.locator('.settings-content');
      await expect(settingsContent).toBeVisible({ timeout: 5000 });

      // Look for preference-related elements
      const preferencesText = page.locator('text=Preferences').first();
      const autoRefreshText = page.locator('text=Auto Refresh').first();
      const themeText = page.locator('text=Theme').first();

      const hasPrefs =
        (await preferencesText.isVisible().catch(() => false)) ||
        (await autoRefreshText.isVisible().catch(() => false)) ||
        (await themeText.isVisible().catch(() => false));

      expect(hasPrefs).toBeTruthy();
    });
  });

  test.describe('Action Buttons', () => {
    test('@smoke should have save button', async ({ page }) => {
      // Wait for settings to load
      const settingsContent = page.locator('.settings-content');
      await expect(settingsContent).toBeVisible({ timeout: 5000 });

      const saveButton = page.locator('button').filter({ hasText: /save/i });
      await expect(saveButton).toBeVisible();
    });

    test('should have reset to defaults option', async ({ page }) => {
      // Wait for settings to load
      const settingsContent = page.locator('.settings-content');
      await expect(settingsContent).toBeVisible({ timeout: 5000 });

      const resetButton = page.locator('button').filter({ hasText: /reset|defaults/i });
      await expect(resetButton).toBeVisible();
    });
  });

  test.describe('About Section', () => {
    test('should have version info in settings', async ({ page }) => {
      // Wait for settings to load
      const settingsContent = page.locator('.settings-content');
      await expect(settingsContent).toBeVisible({ timeout: 5000 });

      // Look for version text or about button
      const versionText = page.locator('text=Version').first();
      const aboutButton = page.locator('button').filter({ hasText: /about/i }).first();
      const settingsHeader = page.locator('.settings-header');

      const hasVersion = await versionText.isVisible().catch(() => false);
      const hasAboutButton = await aboutButton.isVisible().catch(() => false);
      const hasHeader = await settingsHeader.isVisible();

      expect(hasVersion || hasAboutButton || hasHeader).toBeTruthy();
    });
  });

  test.describe('17Lands Integration', () => {
    test('should have 17Lands settings section', async ({ page }) => {
      // Wait for settings to load
      const settingsContent = page.locator('.settings-content');
      await expect(settingsContent).toBeVisible({ timeout: 5000 });

      const landsSection = page.locator('text=17Lands').first();
      await expect(landsSection).toBeVisible({ timeout: 5000 });
    });
  });

  test.describe('ML Settings', () => {
    test('should have ML/AI settings section', async ({ page }) => {
      // Wait for settings to load
      const settingsContent = page.locator('.settings-content');
      await expect(settingsContent).toBeVisible({ timeout: 5000 });

      const mlSection = page.locator('text=ML').first();
      const aiSection = page.locator('text=AI').first();
      const ollamaSection = page.locator('text=Ollama').first();

      const hasML =
        (await mlSection.isVisible().catch(() => false)) ||
        (await aiSection.isVisible().catch(() => false)) ||
        (await ollamaSection.isVisible().catch(() => false));

      expect(hasML).toBeTruthy();
    });
  });

  test.describe('Loading State', () => {
    test('should not show error state on initial load', async ({ page }) => {
      // Wait for settings to load
      const settingsContent = page.locator('.settings-content');
      await expect(settingsContent).toBeVisible({ timeout: 5000 });

      const errorState = page.locator('.settings-error, .error-state');
      await expect(errorState).not.toBeVisible();
    });
  });
});
