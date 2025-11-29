import { test, expect } from '@playwright/test';

/**
 * Match History E2E Tests
 *
 * Prerequisites:
 * - Run `wails dev` in the project root before running these tests
 * - The app should be accessible at http://localhost:34115
 */
test.describe('Match History', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app and wait for it to load
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });
  });

  test.describe('Navigation and Page Load', () => {
    test('should display Match History as the default page', async ({ page }) => {
      // Match History should be the default/home page
      await expect(page.locator('h1.page-title')).toHaveText('Match History');
    });

    test('should display page header with title', async ({ page }) => {
      const header = page.locator('.match-history-header');
      await expect(header).toBeVisible();
      await expect(header.locator('h1')).toHaveText('Match History');
    });
  });

  test.describe('Filter Controls', () => {
    test('should display date range filter', async ({ page }) => {
      const filterRow = page.locator('.filter-row');
      await expect(filterRow).toBeVisible();

      // Check for date range dropdown
      const dateRangeSelect = filterRow.locator('select').first();
      await expect(dateRangeSelect).toBeVisible();

      // Verify options exist
      const options = await dateRangeSelect.locator('option').allTextContents();
      expect(options).toContain('Last 7 Days');
      expect(options).toContain('Last 30 Days');
      expect(options).toContain('All Time');
      expect(options).toContain('Custom Range');
    });

    test('should display card format filter', async ({ page }) => {
      const cardFormatLabel = page.locator('.filter-label').filter({ hasText: 'Card Format' });
      await expect(cardFormatLabel).toBeVisible();

      // Find the associated select
      const filterGroup = cardFormatLabel.locator('..');
      const select = filterGroup.locator('select');
      await expect(select).toBeVisible();

      const options = await select.locator('option').allTextContents();
      expect(options).toContain('All Card Formats');
      expect(options).toContain('Standard');
      expect(options).toContain('Historic');
    });

    test('should display queue type filter', async ({ page }) => {
      const queueTypeLabel = page.locator('.filter-label').filter({ hasText: 'Queue Type' });
      await expect(queueTypeLabel).toBeVisible();

      const filterGroup = queueTypeLabel.locator('..');
      const select = filterGroup.locator('select');
      await expect(select).toBeVisible();

      const options = await select.locator('option').allTextContents();
      expect(options).toContain('All Queues');
      expect(options).toContain('Ranked');
      expect(options).toContain('Play Queue');
    });

    test('should display result filter', async ({ page }) => {
      const resultLabel = page.locator('.filter-label').filter({ hasText: 'Result' });
      await expect(resultLabel).toBeVisible();

      const filterGroup = resultLabel.locator('..');
      const select = filterGroup.locator('select');
      await expect(select).toBeVisible();

      const options = await select.locator('option').allTextContents();
      expect(options).toContain('All Results');
      expect(options).toContain('Wins Only');
      expect(options).toContain('Losses Only');
    });

    test('should show custom date pickers when Custom Range is selected', async ({ page }) => {
      // Select Custom Range
      const dateRangeSelect = page.locator('.filter-group').first().locator('select');
      await dateRangeSelect.selectOption('custom');

      // Custom date pickers should now be visible
      const startDateInput = page.locator('input[type="date"]').first();
      const endDateInput = page.locator('input[type="date"]').last();

      await expect(startDateInput).toBeVisible();
      await expect(endDateInput).toBeVisible();
    });
  });

  test.describe('Content State', () => {
    test('should display either matches table or empty state after loading', async ({ page }) => {
      // Wait for either table or empty state to appear (one must be visible after loading)
      // Using Promise.race to wait for whichever appears first
      const table = page.locator('.match-history-table-container');
      const emptyState = page.locator('.empty-state');

      // Wait for loading to complete by waiting for either content type to appear
      await Promise.race([
        table.waitFor({ state: 'visible', timeout: 10000 }),
        emptyState.waitFor({ state: 'visible', timeout: 10000 }),
      ]).catch(() => {
        // One of them should appear
      });

      // Now check which one is visible
      const hasTable = await table.isVisible();
      const hasEmptyState = await emptyState.isVisible();

      // One of these should be true
      expect(hasTable || hasEmptyState).toBeTruthy();

      if (hasEmptyState) {
        await expect(emptyState.locator('.empty-state-title')).toBeVisible();
        await expect(emptyState.locator('.empty-state-message')).toBeVisible();
      }

      if (hasTable) {
        // Verify the table has expected structure
        await expect(table.locator('table')).toBeVisible();
      }
    });
  });

  test.describe('Match Table', () => {
    test('should display table headers when matches exist', async ({ page }) => {
      const table = page.locator('.match-history-table-container table');
      const hasTable = await table.isVisible().catch(() => false);

      if (hasTable) {
        const headers = table.locator('thead th');
        const headerTexts = await headers.allTextContents();

        // Check for expected column headers (they include sort icons)
        expect(headerTexts.some((h) => h.includes('Time'))).toBeTruthy();
        expect(headerTexts.some((h) => h.includes('Result'))).toBeTruthy();
        expect(headerTexts.some((h) => h.includes('Format'))).toBeTruthy();
        expect(headerTexts.some((h) => h.includes('Event'))).toBeTruthy();
        expect(headerTexts.some((h) => h.includes('Score'))).toBeTruthy();
        expect(headerTexts.some((h) => h.includes('Opponent'))).toBeTruthy();
      }
    });

    test('should have sortable column headers', async ({ page }) => {
      const table = page.locator('.match-history-table-container table');
      const hasTable = await table.isVisible().catch(() => false);

      if (hasTable) {
        // Time column should be sortable (has click handler)
        const timeHeader = table.locator('thead th').first();
        await expect(timeHeader).toHaveCSS('cursor', 'pointer');
      }
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

  test.describe('Match Count', () => {
    test('should display match count when matches exist', async ({ page }) => {
      const hasTable = await page.locator('.match-history-table-container').isVisible().catch(() => false);

      if (hasTable) {
        const matchCount = page.locator('.match-count');
        await expect(matchCount).toBeVisible();
        // Should contain text like "Showing X of Y matches"
        await expect(matchCount).toContainText('Showing');
      }
    });
  });
});
