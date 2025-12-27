import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import Quests from './Quests';
import { mockWailsApp } from '@/test/mocks/apiMock';
import { AppProvider } from '../context/AppContext';
import { models } from '@/types/models';

// Helper function to create mock Quest
function createMockQuest(overrides: Partial<models.Quest> = {}): models.Quest {
  return new models.Quest({
    id: 1,
    quest_id: 'quest_001',
    quest_type: 'Quests/Quest_PlayCards',
    goal: 10,
    starting_progress: 0,
    ending_progress: 5,
    completed: false,
    can_swap: true,
    rewards: '500 Gold',
    assigned_at: new Date('2024-01-15T10:00:00').toISOString(),
    completed_at: undefined,
    rerolled: false,
    ...overrides,
  });
}

// Helper function to create mock Account
function createMockAccount(overrides: Partial<models.Account> = {}): models.Account {
  return new models.Account({
    ID: 1,
    Name: 'TestPlayer',
    DailyWins: 3,
    WeeklyWins: 8,
    MasteryLevel: 45,
    MasteryPass: 'Premium',
    MasteryMax: 100,
    IsDefault: true,
    ...overrides,
  });
}

// Wrapper component with AppProvider
function renderWithProvider(ui: React.ReactElement) {
  return render(<AppProvider>{ui}</AppProvider>);
}

// Helper to get select by finding the label then the next select sibling
function getSelectByLabel(labelText: string): HTMLSelectElement {
  const label = screen.getByText(labelText);
  const filterGroup = label.closest('.filter-group');
  return filterGroup?.querySelector('select') as HTMLSelectElement;
}

// Helper to get input by finding the label
function getInputByLabel(labelText: string): HTMLInputElement {
  const label = screen.getByText(labelText);
  const filterGroup = label.closest('.filter-group');
  return filterGroup?.querySelector('input') as HTMLInputElement;
}

