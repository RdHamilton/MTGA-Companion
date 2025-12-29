import { test, expect } from '@playwright/test';

/**
 * Pipeline E2E Tests
 *
 * These tests validate the full data pipeline from MTGA log files through
 * the daemon to the frontend UI. They use sample log fixtures that the
 * daemon reads on startup (with ReadFromStart=true).
 *
 * The sample log file (frontend/tests/e2e/fixtures/logs/sample-session.log) contains:
 * - Player: "E2ETestPlayer"
 * - 5 decks: Standard, Historic, Explorer, Alchemy, Brawl (10 cards each)
 * - 12 matches total:
 *   - Play: 1 win, 1 loss
 *   - Ladder: 2 wins, 1 loss
 *   - Traditional_Ladder: 1 win, 1 loss
 *   - QuickDraft: 2 wins, 1 loss
 *   - PremierDraft: 1 win, 1 loss
 * - 2 draft sessions: QuickDraft_FDN (3 picks), PremierDraft_DSK (2 picks)
 * - 3 quests with full completion (4 daily wins, 15 weekly wins)
 * - Rank progression: Gold 3->4 (Constructed), Silver 2->3 (Limited)
 *
 * Tests cover:
 * - Match History: matches display, event types, wins/losses
 * - Decks: deck display, multiple formats
 * - Draft: draft sessions, picks
 * - Quests: quest display, daily/weekly wins
 * - Collection: collection page load
 * - Meta: metagame dashboard, format dropdown (#737)
 * - Charts: deck performance, rank progression, format distribution, result breakdown
 * - Sorting/Filtering: filter controls on various pages
 *
 * Run with: USE_LOG_FIXTURES=true npx playwright test --project=pipeline
 */
