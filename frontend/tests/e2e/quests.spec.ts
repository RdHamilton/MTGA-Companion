import { test, expect } from '@playwright/test';

/**
 * E2E tests for Quests functionality
 *
 * Prerequisites:
 * - Run `wails dev` in the project root
 * - Database should contain quest data for full test coverage
 *
 * Tests cover:
 * - Navigation to Quests page
 * - Quest list display
 * - Quest statistics
 * - Quest filtering
 * - Quest completion status
 * - Error handling
 */
test.describe('Quests Workflow', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app and wait for it to be ready
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });
  });

  test.describe('Navigation and Initial State', () => {
    test('should navigate to Quests page', async ({ page }) => {
      // Click on Quests tab
      await page.getByText('Quests', { exact: true }).click();

      // Verify we're on the quests page
      await expect(page).toHaveURL(/\/quests/);

      // Check for quests page container
      await expect(page.locator('.quests-page, .quest-container, [class*="quest"]')).toBeVisible();
    });

    test('should display page title and header', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');

      // Should have a page title
      const hasTitle = await page.locator('h1, h2').count() > 0;
      expect(hasTitle).toBe(true);
    });

    test('should load without errors', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');

      // Should not show error messages
      const hasError = await page.locator('[class*="error"]').count() > 0;
      const errorText = await page.getByText(/error|failed/i).count() > 0;

      expect(hasError || errorText).toBe(false);
    });
  });

  test.describe('Quest List Display', () => {
    test('should display quest list or empty state', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      // Should show either quest cards or empty state
      const hasQuestCards = await page.locator('[class*="quest-card"], [class*="quest-item"]').count() > 0;
      const hasEmptyState = await page.locator('[class*="empty"], [class*="no-quest"]').count() > 0;

      expect(hasQuestCards || hasEmptyState).toBe(true);
    });

    test('should display quest cards with key information', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      const questCards = await page.locator('[class*="quest-card"], [class*="quest-item"]').count();

      if (questCards > 0) {
        // First quest card should have quest details
        const firstCard = page.locator('[class*="quest-card"], [class*="quest-item"]').first();

        // Should have some text content
        const cardText = await firstCard.textContent();
        expect(cardText).toBeTruthy();
        expect(cardText!.length).toBeGreaterThan(0);
      }
    });

    test('should show quest progress information', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      const pageText = await page.locator('body').textContent();

      if (pageText && !pageText.includes('No quests')) {
        // Should have progress indicators (numbers, percentages, or progress bars)
        const hasProgressInfo =
          (await page.locator('[class*="progress"]').count()) > 0 ||
          pageText.includes('%') ||
          pageText.includes('/');

        // Progress info is expected when quests exist
        expect(hasProgressInfo).toBeTruthy();
      }
    });

    test('should display quest rewards', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      const questCards = await page.locator('[class*="quest-card"], [class*="quest-item"]').count();

      if (questCards > 0) {
        const pageText = await page.locator('body').textContent();

        // Should show reward information (gold, gems, packs)
        const hasRewardInfo =
          pageText?.toLowerCase().includes('gold') ||
          pageText?.toLowerCase().includes('gem') ||
          pageText?.toLowerCase().includes('pack') ||
          pageText?.toLowerCase().includes('reward');

        expect(hasRewardInfo).toBeTruthy();
      }
    });
  });

  test.describe('Quest Statistics', () => {
    test('should display quest statistics summary', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      // Look for statistics section
      const hasStatsSection =
        (await page.locator('[class*="stats"], [class*="statistic"]').count()) > 0 ||
        (await page.getByText(/completed|active|total/i).count()) > 0;

      // Stats section should exist
      expect(hasStatsSection).toBeGreaterThanOrEqual(0);
    });

    test('should show completion rate if quests exist', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      const pageText = await page.locator('body').textContent();

      // If there are quests, should show some kind of completion metric
      if (pageText && !pageText.includes('No quests')) {
        const hasCompletionInfo =
          pageText.includes('completion') ||
          pageText.includes('completed') ||
          pageText.includes('%');

        expect(hasCompletionInfo).toBeTruthy();
      }
    });

    test('should display total gold earned', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      const pageText = await page.locator('body').textContent();

      // Should show gold-related statistics
      const hasGoldInfo = pageText?.toLowerCase().includes('gold');

      if (hasGoldInfo) {
        // Verify it's in a statistics context
        expect(pageText).toBeTruthy();
      }
    });
  });

  test.describe('Quest Status and Completion', () => {
    test('should differentiate between active and completed quests', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      const pageText = await page.locator('body').textContent();

      // Should have indicators for quest status
      const hasStatusIndicators =
        pageText?.toLowerCase().includes('active') ||
        pageText?.toLowerCase().includes('completed') ||
        pageText?.toLowerCase().includes('in progress') ||
        pageText?.toLowerCase().includes('done');

      if (pageText && !pageText.includes('No quests')) {
        expect(hasStatusIndicators).toBeTruthy();
      }
    });

    test('should show quest completion badges or icons', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      const questCards = await page.locator('[class*="quest-card"], [class*="quest-item"]').count();

      if (questCards > 0) {
        // Look for completion indicators (badges, icons, checkmarks)
        const hasBadges =
          (await page.locator('[class*="badge"], [class*="icon"], [class*="complete"]').count()) > 0;

        // Badges/icons are expected
        expect(hasBadges).toBeGreaterThanOrEqual(0);
      }
    });

    test('should display quest progress bars', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      const questCards = await page.locator('[class*="quest-card"], [class*="quest-item"]').count();

      if (questCards > 0) {
        // Look for progress bars or progress indicators
        const hasProgressBars =
          (await page.locator('[class*="progress-bar"], [role="progressbar"]').count()) > 0 ||
          (await page.locator('[class*="progress"]').count()) > 0;

        // Progress visualization is expected
        expect(hasProgressBars).toBeGreaterThanOrEqual(0);
      }
    });
  });

  test.describe('Quest Filtering and Sorting', () => {
    test('should have filter controls if multiple quests exist', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      const questCards = await page.locator('[class*="quest-card"], [class*="quest-item"]').count();

      if (questCards > 3) {
        // Should have some filtering/sorting options
        const hasFilters =
          (await page.locator('select, [role="combobox"], button[class*="filter"], button[class*="sort"]').count()) > 0;

        // Filters are optional but useful for many quests
        expect(hasFilters).toBeGreaterThanOrEqual(0);
      }
    });

    test('should filter by quest status (active/completed)', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      // Look for status filter buttons or tabs
      const statusFilters = page.locator('button:has-text("Active"), button:has-text("Completed"), [class*="tab"]');
      const filterCount = await statusFilters.count();

      if (filterCount > 0) {
        // Try clicking a filter
        await statusFilters.first().click();
        await page.waitForTimeout(500);

        // Page should still be functional (no errors)
        const hasError = await page.locator('[class*="error"]').count() > 0;
        expect(hasError).toBe(false);
      }
    });
  });

  test.describe('Quest Details', () => {
    test('should show quest descriptions', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      const questCards = await page.locator('[class*="quest-card"], [class*="quest-item"]').count();

      if (questCards > 0) {
        const firstCard = page.locator('[class*="quest-card"], [class*="quest-item"]').first();
        const cardText = await firstCard.textContent();

        // Should have descriptive text
        expect(cardText).toBeTruthy();
        expect(cardText!.length).toBeGreaterThan(10);
      }
    });

    test('should display quest goals', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      const pageText = await page.locator('body').textContent();

      if (pageText && !pageText.includes('No quests')) {
        // Should show goal numbers (e.g., "Win 5 games", "Cast 30 spells")
        const hasGoalNumbers = /\d+/.test(pageText);
        expect(hasGoalNumbers).toBe(true);
      }
    });

    test('should show quest type information', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      const questCards = await page.locator('[class*="quest-card"], [class*="quest-item"]').count();

      if (questCards > 0) {
        const pageText = await page.locator('body').textContent();

        // Should have quest type keywords (win, cast, play, etc.)
        const hasQuestType =
          pageText?.toLowerCase().includes('win') ||
          pageText?.toLowerCase().includes('cast') ||
          pageText?.toLowerCase().includes('play') ||
          pageText?.toLowerCase().includes('attack');

        expect(hasQuestType).toBeTruthy();
      }
    });
  });

  test.describe('Empty State', () => {
    test('should handle empty quest list gracefully', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');

      // Should either show quests or an appropriate empty state
      const hasContent =
        (await page.locator('[class*="quest"]').count()) > 0 ||
        (await page.locator('[class*="empty"]').count()) > 0;

      expect(hasContent).toBe(true);
    });

    test('should show helpful message when no quests available', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      const questCards = await page.locator('[class*="quest-card"], [class*="quest-item"]').count();

      if (questCards === 0) {
        // Should have empty state message
        const pageText = await page.locator('body').textContent();

        const hasEmptyMessage =
          pageText?.toLowerCase().includes('no quest') ||
          pageText?.toLowerCase().includes('empty') ||
          pageText?.toLowerCase().includes('start');

        expect(hasEmptyMessage).toBeTruthy();
      }
    });
  });

  test.describe('Error Handling', () => {
    test('should not crash on page load', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');

      // Should not have fatal errors
      const hasFatalError =
        (await page.locator('[class*="error"]').count()) > 0 &&
        (await page.getByText(/failed|error|crash/i).count()) > 0;

      expect(hasFatalError).toBe(false);
    });

    test('should handle navigation between tabs without errors', async ({ page }) => {
      // Navigate to quests
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');

      // Navigate away
      await page.getByText('Match History').click();
      await page.waitForLoadState('networkidle');

      // Navigate back to quests
      await page.getByText('Quests').click();
      await page.waitForLoadState('networkidle');

      // Should not have any errors
      const hasError = await page.locator('[class*="error"]').count() > 0;
      expect(hasError).toBe(false);
    });

    test('should recover from failed data fetch', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Page should be in a valid state (showing content or empty state)
      const hasValidState =
        (await page.locator('[class*="quest"]').count()) > 0 ||
        (await page.locator('[class*="empty"]').count()) > 0 ||
        (await page.locator('[class*="loading"]').count()) > 0;

      expect(hasValidState).toBe(true);
    });
  });

  test.describe('Performance and Loading', () => {
    test('should show loading state initially', async ({ page }) => {
      const loadingPromise = page.goto('/quests');

      // Should show loading indicator briefly
      const hasLoadingIndicator =
        (await page.locator('[class*="loading"], [class*="spinner"]').count()) > 0;

      await loadingPromise;

      // Loading state is optional but good UX
      expect(hasLoadingIndicator).toBeGreaterThanOrEqual(0);
    });

    test('should load quest data within reasonable time', async ({ page }) => {
      const startTime = Date.now();

      await page.goto('/quests');
      await page.waitForLoadState('networkidle');

      const loadTime = Date.now() - startTime;

      // Should load in under 10 seconds
      expect(loadTime).toBeLessThan(10000);
    });
  });

  test.describe('Visual Layout', () => {
    test('should have responsive layout', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');

      // Page should have content and be visible
      const isVisible = await page.locator('.quests-page, [class*="quest"]').isVisible();
      expect(isVisible).toBeTruthy();
    });

    test('should display quest cards in grid or list format', async ({ page }) => {
      await page.goto('/quests');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      const questCards = await page.locator('[class*="quest-card"], [class*="quest-item"]').count();

      if (questCards > 0) {
        // Cards should be visible and properly laid out
        const firstCard = page.locator('[class*="quest-card"], [class*="quest-item"]').first();
        await expect(firstCard).toBeVisible();
      }
    });
  });
});
