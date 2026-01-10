import type { CFBLimitedGrade } from '@/services/api/cards';
import { getCFBGradeColor } from '@/services/api/cards';
import './CFBRatingBadge.css';

export interface CFBRatingBadgeProps {
  /** The CFB limited rating grade (A+, A, A-, B+, etc.) */
  grade: CFBLimitedGrade;
  /** Optional commentary/tooltip text */
  commentary?: string;
  /** Size variant */
  size?: 'small' | 'medium' | 'large';
  /** Whether to show the "CFB" label */
  showLabel?: boolean;
}

/**
 * Displays a ChannelFireball rating grade as a colored badge.
 */
export function CFBRatingBadge({
  grade,
  commentary,
  size = 'medium',
  showLabel = true,
}: CFBRatingBadgeProps) {
  const color = getCFBGradeColor(grade);

  return (
    <span
      className={`cfb-rating-badge cfb-rating-badge--${size}`}
      style={{ backgroundColor: color }}
      title={commentary || `CFB Rating: ${grade}`}
      data-testid="cfb-rating-badge"
    >
      {showLabel && <span className="cfb-rating-badge__label">CFB</span>}
      <span className="cfb-rating-badge__grade">{grade}</span>
    </span>
  );
}

export default CFBRatingBadge;
