import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { render } from '../test/utils/testUtils';
import CurrentPackPicker from './CurrentPackPicker';
import { mockWailsApp } from '@/test/mocks/apiMock';
import { gui } from '@/types/models';

function createMockPackCard(overrides: Partial<gui.PackCardWithRating> = {}): gui.PackCardWithRating {
  return new gui.PackCardWithRating({
    arena_id: '12345',
    name: 'Lightning Bolt',
    image_url: 'https://example.com/bolt.jpg',
    rarity: 'common',
    colors: ['R'],
    mana_cost: '{R}',
    cmc: 1,
    type_line: 'Instant',
    gihwr: 58.5,
    alsa: 3.2,
    tier: 'S',
    is_recommended: false,
    score: 0.85,
    reasoning: 'This card high win rate card.',
    ...overrides,
  });
}

function createMockPackResponse(overrides: Partial<gui.CurrentPackResponse> = {}): gui.CurrentPackResponse {
  return new gui.CurrentPackResponse({
    session_id: 'session-123',
    pack_number: 0,
    pick_number: 0,
    pack_label: 'Pack 1, Pick 1',
    cards: [
      createMockPackCard({ arena_id: '1', name: 'Lightning Bolt', score: 0.9, is_recommended: true }),
      createMockPackCard({ arena_id: '2', name: 'Counterspell', score: 0.8, colors: ['U'] }),
      createMockPackCard({ arena_id: '3', name: 'Llanowar Elves', score: 0.7, colors: ['G'] }),
    ],
    recommended_card: createMockPackCard({ arena_id: '1', name: 'Lightning Bolt', score: 0.9, is_recommended: true }),
    pool_colors: [],
    pool_size: 0,
    ...overrides,
  });
}