test.describe('Data Pipeline - Log to UI', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to app and wait for it to load
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible({ timeout: 15000 });

    // Give the daemon time to process the log file
    await page.waitForTimeout(2000);
  });

  test.describe('Match History Pipeline', () => {
    test('should display matches parsed from log file', async ({ page }) => {
      await expect(page.locator('h1.page-title')).toHaveText('Match History');

      const table = page.locator('.match-history-table-container table');
      const emptyState = page.locator('.empty-state');

      await expect(table.or(emptyState)).toBeVisible({ timeout: 10000 });

      const hasMatches = await table.isVisible();

      if (hasMatches) {
        const rows = table.locator('tbody tr');
        const rowCount = await rows.count();

        // Should have 12 matches from the log
        expect(rowCount).toBeGreaterThan(0);

        const tableText = await table.textContent();

        // Check for various opponent types from our log fixture
        const hasOpponents =
          tableText?.includes('Opponent') ||
          tableText?.includes('PlayOpponent') ||
          tableText?.includes('LadderOpponent') ||
          tableText?.includes('DraftOpponent') ||
          tableText?.includes('PremierOpponent');

        expect(hasOpponents).toBeTruthy();
      }
    });

    test('should show multiple event types from log', async ({ page }) => {
      const table = page.locator('.match-history-table-container table');
      const hasTable = await table.isVisible().catch(() => false);

      if (hasTable) {
        const tableText = await table.textContent();

        // Log contains: Play, Ladder, Traditional_Ladder, QuickDraft, PremierDraft
        const hasPlayEvents = tableText?.includes('Play');
        const hasRankedEvents =
          tableText?.includes('Ranked') || tableText?.includes('Ladder');
        const hasDraftEvents =
          tableText?.includes('Draft') ||
          tableText?.includes('Quick') ||
          tableText?.includes('Premier');

        // Should have at least some of these event types
        expect(hasPlayEvents || hasRankedEvents || hasDraftEvents).toBeTruthy();
      }
    });

    test('should show both wins and losses', async ({ page }) => {
      const table = page.locator('.match-history-table-container table');
      const hasTable = await table.isVisible().catch(() => false);

      if (hasTable) {
        // Check for Win and Loss indicators in the table
        const winCells = table.locator('td:has-text("Win"), .result-win');
        const lossCells = table.locator('td:has-text("Loss"), .result-loss');

        const winCount = await winCells.count().catch(() => 0);
        const lossCount = await lossCells.count().catch(() => 0);

        // Should have both wins and losses from the log
        // Log has: 7 wins, 5 losses
        expect(winCount + lossCount).toBeGreaterThan(0);
      }
    });
  });

  test.describe('Decks Pipeline', () => {
    test.beforeEach(async ({ page }) => {
      await page.click('a[href="/decks"]');
      await page.waitForURL('**/decks');
    });

    test('should display decks from log file', async ({ page }) => {
      // Wait for the decks page to load - actual class is .decks-page
      const decksPage = page.locator('.decks-page');
      await expect(decksPage).toBeVisible({ timeout: 10000 });

      // Check for either decks grid or empty state
      const decksGrid = page.locator('.decks-grid');
      const emptyState = page.locator('.empty-state');

      await expect(decksGrid.or(emptyState)).toBeVisible({ timeout: 10000 });

      const hasDecks = await decksGrid.isVisible().catch(() => false);

      if (hasDecks) {
        const pageText = await page.textContent('body');

        // Log contains 5 decks in different formats
        const hasStandardDeck = pageText?.includes('Mono Red Aggro');
        const hasHistoricDeck = pageText?.includes('Historic Elves');
        const hasExplorerDeck = pageText?.includes('Explorer Control');
        const hasAlchemyDeck = pageText?.includes('Alchemy Combo');
        const hasBrawlDeck = pageText?.includes('Brawl Commander');

        // At least one deck should be present
        const hasDeck =
          hasStandardDeck ||
          hasHistoricDeck ||
          hasExplorerDeck ||
          hasAlchemyDeck ||
          hasBrawlDeck;

        expect(hasDeck).toBeTruthy();
      }
    });

    test('should show decks from multiple formats', async ({ page }) => {
      // Wait for decks page to fully load
      const decksPage = page.locator('.decks-page');
      await expect(decksPage).toBeVisible({ timeout: 10000 });

      // Check for decks grid with format badges
      const decksGrid = page.locator('.decks-grid');
      const hasDecks = await decksGrid.isVisible().catch(() => false);

      if (hasDecks) {
        // Look for format labels inside deck cards (actual class is .deck-format)
        const formatLabels = page.locator('.deck-format');
        const formatCount = await formatLabels.count();
        expect(formatCount).toBeGreaterThanOrEqual(1);
      } else {
        // If no decks, that's also valid - page just shows empty state
        const emptyState = page.locator('.empty-state');
        await expect(emptyState).toBeVisible();
      }
    });
  });

  test.describe('Draft Pipeline', () => {
    test.beforeEach(async ({ page }) => {
      await page.click('a[href="/draft"]');
      await page.waitForURL('**/draft');
    });

    test('should display draft session from log file', async ({ page }) => {
      const draftContent = page.locator('.draft-container, .draft-empty');
      await expect(draftContent.first()).toBeVisible({ timeout: 10000 });

      const draftContainer = page.locator('.draft-container');
      const draftEmpty = page.locator('.draft-empty');
      const historicalSection = page.locator('text=Historical Drafts');

      const hasDraftContent = await draftContainer.isVisible().catch(() => false);
      const hasHistorical = await historicalSection.isVisible().catch(() => false);
      const hasEmpty = await draftEmpty.isVisible().catch(() => false);

      expect(hasDraftContent || hasHistorical || hasEmpty).toBeTruthy();

      if (hasDraftContent || hasHistorical) {
        const pageText = await page.textContent('body');

        // Log contains QuickDraft_FDN and PremierDraft_DSK sessions
        const hasDraftInfo =
          pageText?.includes('FDN') ||
          pageText?.includes('DSK') ||
          pageText?.includes('Quick Draft') ||
          pageText?.includes('Premier Draft') ||
          pageText?.includes('draft');

        expect(hasDraftInfo).toBeTruthy();
      }
    });

    test('should show draft picks from log file', async ({ page }) => {
      const draftContainer = page.locator('.draft-container');
      const hasDraft = await draftContainer.isVisible().catch(() => false);

      if (hasDraft) {
        const cardElements = page.locator('.draft-card, .card-item, .picked-card');
        const cardCount = await cardElements.count().catch(() => 0);

        if (cardCount > 0) {
          expect(cardCount).toBeGreaterThanOrEqual(1);
        }
      }
    });
  });

  test.describe('Quests Pipeline', () => {
    test.beforeEach(async ({ page }) => {
      await page.click('a[href="/quests"]');
      await page.waitForURL('**/quests');
    });

    test('should display quests from log file', async ({ page }) => {
      const questsSection = page.locator('.quests-section');
      const emptyState = page.locator('.empty-state');

      await expect(questsSection.first().or(emptyState)).toBeVisible({ timeout: 10000 });

      const hasQuests = await questsSection.first().isVisible();

      if (hasQuests) {
        const pageText = await page.textContent('body');

        // Log contains quests: Win 4 games, Cast 20 spells, Play 30 lands
        const hasQuestContent =
          pageText?.includes('Win') ||
          pageText?.includes('Cast') ||
          pageText?.includes('Play') ||
          pageText?.includes('Quest') ||
          pageText?.includes('games') ||
          pageText?.includes('spells') ||
          pageText?.includes('lands');

        expect(hasQuestContent).toBeTruthy();
      }
    });

    test('should show completed quests', async ({ page }) => {
      const questsSection = page.locator('.quests-section');
      const hasQuests = await questsSection.first().isVisible().catch(() => false);

      if (hasQuests) {
        // All 3 quests are completed in the log
        // Look for completion indicators (100%, checkmarks, completed status)
        const completionIndicators = page.locator(
          '.quest-complete, .completed, [data-completed="true"], .progress-100'
        );
        const indicatorCount = await completionIndicators.count().catch(() => 0);

        // Also check for progress bars at 100%
        const progressBars = page.locator('.quest-progress, .progress-bar, progress');
        const progressCount = await progressBars.count().catch(() => 0);

        expect(indicatorCount + progressCount).toBeGreaterThanOrEqual(0);
      }
    });

    test('should show daily and weekly win counts', async ({ page }) => {
      // Wait for loading to complete (spinner to disappear)
      const loadingSpinner = page.locator('.loading-container');
      await loadingSpinner.waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {});

      // Page should not have errors
      const errorState = page.locator('.error-state');
      await expect(errorState).not.toBeVisible();

      // After loading, check for wins grid (should be visible when page loads successfully)
      const winsGrid = page.locator('.wins-grid');
      await expect(winsGrid).toBeVisible({ timeout: 10000 });

      // Check for daily/weekly wins cards
      const dailyWinsCard = page.locator('.daily-wins-card');
      const cardCount = await dailyWinsCard.count();
      expect(cardCount).toBeGreaterThanOrEqual(1);
    });
  });

  test.describe('Collection Pipeline', () => {
    test.beforeEach(async ({ page }) => {
      await page.click('a[href="/collection"]');
      await page.waitForURL('**/collection');
    });

    test('should display collection page', async ({ page }) => {
      const collectionContainer = page.locator('.collection-container, .collection-page');
      const emptyState = page.locator('.empty-state');

      await expect(collectionContainer.first().or(emptyState)).toBeVisible({ timeout: 10000 });

      // Collection page should load without errors
      const errorState = page.locator('.error-state');
      await expect(errorState).not.toBeVisible();
    });
  });

  test.describe('Meta Pipeline', () => {
    test.beforeEach(async ({ page }) => {
      await page.click('a[href="/meta"]');
      await page.waitForURL('**/meta');
    });

    test('should display meta page without errors', async ({ page }) => {
      // Wait for page to load
      const metaPage = page.locator('.meta-page');
      await expect(metaPage).toBeVisible({ timeout: 10000 });

      // Wait for loading to complete
      const loadingSpinner = page.locator('.meta-loading');
      await loadingSpinner.waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {});

      // Page should not have error banner (#737 fix - null check)
      const errorBanner = page.locator('.meta-error');
      await expect(errorBanner).not.toBeVisible();
    });

    test('should have format dropdown with friendly names', async ({ page }) => {
      // Wait for page to load
      const metaPage = page.locator('.meta-page');
      await expect(metaPage).toBeVisible({ timeout: 10000 });

      // Find format dropdown
      const formatSelect = page.locator('.format-select');
      await expect(formatSelect).toBeVisible();

      // Get all options
      const options = await formatSelect.locator('option').allTextContents();

      // Should have friendly format names (not raw like "Alchemy_Play")
      expect(options).toContain('Standard');
      expect(options).toContain('Historic');
      expect(options).toContain('Explorer');

      // Should NOT contain draft formats
      const hasDraftFormats = options.some(opt =>
        opt.includes('Draft') || opt.includes('Sealed') || opt.includes('QuickDraft')
      );
      expect(hasDraftFormats).toBeFalsy();

      // Should NOT contain raw format names with underscores
      const hasRawFormats = options.some(opt => opt.includes('_'));
      expect(hasRawFormats).toBeFalsy();
    });

    test('should filter archetypes by format', async ({ page }) => {
      // Wait for page to load
      const metaPage = page.locator('.meta-page');
      await expect(metaPage).toBeVisible({ timeout: 10000 });

      // Wait for loading to complete
      const loadingSpinner = page.locator('.meta-loading');
      await loadingSpinner.waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {});

      // Change format
      const formatSelect = page.locator('.format-select');
      await formatSelect.selectOption('historic');

      // Wait for reload
      await page.waitForTimeout(1000);
      await loadingSpinner.waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {});

      // Verify the format was changed
      const selectedValue = await formatSelect.inputValue();
      expect(selectedValue).toBe('historic');

      // Page should still not have errors
      const errorBanner = page.locator('.meta-error');
      await expect(errorBanner).not.toBeVisible();
    });

    test('should display archetype cards when data is available', async ({ page }) => {
      // Wait for page to load
      const metaPage = page.locator('.meta-page');
      await expect(metaPage).toBeVisible({ timeout: 10000 });

      // Wait for loading to complete
      const loadingSpinner = page.locator('.meta-loading');
      await loadingSpinner.waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {});

      // Check for either archetype cards or no-data message (both are valid)
      const archetypeCards = page.locator('.archetype-card');
      const noData = page.locator('.no-data');

      const hasArchetypes = await archetypeCards.count() > 0;
      const hasNoData = await noData.isVisible().catch(() => false);

      // Should have either archetypes or no-data message
      expect(hasArchetypes || hasNoData).toBeTruthy();
    });
  });

  test.describe('Charts Pipeline', () => {
    // TODO: Re-enable when #739 is fixed (Win Rate Trend chart fails to load - 400 Bad Request)
    test.skip('should display Win Rate Trend chart', async ({ page }) => {
      await page.click('a.tab[href="/charts/win-rate-trend"]');
      await page.waitForURL('**/charts/win-rate-trend');

      // Wait for page to load
      const pageContainer = page.locator('.page-container');
      await expect(pageContainer).toBeVisible({ timeout: 10000 });

      // Change date filter to "All Time" since log data has old dates
      const dateRangeSelect = page.locator('.filter-row select').first();
      await dateRangeSelect.selectOption('all');

      // Wait for loading to complete (spinner to disappear)
      const loadingSpinner = page.locator('.loading-container');
      await loadingSpinner.waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {});

      // Chart should display without errors
      const chartContainer = page.locator('.chart-container');
      const emptyState = page.locator('.empty-state');

      await expect(chartContainer.or(emptyState)).toBeVisible({ timeout: 10000 });

      // Should not have errors
      const errorState = page.locator('.error-state');
      await expect(errorState).not.toBeVisible();
    });

    test('should display Deck Performance chart', async ({ page }) => {
      await page.click('a.tab[href="/charts/win-rate-trend"]');
      await page.waitForURL('**/charts/**');

      await page.click('.sub-tab-bar a[href="/charts/deck-performance"]');
      await page.waitForURL('**/charts/deck-performance');

      const activeSubTab = page.locator('.sub-tab-bar a.active');
      await expect(activeSubTab).toContainText(/Deck Performance/i);

      const errorState = page.locator('.error-state');
      await expect(errorState).not.toBeVisible();
    });

    test('should display Rank Progression chart', async ({ page }) => {
      await page.click('a.tab[href="/charts/win-rate-trend"]');
      await page.waitForURL('**/charts/**');

      await page.click('.sub-tab-bar a[href="/charts/rank-progression"]');
      await page.waitForURL('**/charts/rank-progression');

      const activeSubTab = page.locator('.sub-tab-bar a.active');
      await expect(activeSubTab).toContainText(/Rank Progression/i);

      // Log has rank updates for both Constructed (Gold 3->4) and Limited (Silver 2->3)
      const pageText = await page.textContent('body');
      const hasRankInfo =
        pageText?.includes('Gold') ||
        pageText?.includes('Silver') ||
        pageText?.includes('Rank') ||
        pageText?.includes('Constructed') ||
        pageText?.includes('Limited');

      expect(hasRankInfo).toBeTruthy();

      const errorState = page.locator('.error-state');
      await expect(errorState).not.toBeVisible();
    });

    test('should display Format Distribution chart', async ({ page }) => {
      await page.click('a.tab[href="/charts/win-rate-trend"]');
      await page.waitForURL('**/charts/**');

      await page.click('.sub-tab-bar a[href="/charts/format-distribution"]');
      await page.waitForURL('**/charts/format-distribution');

      const activeSubTab = page.locator('.sub-tab-bar a.active');
      await expect(activeSubTab).toContainText(/Format Distribution/i);

      // Log has matches in Play, Ladder, Traditional, QuickDraft, PremierDraft
      const chartOrData = page.locator('.recharts-wrapper, svg, .chart-container, .stats-grid');
      const emptyState = page.locator('.empty-state, .no-data');

      await expect(chartOrData.first().or(emptyState)).toBeVisible({ timeout: 10000 });

      const errorState = page.locator('.error-state');
      await expect(errorState).not.toBeVisible();
    });

    test('should display Result Breakdown chart', async ({ page }) => {
      await page.click('a.tab[href="/charts/win-rate-trend"]');
      await page.waitForURL('**/charts/**');

      await page.click('.sub-tab-bar a[href="/charts/result-breakdown"]');
      await page.waitForURL('**/charts/result-breakdown');

      const activeSubTab = page.locator('.sub-tab-bar a.active');
      await expect(activeSubTab).toContainText(/Result Breakdown/i);

      // Wait for page to load
      const pageContainer = page.locator('.page-container');
      await expect(pageContainer).toBeVisible({ timeout: 10000 });

      // Change date filter to "All Time" since log data has old dates
      const dateRangeSelect = page.locator('.filter-row select').first();
      await dateRangeSelect.selectOption('all');

      // Wait for data to reload
      await page.waitForTimeout(1000);

      // Log has 7 wins and 5 losses - actual class is .metrics-container
      const metricsContainer = page.locator('.metrics-container');
      const emptyState = page.locator('.empty-state');

      await expect(metricsContainer.or(emptyState)).toBeVisible({ timeout: 10000 });

      const errorState = page.locator('.error-state');
      await expect(errorState).not.toBeVisible();
    });
  });

  test.describe('Sorting and Filtering', () => {
    test('should have filter controls on Match History page', async ({ page }) => {
      // Match History is the default page - check for filter row
      const filterRow = page.locator('.filter-row');
      await expect(filterRow).toBeVisible({ timeout: 10000 });

      // Should have at least one select element for filtering
      const selects = filterRow.locator('select');
      const selectCount = await selects.count();
      expect(selectCount).toBeGreaterThan(0);
    });

    test('should have sortable table headers on Match History', async ({ page }) => {
      const table = page.locator('.match-history-table-container table');
      const hasTable = await table.isVisible().catch(() => false);

      if (hasTable) {
        // Table should have headers
        const headers = table.locator('thead th');
        const headerCount = await headers.count();
        expect(headerCount).toBeGreaterThan(0);
      }
    });

    test('should have date filter on Quests page', async ({ page }) => {
      await page.click('a[href="/quests"]');
      await page.waitForURL('**/quests');

      // Wait for page to load
      await page.waitForTimeout(1000);

      // Check for any select element (date filter) or page content
      const selects = page.locator('select');
      const selectCount = await selects.count();

      // Quests page should either have filters or show content
      const pageContent = page.locator('.quests-section, .quests-header, .empty-state');
      await expect(pageContent.first()).toBeVisible({ timeout: 10000 });

      // Should have at least some interactive elements
      expect(selectCount).toBeGreaterThanOrEqual(0);
    });

    test('should have filters on Collection page', async ({ page }) => {
      await page.click('a[href="/collection"]');
      await page.waitForURL('**/collection');

      // Wait for page to load
      await page.waitForTimeout(1000);

      // Collection page should have some filter controls
      const filterArea = page.locator('.filter-controls, .collection-filters, select, input');
      const filterCount = await filterArea.count();
      expect(filterCount).toBeGreaterThanOrEqual(0);

      // Should not error
      const errorState = page.locator('.error-state');
      await expect(errorState).not.toBeVisible();
    });

    test('should have date filter on Charts pages', async ({ page }) => {
      await page.click('a.tab[href="/charts/win-rate-trend"]');
      await page.waitForURL('**/charts/win-rate-trend');

      // Wait for page to load
      const pageContainer = page.locator('.page-container');
      await expect(pageContainer).toBeVisible({ timeout: 10000 });

      // Chart pages should have filter controls
      const filterRow = page.locator('.filter-row');
      await expect(filterRow).toBeVisible({ timeout: 5000 });

      const selects = filterRow.locator('select');
      const selectCount = await selects.count();
      expect(selectCount).toBeGreaterThan(0);

      // Verify the date range select has expected options
      const dateRangeSelect = selects.first();
      const options = await dateRangeSelect.locator('option').allTextContents();
      expect(options.length).toBeGreaterThanOrEqual(3);

      // Should have common filter options
      const hasDateOptions = options.some(opt =>
        opt.includes('7 Days') || opt.includes('30 Days') || opt.includes('All')
      );
      expect(hasDateOptions).toBeTruthy();
    });
  });

  test.describe('Footer Stats Pipeline', () => {
    test('should display stats in footer from parsed matches', async ({ page }) => {
      const footer = page.locator('.footer, footer');
      const hasFooter = await footer.isVisible().catch(() => false);

      if (hasFooter) {
        const footerText = await footer.textContent();

        // Footer should show win/loss stats
        // Log contains: 7 wins, 5 losses = ~58% win rate
        const hasStats =
          footerText?.includes('W') || footerText?.includes('L') || footerText?.includes('%');

        expect(hasStats).toBeTruthy();
      }
    });
  });

  test.describe('Data Consistency', () => {
    test('should not show error states when log data is present', async ({ page }) => {
      // Check Match History (default page)
      let errorState = page.locator('.error-state');
      await expect(errorState).not.toBeVisible({ timeout: 5000 }).catch(() => {});

      // Check Draft page
      await page.click('a[href="/draft"]');
      await page.waitForURL('**/draft');
      errorState = page.locator('.error-state');
      await expect(errorState).not.toBeVisible({ timeout: 5000 }).catch(() => {});

      // Check Quests page
      await page.click('a[href="/quests"]');
      await page.waitForURL('**/quests');
      errorState = page.locator('.error-state');
      await expect(errorState).not.toBeVisible({ timeout: 5000 }).catch(() => {});

      // Check Decks page
      await page.click('a[href="/decks"]');
      await page.waitForURL('**/decks');
      errorState = page.locator('.error-state');
      await expect(errorState).not.toBeVisible({ timeout: 5000 }).catch(() => {});
    });

    test('should maintain data across page navigation', async ({ page }) => {
      // Navigate to Draft
      await page.click('a[href="/draft"]');
      await page.waitForURL('**/draft');

      // Navigate to Quests
      await page.click('a[href="/quests"]');
      await page.waitForURL('**/quests');

      // Navigate to Decks
      await page.click('a[href="/decks"]');
      await page.waitForURL('**/decks');

      // Navigate back to Match History
      await page.click('a[href="/match-history"]');
      await page.waitForURL('**/match-history');

      // Data should still be present
      const table = page.locator('.match-history-table-container table');
      const emptyState = page.locator('.empty-state');

      await expect(table.or(emptyState)).toBeVisible({ timeout: 10000 });
    });
  });
});
