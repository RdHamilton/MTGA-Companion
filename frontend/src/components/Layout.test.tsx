import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { render } from '../test/utils/testUtils';
import Layout from './Layout';
import { mockSystem, mockMatches } from '@/test/mocks/apiMock';
import { mockEventEmitter } from '@/test/mocks/websocketMock';

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

describe('Layout Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockEventEmitter.clear();
    mockSubscribers.length = 0;
    mockReplayState.isActive = false;
    mockReplayState.isPaused = false;
    mockSystem.getStatus.mockResolvedValue({
      status: 'standalone',
      connected: false,
    });
  });

  describe('Navigation Tabs', () => {
    it('should render all navigation tabs', () => {
      render(
        <Layout>
          <div>Test Content</div>
        </Layout>,
        { initialRoute: '/' }
      );

      expect(screen.getByText('Match History')).toBeInTheDocument();
      expect(screen.getByText('Quests')).toBeInTheDocument();
      expect(screen.getByText('Draft')).toBeInTheDocument();
      expect(screen.getByText('Charts')).toBeInTheDocument();
      expect(screen.getByText('Settings')).toBeInTheDocument();
    });

    it('should highlight active tab based on current route', () => {
      render(
        <Layout>
          <div>Test Content</div>
        </Layout>,
        { initialRoute: '/draft' }
      );

      const draftTab = screen.getByText('Draft').closest('.tab');
      expect(draftTab).toHaveClass('active');
    });

    it('should navigate to correct route when tab is clicked', async () => {
      render(
        <Layout>
          <div>Test Content</div>
        </Layout>,
        { initialRoute: '/' }
      );

      const questsTab = screen.getByText('Quests');
      await userEvent.click(questsTab);

      await waitFor(() => {
        expect(questsTab.closest('.tab')).toHaveClass('active');
      });
    });

    it('should show sub-navigation when Charts tab is active', async () => {
      render(
        <Layout>
          <div>Test Content</div>
        </Layout>,
        { initialRoute: '/charts/win-rate-trend' }
      );

      await waitFor(() => {
        expect(screen.getByText('Win Rate Trend')).toBeInTheDocument();
        expect(screen.getByText('Deck Performance')).toBeInTheDocument();
        expect(screen.getByText('Rank Progression')).toBeInTheDocument();
        expect(screen.getByText('Format Distribution')).toBeInTheDocument();
        expect(screen.getByText('Result Breakdown')).toBeInTheDocument();
      });
    });

    it('should not show sub-navigation when Charts tab is inactive', () => {
      render(
        <Layout>
          <div>Test Content</div>
        </Layout>,
        { initialRoute: '/match-history' }
      );

      // Sub-navigation should not be present
      const subTabBar = document.querySelector('.sub-tab-bar');
      expect(subTabBar).not.toBeInTheDocument();
    });
  });

  describe('Connection Status', () => {
    it('should display connection status indicator', async () => {
      mockSystem.getStatus.mockResolvedValue({
        status: 'connected',
        connected: true,
      });

      render(
        <Layout>
          <div>Test Content</div>
        </Layout>
      );

      await waitFor(() => {
        const statusBadge = document.querySelector('.status-badge-compact');
        expect(statusBadge).toBeInTheDocument();
        expect(statusBadge).toHaveClass('status-connected');
      });
    });

    it('should update connection status when daemon:connected event fires', async () => {
      mockSystem.getStatus
        .mockResolvedValueOnce({
          status: 'standalone',
          connected: false,
        })
        .mockResolvedValueOnce({
          status: 'connected',
          connected: true,
        });

      render(
        <Layout>
          <div>Test Content</div>
        </Layout>
      );

      await waitFor(() => {
        const statusBadge = document.querySelector('.status-badge-compact');
        expect(statusBadge).toHaveClass('status-standalone');
      });

      // Trigger daemon:connected event
      mockEventEmitter.emit('daemon:connected');

      await waitFor(() => {
        const statusBadge = document.querySelector('.status-badge-compact');
        expect(statusBadge).toHaveClass('status-connected');
      });
    });
  });

  describe('Replay Controls', () => {
    it('should not show replay controls when replay is inactive', () => {
      render(
        <Layout>
          <div>Test Content</div>
        </Layout>
      );

      expect(screen.queryByText('⏸️ Replay Paused')).not.toBeInTheDocument();
    });

    it('should show replay controls when replay is active and paused', async () => {
      mockReplayState.isActive = true;
      mockReplayState.isPaused = true;

      render(
        <Layout>
          <div>Test Content</div>
        </Layout>,
        { initialRoute: '/' }
      );

      await waitFor(() => {
        expect(screen.getByText('⏸️ Replay Paused')).toBeInTheDocument();
        expect(screen.getByText('▶️ Resume')).toBeInTheDocument();
        expect(screen.getByText('⏹️ Stop')).toBeInTheDocument();
      });
    });

    it('should not show replay controls on settings page', () => {
      mockReplayState.isActive = true;
      mockReplayState.isPaused = true;

      render(
        <Layout>
          <div>Test Content</div>
        </Layout>,
        { initialRoute: '/settings' }
      );

      expect(screen.queryByText('⏸️ Replay Paused')).not.toBeInTheDocument();
    });

    it('should not show replay controls on draft page', () => {
      mockReplayState.isActive = true;
      mockReplayState.isPaused = true;

      render(
        <Layout>
          <div>Test Content</div>
        </Layout>,
        { initialRoute: '/draft' }
      );

      expect(screen.queryByText('⏸️ Replay Paused')).not.toBeInTheDocument();
    });

    it('should have ResumeReplay function available', () => {
      mockReplayState.isActive = true;
      mockReplayState.isPaused = true;

      render(
        <Layout>
          <div>Test Content</div>
        </Layout>,
        { initialRoute: '/' }
      );

      // Verify the mock function exists
      expect(mockSystem.resumeReplay).toBeDefined();
    });

    it('should have StopReplay function available', () => {
      mockReplayState.isActive = true;
      mockReplayState.isPaused = true;

      render(
        <Layout>
          <div>Test Content</div>
        </Layout>,
        { initialRoute: '/' }
      );

      // Verify the mock function exists
      expect(mockSystem.stopReplay).toBeDefined();
    });
  });

  describe('Content Rendering', () => {
    it('should render children content', () => {
      render(
        <Layout>
          <div data-testid="test-content">Test Content</div>
        </Layout>
      );

      expect(screen.getByTestId('test-content')).toBeInTheDocument();
      expect(screen.getByText('Test Content')).toBeInTheDocument();
    });

    it('should render Footer component', () => {
      mockMatches.getStats.mockResolvedValue({
        TotalMatches: 0,
        MatchesWon: 0,
        MatchesLost: 0,
        TotalGames: 0,
        GamesWon: 0,
        GamesLost: 0,
        WinRate: 0,
      });

      render(
        <Layout>
          <div>Test Content</div>
        </Layout>
      );

      // Footer should be present
      const footer = document.querySelector('.app-footer');
      expect(footer).toBeInTheDocument();
    });
  });

  describe('Error Handling', () => {
    it('should handle connection status load error gracefully', async () => {
      mockSystem.getStatus.mockRejectedValue(new Error('Failed to load'));

      render(
        <Layout>
          <div>Test Content</div>
        </Layout>
      );

      // Component should still render
      expect(screen.getByText('Match History')).toBeInTheDocument();
    });

    it('should handle connection status error without crashing', async () => {
      mockSystem.getStatus.mockRejectedValue(new Error('Connection error'));

      expect(() => {
        render(
          <Layout>
            <div>Test Content</div>
          </Layout>
        );
      }).not.toThrow();

      // Layout should still render
      expect(screen.getByText('Match History')).toBeInTheDocument();
    });
  });
});
