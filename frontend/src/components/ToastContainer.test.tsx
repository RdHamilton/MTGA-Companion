import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor, act } from '@testing-library/react';
import { render } from '../test/utils/testUtils';
import ToastContainer, { showToast } from './ToastContainer';
import { mockEventEmitter } from '../test/mocks/wailsRuntime';

// Mock the getReplayState and subscribeToReplayState functions
const mockReplayState = {
  isActive: false,
  isPaused: false,
  progress: null,
};

interface ReplayState {
  isActive: boolean;
  isPaused: boolean;
  progress: null | unknown;
}

const mockSubscribers: Array<(state: ReplayState) => void> = [];

vi.mock('../App', () => ({
  getReplayState: vi.fn(() => mockReplayState),
  subscribeToReplayState: vi.fn((callback) => {
    mockSubscribers.push(callback);
    return () => {
      const index = mockSubscribers.indexOf(callback);
      if (index > -1) {
        mockSubscribers.splice(index, 1);
      }
    };
  }),
}));

describe('ToastContainer Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockEventEmitter.clear();
    mockSubscribers.length = 0;
    mockReplayState.isActive = false;
    mockReplayState.isPaused = false;
  });

  describe('Toast Display', () => {
    it('should not display toasts initially', () => {
      render(<ToastContainer />);

      expect(screen.queryByText(/New match detected/i)).not.toBeInTheDocument();
    });

    it('should display toast when stats:updated event fires', async () => {
      render(<ToastContainer />);

      mockEventEmitter.emit('stats:updated', { matches: 1, games: 2 });

      await waitFor(() => {
        expect(screen.getByText(/New match detected! 1 match, 2 games/i)).toBeInTheDocument();
      });
    });

    it('should display plural form for multiple matches', async () => {
      render(<ToastContainer />);

      mockEventEmitter.emit('stats:updated', { matches: 3, games: 5 });

      await waitFor(() => {
        expect(screen.getByText(/3 matches, 5 games/i)).toBeInTheDocument();
      });
    });

    it('should display rank update toast', async () => {
      render(<ToastContainer />);

      mockEventEmitter.emit('rank:updated', {
        format: 'Standard',
        tier: 'Gold',
        step: '3',
      });

      await waitFor(() => {
        expect(screen.getByText(/Rank updated: Standard Gold 3/i)).toBeInTheDocument();
      });
    });

    it('should display quest completed toast', async () => {
      render(<ToastContainer />);

      mockEventEmitter.emit('quest:updated', { completed: 1, count: 1 });

      await waitFor(() => {
        expect(screen.getByText(/Quest completed!/i)).toBeInTheDocument();
      });
    });

    it('should display multiple quests completed toast', async () => {
      render(<ToastContainer />);

      mockEventEmitter.emit('quest:updated', { completed: 3, count: 3 });

      await waitFor(() => {
        expect(screen.getByText(/Quests completed! \(3\)/i)).toBeInTheDocument();
      });
    });

    it('should display quest updated toast when not completed', async () => {
      render(<ToastContainer />);

      mockEventEmitter.emit('quest:updated', { completed: 0, count: 2 });

      await waitFor(() => {
        expect(screen.getByText(/Quests updated \(2\)/i)).toBeInTheDocument();
      });
    });
  });

  describe('Draft Updates', () => {
    it('should display draft update toast in normal mode', async () => {
      render(<ToastContainer />);

      mockEventEmitter.emit('draft:updated', { count: 1, picks: 5 });

      await waitFor(() => {
        expect(screen.getByText(/Draft session stored! \(1 session, 5 picks\)/i)).toBeInTheDocument();
      });
    });

    it('should display plural form for multiple draft sessions', async () => {
      render(<ToastContainer />);

      mockEventEmitter.emit('draft:updated', { count: 2, picks: 10 });

      await waitFor(() => {
        expect(screen.getByText(/Draft sessions stored! \(2 sessions, 10 picks\)/i)).toBeInTheDocument();
      });
    });

    it('should batch draft updates during replay mode', async () => {
      mockReplayState.isActive = true;

      render(<ToastContainer />);

      // Emit multiple draft updates
      mockEventEmitter.emit('draft:updated', { count: 1, picks: 1 });
      mockEventEmitter.emit('draft:updated', { count: 1, picks: 1 });
      mockEventEmitter.emit('draft:updated', { count: 1, picks: 1 });

      // Should not show individual toasts
      expect(screen.queryByText(/Draft session stored!/i)).not.toBeInTheDocument();

      // Wait for batched toast
      await waitFor(
        () => {
          expect(screen.getByText(/Replay: 3 draft updates processed/i)).toBeInTheDocument();
        },
        { timeout: 3000 }
      );
    });

    it('should clear draft update count when replay stops', async () => {
      mockReplayState.isActive = true;

      render(<ToastContainer />);

      // Emit draft updates during replay
      mockEventEmitter.emit('draft:updated', { count: 1, picks: 1 });
      mockEventEmitter.emit('draft:updated', { count: 1, picks: 1 });

      // Stop replay
      mockReplayState.isActive = false;
      mockSubscribers.forEach(sub => sub(mockReplayState));

      // New update should show immediately
      mockEventEmitter.emit('draft:updated', { count: 1, picks: 5 });

      await waitFor(() => {
        expect(screen.getByText(/Draft session stored! \(1 session, 5 picks\)/i)).toBeInTheDocument();
      });
    });
  });

  describe('Toast Types', () => {
    it('should display success toast for match updates', async () => {
      render(<ToastContainer />);

      mockEventEmitter.emit('stats:updated', { matches: 1, games: 1 });

      await waitFor(() => {
        const toast = screen.getByText(/New match detected/i).closest('.toast');
        expect(toast).toHaveClass('toast-success');
      });
    });

    it('should display info toast for rank updates', async () => {
      render(<ToastContainer />);

      mockEventEmitter.emit('rank:updated', {
        format: 'Standard',
        tier: 'Gold',
        step: '3',
      });

      await waitFor(() => {
        const toast = screen.getByText(/Rank updated/i).closest('.toast');
        expect(toast).toHaveClass('toast-info');
      });
    });
  });

  describe('Toast Auto-removal', () => {
    it('should display multiple toasts simultaneously', async () => {
      render(<ToastContainer />);

      mockEventEmitter.emit('stats:updated', { matches: 1, games: 1 });

      await waitFor(() => {
        expect(screen.getByText(/New match detected/i)).toBeInTheDocument();
      });

      mockEventEmitter.emit('rank:updated', {
        format: 'Standard',
        tier: 'Gold',
        step: '3',
      });

      await waitFor(() => {
        expect(screen.getByText(/Rank updated/i)).toBeInTheDocument();
      });

      // Both toasts should be visible
      expect(screen.getByText(/New match detected/i)).toBeInTheDocument();
      expect(screen.getByText(/Rank updated/i)).toBeInTheDocument();
    });
  });

  describe('Toast Position', () => {
    it('should render toasts in a container', () => {
      render(<ToastContainer />);

      // Container should exist
      const container = document.querySelector('[style*="position: fixed"]');
      expect(container).toBeInTheDocument();
    });
  });

  describe('Cleanup', () => {
    it('should not crash on unmount', () => {
      const { unmount } = render(<ToastContainer />);

      // Should unmount without errors
      expect(() => unmount()).not.toThrow();
    });
  });
});
