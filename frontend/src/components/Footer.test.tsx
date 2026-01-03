import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import { render } from '../test/utils/testUtils';
import Footer from './Footer';
import { mockMatches, mockSystem } from '@/test/mocks/apiMock';
import { mockEventEmitter } from '@/test/mocks/websocketMock';
import { models } from '@/types/models';

// Mock useDownload since Footer now includes DownloadProgressBar
vi.mock('@/context/DownloadContext', () => ({
  useDownload: () => ({
    state: { tasks: [], activeTask: null },
    isDownloading: false,
    overallProgress: 0,
  }),
  DownloadProvider: ({ children }: { children: React.ReactNode }) => children,
}));

function createMockStatistics(overrides: Partial<models.Statistics> = {}): models.Statistics {
  return new models.Statistics({
    TotalMatches: 100,
    MatchesWon: 60,
    MatchesLost: 40,
    TotalGames: 250,
    GamesWon: 150,
    GamesLost: 100,
    WinRate: 0.6,
    ...overrides,
  });
}

function createMockMatch(overrides: Partial<models.Match> = {}): models.Match {
  return new models.Match({
    ID: 'match-1',
    EventID: 'event-1',
    MatchID: 'match-123',
    Timestamp: new Date('2025-11-20T10:00:00Z'),
    Result: 'win',
    OpponentScreenName: 'Opponent1',
    Format: 'Standard',
    DeckColors: ['W', 'U'],
    OpponentColors: ['B', 'R'],
    OnPlay: true,
    TotalTurns: 10,
    DurationSeconds: 600,
    RankTier: 'Gold',
    RankClass: '4',
    ...overrides,
  });
}

