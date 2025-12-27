import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { WinRatePrediction } from './WinRatePrediction';

// Mock drafts API module
vi.mock('@/services/api', () => ({
  drafts: {
    getDraftWinRatePrediction: vi.fn(),
  },
}));

import { drafts } from '@/services/api';

const mockPrediction = {
  PredictedWinRate: 0.55,
  PredictedWinRateMin: 0.50,
  PredictedWinRateMax: 0.60,
  Factors: {
    deck_average_gihwr: 0.52,
    curve_score: 0.75,
    color_adjustment: 0.02,
    bomb_bonus: 0.05,
    confidence_level: 'medium',
    explanation: 'Your deck has good card quality and a solid mana curve.',
    high_performers: ['Lightning Bolt', 'Counterspell'],
    low_performers: ['Grizzly Bears'],
    color_distribution: { W: 10, U: 12, B: 0, R: 0, G: 0, C: 3 },
    curve_distribution: { '0': 0, '1': 2, '2': 8, '3': 6, '4': 4, '5': 2, '6': 1, '7': 0 },
    total_cards: 40,
  },
};

describe('WinRatePrediction', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Loading State', () => {
    it('should show loading state initially', async () => {
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      render(<WinRatePrediction sessionID="test-session" />);

      expect(screen.getByText('Loading prediction...')).toBeInTheDocument();
    });

    it('should show loading state when calculating prediction', async () => {
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Not found'));
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      render(<WinRatePrediction sessionID="test-session" showPredictButton={true} />);

      await waitFor(() => {
        expect(screen.queryByText('Loading prediction...')).not.toBeInTheDocument();
      });

      const button = screen.getByText(/Predict Win Rate/);
      fireEvent.click(button);

      expect(screen.getByText('Loading prediction...')).toBeInTheDocument();
    });
  });

  describe('Error State', () => {
    it('should show error message when calculation fails', async () => {
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Not found'));
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockRejectedValue(new Error('Calculation failed'));

      render(<WinRatePrediction sessionID="test-session" showPredictButton={true} />);

      await waitFor(() => {
        const button = screen.queryByText(/Predict Win Rate/);
        expect(button).toBeInTheDocument();
      });

      const button = screen.getByText(/Predict Win Rate/);
      fireEvent.click(button);

      await waitFor(() => {
        expect(screen.getByText('Calculation failed')).toBeInTheDocument();
      });
    });

    it('should show generic error for non-Error exceptions', async () => {
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Not found'));
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockRejectedValue('String error');

      render(<WinRatePrediction sessionID="test-session" showPredictButton={true} />);

      await waitFor(() => {
        const button = screen.queryByText(/Predict Win Rate/);
        expect(button).toBeInTheDocument();
      });

      const button = screen.getByText(/Predict Win Rate/);
      fireEvent.click(button);

      await waitFor(() => {
        expect(screen.getByText('Failed to calculate prediction')).toBeInTheDocument();
      });
    });
  });

  describe('No Prediction State', () => {
    it('should return null when no prediction and no predict button', async () => {
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockRejectedValue(new Error('Not found'));

      const { container } = render(<WinRatePrediction sessionID="test-session" showPredictButton={false} />);

      await waitFor(() => {
        expect(container.firstChild).toBeNull();
      });
    });

    it('should show predict button when no prediction exists', async () => {
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockRejectedValue(new Error('Not found'));

      render(<WinRatePrediction sessionID="test-session" showPredictButton={true} />);

      await waitFor(() => {
        expect(screen.getByText(/Predict Win Rate/)).toBeInTheDocument();
      });
    });

    it('should call PredictDraftWinRate when predict button clicked', async () => {
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Not found'));
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockResolvedValue(mockPrediction);

      render(<WinRatePrediction sessionID="test-session" showPredictButton={true} />);

      await waitFor(() => {
        const button = screen.queryByText(/Predict Win Rate/);
        expect(button).toBeInTheDocument();
      });

      const button = screen.getByText(/Predict Win Rate/);
      fireEvent.click(button);

      await waitFor(() => {
        expect(drafts.getDraftWinRatePrediction).toHaveBeenCalledWith('test-session');
      });
    });
  });

  describe('Prediction Display', () => {
    beforeEach(() => {
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockResolvedValue(mockPrediction);
    });

    it('should display prediction win rate', async () => {
      render(<WinRatePrediction sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('55%')).toBeInTheDocument();
      });
    });

    it('should display prediction range', async () => {
      render(<WinRatePrediction sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('(50-60%)')).toBeInTheDocument();
      });
    });

    it('should display explanation', async () => {
      render(<WinRatePrediction sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText(mockPrediction.Factors.explanation)).toBeInTheDocument();
      });
    });

    it('should display predicted win rate heading', async () => {
      render(<WinRatePrediction sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('Predicted Win Rate')).toBeInTheDocument();
      });
    });

    it('should display click hint', async () => {
      render(<WinRatePrediction sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('Click for details')).toBeInTheDocument();
      });
    });
  });

  describe('Compact Mode', () => {
    beforeEach(() => {
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockResolvedValue(mockPrediction);
    });

    it('should render in compact mode', async () => {
      const { container } = render(<WinRatePrediction sessionID="test-session" compact={true} />);

      await waitFor(() => {
        expect(container.querySelector('.win-rate-badge')).toBeInTheDocument();
      });
    });

    it('should display win rate in compact mode', async () => {
      render(<WinRatePrediction sessionID="test-session" compact={true} />);

      await waitFor(() => {
        expect(screen.getByText('55%')).toBeInTheDocument();
      });
    });

    it('should display "Win Rate" label in compact mode', async () => {
      render(<WinRatePrediction sessionID="test-session" compact={true} />);

      await waitFor(() => {
        expect(screen.getByText('Win Rate')).toBeInTheDocument();
      });
    });

    it('should have title attribute in compact mode', async () => {
      const { container } = render(<WinRatePrediction sessionID="test-session" compact={true} />);

      await waitFor(() => {
        const badge = container.querySelector('.win-rate-badge');
        expect(badge).toHaveAttribute('title', 'Expected 55% win rate (50-60%)');
      });
    });
  });

  describe('Modal Interaction', () => {
    beforeEach(() => {
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockResolvedValue(mockPrediction);
    });

    it('should open breakdown modal on click', async () => {
      render(<WinRatePrediction sessionID="test-session" />);

      await waitFor(() => {
        const card = screen.getByText('Predicted Win Rate').closest('.prediction-card');
        expect(card).toBeInTheDocument();
      });

      const card = screen.getByText('Predicted Win Rate').closest('.prediction-card');
      fireEvent.click(card!);

      await waitFor(() => {
        expect(screen.getByText('Win Rate Prediction Breakdown')).toBeInTheDocument();
      });
    });

    it('should close modal on close button click', async () => {
      render(<WinRatePrediction sessionID="test-session" />);

      await waitFor(() => {
        const card = screen.getByText('Predicted Win Rate').closest('.prediction-card');
        fireEvent.click(card!);
      });

      await waitFor(() => {
        expect(screen.getByText('Win Rate Prediction Breakdown')).toBeInTheDocument();
      });

      const closeButton = screen.getByText('×');
      fireEvent.click(closeButton);

      await waitFor(() => {
        expect(screen.queryByText('Win Rate Prediction Breakdown')).not.toBeInTheDocument();
      });
    });

    it('should close modal on overlay click', async () => {
      const { container } = render(<WinRatePrediction sessionID="test-session" />);

      await waitFor(() => {
        const card = screen.getByText('Predicted Win Rate').closest('.prediction-card');
        fireEvent.click(card!);
      });

      await waitFor(() => {
        expect(screen.getByText('Win Rate Prediction Breakdown')).toBeInTheDocument();
      });

      const overlay = container.querySelector('.modal-overlay');
      fireEvent.click(overlay!);

      await waitFor(() => {
        expect(screen.queryByText('Win Rate Prediction Breakdown')).not.toBeInTheDocument();
      });
    });

    it('should not close modal on content click', async () => {
      const { container } = render(<WinRatePrediction sessionID="test-session" />);

      await waitFor(() => {
        const card = screen.getByText('Predicted Win Rate').closest('.prediction-card');
        fireEvent.click(card!);
      });

      await waitFor(() => {
        expect(screen.getByText('Win Rate Prediction Breakdown')).toBeInTheDocument();
      });

      const modalContent = container.querySelector('.modal-content');
      fireEvent.click(modalContent!);

      expect(screen.getByText('Win Rate Prediction Breakdown')).toBeInTheDocument();
    });
  });

  describe('Modal Content', () => {
    beforeEach(() => {
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockResolvedValue(mockPrediction);
    });

    it('should display prediction factors', async () => {
      render(<WinRatePrediction sessionID="test-session" />);

      await waitFor(() => {
        const card = screen.getByText('Predicted Win Rate').closest('.prediction-card');
        fireEvent.click(card!);
      });

      await waitFor(() => {
        expect(screen.getByText('Card Quality (GIHWR)')).toBeInTheDocument();
        expect(screen.getByText('Curve Score')).toBeInTheDocument();
        expect(screen.getByText('Color Discipline')).toBeInTheDocument();
        expect(screen.getByText('Bomb Bonus')).toBeInTheDocument();
      });
    });

    it('should display high performers', async () => {
      render(<WinRatePrediction sessionID="test-session" />);

      await waitFor(() => {
        const card = screen.getByText('Predicted Win Rate').closest('.prediction-card');
        fireEvent.click(card!);
      });

      await waitFor(() => {
        expect(screen.getByText('🌟 Premium Cards')).toBeInTheDocument();
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
        expect(screen.getByText('Counterspell')).toBeInTheDocument();
      });
    });

    it('should display low performers', async () => {
      render(<WinRatePrediction sessionID="test-session" />);

      await waitFor(() => {
        const card = screen.getByText('Predicted Win Rate').closest('.prediction-card');
        fireEvent.click(card!);
      });

      await waitFor(() => {
        expect(screen.getByText('⚠️ Weak Cards')).toBeInTheDocument();
        expect(screen.getByText('Grizzly Bears')).toBeInTheDocument();
      });
    });

    it('should display color distribution', async () => {
      render(<WinRatePrediction sessionID="test-session" />);

      await waitFor(() => {
        const card = screen.getByText('Predicted Win Rate').closest('.prediction-card');
        fireEvent.click(card!);
      });

      await waitFor(() => {
        expect(screen.getByText('Color Distribution')).toBeInTheDocument();
        expect(screen.getByText('W:')).toBeInTheDocument();
        expect(screen.getByText('U:')).toBeInTheDocument();
      });
    });

    it('should display mana curve', async () => {
      render(<WinRatePrediction sessionID="test-session" />);

      await waitFor(() => {
        const card = screen.getByText('Predicted Win Rate').closest('.prediction-card');
        fireEvent.click(card!);
      });

      await waitFor(() => {
        expect(screen.getByText('Mana Curve')).toBeInTheDocument();
      });
    });

    it('should display confidence level', async () => {
      render(<WinRatePrediction sessionID="test-session" />);

      await waitFor(() => {
        const card = screen.getByText('Predicted Win Rate').closest('.prediction-card');
        fireEvent.click(card!);
      });

      await waitFor(() => {
        expect(screen.getByText('medium confidence')).toBeInTheDocument();
      });
    });
  });

  describe('Callback Handling', () => {
    it('should call onPredictionCalculated when prediction calculated', async () => {
      const onPredictionCalculated = vi.fn();
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Not found'));
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockResolvedValue(mockPrediction);

      render(
        <WinRatePrediction
          sessionID="test-session"
          showPredictButton={true}
          onPredictionCalculated={onPredictionCalculated}
        />
      );

      await waitFor(() => {
        const button = screen.queryByText(/Predict Win Rate/);
        expect(button).toBeInTheDocument();
      });

      const button = screen.getByText(/Predict Win Rate/);
      fireEvent.click(button);

      await waitFor(() => {
        expect(onPredictionCalculated).toHaveBeenCalledWith(mockPrediction);
      });
    });

    it('should not error when onPredictionCalculated not provided', async () => {
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Not found'));
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockResolvedValue(mockPrediction);

      render(<WinRatePrediction sessionID="test-session" showPredictButton={true} />);

      await waitFor(() => {
        const button = screen.queryByText(/Predict Win Rate/);
        expect(button).toBeInTheDocument();
      });

      const button = screen.getByText(/Predict Win Rate/);

      expect(() => {
        fireEvent.click(button);
      }).not.toThrow();

      await waitFor(() => {
        expect(screen.getByText('55%')).toBeInTheDocument();
      });
    });
  });

  describe('Win Rate Color Coding', () => {
    it('should use green for high win rate (60%+)', async () => {
      const highWinPrediction = { ...mockPrediction, PredictedWinRate: 0.65 };
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockResolvedValue(highWinPrediction);

      const { container } = render(<WinRatePrediction sessionID="test-session" />);

      await waitFor(() => {
        const winRateElement = container.querySelector('.win-rate-main');
        expect(winRateElement).toHaveStyle({ color: '#44ff88' });
      });
    });

    it('should use blue for good win rate (55-60%)', async () => {
      const goodWinPrediction = { ...mockPrediction, PredictedWinRate: 0.57 };
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockResolvedValue(goodWinPrediction);

      const { container } = render(<WinRatePrediction sessionID="test-session" />);

      await waitFor(() => {
        const winRateElement = container.querySelector('.win-rate-main');
        expect(winRateElement).toHaveStyle({ color: '#4a9eff' });
      });
    });

    it('should use orange for average win rate (50-55%)', async () => {
      const avgWinPrediction = { ...mockPrediction, PredictedWinRate: 0.52 };
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockResolvedValue(avgWinPrediction);

      const { container } = render(<WinRatePrediction sessionID="test-session" />);

      await waitFor(() => {
        const winRateElement = container.querySelector('.win-rate-main');
        expect(winRateElement).toHaveStyle({ color: '#ffaa44' });
      });
    });

    it('should use red for low win rate (<50%)', async () => {
      const lowWinPrediction = { ...mockPrediction, PredictedWinRate: 0.45 };
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockResolvedValue(lowWinPrediction);

      const { container } = render(<WinRatePrediction sessionID="test-session" />);

      await waitFor(() => {
        const winRateElement = container.querySelector('.win-rate-main');
        expect(winRateElement).toHaveStyle({ color: '#ff4444' });
      });
    });
  });

  describe('Session ID Changes', () => {
    it('should reload prediction when sessionID changes', async () => {
      (drafts.getDraftWinRatePrediction as ReturnType<typeof vi.fn>).mockResolvedValue(mockPrediction);

      const { rerender } = render(<WinRatePrediction sessionID="session-1" />);

      await waitFor(() => {
        expect(drafts.getDraftWinRatePrediction).toHaveBeenCalledWith('session-1');
      });

      rerender(<WinRatePrediction sessionID="session-2" />);

      await waitFor(() => {
        expect(drafts.getDraftWinRatePrediction).toHaveBeenCalledWith('session-2');
      });
    });
  });
});
