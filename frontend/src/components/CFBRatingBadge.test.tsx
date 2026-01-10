import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { CFBRatingBadge } from './CFBRatingBadge';
import type { CFBLimitedGrade } from '@/services/api/cards';
import { getCFBGradeColor } from '@/services/api/cards';

describe('CFBRatingBadge', () => {
  it('renders the grade text', () => {
    render(<CFBRatingBadge grade="A" />);
    expect(screen.getByText('A')).toBeInTheDocument();
  });

  it('renders the CFB label by default', () => {
    render(<CFBRatingBadge grade="B+" />);
    expect(screen.getByText('CFB')).toBeInTheDocument();
    expect(screen.getByText('B+')).toBeInTheDocument();
  });

  it('hides the CFB label when showLabel is false', () => {
    render(<CFBRatingBadge grade="A-" showLabel={false} />);
    expect(screen.queryByText('CFB')).not.toBeInTheDocument();
    expect(screen.getByText('A-')).toBeInTheDocument();
  });

  it('applies the correct background color for each grade', () => {
    const grades: CFBLimitedGrade[] = ['A+', 'A', 'A-', 'B+', 'B', 'B-', 'C+', 'C', 'C-', 'D', 'F'];

    grades.forEach((grade) => {
      const { unmount } = render(<CFBRatingBadge grade={grade} />);
      const badge = screen.getByTestId('cfb-rating-badge');
      expect(badge).toHaveStyle({ backgroundColor: getCFBGradeColor(grade) });
      unmount();
    });
  });

  it('shows default tooltip with grade', () => {
    render(<CFBRatingBadge grade="B" />);
    const badge = screen.getByTestId('cfb-rating-badge');
    expect(badge).toHaveAttribute('title', 'CFB Rating: B');
  });

  it('shows custom commentary as tooltip', () => {
    render(<CFBRatingBadge grade="A+" commentary="Best card in the set!" />);
    const badge = screen.getByTestId('cfb-rating-badge');
    expect(badge).toHaveAttribute('title', 'Best card in the set!');
  });

  it('applies small size class', () => {
    render(<CFBRatingBadge grade="C" size="small" />);
    const badge = screen.getByTestId('cfb-rating-badge');
    expect(badge).toHaveClass('cfb-rating-badge--small');
  });

  it('applies medium size class by default', () => {
    render(<CFBRatingBadge grade="C" />);
    const badge = screen.getByTestId('cfb-rating-badge');
    expect(badge).toHaveClass('cfb-rating-badge--medium');
  });

  it('applies large size class', () => {
    render(<CFBRatingBadge grade="C" size="large" />);
    const badge = screen.getByTestId('cfb-rating-badge');
    expect(badge).toHaveClass('cfb-rating-badge--large');
  });
});

describe('getCFBGradeColor', () => {
  it('returns gold for A+', () => {
    expect(getCFBGradeColor('A+')).toBe('#ffd700');
  });

  it('returns silver for A', () => {
    expect(getCFBGradeColor('A')).toBe('#c0c0c0');
  });

  it('returns bronze for B+', () => {
    expect(getCFBGradeColor('B+')).toBe('#cd7f32');
  });

  it('returns blue for C+', () => {
    expect(getCFBGradeColor('C+')).toBe('#4a9eff');
  });

  it('returns gray for D', () => {
    expect(getCFBGradeColor('D')).toBe('#888888');
  });

  it('returns red for F', () => {
    expect(getCFBGradeColor('F')).toBe('#ff4444');
  });
});