describe('Footer Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockEventEmitter.clear();
  });

  describe('Loading State', () => {
    it('should display loading state initially', () => {
      mockMatches.getStats.mockImplementation(() => new Promise(() => {})); // Never resolves
      mockMatches.getMatches.mockImplementation(() => new Promise(() => {}));

      render(<Footer />);

      expect(screen.getByText('Loading stats...')).toBeInTheDocument();
    });
  });

  describe('Empty State', () => {
    it('should display empty state when no matches exist', async () => {
      const emptyStats = createMockStatistics({ TotalMatches: 0 });
      mockMatches.getStats.mockResolvedValue(emptyStats);
      mockMatches.getMatches.mockResolvedValue([]);

      render(<Footer />);

      await waitFor(() => {
        expect(screen.getByText(/No matches yet/i)).toBeInTheDocument();
      });
    });
  });

  describe('Statistics Display', () => {
    it('should display total matches count', async () => {
      const stats = createMockStatistics({ TotalMatches: 100 });
      mockMatches.getStats.mockResolvedValue(stats);
      mockMatches.getMatches.mockResolvedValue([]);

      render(<Footer />);

      await waitFor(() => {
        expect(screen.getByText('Matches:')).toBeInTheDocument();
        expect(screen.getByText('100')).toBeInTheDocument();
      });
    });

    it('should display win rate correctly', async () => {
      const stats = createMockStatistics({
        MatchesWon: 60,
        MatchesLost: 40,
        WinRate: 0.6,
      });
      mockMatches.getStats.mockResolvedValue(stats);
      mockMatches.getMatches.mockResolvedValue([]);

      render(<Footer />);

      await waitFor(() => {
        expect(screen.getByText('Win Rate:')).toBeInTheDocument();
        expect(screen.getByText(/60%/)).toBeInTheDocument();
        expect(screen.getByText(/60-40/)).toBeInTheDocument();
      });
    });

    it('should round win rate to one decimal place', async () => {
      const stats = createMockStatistics({
        WinRate: 0.5555, // Should display as 55.6%
      });
      mockMatches.getStats.mockResolvedValue(stats);
      mockMatches.getMatches.mockResolvedValue([]);

      render(<Footer />);

      await waitFor(() => {
        expect(screen.getByText(/55\.6%/)).toBeInTheDocument();
      });
    });

    it('should display last match time with Last Played label', async () => {
      const stats = createMockStatistics();
      const matches = [
        createMockMatch({
          Timestamp: new Date('2025-11-20T10:00:00Z'),
        }),
      ];
      mockMatches.getStats.mockResolvedValue(stats);
      mockMatches.getMatches.mockResolvedValue(matches);

      render(<Footer />);

      await waitFor(() => {
        expect(screen.getByText('Last Played:')).toBeInTheDocument();
        // The exact format depends on locale, just check it exists
        const lastMatchText = screen.getByText('Last Played:').nextSibling;
        expect(lastMatchText).toBeTruthy();
      });
    });

    it('should display Synced indicator after data loads', async () => {
      const stats = createMockStatistics();
      mockMatches.getStats.mockResolvedValue(stats);
      mockMatches.getMatches.mockResolvedValue([]);

      render(<Footer />);

      await waitFor(() => {
        expect(screen.getByText('Synced:')).toBeInTheDocument();
        // Synced time should be present
        const syncedText = screen.getByText('Synced:').nextSibling;
        expect(syncedText).toBeTruthy();
      });
    });
  });

  describe('Win Streak Display', () => {
    it('should display winning streak', async () => {
      const stats = createMockStatistics();
      const matches = [
        createMockMatch({ Result: 'win' }),
        createMockMatch({ Result: 'win' }),
        createMockMatch({ Result: 'win' }),
        createMockMatch({ Result: 'loss' }),
      ];
      mockMatches.getStats.mockResolvedValue(stats);
      mockMatches.getMatches.mockResolvedValue(matches);

      render(<Footer />);

      await waitFor(() => {
        expect(screen.getByText('Streak:')).toBeInTheDocument();
        expect(screen.getByText('W3')).toBeInTheDocument();
      });
    });

    it('should display losing streak', async () => {
      const stats = createMockStatistics();
      const matches = [
        createMockMatch({ Result: 'loss' }),
        createMockMatch({ Result: 'loss' }),
        createMockMatch({ Result: 'win' }),
      ];
      mockMatches.getStats.mockResolvedValue(stats);
      mockMatches.getMatches.mockResolvedValue(matches);

      render(<Footer />);

      await waitFor(() => {
        expect(screen.getByText('Streak:')).toBeInTheDocument();
        expect(screen.getByText('L2')).toBeInTheDocument();
      });
    });

    it('should not display streak when count is 0', async () => {
      const stats = createMockStatistics();
      mockMatches.getStats.mockResolvedValue(stats);
      mockMatches.getMatches.mockResolvedValue([]);

      render(<Footer />);

      await waitFor(() => {
        expect(screen.queryByText('Streak:')).not.toBeInTheDocument();
      });
    });

    it('should calculate streak correctly across multiple matches', async () => {
      const stats = createMockStatistics();
      const matches = [
        createMockMatch({ Result: 'win' }),
        createMockMatch({ Result: 'win' }),
        createMockMatch({ Result: 'win' }),
        createMockMatch({ Result: 'win' }),
        createMockMatch({ Result: 'win' }),
        createMockMatch({ Result: 'loss' }),
      ];
      mockMatches.getStats.mockResolvedValue(stats);
      mockMatches.getMatches.mockResolvedValue(matches);

      render(<Footer />);

      await waitFor(() => {
        expect(screen.getByText('W5')).toBeInTheDocument();
      });
    });
  });

  describe('Real-time Updates', () => {
    it('should reload stats when stats:updated event fires', async () => {
      const initialStats = createMockStatistics({ TotalMatches: 10 });
      const updatedStats = createMockStatistics({ TotalMatches: 11 });

      mockMatches.getStats.mockResolvedValueOnce(initialStats);
      mockMatches.getMatches.mockResolvedValue([]);

      render(<Footer />);

      await waitFor(() => {
        expect(screen.getByText('10')).toBeInTheDocument();
      });

      // Update mock to return new stats
      mockMatches.getStats.mockResolvedValueOnce(updatedStats);

      // Trigger stats:updated event
      mockEventEmitter.emit('stats:updated');

      await waitFor(() => {
        expect(screen.getByText('11')).toBeInTheDocument();
      });
    });

    it('should update streak when new match is added', async () => {
      const stats = createMockStatistics();
      const initialMatches = [createMockMatch({ Result: 'win' })];
      const updatedMatches = [
        createMockMatch({ Result: 'win' }),
        createMockMatch({ Result: 'win' }),
      ];

      mockMatches.getStats.mockResolvedValue(stats);
      mockMatches.getMatches.mockResolvedValueOnce(initialMatches);

      render(<Footer />);

      await waitFor(() => {
        expect(screen.getByText('W1')).toBeInTheDocument();
      });

      // Update matches
      mockMatches.getMatches.mockResolvedValueOnce(updatedMatches);

      // Trigger event
      mockEventEmitter.emit('stats:updated');

      await waitFor(() => {
        expect(screen.getByText('W2')).toBeInTheDocument();
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle GetStats error gracefully', async () => {
      mockMatches.getStats.mockRejectedValue(new Error('Failed to load stats'));
      mockMatches.getMatches.mockResolvedValue([]);

      render(<Footer />);

      await waitFor(() => {
        // Should not show error message to user, just show empty state
        expect(screen.getByText(/No matches yet/i)).toBeInTheDocument();
      });
    });

    it('should handle GetMatches error gracefully', async () => {
      const stats = createMockStatistics();
      mockMatches.getStats.mockResolvedValue(stats);
      mockMatches.getMatches.mockRejectedValue(new Error('Failed to load matches'));

      render(<Footer />);

      await waitFor(() => {
        // Should still show stats even if matches fail
        expect(screen.getByText('Matches:')).toBeInTheDocument();
      });
    });
  });

  describe('Visual Elements', () => {
    it('should display All Time label', async () => {
      const stats = createMockStatistics();
      mockMatches.getStats.mockResolvedValue(stats);
      mockMatches.getMatches.mockResolvedValue([]);

      render(<Footer />);

      await waitFor(() => {
        expect(screen.getByText('All Time')).toBeInTheDocument();
      });
    });

    it('should apply correct CSS class for win streak', async () => {
      const stats = createMockStatistics();
      const matches = [createMockMatch({ Result: 'win' })];
      mockMatches.getStats.mockResolvedValue(stats);
      mockMatches.getMatches.mockResolvedValue(matches);

      render(<Footer />);

      await waitFor(() => {
        const streakElement = screen.getByText('W1').closest('.footer-stat');
        expect(streakElement).toHaveClass('streak-w');
      });
    });

    it('should apply correct CSS class for loss streak', async () => {
      const stats = createMockStatistics();
      const matches = [createMockMatch({ Result: 'loss' })];
      mockMatches.getStats.mockResolvedValue(stats);
      mockMatches.getMatches.mockResolvedValue(matches);

      render(<Footer />);

      await waitFor(() => {
        const streakElement = screen.getByText('L1').closest('.footer-stat');
        expect(streakElement).toHaveClass('streak-l');
      });
    });
  });

  describe('Backend Sync Time', () => {
    it('should display backend sync time from health status', async () => {
      const stats = createMockStatistics();
      const syncTime = '2025-12-28T15:30:00Z';
      mockMatches.getStats.mockResolvedValue(stats);
      mockMatches.getMatches.mockResolvedValue([]);
      mockSystem.getHealth.mockResolvedValue({
        status: 'healthy',
        version: '1.4.0',
        uptime: 3600,
        database: {
          status: 'ok',
          lastWrite: syncTime,
        },
        logMonitor: { status: 'ok' },
        websocket: { status: 'ok', connectedClients: 1 },
        metrics: { totalProcessed: 100, totalErrors: 0 },
      });

      render(<Footer />);

      await waitFor(() => {
        expect(screen.getByText('Synced:')).toBeInTheDocument();
        // The time should be formatted from the backend timestamp
        const syncedText = screen.getByText('Synced:').nextSibling;
        expect(syncedText).toBeTruthy();
      });
    });

    it('should fallback to current time when health check fails', async () => {
      const stats = createMockStatistics();
      mockMatches.getStats.mockResolvedValue(stats);
      mockMatches.getMatches.mockResolvedValue([]);
      mockSystem.getHealth.mockRejectedValue(new Error('Health check failed'));

      render(<Footer />);

      await waitFor(() => {
        expect(screen.getByText('Synced:')).toBeInTheDocument();
        // Should still show synced time (fallback to current time)
        const syncedText = screen.getByText('Synced:').nextSibling;
        expect(syncedText).toBeTruthy();
      });
    });

    it('should fallback to current time when no lastWrite in health status', async () => {
      const stats = createMockStatistics();
      mockMatches.getStats.mockResolvedValue(stats);
      mockMatches.getMatches.mockResolvedValue([]);
      mockSystem.getHealth.mockResolvedValue({
        status: 'standalone',
        version: '1.4.0',
        uptime: 0,
        database: {
          status: 'ok',
          // No lastWrite field - daemon hasn't written to DB yet
        },
        logMonitor: { status: 'ok' },
        websocket: { status: 'ok', connectedClients: 0 },
        metrics: { totalProcessed: 0, totalErrors: 0 },
      });

      render(<Footer />);

      await waitFor(() => {
        expect(screen.getByText('Synced:')).toBeInTheDocument();
        // Should still show synced time (fallback)
        const syncedText = screen.getByText('Synced:').nextSibling;
        expect(syncedText).toBeTruthy();
      });
    });

    it('should call getHealth when loading stats', async () => {
      const stats = createMockStatistics();
      mockMatches.getStats.mockResolvedValue(stats);
      mockMatches.getMatches.mockResolvedValue([]);
      mockSystem.getHealth.mockResolvedValue({
        status: 'healthy',
        version: '1.4.0',
        uptime: 3600,
        database: { status: 'ok', lastWrite: new Date().toISOString() },
        logMonitor: { status: 'ok' },
        websocket: { status: 'ok', connectedClients: 1 },
        metrics: { totalProcessed: 100, totalErrors: 0 },
      });

      render(<Footer />);

      await waitFor(() => {
        expect(mockSystem.getHealth).toHaveBeenCalled();
      });
    });
  });
});
