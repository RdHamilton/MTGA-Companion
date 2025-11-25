import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { screen, waitFor, act, render } from '@testing-library/react';
import App from './App';
import { mockWailsApp } from './test/mocks/wailsApp';
import { mockEventEmitter } from './test/mocks/wailsRuntime';
import { resetReplayState, getReplayState } from './utils/replayState';
import { gui } from '../wailsjs/go/models';

// Helper to set initial route before rendering App (which has its own Router)
function renderAppWithRoute(route: string = '/') {
  window.history.pushState({}, 'Test page', route);
  return render(<App />);
}

// Mock child components to simplify testing
vi.mock('./components/Layout', () => ({
  default: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="layout">{children}</div>
  ),
}));

vi.mock('./components/ToastContainer', () => ({
  default: () => <div data-testid="toast-container" />,
}));

vi.mock('./components/KeyboardShortcutsHandler', () => ({
  default: () => <div data-testid="keyboard-shortcuts-handler" />,
}));

vi.mock('./pages/MatchHistory', () => ({
  default: () => <div data-testid="match-history-page">Match History</div>,
}));

vi.mock('./pages/Quests', () => ({
  default: () => <div data-testid="quests-page">Quests</div>,
}));

vi.mock('./pages/Draft', () => ({
  default: () => <div data-testid="draft-page">Draft</div>,
}));

vi.mock('./pages/Decks', () => ({
  default: () => <div data-testid="decks-page">Decks</div>,
}));

vi.mock('./pages/DeckBuilder', () => ({
  default: () => <div data-testid="deck-builder-page">Deck Builder</div>,
}));

vi.mock('./pages/Settings', () => ({
  default: () => <div data-testid="settings-page">Settings</div>,
}));

vi.mock('./pages/WinRateTrend', () => ({
  default: () => <div data-testid="win-rate-trend-page">Win Rate Trend</div>,
}));

vi.mock('./pages/DeckPerformance', () => ({
  default: () => <div data-testid="deck-performance-page">Deck Performance</div>,
}));

vi.mock('./pages/RankProgression', () => ({
  default: () => <div data-testid="rank-progression-page">Rank Progression</div>,
}));

vi.mock('./pages/FormatDistribution', () => ({
  default: () => <div data-testid="format-distribution-page">Format Distribution</div>,
}));

vi.mock('./pages/ResultBreakdown', () => ({
  default: () => <div data-testid="result-breakdown-page">Result Breakdown</div>,
}));

