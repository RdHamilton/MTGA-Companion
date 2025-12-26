import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { render } from '../test/utils/testUtils';
import { DraftGrade } from './DraftGrade';
import { mockWailsApp } from '@/test/mocks/apiMock';
import { grading } from '@/types/models';

function createMockDraftGrade(overrides: Partial<grading.DraftGrade> = {}): grading.DraftGrade {
  return new grading.DraftGrade({
    overall_grade: 'B+',
    overall_score: 85.5,
    pick_quality_score: 88,
    color_discipline_score: 82,
    deck_composition_score: 86,
    strategic_score: 84,
    best_picks: ['Lightning Bolt - P1P1', 'Counterspell - P1P3'],
    worst_picks: ['Vanilla Bear - P2P8'],
    suggestions: ['Focus on 2-color decks for better consistency', 'Pick more removal spells'],
    ...overrides,
  });
}

describe('DraftGrade Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Initial State', () => {
    it('should show loading state initially', () => {
      mockWailsApp.GetDraftGrade.mockImplementation(() => new Promise(() => {})); // Never resolves

      render(<DraftGrade sessionID="test-session" />);

      expect(screen.getByText('Loading grade...')).toBeInTheDocument();
    });

    it('should hide when no grade exists and showCalculateButton is false', async () => {
      mockWailsApp.GetDraftGrade.mockRejectedValue(new Error('No grade found'));

      const { container } = render(<DraftGrade sessionID="test-session" showCalculateButton={false} />);

      await waitFor(() => {
        expect(container.firstChild).toBeNull();
      });
    });

    it('should show calculate button when no grade exists and showCalculateButton is true', async () => {
      mockWailsApp.GetDraftGrade.mockRejectedValue(new Error('No grade found'));

      render(<DraftGrade sessionID="test-session" showCalculateButton={true} />);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /Calculate Draft Grade/i })).toBeInTheDocument();
      });
    });
  });

  describe('Display Grade', () => {
    it('should display grade in full mode', async () => {
      const grade = createMockDraftGrade();
      mockWailsApp.GetDraftGrade.mockResolvedValue(grade);

      render(<DraftGrade sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('B+')).toBeInTheDocument();
        expect(screen.getByText('86/100')).toBeInTheDocument();
      });
    });

    it('should display grade in compact mode', async () => {
      const grade = createMockDraftGrade({ overall_grade: 'A' });
      mockWailsApp.GetDraftGrade.mockResolvedValue(grade);

      render(<DraftGrade sessionID="test-session" compact={true} />);

      await waitFor(() => {
        const badge = screen.getByText('A');
        expect(badge).toBeInTheDocument();
        expect(badge.className).toContain('draft-grade-badge');
      });
    });

    it('should apply correct color styling based on grade', async () => {
      const testCases = [
        { grade: 'A', expectedColor: 'rgb(68, 255, 136)' }, // #44ff88
        { grade: 'B+', expectedColor: 'rgb(74, 158, 255)' }, // #4a9eff
        { grade: 'C', expectedColor: 'rgb(255, 170, 68)' }, // #ffaa44
        { grade: 'D', expectedColor: 'rgb(255, 68, 68)' }, // #ff4444
        { grade: 'F', expectedColor: 'rgb(255, 68, 68)' }, // #ff4444
      ];

      for (const testCase of testCases) {
        vi.clearAllMocks();
        const grade = createMockDraftGrade({ overall_grade: testCase.grade });
        mockWailsApp.GetDraftGrade.mockResolvedValue(grade);

        const { container } = render(<DraftGrade sessionID="test-session" compact={true} />);

        await waitFor(() => {
          const badge = container.querySelector('.draft-grade-badge') as HTMLElement;
          expect(badge).toBeInTheDocument();
          expect(badge.style.backgroundColor).toBe(testCase.expectedColor);
        });
      }
    });
  });

  describe('Calculate Grade', () => {
    it('should calculate grade when button is clicked', async () => {
      mockWailsApp.GetDraftGrade.mockRejectedValue(new Error('No grade'));
      const newGrade = createMockDraftGrade();
      mockWailsApp.CalculateDraftGrade.mockResolvedValue(newGrade);

      render(<DraftGrade sessionID="test-session" showCalculateButton={true} />);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /Calculate Draft Grade/i })).toBeInTheDocument();
      });

      const calculateButton = screen.getByRole('button', { name: /Calculate Draft Grade/i });
      await userEvent.click(calculateButton);

      await waitFor(() => {
        expect(mockWailsApp.CalculateDraftGrade).toHaveBeenCalledWith('test-session');
        expect(screen.getByText('B+')).toBeInTheDocument();
      });
    });

    it('should call onGradeCalculated callback when grade is calculated', async () => {
      mockWailsApp.GetDraftGrade.mockRejectedValue(new Error('No grade'));
      const newGrade = createMockDraftGrade();
      mockWailsApp.CalculateDraftGrade.mockResolvedValue(newGrade);
      const onGradeCalculated = vi.fn();

      render(
        <DraftGrade
          sessionID="test-session"
          showCalculateButton={true}
          onGradeCalculated={onGradeCalculated}
        />
      );

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /Calculate Draft Grade/i })).toBeInTheDocument();
      });

      const calculateButton = screen.getByRole('button', { name: /Calculate Draft Grade/i });
      await userEvent.click(calculateButton);

      await waitFor(() => {
        expect(onGradeCalculated).toHaveBeenCalledWith(newGrade);
      });
    });

    it('should display error when grade calculation fails', async () => {
      mockWailsApp.GetDraftGrade.mockRejectedValue(new Error('No grade'));
      mockWailsApp.CalculateDraftGrade.mockRejectedValue(new Error('Calculation failed'));

      render(<DraftGrade sessionID="test-session" showCalculateButton={true} />);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /Calculate Draft Grade/i })).toBeInTheDocument();
      });

      const calculateButton = screen.getByRole('button', { name: /Calculate Draft Grade/i });
      await userEvent.click(calculateButton);

      await waitFor(() => {
        expect(screen.getByText(/Calculation failed/i)).toBeInTheDocument();
      });
    });
  });

  describe('Grade Breakdown Modal', () => {
    it('should open breakdown modal when grade is clicked', async () => {
      const grade = createMockDraftGrade();
      mockWailsApp.GetDraftGrade.mockResolvedValue(grade);

      render(<DraftGrade sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('B+')).toBeInTheDocument();
      });

      const gradeCard = screen.getByText('Click for breakdown');
      await userEvent.click(gradeCard.closest('.grade-card')!);

      await waitFor(() => {
        expect(screen.getByText('Draft Grade Breakdown')).toBeInTheDocument();
      });
    });

    it('should display component scores in breakdown modal', async () => {
      const grade = createMockDraftGrade();
      mockWailsApp.GetDraftGrade.mockResolvedValue(grade);

      render(<DraftGrade sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('B+')).toBeInTheDocument();
      });

      const gradeCard = screen.getByText('Click for breakdown');
      await userEvent.click(gradeCard.closest('.grade-card')!);

      await waitFor(() => {
        expect(screen.getByText('Component Scores')).toBeInTheDocument();
        expect(screen.getByText('Pick Quality')).toBeInTheDocument();
        expect(screen.getByText('Color Discipline')).toBeInTheDocument();
        expect(screen.getByText('Deck Composition')).toBeInTheDocument();
        expect(screen.getByText('Strategic Picks')).toBeInTheDocument();
      });
    });

    it('should display best and worst picks in breakdown modal', async () => {
      const grade = createMockDraftGrade({
        best_picks: ['Lightning Bolt - P1P1', 'Counterspell - P1P3'],
        worst_picks: ['Vanilla Bear - P2P8'],
      });
      mockWailsApp.GetDraftGrade.mockResolvedValue(grade);

      render(<DraftGrade sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('B+')).toBeInTheDocument();
      });

      const gradeCard = screen.getByText('Click for breakdown');
      await userEvent.click(gradeCard.closest('.grade-card')!);

      await waitFor(() => {
        expect(screen.getByText('✅ Best Picks')).toBeInTheDocument();
        expect(screen.getByText('Lightning Bolt - P1P1')).toBeInTheDocument();
        expect(screen.getByText('Counterspell - P1P3')).toBeInTheDocument();

        expect(screen.getByText('⚠️ Worst Picks')).toBeInTheDocument();
        expect(screen.getByText('Vanilla Bear - P2P8')).toBeInTheDocument();
      });
    });

    it('should display suggestions in breakdown modal', async () => {
      const grade = createMockDraftGrade({
        suggestions: ['Focus on 2-color decks', 'Pick more removal spells'],
      });
      mockWailsApp.GetDraftGrade.mockResolvedValue(grade);

      render(<DraftGrade sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('B+')).toBeInTheDocument();
      });

      const gradeCard = screen.getByText('Click for breakdown');
      await userEvent.click(gradeCard.closest('.grade-card')!);

      await waitFor(() => {
        expect(screen.getByText('💡 Improvement Suggestions')).toBeInTheDocument();
        expect(screen.getByText('Focus on 2-color decks')).toBeInTheDocument();
        expect(screen.getByText('Pick more removal spells')).toBeInTheDocument();
      });
    });

    it('should close breakdown modal when close button is clicked', async () => {
      const grade = createMockDraftGrade();
      mockWailsApp.GetDraftGrade.mockResolvedValue(grade);

      render(<DraftGrade sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('B+')).toBeInTheDocument();
      });

      const gradeCard = screen.getByText('Click for breakdown');
      await userEvent.click(gradeCard.closest('.grade-card')!);

      await waitFor(() => {
        expect(screen.getByText('Draft Grade Breakdown')).toBeInTheDocument();
      });

      const closeButton = screen.getByRole('button', { name: '×' });
      await userEvent.click(closeButton);

      await waitFor(() => {
        expect(screen.queryByText('Draft Grade Breakdown')).not.toBeInTheDocument();
      });
    });

    it('should close breakdown modal when overlay is clicked', async () => {
      const grade = createMockDraftGrade();
      mockWailsApp.GetDraftGrade.mockResolvedValue(grade);

      render(<DraftGrade sessionID="test-session" />);

      await waitFor(() => {
        expect(screen.getByText('B+')).toBeInTheDocument();
      });

      const gradeCard = screen.getByText('Click for breakdown');
      await userEvent.click(gradeCard.closest('.grade-card')!);

      await waitFor(() => {
        expect(screen.getByText('Draft Grade Breakdown')).toBeInTheDocument();
      });

      const overlay = document.querySelector('.modal-overlay');
      if (overlay) {
        await userEvent.click(overlay);

        await waitFor(() => {
          expect(screen.queryByText('Draft Grade Breakdown')).not.toBeInTheDocument();
        });
      }
    });
  });

  describe('Compact Mode', () => {
    it('should render badge in compact mode', async () => {
      const grade = createMockDraftGrade();
      mockWailsApp.GetDraftGrade.mockResolvedValue(grade);

      render(<DraftGrade sessionID="test-session" compact={true} showCalculateButton={true} />);

      await waitFor(() => {
        expect(screen.getByText('B+')).toBeInTheDocument();
      });

      // The compact badge should have the grade displayed
      expect(screen.getByText('B+')).toBeInTheDocument();
      expect(screen.getByTitle(/Click to view breakdown/i)).toBeInTheDocument();
    });

    it('should display tooltip on hover in compact mode', async () => {
      const grade = createMockDraftGrade({ overall_score: 85.5 });
      mockWailsApp.GetDraftGrade.mockResolvedValue(grade);

      render(<DraftGrade sessionID="test-session" compact={true} />);

      await waitFor(() => {
        const badge = screen.getByText('B+');
        expect(badge).toHaveAttribute('title', 'Click to view breakdown (85.5/100)');
      });
    });

    it('should open breakdown modal when compact badge is clicked', async () => {
      const grade = createMockDraftGrade();
      mockWailsApp.GetDraftGrade.mockResolvedValue(grade);

      render(<DraftGrade sessionID="test-session" compact={true} />);

      await waitFor(() => {
        expect(screen.getByText('B+')).toBeInTheDocument();
      });

      // Click the compact badge
      const badge = screen.getByText('B+');
      await userEvent.click(badge);

      // Modal should open
      await waitFor(() => {
        expect(screen.getByText('Draft Grade Breakdown')).toBeInTheDocument();
        expect(screen.getByText('Component Scores')).toBeInTheDocument();
      });
    });

    it('should close breakdown modal when close button is clicked in compact mode', async () => {
      const grade = createMockDraftGrade();
      mockWailsApp.GetDraftGrade.mockResolvedValue(grade);

      render(<DraftGrade sessionID="test-session" compact={true} />);

      await waitFor(() => {
        expect(screen.getByText('B+')).toBeInTheDocument();
      });

      // Click the compact badge to open modal
      const badge = screen.getByText('B+');
      await userEvent.click(badge);

      await waitFor(() => {
        expect(screen.getByText('Draft Grade Breakdown')).toBeInTheDocument();
      });

      // Click close button
      const closeButton = screen.getByRole('button', { name: '×' });
      await userEvent.click(closeButton);

      await waitFor(() => {
        expect(screen.queryByText('Draft Grade Breakdown')).not.toBeInTheDocument();
      });
    });
  });
});
