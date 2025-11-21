import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import { render } from '../test/utils/testUtils';
import DraftStatistics from './DraftStatistics';
import { mockWailsApp } from '../test/mocks/wailsApp';
import { models } from '../../wailsjs/go/models';

function createMockDeckMetrics(overrides: Partial<models.DeckMetrics> = {}): models.DeckMetrics {
  return new models.DeckMetrics({
    total_cards: 23,
    total_non_land_cards: 23,
    creature_count: 15,
    noncreature_count: 8,
    cmc_average: 2.87,
    distribution_all: [0, 2, 5, 7, 6, 2, 1], // Cards at CMC 0-6+
    distribution_creatures: [0, 1, 3, 5, 4, 2, 0],
    distribution_noncreatures: [0, 1, 2, 2, 2, 0, 1],
    type_breakdown: {
      Creature: 15,
      Instant: 5,
      Sorcery: 2,
      Enchantment: 1,
    },
    color_distribution: {
      W: 0.4,
      U: 0.3,
      B: 0,
      R: 0,
      G: 0.3,
    },
    color_counts: {
      W: 9,
      U: 7,
      B: 0,
      R: 0,
      G: 7,
    },
    multi_color_count: 2,
    colorless_count: 0,
    ...overrides,
  });
}

describe('DraftStatistics Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Loading and Error States', () => {
    it('should display loading state while fetching metrics', () => {
      mockWailsApp.GetDraftDeckMetrics.mockImplementation(() => new Promise(() => {})); // Never resolves

      render(<DraftStatistics sessionID="test-session" pickCount={5} />);

      expect(screen.getByText('Loading statistics...')).toBeInTheDocument();
    });

    it('should display error message when fetching fails', async () => {
      mockWailsApp.GetDraftDeckMetrics.mockRejectedValue(new Error('Failed to load metrics'));

      render(<DraftStatistics sessionID="test-session" pickCount={5} />);

      await waitFor(() => {
        expect(screen.getByText(/Error: Failed to load metrics/i)).toBeInTheDocument();
      });
    });

    it('should display empty state when no cards are drafted', async () => {
      const emptyMetrics = createMockDeckMetrics({ total_cards: 0 });
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(emptyMetrics);

      render(<DraftStatistics sessionID="test-session" pickCount={0} />);

      await waitFor(() => {
        expect(screen.getByText('No cards drafted yet')).toBeInTheDocument();
      });
    });
  });

  describe('Summary Statistics Display', () => {
    it('should display total cards count', async () => {
      const metrics = createMockDeckMetrics({ total_cards: 23 });
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(metrics);

      render(<DraftStatistics sessionID="test-session" pickCount={23} />);

      await waitFor(() => {
        expect(screen.getByText('Total Cards')).toBeInTheDocument();
        expect(screen.getByText('23')).toBeInTheDocument();
      });
    });

    it('should display average CMC correctly', async () => {
      const metrics = createMockDeckMetrics({ cmc_average: 2.87 });
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(metrics);

      render(<DraftStatistics sessionID="test-session" pickCount={23} />);

      await waitFor(() => {
        expect(screen.getByText('Avg CMC')).toBeInTheDocument();
        expect(screen.getByText('2.87')).toBeInTheDocument();
      });
    });

    it('should display creature count and percentage', async () => {
      const metrics = createMockDeckMetrics({
        total_cards: 23,
        creature_count: 15,
      });
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(metrics);

      render(<DraftStatistics sessionID="test-session" pickCount={23} />);

      await waitFor(() => {
        expect(screen.getByText('Creatures')).toBeInTheDocument();
        expect(screen.getByText(/15.*65\.2%/)).toBeInTheDocument();
      });
    });

    it('should display spell count and percentage', async () => {
      const metrics = createMockDeckMetrics({
        total_cards: 23,
        noncreature_count: 8,
      });
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(metrics);

      render(<DraftStatistics sessionID="test-session" pickCount={23} />);

      await waitFor(() => {
        expect(screen.getByText('Spells')).toBeInTheDocument();
        expect(screen.getByText(/8.*34\.8%/)).toBeInTheDocument();
      });
    });
  });

  describe('Mana Curve Display', () => {
    it('should render mana curve chart', async () => {
      const metrics = createMockDeckMetrics();
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(metrics);

      render(<DraftStatistics sessionID="test-session" pickCount={23} />);

      await waitFor(() => {
        expect(screen.getByText('Mana Curve')).toBeInTheDocument();
      });
    });

    it('should render creatures vs spells chart', async () => {
      const metrics = createMockDeckMetrics();
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(metrics);

      render(<DraftStatistics sessionID="test-session" pickCount={23} />);

      await waitFor(() => {
        expect(screen.getByText('Creatures vs Spells by CMC')).toBeInTheDocument();
      });
    });
  });

  describe('Color Distribution', () => {
    it('should display color distribution pie chart when colors exist', async () => {
      const metrics = createMockDeckMetrics({
        color_counts: {
          W: 9,
          U: 7,
          B: 0,
          R: 0,
          G: 7,
        },
      });
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(metrics);

      render(<DraftStatistics sessionID="test-session" pickCount={23} />);

      await waitFor(() => {
        expect(screen.getByText('Color Distribution')).toBeInTheDocument();
      });
    });

    it('should display multi-color card count when present', async () => {
      const metrics = createMockDeckMetrics({
        multi_color_count: 2,
      });
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(metrics);

      render(<DraftStatistics sessionID="test-session" pickCount={23} />);

      await waitFor(() => {
        expect(screen.getByText(/Multi-color cards: 2/i)).toBeInTheDocument();
      });
    });

    it('should display colorless card count when present', async () => {
      const metrics = createMockDeckMetrics({
        colorless_count: 3,
      });
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(metrics);

      render(<DraftStatistics sessionID="test-session" pickCount={23} />);

      await waitFor(() => {
        expect(screen.getByText(/Colorless cards: 3/i)).toBeInTheDocument();
      });
    });
  });

  describe('Type Breakdown', () => {
    it('should display top 5 card types', async () => {
      const metrics = createMockDeckMetrics({
        type_breakdown: {
          Creature: 15,
          Instant: 5,
          Sorcery: 2,
          Enchantment: 1,
          Artifact: 1,
          Planeswalker: 1,
        },
      });
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(metrics);

      render(<DraftStatistics sessionID="test-session" pickCount={25} />);

      await waitFor(() => {
        expect(screen.getByText('Card Types')).toBeInTheDocument();
        expect(screen.getByText('Creature')).toBeInTheDocument();
        expect(screen.getByText('Instant')).toBeInTheDocument();
        expect(screen.getByText('Sorcery')).toBeInTheDocument();
        expect(screen.getByText('Enchantment')).toBeInTheDocument();
        expect(screen.getByText('Artifact')).toBeInTheDocument();
      });
    });

    it('should display card type counts correctly', async () => {
      const metrics = createMockDeckMetrics({
        type_breakdown: {
          Creature: 15,
          Instant: 5,
        },
      });
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(metrics);

      render(<DraftStatistics sessionID="test-session" pickCount={20} />);

      await waitFor(() => {
        const typeRows = document.querySelectorAll('.type-row');
        expect(typeRows.length).toBeGreaterThanOrEqual(2);
      });
    });
  });

  describe('Reactive Updates', () => {
    it('should reload metrics when pickCount changes', async () => {
      const metrics1 = createMockDeckMetrics({ total_cards: 10, creature_count: 5 });
      const metrics2 = createMockDeckMetrics({ total_cards: 20, creature_count: 12 });

      mockWailsApp.GetDraftDeckMetrics.mockResolvedValueOnce(metrics1);

      const { rerender } = render(<DraftStatistics sessionID="test-session" pickCount={10} />);

      await waitFor(() => {
        expect(screen.getByText('Draft Statistics')).toBeInTheDocument();
        const totalCards = screen.getAllByText('10')[0];
        expect(totalCards).toBeInTheDocument();
      });

      mockWailsApp.GetDraftDeckMetrics.mockResolvedValueOnce(metrics2);
      rerender(<DraftStatistics sessionID="test-session" pickCount={20} />);

      await waitFor(() => {
        const totalCards = screen.getAllByText('20')[0];
        expect(totalCards).toBeInTheDocument();
      });

      expect(mockWailsApp.GetDraftDeckMetrics).toHaveBeenCalledTimes(2);
    });

    it('should reload metrics when sessionID changes', async () => {
      const metrics1 = createMockDeckMetrics({ total_cards: 10 });
      const metrics2 = createMockDeckMetrics({ total_cards: 12 });

      mockWailsApp.GetDraftDeckMetrics.mockResolvedValueOnce(metrics1);

      const { rerender } = render(<DraftStatistics sessionID="session-1" pickCount={10} />);

      await waitFor(() => {
        expect(screen.getByText('10')).toBeInTheDocument();
      });

      mockWailsApp.GetDraftDeckMetrics.mockResolvedValueOnce(metrics2);
      rerender(<DraftStatistics sessionID="session-2" pickCount={12} />);

      await waitFor(() => {
        expect(screen.getByText('12')).toBeInTheDocument();
      });

      expect(mockWailsApp.GetDraftDeckMetrics).toHaveBeenCalledWith('session-1');
      expect(mockWailsApp.GetDraftDeckMetrics).toHaveBeenCalledWith('session-2');
    });
  });

  describe('Data Validation', () => {
    it('should handle zero creature count', async () => {
      const metrics = createMockDeckMetrics({
        total_cards: 10,
        creature_count: 0,
        noncreature_count: 10,
      });
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(metrics);

      render(<DraftStatistics sessionID="test-session" pickCount={10} />);

      await waitFor(() => {
        expect(screen.getByText('Creatures')).toBeInTheDocument();
        expect(screen.getByText('0 (0.0%)')).toBeInTheDocument();
      });
    });

    it('should handle missing type breakdown gracefully', async () => {
      const metrics = createMockDeckMetrics({
        type_breakdown: {},
      });
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(metrics);

      render(<DraftStatistics sessionID="test-session" pickCount={10} />);

      await waitFor(() => {
        expect(screen.queryByText('Card Types')).not.toBeInTheDocument();
      });
    });

    it('should handle empty color distribution', async () => {
      const metrics = createMockDeckMetrics({
        color_counts: {
          W: 0,
          U: 0,
          B: 0,
          R: 0,
          G: 0,
        },
      });
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(metrics);

      render(<DraftStatistics sessionID="test-session" pickCount={10} />);

      await waitFor(() => {
        expect(screen.queryByText('Color Distribution')).not.toBeInTheDocument();
      });
    });
  });

  describe('Chart Rendering', () => {
    it('should render all required charts when data is present', async () => {
      const metrics = createMockDeckMetrics();
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(metrics);

      render(<DraftStatistics sessionID="test-session" pickCount={23} />);

      await waitFor(() => {
        expect(screen.getByText('Draft Statistics')).toBeInTheDocument();
        expect(screen.getByText('Mana Curve')).toBeInTheDocument();
        expect(screen.getByText('Creatures vs Spells by CMC')).toBeInTheDocument();
        expect(screen.getByText('Color Distribution')).toBeInTheDocument();
        expect(screen.getByText('Card Types')).toBeInTheDocument();
      });
    });
  });
});