describe('CurrentPackPicker Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Loading State', () => {
    it('should show loading state initially', () => {
      mockWailsApp.GetCurrentPackWithRecommendation.mockImplementation(() => new Promise(() => {}));

      render(<CurrentPackPicker sessionID="test-session" />);

      expect(screen.getByText('Loading current pack...')).toBeInTheDocument();
    });
  });

  describe('Error State', () => {
    it('should show error message when loading fails', async () => {
      mockWailsApp.GetCurrentPackWithRecommendation.mockRejectedValue(new Error('Network error'));

      render(<CurrentPackPicker sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('Network error')).toBeInTheDocument();
      });
    });

    it('should show retry button when error occurs', async () => {
      mockWailsApp.GetCurrentPackWithRecommendation.mockRejectedValue(new Error('Network error'));

      render(<CurrentPackPicker sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /Retry/i })).toBeInTheDocument();
      });
    });

    it('should reload data when retry button is clicked', async () => {
      mockWailsApp.GetCurrentPackWithRecommendation.mockRejectedValueOnce(new Error('Network error'));
      const packData = createMockPackResponse();
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValueOnce(packData);

      render(<CurrentPackPicker sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /Retry/i })).toBeInTheDocument();
      });

      await userEvent.click(screen.getByRole('button', { name: /Retry/i }));

      await waitFor(() => {
        expect(screen.getByText('Pack 1, Pick 1')).toBeInTheDocument();
      });
    });
  });

  describe('Empty State', () => {
    it('should show empty state when no pack data available', async () => {
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValue(null);

      render(<CurrentPackPicker sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('No pack data available')).toBeInTheDocument();
      });
    });

    it('should show help text when no pack data', async () => {
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValue(null);

      render(<CurrentPackPicker sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('Pack data will appear when you start a draft pick')).toBeInTheDocument();
      });
    });
  });

  describe('Display Pack Data', () => {
    it('should display pack label', async () => {
      const packData = createMockPackResponse({ pack_label: 'Pack 2, Pick 5' });
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValue(packData);

      render(<CurrentPackPicker sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('Pack 2, Pick 5')).toBeInTheDocument();
      });
    });

    it('should display pool size', async () => {
      const packData = createMockPackResponse({ pool_size: 10 });
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValue(packData);

      render(<CurrentPackPicker sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('Pool: 10 cards')).toBeInTheDocument();
      });
    });

    it('should display cards in the pack', async () => {
      const packData = createMockPackResponse();
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValue(packData);

      render(<CurrentPackPicker sessionID="test-session" />);

      await waitFor(() => {
        // Lightning Bolt appears twice (in banner and grid)
        expect(screen.getAllByText('Lightning Bolt').length).toBeGreaterThanOrEqual(1);
        expect(screen.getByText('Counterspell')).toBeInTheDocument();
        expect(screen.getByText('Llanowar Elves')).toBeInTheDocument();
      });
    });

    it('should display tier badges for each card', async () => {
      const packData = createMockPackResponse({
        cards: [
          createMockPackCard({ arena_id: '1', name: 'Card A', tier: 'S' }),
          createMockPackCard({ arena_id: '2', name: 'Card B', tier: 'A' }),
          createMockPackCard({ arena_id: '3', name: 'Card C', tier: 'B' }),
        ],
      });
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValue(packData);

      render(<CurrentPackPicker sessionID="test-session" />);

      await waitFor(() => {
        const tierBadges = screen.getAllByText(/^[SABCDF]$/);
        expect(tierBadges.length).toBeGreaterThan(0);
      });
    });
  });

  describe('Recommended Pick', () => {
    it('should display recommended card banner', async () => {
      const packData = createMockPackResponse();
      // Verify the recommended_card is set
      expect(packData.recommended_card).toBeDefined();
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValue(packData);

      render(<CurrentPackPicker sessionID="test-session" />);

      await waitFor(() => {
        // Check that the recommended banner shows up
        const banner = screen.getByText('Recommended Pick:');
        expect(banner).toBeInTheDocument();
      });
    });

    it('should highlight the recommended card in the grid', async () => {
      const packData = createMockPackResponse();
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValue(packData);

      const { container } = render(<CurrentPackPicker sessionID="test-session" />);

      await waitFor(() => {
        const recommendedCard = container.querySelector('.pack-card.recommended');
        expect(recommendedCard).toBeInTheDocument();
      });
    });

    it('should display Best Pick indicator on recommended card', async () => {
      const packData = createMockPackResponse();
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValue(packData);

      render(<CurrentPackPicker sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('Best Pick')).toBeInTheDocument();
      });
    });
  });

  describe('Refresh Functionality', () => {
    it('should have a refresh button', async () => {
      const packData = createMockPackResponse();
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValue(packData);

      render(<CurrentPackPicker sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /Refresh/i })).toBeInTheDocument();
      });
    });

    it('should reload data when refresh button is clicked', async () => {
      const packData1 = createMockPackResponse({ pack_label: 'Pack 1, Pick 1' });
      const packData2 = createMockPackResponse({ pack_label: 'Pack 1, Pick 2' });
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValueOnce(packData1);
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValueOnce(packData2);

      render(<CurrentPackPicker sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('Pack 1, Pick 1')).toBeInTheDocument();
      });

      await userEvent.click(screen.getByRole('button', { name: /Refresh/i }));

      await waitFor(() => {
        expect(screen.getByText('Pack 1, Pick 2')).toBeInTheDocument();
      });
    });

    it('should call onRefresh callback when refresh is clicked', async () => {
      const packData = createMockPackResponse();
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValue(packData);
      const onRefresh = vi.fn();

      render(<CurrentPackPicker sessionID="test-session" onRefresh={onRefresh} />);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /Refresh/i })).toBeInTheDocument();
      });

      await userEvent.click(screen.getByRole('button', { name: /Refresh/i }));

      expect(onRefresh).toHaveBeenCalled();
    });
  });

  describe('Card Statistics', () => {
    it('should display GIHWR for each card', async () => {
      const packData = createMockPackResponse({
        cards: [
          createMockPackCard({ arena_id: '1', name: 'Card A', gihwr: 58.5 }),
        ],
      });
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValue(packData);

      render(<CurrentPackPicker sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('58.5%')).toBeInTheDocument();
      });
    });

    it('should display ALSA for each card', async () => {
      const packData = createMockPackResponse({
        cards: [
          createMockPackCard({ arena_id: '1', name: 'Card A', alsa: 3.2 }),
        ],
      });
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValue(packData);

      render(<CurrentPackPicker sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('ALSA: 3.2')).toBeInTheDocument();
      });
    });
  });

  describe('Color Indicators', () => {
    it('should display color indicators for colored cards', async () => {
      const packData = createMockPackResponse({
        cards: [
          createMockPackCard({ arena_id: '1', name: 'Card A', colors: ['R', 'U'] }),
        ],
      });
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValue(packData);

      const { container } = render(<CurrentPackPicker sessionID="test-session" />);

      await waitFor(() => {
        const colorIndicators = container.querySelectorAll('.color-indicator');
        expect(colorIndicators.length).toBeGreaterThan(0);
      });
    });

    it('should display colorless indicator for colorless cards', async () => {
      const packData = createMockPackResponse({
        cards: [
          createMockPackCard({ arena_id: '1', name: 'Artifact', colors: [] }),
        ],
      });
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValue(packData);

      const { container } = render(<CurrentPackPicker sessionID="test-session" />);

      await waitFor(() => {
        const colorlessIndicator = container.querySelector('.color-indicator.colorless');
        expect(colorlessIndicator).toBeInTheDocument();
      });
    });
  });

  describe('Session ID Changes', () => {
    it('should reload data when sessionID changes', async () => {
      const packData1 = createMockPackResponse({ session_id: 'session-1' });
      const packData2 = createMockPackResponse({ session_id: 'session-2' });
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValueOnce(packData1);
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValueOnce(packData2);

      const { rerender } = render(<CurrentPackPicker sessionID="session-1" />);

      await waitFor(() => {
        expect(mockWailsApp.GetCurrentPackWithRecommendation).toHaveBeenCalledWith('session-1');
      });

      rerender(<CurrentPackPicker sessionID="session-2" />);

      await waitFor(() => {
        expect(mockWailsApp.GetCurrentPackWithRecommendation).toHaveBeenCalledWith('session-2');
      });
    });
  });

  describe('Card Reasoning', () => {
    it('should display reasoning when available', async () => {
      const packData = createMockPackResponse({
        cards: [
          createMockPackCard({
            arena_id: '1',
            name: 'Card A',
            reasoning: 'This card high win rate card and matches your colors.'
          }),
        ],
      });
      mockWailsApp.GetCurrentPackWithRecommendation.mockResolvedValue(packData);

      render(<CurrentPackPicker sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('This card high win rate card and matches your colors.')).toBeInTheDocument();
      });
    });
  });
});