describe('Quests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  describe('Loading State', () => {
    it('should show loading spinner while fetching data', async () => {
      let resolveQuests: (value: models.Quest[]) => void;
      const loadingPromise = new Promise<models.Quest[]>((resolve) => {
        resolveQuests = resolve;
      });
      mockWailsApp.GetActiveQuests.mockReturnValue(loadingPromise);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(createMockAccount());

      renderWithProvider(<Quests />);

      expect(screen.getByText('Loading quest data...')).toBeInTheDocument();

      resolveQuests!([]);
      await waitFor(() => {
        expect(screen.queryByText('Loading quest data...')).not.toBeInTheDocument();
      });
    });
  });

  describe('Error State', () => {
    it('should show error state when GetActiveQuests fails', async () => {
      mockWailsApp.GetActiveQuests.mockRejectedValue(new Error('Database error'));
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Failed to load quest data' })).toBeInTheDocument();
      });
      expect(screen.getByText(/Failed to load active quests: Database error/)).toBeInTheDocument();
    });

    it('should show error state when GetQuestHistory fails', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockRejectedValue(new Error('History error'));
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Failed to load quest data' })).toBeInTheDocument();
      });
      expect(screen.getByText(/Failed to load quest history: History error/)).toBeInTheDocument();
    });

    it('should continue loading when GetCurrentAccount fails (account is optional)', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockRejectedValue(new Error('Account error'));

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('No active quests')).toBeInTheDocument();
      });
      // Should not show error since account is optional
      expect(screen.queryByRole('heading', { name: 'Failed to load quest data' })).not.toBeInTheDocument();
    });

    it('should show generic error for non-Error rejections', async () => {
      mockWailsApp.GetActiveQuests.mockRejectedValue('Unknown error');
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Failed to load quest data' })).toBeInTheDocument();
      });
    });
  });

  describe('Empty State', () => {
    it('should show empty state for active quests when none exist', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('No active quests')).toBeInTheDocument();
      });
      expect(screen.getByText("You don't have any active daily quests at the moment.")).toBeInTheDocument();
    });

    it('should show empty state for quest history when none exist', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('No quest history')).toBeInTheDocument();
      });
      expect(screen.getByText('No completed quests found for the selected time period.')).toBeInTheDocument();
    });

    it('should show empty state when API returns null', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue(null);
      mockWailsApp.GetQuestHistory.mockResolvedValue(null);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('No active quests')).toBeInTheDocument();
      });
    });
  });

  describe('Active Quests Display', () => {
    it('should render active quest cards', async () => {
      const quests = [
        createMockQuest({ id: 1, quest_type: 'Quests/Quest_PlayCards', goal: 10, ending_progress: 5 }),
        createMockQuest({ id: 2, quest_type: 'Quests/Quest_WinGames', goal: 5, ending_progress: 2 }),
      ];
      mockWailsApp.GetActiveQuests.mockResolvedValue(quests);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        // Text includes "(500 Gold)" badge, so use partial text match
        expect(screen.getByText(/PlayCards/)).toBeInTheDocument();
      });
      expect(screen.getByText(/WinGames/)).toBeInTheDocument();
    });

    it('should display quest progress', async () => {
      const quest = createMockQuest({ goal: 10, ending_progress: 7 });
      mockWailsApp.GetActiveQuests.mockResolvedValue([quest]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('7 / 10')).toBeInTheDocument();
      });
      expect(screen.getByText('70%')).toBeInTheDocument();
    });

    it('should display 750 gold reward badge', async () => {
      const quest = createMockQuest({ rewards: '750 Gold' });
      mockWailsApp.GetActiveQuests.mockResolvedValue([quest]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText(/750 Gold/)).toBeInTheDocument();
      });
    });

    it('should display 500 gold reward badge', async () => {
      const quest = createMockQuest({ rewards: '500 Gold' });
      mockWailsApp.GetActiveQuests.mockResolvedValue([quest]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText(/500 Gold/)).toBeInTheDocument();
      });
    });

    it('should display assigned date', async () => {
      const quest = createMockQuest({ assigned_at: new Date('2024-01-15T10:00:00').toISOString() });
      mockWailsApp.GetActiveQuests.mockResolvedValue([quest]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText(/Assigned:/)).toBeInTheDocument();
      });
    });

    it('should cap progress at 100%', async () => {
      const quest = createMockQuest({ goal: 10, ending_progress: 15 }); // Over 100%
      mockWailsApp.GetActiveQuests.mockResolvedValue([quest]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('100%')).toBeInTheDocument();
      });
    });

    it('should handle quest with zero goal', async () => {
      const quest = createMockQuest({ goal: 0, ending_progress: 0 });
      mockWailsApp.GetActiveQuests.mockResolvedValue([quest]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('0 / 0')).toBeInTheDocument();
      });
      expect(screen.getByText('0%')).toBeInTheDocument();
    });
  });

  describe('Quest History Display', () => {
    it('should render quest history table', async () => {
      const history = [
        createMockQuest({ id: 1, completed: true, completed_at: new Date('2024-01-16T12:00:00').toISOString() }),
        createMockQuest({ id: 2, completed: false }),
      ];
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue(history);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('Quest History')).toBeInTheDocument();
      });
      expect(screen.getByRole('table')).toBeInTheDocument();
    });

    it('should display completion status badges', async () => {
      const history = [
        createMockQuest({ id: 1, completed: true, completed_at: new Date('2024-01-16T12:00:00').toISOString() }),
        createMockQuest({ id: 2, completed: false }),
      ];
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue(history);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('COMPLETED')).toBeInTheDocument();
      });
      expect(screen.getByText('INCOMPLETE')).toBeInTheDocument();
    });

    it('should display completion duration', async () => {
      const quest = createMockQuest({
        completed: true,
        assigned_at: new Date('2024-01-15T10:00:00').toISOString(),
        completed_at: new Date('2024-01-15T12:30:00').toISOString(), // 2.5 hours later
      });
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([quest]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('2h 30m')).toBeInTheDocument();
      });
    });

    it('should display N/A for incomplete quest duration', async () => {
      const quest = createMockQuest({ completed: false, completed_at: undefined });
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([quest]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('N/A')).toBeInTheDocument();
      });
    });

    it('should display minutes only for short completion time', async () => {
      const quest = createMockQuest({
        completed: true,
        assigned_at: new Date('2024-01-15T10:00:00').toISOString(),
        completed_at: new Date('2024-01-15T10:45:00').toISOString(), // 45 minutes
      });
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([quest]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('45m')).toBeInTheDocument();
      });
    });

    it('should display REROLLED status badge for rerolled quests', async () => {
      const history = [
        createMockQuest({ id: 1, completed: true, completed_at: new Date('2024-01-16T12:00:00').toISOString() }),
        createMockQuest({ id: 2, completed: false, rerolled: true }),
        createMockQuest({ id: 3, completed: false, rerolled: false }),
      ];
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue(history);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('COMPLETED')).toBeInTheDocument();
      });
      expect(screen.getByText('REROLLED')).toBeInTheDocument();
      expect(screen.getByText('INCOMPLETE')).toBeInTheDocument();
    });
  });

  describe('Pagination', () => {
    it('should show pagination when more than 10 history items', async () => {
      const history = Array.from({ length: 15 }, (_, i) =>
        createMockQuest({ id: i + 1, completed: true, completed_at: new Date().toISOString() })
      );
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue(history);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('Page 1 of 2')).toBeInTheDocument();
      });
      expect(screen.getByRole('button', { name: 'First' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Previous' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Next' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Last' })).toBeInTheDocument();
    });

    it('should navigate to next page', async () => {
      const history = Array.from({ length: 15 }, (_, i) =>
        createMockQuest({ id: i + 1, completed: true, completed_at: new Date().toISOString() })
      );
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue(history);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('Page 1 of 2')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: 'Next' }));

      await waitFor(() => {
        expect(screen.getByText('Page 2 of 2')).toBeInTheDocument();
      });
    });

    it('should navigate to last page', async () => {
      const history = Array.from({ length: 25 }, (_, i) =>
        createMockQuest({ id: i + 1, completed: true, completed_at: new Date().toISOString() })
      );
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue(history);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('Page 1 of 3')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: 'Last' }));

      await waitFor(() => {
        expect(screen.getByText('Page 3 of 3')).toBeInTheDocument();
      });
    });

    it('should navigate to previous page', async () => {
      const history = Array.from({ length: 15 }, (_, i) =>
        createMockQuest({ id: i + 1, completed: true, completed_at: new Date().toISOString() })
      );
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue(history);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('Page 1 of 2')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: 'Next' }));
      await waitFor(() => {
        expect(screen.getByText('Page 2 of 2')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: 'Previous' }));
      await waitFor(() => {
        expect(screen.getByText('Page 1 of 2')).toBeInTheDocument();
      });
    });

    it('should navigate to first page', async () => {
      const history = Array.from({ length: 25 }, (_, i) =>
        createMockQuest({ id: i + 1, completed: true, completed_at: new Date().toISOString() })
      );
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue(history);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('Page 1 of 3')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: 'Last' }));
      await waitFor(() => {
        expect(screen.getByText('Page 3 of 3')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: 'First' }));
      await waitFor(() => {
        expect(screen.getByText('Page 1 of 3')).toBeInTheDocument();
      });
    });

    it('should disable First and Previous buttons on first page', async () => {
      const history = Array.from({ length: 15 }, (_, i) =>
        createMockQuest({ id: i + 1, completed: true, completed_at: new Date().toISOString() })
      );
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue(history);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'First' })).toBeDisabled();
        expect(screen.getByRole('button', { name: 'Previous' })).toBeDisabled();
      });
    });

    it('should disable Next and Last buttons on last page', async () => {
      const history = Array.from({ length: 15 }, (_, i) =>
        createMockQuest({ id: i + 1, completed: true, completed_at: new Date().toISOString() })
      );
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue(history);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('Page 1 of 2')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: 'Last' }));

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Next' })).toBeDisabled();
        expect(screen.getByRole('button', { name: 'Last' })).toBeDisabled();
      });
    });

    it('should not show pagination when 10 or fewer items', async () => {
      const history = Array.from({ length: 8 }, (_, i) =>
        createMockQuest({ id: i + 1, completed: true, completed_at: new Date().toISOString() })
      );
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue(history);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByRole('table')).toBeInTheDocument();
      });
      expect(screen.queryByText(/Page \d+ of/)).not.toBeInTheDocument();
    });
  });

  describe('Date Range Filter', () => {
    it('should render date range filter with default value', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        const dateRangeSelect = getSelectByLabel('Date Range');
        expect(dateRangeSelect.value).toBe('30days');
      });
    });

    it('should show custom date inputs when custom range selected', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('Active Quests')).toBeInTheDocument();
      });

      const dateRangeSelect = getSelectByLabel('Date Range');
      fireEvent.change(dateRangeSelect, { target: { value: 'custom' } });

      await waitFor(() => {
        expect(getInputByLabel('Start Date')).toBeInTheDocument();
        expect(getInputByLabel('End Date')).toBeInTheDocument();
      });
    });

    it('should refetch data when date range changes', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(mockWailsApp.GetQuestHistory).toHaveBeenCalled();
      });

      const initialCallCount = mockWailsApp.GetQuestHistory.mock.calls.length;

      const dateRangeSelect = getSelectByLabel('Date Range');
      fireEvent.change(dateRangeSelect, { target: { value: '7days' } });

      await waitFor(() => {
        expect(mockWailsApp.GetQuestHistory.mock.calls.length).toBeGreaterThan(initialCallCount);
      });
    });

    it('should have all date range options', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        const dateRangeSelect = getSelectByLabel('Date Range');
        const options = Array.from(dateRangeSelect.options).map((o) => o.value);
        expect(options).toContain('7days');
        expect(options).toContain('30days');
        expect(options).toContain('90days');
        expect(options).toContain('all');
        expect(options).toContain('custom');
      });
    });
  });

  describe('Win Progress Section', () => {
    it('should display daily wins progress', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(createMockAccount({ DailyWins: 7 }));

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('Daily Wins')).toBeInTheDocument();
      });
      expect(screen.getByText('7 / 15')).toBeInTheDocument();
    });

    it('should display weekly wins progress', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(createMockAccount({ WeeklyWins: 10 }));

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('Weekly Wins')).toBeInTheDocument();
      });
      expect(screen.getByText('10 / 15')).toBeInTheDocument();
    });

    it('should show goal message for under 5 daily wins', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(createMockAccount({ DailyWins: 3 }));

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('Goal: 5 wins for mastery')).toBeInTheDocument();
      });
    });

    it('should show gold reward message for 5+ daily wins', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(createMockAccount({ DailyWins: 6 }));

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('Earn up to 1,250 gold')).toBeInTheDocument();
      });
    });

    it('should not show win progress when no account data', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('No active quests')).toBeInTheDocument();
      });
      expect(screen.queryByText('Win Progress')).not.toBeInTheDocument();
    });
  });

  describe('Mastery Pass Summary', () => {
    it('should display mastery level', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(createMockAccount({ MasteryLevel: 75 }));

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('Mastery Level')).toBeInTheDocument();
      });
      expect(screen.getByText('75')).toBeInTheDocument();
    });

    it('should display pass type', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(createMockAccount({ MasteryPass: 'Premium' }));

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('Pass Type')).toBeInTheDocument();
      });
      expect(screen.getByText('Premium')).toBeInTheDocument();
    });

    it('should display mastery progress percentage', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(
        createMockAccount({ MasteryLevel: 50, MasteryMax: 100 })
      );

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('Progress')).toBeInTheDocument();
      });
      expect(screen.getByText('50.0%')).toBeInTheDocument();
    });

    it('should display N/A for progress when MasteryMax is 0', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(
        createMockAccount({ MasteryLevel: 50, MasteryMax: 0 })
      );

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('Progress')).toBeInTheDocument();
      });
      // Find stat-value containing N/A by looking for multiple N/A elements (one in progress, one in daily goal)
      const naElements = screen.getAllByText('N/A');
      expect(naElements.length).toBeGreaterThan(0);
    });

    it('should display daily goal checkmark when >= 5 wins', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(createMockAccount({ DailyWins: 5, MasteryMax: 100 }));

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('Daily Goal')).toBeInTheDocument();
      });
      // The checkmark should be present
      expect(screen.getByText('✓')).toBeInTheDocument();
    });

    it('should display daily goal progress when < 5 wins', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(createMockAccount({ DailyWins: 3, MasteryMax: 100 }));

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('Daily Goal')).toBeInTheDocument();
      });
      expect(screen.getByText('3/5')).toBeInTheDocument();
    });
  });

  describe('Page Header', () => {
    it('should display page title', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Daily Quests');
      });
    });
  });

  describe('Quest Type Formatting', () => {
    it('should format quest type by removing prefix', async () => {
      const quest = createMockQuest({ quest_type: 'Quests/Quest_CastSpells' });
      mockWailsApp.GetActiveQuests.mockResolvedValue([quest]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        // Text includes gold badge, so use partial match
        expect(screen.getByText(/CastSpells/)).toBeInTheDocument();
      });
    });

    it('should replace underscores with spaces', async () => {
      const quest = createMockQuest({ quest_type: 'Quests/Quest_Kill_Creatures_Black' });
      mockWailsApp.GetActiveQuests.mockResolvedValue([quest]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        // Text includes gold badge, so use partial match
        expect(screen.getByText(/Kill Creatures Black/)).toBeInTheDocument();
      });
    });

    it('should handle empty quest type gracefully', async () => {
      const quest = createMockQuest({ quest_type: '' });
      mockWailsApp.GetActiveQuests.mockResolvedValue([quest]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      // Quest card should still render even with empty quest type
      await waitFor(() => {
        expect(screen.getByText('5 / 10')).toBeInTheDocument(); // progress should be shown
      });
    });
  });

  describe('API Calls', () => {
    it('should call GetQuestHistory with correct date parameters', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(mockWailsApp.GetQuestHistory).toHaveBeenCalled();
      });

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const call = mockWailsApp.GetQuestHistory.mock.calls[0] as any[];
      // Default is 30days, so dates should be strings
      expect(typeof call[0]).toBe('string'); // start date
      expect(typeof call[1]).toBe('string'); // end date
      expect(call[2]).toBe(50); // history limit
    });

    it('should call GetActiveQuests', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(mockWailsApp.GetActiveQuests).toHaveBeenCalled();
      });
    });

    it('should call GetCurrentAccount', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(mockWailsApp.GetCurrentAccount).toHaveBeenCalled();
      });
    });

    it('should pass empty strings for all-time date range', async () => {
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue([]);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.queryByText('Loading quest data...')).not.toBeInTheDocument();
      });

      const dateRangeSelect = getSelectByLabel('Date Range');
      fireEvent.change(dateRangeSelect, { target: { value: 'all' } });

      await waitFor(() => {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const calls = mockWailsApp.GetQuestHistory.mock.calls as any[];
        const lastCall = calls[calls.length - 1];
        expect(lastCall[0]).toBe(''); // empty start date for all time
        expect(lastCall[1]).toBe(''); // empty end date for all time
      });
    });
  });

  describe('Table Headers', () => {
    it('should display all table headers with tooltips', async () => {
      const history = [createMockQuest({ completed: true, completed_at: new Date().toISOString() })];
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue(history);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        // Check for table headers (may have duplicates with filter labels)
        const typeElements = screen.getAllByText('Type');
        expect(typeElements.length).toBeGreaterThanOrEqual(1);
      });
      // Status appears in both filter label and table header
      const statusElements = screen.getAllByText('Status');
      expect(statusElements.length).toBeGreaterThanOrEqual(1);
      expect(screen.getByText('Assigned')).toBeInTheDocument();
      expect(screen.getByText('Progress')).toBeInTheDocument();
      expect(screen.getByText('Duration')).toBeInTheDocument();
    });
  });

  describe('Quest History Filtering', () => {
    it('should display status and type filter controls', async () => {
      const history = [createMockQuest({ completed: true, completed_at: new Date().toISOString() })];
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue(history);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByDisplayValue('All Status')).toBeInTheDocument();
      });
      expect(screen.getByPlaceholderText('Search quests...')).toBeInTheDocument();
    });

    it('should filter by status', async () => {
      const history = [
        createMockQuest({ id: 1, quest_type: 'Quest_Complete', completed: true, completed_at: new Date().toISOString() }),
        createMockQuest({ id: 2, quest_type: 'Quest_Incomplete', completed: false, rerolled: false }),
        createMockQuest({ id: 3, quest_type: 'Quest_Rerolled', completed: false, rerolled: true }),
      ];
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue(history);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('COMPLETED')).toBeInTheDocument();
      });

      // Filter to show only completed
      const statusSelect = screen.getByDisplayValue('All Status');
      fireEvent.change(statusSelect, { target: { value: 'completed' } });

      await waitFor(() => {
        expect(screen.getByText('COMPLETED')).toBeInTheDocument();
        expect(screen.queryByText('INCOMPLETE')).not.toBeInTheDocument();
      });
    });

    it('should filter by quest type search', async () => {
      const history = [
        createMockQuest({ id: 1, quest_type: 'Quests/Quest_Play_Cards', completed: true, completed_at: new Date().toISOString() }),
        createMockQuest({ id: 2, quest_type: 'Quests/Quest_Win_Games', completed: true, completed_at: new Date().toISOString() }),
      ];
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue(history);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('Play Cards')).toBeInTheDocument();
      });
      expect(screen.getByText('Win Games')).toBeInTheDocument();

      // Filter by type
      const typeInput = screen.getByPlaceholderText('Search quests...');
      fireEvent.change(typeInput, { target: { value: 'play' } });

      await waitFor(() => {
        expect(screen.getByText('Play Cards')).toBeInTheDocument();
        expect(screen.queryByText('Win Games')).not.toBeInTheDocument();
      });
    });

    it('should show empty state when filters return no results', async () => {
      const history = [
        createMockQuest({ id: 1, completed: true, completed_at: new Date().toISOString() }),
      ];
      mockWailsApp.GetActiveQuests.mockResolvedValue([]);
      mockWailsApp.GetQuestHistory.mockResolvedValue(history);
      mockWailsApp.GetCurrentAccount.mockResolvedValue(null);

      renderWithProvider(<Quests />);

      await waitFor(() => {
        expect(screen.getByText('COMPLETED')).toBeInTheDocument();
      });

      // Filter to show only incomplete (no results)
      const statusSelect = screen.getByDisplayValue('All Status');
      fireEvent.change(statusSelect, { target: { value: 'incomplete' } });

      await waitFor(() => {
        expect(screen.getByText('No matching quests')).toBeInTheDocument();
      });
    });
  });
});