describe('App', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockEventEmitter.clear();
    resetReplayState();

    // Setup default mocks
    mockWailsApp.GetStats.mockResolvedValue({
      TotalMatches: 0,
      MatchesWon: 0,
      MatchesLost: 0,
      TotalGames: 0,
      GamesWon: 0,
      GamesLost: 0,
      WinRate: 0,
    });
    mockWailsApp.GetConnectionStatus.mockResolvedValue({
      status: 'standalone',
      connected: false,
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('Routing', () => {
    it('should redirect root to /match-history', async () => {
      renderAppWithRoute('/');

      await waitFor(() => {
        expect(screen.getByTestId('match-history-page')).toBeInTheDocument();
      });
    });

    it('should render MatchHistory page at /match-history', async () => {
      renderAppWithRoute('/match-history');

      await waitFor(() => {
        expect(screen.getByTestId('match-history-page')).toBeInTheDocument();
      });
    });

    it('should render Quests page at /quests', async () => {
      renderAppWithRoute('/quests');

      await waitFor(() => {
        expect(screen.getByTestId('quests-page')).toBeInTheDocument();
      });
    });

    it('should render Draft page at /draft', async () => {
      renderAppWithRoute('/draft');

      await waitFor(() => {
        expect(screen.getByTestId('draft-page')).toBeInTheDocument();
      });
    });

    it('should render Decks page at /decks', async () => {
      renderAppWithRoute('/decks');

      await waitFor(() => {
        expect(screen.getByTestId('decks-page')).toBeInTheDocument();
      });
    });

    it('should render DeckBuilder page at /deck-builder/:deckID', async () => {
      renderAppWithRoute('/deck-builder/123');

      await waitFor(() => {
        expect(screen.getByTestId('deck-builder-page')).toBeInTheDocument();
      });
    });

    it('should render DeckBuilder page at /deck-builder/draft/:draftEventID', async () => {
      renderAppWithRoute('/deck-builder/draft/456');

      await waitFor(() => {
        expect(screen.getByTestId('deck-builder-page')).toBeInTheDocument();
      });
    });

    it('should render Settings page at /settings', async () => {
      renderAppWithRoute('/settings');

      await waitFor(() => {
        expect(screen.getByTestId('settings-page')).toBeInTheDocument();
      });
    });

    it('should render WinRateTrend chart at /charts/win-rate-trend', async () => {
      renderAppWithRoute('/charts/win-rate-trend');

      await waitFor(() => {
        expect(screen.getByTestId('win-rate-trend-page')).toBeInTheDocument();
      });
    });

    it('should render DeckPerformance chart at /charts/deck-performance', async () => {
      renderAppWithRoute('/charts/deck-performance');

      await waitFor(() => {
        expect(screen.getByTestId('deck-performance-page')).toBeInTheDocument();
      });
    });

    it('should render RankProgression chart at /charts/rank-progression', async () => {
      renderAppWithRoute('/charts/rank-progression');

      await waitFor(() => {
        expect(screen.getByTestId('rank-progression-page')).toBeInTheDocument();
      });
    });

    it('should render FormatDistribution chart at /charts/format-distribution', async () => {
      renderAppWithRoute('/charts/format-distribution');

      await waitFor(() => {
        expect(screen.getByTestId('format-distribution-page')).toBeInTheDocument();
      });
    });

    it('should render ResultBreakdown chart at /charts/result-breakdown', async () => {
      renderAppWithRoute('/charts/result-breakdown');

      await waitFor(() => {
        expect(screen.getByTestId('result-breakdown-page')).toBeInTheDocument();
      });
    });
  });

  describe('Component Structure', () => {
    it('should render Layout component', async () => {
      renderAppWithRoute('/');

      await waitFor(() => {
        expect(screen.getByTestId('layout')).toBeInTheDocument();
      });
    });

    it('should render ToastContainer component', async () => {
      renderAppWithRoute('/');

      await waitFor(() => {
        expect(screen.getByTestId('toast-container')).toBeInTheDocument();
      });
    });

    it('should render KeyboardShortcutsHandler component', async () => {
      renderAppWithRoute('/');

      await waitFor(() => {
        expect(screen.getByTestId('keyboard-shortcuts-handler')).toBeInTheDocument();
      });
    });
  });

  describe('Replay Event Handling', () => {
    it('should update replay state when replay:started event fires', async () => {
      renderAppWithRoute('/');

      const replayStatus = new gui.ReplayStatus({
        isActive: true,
        isPaused: false,
        currentEntry: 0,
        totalEntries: 100,
        percentComplete: 0,
        elapsed: '00:00:00',
        speed: 1.0,
        filter: 'all',
      });

      await act(async () => {
        mockEventEmitter.emit('replay:started', replayStatus);
      });

      const state = getReplayState();
      expect(state.isActive).toBe(true);
      expect(state.isPaused).toBe(false);
    });

    it('should update replay state when replay:progress event fires', async () => {
      renderAppWithRoute('/');

      const replayStatus = new gui.ReplayStatus({
        isActive: true,
        isPaused: false,
        currentEntry: 50,
        totalEntries: 100,
        percentComplete: 50,
        elapsed: '00:02:30',
        speed: 1.0,
        filter: 'all',
      });

      await act(async () => {
        mockEventEmitter.emit('replay:progress', replayStatus);
      });

      const state = getReplayState();
      expect(state.progress).toBeDefined();
    });

    it('should update replay state when replay:paused event fires', async () => {
      renderAppWithRoute('/');

      // First start the replay
      await act(async () => {
        mockEventEmitter.emit('replay:started', new gui.ReplayStatus({
          isActive: true,
          isPaused: false,
          currentEntry: 0,
          totalEntries: 100,
          percentComplete: 0,
          elapsed: '00:00:00',
          speed: 1.0,
          filter: 'all',
        }));
      });

      // Then pause it
      await act(async () => {
        mockEventEmitter.emit('replay:paused', new gui.ReplayStatus({
          isActive: true,
          isPaused: true,
          currentEntry: 50,
          totalEntries: 100,
          percentComplete: 50,
          elapsed: '00:02:30',
          speed: 1.0,
          filter: 'all',
        }));
      });

      const state = getReplayState();
      expect(state.isPaused).toBe(true);
    });

    it('should update replay state when replay:resumed event fires', async () => {
      renderAppWithRoute('/');

      // First start and pause
      await act(async () => {
        mockEventEmitter.emit('replay:started', new gui.ReplayStatus({
          isActive: true,
          isPaused: false,
          currentEntry: 0,
          totalEntries: 100,
          percentComplete: 0,
          elapsed: '00:00:00',
          speed: 1.0,
          filter: 'all',
        }));
      });

      await act(async () => {
        mockEventEmitter.emit('replay:paused', new gui.ReplayStatus({
          isActive: true,
          isPaused: true,
          currentEntry: 50,
          totalEntries: 100,
          percentComplete: 50,
          elapsed: '00:02:30',
          speed: 1.0,
          filter: 'all',
        }));
      });

      // Then resume
      await act(async () => {
        mockEventEmitter.emit('replay:resumed', new gui.ReplayStatus({
          isActive: true,
          isPaused: false,
          currentEntry: 50,
          totalEntries: 100,
          percentComplete: 50,
          elapsed: '00:02:30',
          speed: 1.0,
          filter: 'all',
        }));
      });

      const state = getReplayState();
      expect(state.isPaused).toBe(false);
    });

    it('should update replay state when replay:completed event fires', async () => {
      renderAppWithRoute('/');

      // First start the replay
      await act(async () => {
        mockEventEmitter.emit('replay:started', new gui.ReplayStatus({
          isActive: true,
          isPaused: false,
          currentEntry: 0,
          totalEntries: 100,
          percentComplete: 0,
          elapsed: '00:00:00',
          speed: 1.0,
          filter: 'all',
        }));
      });

      expect(getReplayState().isActive).toBe(true);

      // Then complete it
      await act(async () => {
        mockEventEmitter.emit('replay:completed', new gui.ReplayStatus({
          isActive: false,
          isPaused: false,
          currentEntry: 100,
          totalEntries: 100,
          percentComplete: 100,
          elapsed: '00:05:00',
          speed: 1.0,
          filter: 'all',
        }));
      });

      const state = getReplayState();
      expect(state.isActive).toBe(false);
      expect(state.isPaused).toBe(false);
    });

    it('should update replay state when replay:error event fires', async () => {
      renderAppWithRoute('/');

      // First start the replay
      await act(async () => {
        mockEventEmitter.emit('replay:started', new gui.ReplayStatus({
          isActive: true,
          isPaused: false,
          currentEntry: 0,
          totalEntries: 100,
          percentComplete: 0,
          elapsed: '00:00:00',
          speed: 1.0,
          filter: 'all',
        }));
      });

      expect(getReplayState().isActive).toBe(true);

      // Then trigger an error
      await act(async () => {
        mockEventEmitter.emit('replay:error', {
          error: 'Test error',
          code: 'ERR_TEST',
          details: 'Test error details',
        });
      });

      const state = getReplayState();
      expect(state.isActive).toBe(false);
      expect(state.isPaused).toBe(false);
    });

    it('should navigate to /draft when replay:draft_detected event fires', async () => {
      renderAppWithRoute('/match-history');

      await waitFor(() => {
        expect(screen.getByTestId('match-history-page')).toBeInTheDocument();
      });

      await act(async () => {
        mockEventEmitter.emit('replay:draft_detected', {
          draftId: 'draft-123',
          setCode: 'DSK',
          eventType: 'PremierDraft',
        });
      });

      await waitFor(() => {
        expect(screen.getByTestId('draft-page')).toBeInTheDocument();
      });
    });
  });

  describe('Exports', () => {
    it('should export getReplayState', async () => {
      const { getReplayState: exportedGetReplayState } = await import('./App');
      expect(typeof exportedGetReplayState).toBe('function');
    });

    it('should export subscribeToReplayState', async () => {
      const { subscribeToReplayState: exportedSubscribeToReplayState } = await import('./App');
      expect(typeof exportedSubscribeToReplayState).toBe('function');
    });
  });
});
