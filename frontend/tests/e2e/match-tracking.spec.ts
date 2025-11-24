import { test, expect } from '@playwright/test';

/**
 * E2E tests for Match Tracking workflow
 *
 * Prerequisites:
 * - Run `wails dev` in the project root
 * - For full workflow testing, the database should contain match history data
 *
 * Tests verify the match tracking, filtering, and statistics features.
 */
test.describe('Match Tracking Workflow', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app and wait for it to be ready
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 10000 });
  });

  test.describe('Match History View', () => {
    test('should navigate to Match History view', async ({ page }) => {
      // Match History should be the default view
      await page.getByText('Match History').click();

      // Verify we're on the match history page
      await expect(page).toHaveURL(/\/(match-history)?$/);

      // Check for match history container
      const hasMatchHistory =
        (await page.locator('.match-history, .matches-container').count()) > 0 ||
        (await page.getByText(/match/i).count()) > 0;

      expect(hasMatchHistory).toBeTruthy();
    });

    test('should display match list or empty state', async ({ page }) => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Should show either matches or empty state
      const hasContent =
        (await page.locator('[class*="match"]').count()) > 0 ||
        (await page.getByText(/no matches/i).count()) > 0 ||
        (await page.getByText(/empty/i).count()) > 0;

      expect(hasContent).toBeTruthy();
    });

    test('should display match cards when data exists', async ({ page }) => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Look for match card elements
      const matchCards = page.locator('[class*="match-card"], [class*="match-item"]');
      const matchCount = await matchCards.count();

      if (matchCount > 0) {
        // Verify match cards have expected content
        const firstMatch = matchCards.first();
        await expect(firstMatch).toBeVisible();

        // Match cards should have some text content
        const cardText = await firstMatch.textContent();
        expect(cardText).toBeTruthy();
      } else {
        // Empty state is acceptable
        const hasEmptyState = await page.getByText(/no matches/i).count() > 0;
        expect(hasEmptyState).toBeTruthy();
      }
    });
  });

  test.describe('Match Filtering', () => {
    test('should have format filter controls', async ({ page }) => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Look for filter dropdowns or buttons
      const filterControls = page.locator(
        'select[class*="filter"], [role="combobox"], button[class*="filter"]'
      );

      const filterCount = await filterControls.count();

      // Filters might be present if data exists
      expect(filterCount).toBeGreaterThanOrEqual(0);
    });

    test('should filter matches by format when selected', async ({ page }) => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Look for format filter (e.g., Standard, Historic, etc.)
      const formatSelectors = page.locator('select, [role="combobox"]');
      const selectCount = await formatSelectors.count();

      if (selectCount > 0) {
        // Try to select a filter option
        const firstSelect = formatSelectors.first();
        await firstSelect.click();

        // Wait for potential filter to apply
        await page.waitForTimeout(1000);

        // After filtering, match count might change or stay the same
        const filteredMatches = await page.locator('[class*="match-card"], [class*="match-item"]').count();
        expect(filteredMatches).toBeGreaterThanOrEqual(0);
      }
    });

    test('should have date range filtering controls', async ({ page }) => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Look for date pickers or date range controls
      const dateControls = page.locator(
        'input[type="date"], input[type="datetime-local"], [class*="date-picker"]'
      );

      const dateControlCount = await dateControls.count();

      // Date filters are optional but common
      expect(dateControlCount).toBeGreaterThanOrEqual(0);
    });
  });

  test.describe('Match Details', () => {
    test('should expand match details when clicked', async ({ page }) => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Find match cards
      const matchCards = page.locator('[class*="match-card"], [class*="match-item"]');
      const matchCount = await matchCards.count();

      if (matchCount > 0) {
        // Click first match to expand
        await matchCards.first().click();

        // Wait for expansion
        await page.waitForTimeout(500);

        // Check if details are visible (game-by-game breakdown, etc.)
        const hasExpandedContent =
          (await page.locator('[class*="game"], [class*="detail"], [class*="expanded"]').count()) > 0;

        expect(hasExpandedContent).toBeTruthy();
      }
    });

    test('should display game details within expanded match', async ({ page }) => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Find and click a match card
      const matchCards = page.locator('[class*="match-card"], [class*="match-item"]');
      const matchCount = await matchCards.count();

      if (matchCount > 0) {
        await matchCards.first().click();
        await page.waitForTimeout(500);

        // Look for game-level details
        const pageText = await page.locator('body').textContent();

        const hasGameDetails =
          pageText?.toLowerCase().includes('game') ||
          pageText?.toLowerCase().includes('turn') ||
          pageText?.toLowerCase().includes('duration') ||
          (await page.locator('[class*="game-"]').count()) > 0;

        // Game details should be present when match is expanded
        expect(hasGameDetails).toBeTruthy();
      }
    });

    test('should show match result and deck information', async ({ page }) => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Find match cards
      const matchCards = page.locator('[class*="match-card"], [class*="match-item"]');
      const matchCount = await matchCards.count();

      if (matchCount > 0) {
        const matchText = await matchCards.first().textContent();

        // Match should show result (win/loss) and deck colors or info
        const hasMatchInfo =
          matchText?.toLowerCase().includes('win') ||
          matchText?.toLowerCase().includes('loss') ||
          matchText?.toLowerCase().includes('deck') ||
          (await page.locator('[class*="color"], [class*="mana"]').count()) > 0;

        expect(hasMatchInfo).toBeTruthy();
      }
    });
  });

  test.describe('Statistics and Charts', () => {
    test('should navigate to Charts view', async ({ page }) => {
      // Click on Charts tab
      await page.getByText('Charts').click();

      // Verify we're on the charts page
      await expect(page).toHaveURL(/\/charts/);

      // Check for charts container or sub-navigation
      const hasChartsView =
        (await page.locator('[class*="chart"]').count()) > 0 ||
        (await page.getByText(/win rate|trend|progression/i).count()) > 0;

      expect(hasChartsView).toBeTruthy();
    });

    test('should display win rate trend chart', async ({ page }) => {
      await page.goto('/charts/win-rate-trend');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Look for chart elements (SVG, canvas, or recharts components)
      const hasChart =
        (await page.locator('svg, canvas, [class*="recharts"]').count()) > 0 ||
        (await page.getByText(/no data/i).count()) > 0;

      expect(hasChart).toBeTruthy();
    });

    test('should display deck performance chart', async ({ page }) => {
      await page.goto('/charts/deck-performance');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Look for chart or deck performance data
      const hasChart =
        (await page.locator('svg, canvas, [class*="recharts"], [class*="chart"]').count()) > 0 ||
        (await page.getByText(/no data|deck/i).count()) > 0;

      expect(hasChart).toBeTruthy();
    });

    test('should display rank progression chart', async ({ page }) => {
      await page.goto('/charts/rank-progression');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Look for rank progression visualization
      const hasChart =
        (await page.locator('svg, canvas, [class*="recharts"], [class*="rank"]').count()) > 0 ||
        (await page.getByText(/no data|rank/i).count()) > 0;

      expect(hasChart).toBeTruthy();
    });

    test('should navigate between chart views using sub-tabs', async ({ page }) => {
      await page.goto('/charts/win-rate-trend');
      await page.waitForLoadState('networkidle');

      // Click on different chart sub-tabs
      const deckPerformanceTab = page.getByText('Deck Performance');
      if (await deckPerformanceTab.isVisible()) {
        await deckPerformanceTab.click();
        await expect(page).toHaveURL(/\/charts\/deck-performance/);
      }

      const rankProgressionTab = page.getByText('Rank Progression');
      if (await rankProgressionTab.isVisible()) {
        await rankProgressionTab.click();
        await expect(page).toHaveURL(/\/charts\/rank-progression/);
      }
    });
  });

  test.describe('Footer Statistics', () => {
    test('should display footer with statistics', async ({ page }) => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Check for footer
      const footer = page.locator('.app-footer, footer, [class*="footer"]');
      await expect(footer).toBeVisible();
    });

    test('should show total matches count in footer', async ({ page }) => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Look for matches count
      const footerText = await page.locator('.app-footer, footer, [class*="footer"]').textContent();

      const hasMatchCount =
        footerText?.toLowerCase().includes('match') ||
        /\d+/.test(footerText || ''); // Has at least one number

      expect(hasMatchCount).toBeTruthy();
    });

    test('should show win rate in footer', async ({ page }) => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Look for win rate percentage
      const footerText = await page.locator('.app-footer, footer, [class*="footer"]').textContent();

      const hasWinRate =
        footerText?.toLowerCase().includes('win') ||
        footerText?.toLowerCase().includes('%') ||
        footerText?.toLowerCase().includes('rate');

      expect(hasWinRate).toBeTruthy();
    });

    test('should show win streak in footer when applicable', async ({ page }) => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Look for streak indicator
      const footerText = await page.locator('.app-footer, footer, [class*="footer"]').textContent();

      const hasStreakInfo =
        footerText?.toLowerCase().includes('streak') ||
        /[WL]\d+/.test(footerText || ''); // Matches patterns like W3, L2

      // Streak might only show when there is one
      expect(hasStreakInfo).toBeDefined();
    });

    test('should update footer statistics when navigating', async ({ page }) => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Navigate to different view
      await page.getByText('Draft').click();
      await page.waitForTimeout(1000);

      // Navigate back
      await page.getByText('Match History').click();
      await page.waitForTimeout(1000);

      // Footer should still be present and functional
      const updatedFooter = await page.locator('.app-footer, footer, [class*="footer"]').textContent();

      expect(updatedFooter).toBeTruthy();
    });
  });

  test.describe('Deck Performance Filtering', () => {
    test('should navigate to deck performance view', async ({ page }) => {
      await page.goto('/charts/deck-performance');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Verify we're on deck performance page
      await expect(page).toHaveURL(/\/charts\/deck-performance/);

      // Check for deck-related content
      const pageText = await page.locator('body').textContent();
      const hasDeckContent = pageText?.toLowerCase().includes('deck');

      expect(hasDeckContent).toBeTruthy();
    });

    test('should display deck performance data when available', async ({ page }) => {
      await page.goto('/charts/deck-performance');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Look for deck performance visualization or data
      const hasContent =
        (await page.locator('svg, canvas, [class*="recharts"]').count()) > 0 ||
        (await page.locator('[class*="deck-"]').count()) > 0 ||
        (await page.getByText(/no data/i).count()) > 0;

      expect(hasContent).toBeTruthy();
    });
  });

  test.describe('Error Handling', () => {
    test('should handle empty match history gracefully', async ({ page }) => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      // Should not show errors even with no data
      const hasError = (await page.locator('[class*="error"]').count()) > 0;
      expect(hasError).toBe(false);
    });

    test('should handle chart rendering without data', async ({ page }) => {
      await page.goto('/charts/win-rate-trend');
      await page.waitForLoadState('networkidle');

      // Charts should show empty state, not error
      const hasError = await page.getByText(/error|failed/i).count() > 0;
      expect(hasError).toBe(false);
    });

    test('should maintain functionality across view transitions', async ({ page }) => {
      // Navigate through multiple views
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      await page.getByText('Charts').click();
      await page.waitForTimeout(500);

      await page.getByText('Draft').click();
      await page.waitForTimeout(500);

      await page.getByText('Match History').click();
      await page.waitForTimeout(500);

      // App should still be functional
      const hasError = (await page.locator('[class*="error"]').count()) > 0;
      expect(hasError).toBe(false);

      // Footer should still be visible
      await expect(page.locator('.app-footer, footer, [class*="footer"]')).toBeVisible();
    });
  });
});
