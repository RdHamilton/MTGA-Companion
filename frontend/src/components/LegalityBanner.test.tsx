import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { LegalityBanner } from './LegalityBanner';
import type { ValidationError, ValidationWarning } from '@/services/api/standard';

const mockBannedErrors: ValidationError[] = [
  {
    cardId: 12345,
    cardName: 'Omnath, Locus of Creation',
    reason: 'banned',
    details: 'Card is banned in Standard',
  },
];

const mockNotLegalErrors: ValidationError[] = [
  {
    cardId: 67890,
    cardName: 'Some Old Card',
    reason: 'not_legal',
    details: 'Card is not legal in Standard',
  },
];

const mockCopyErrors: ValidationError[] = [
  {
    cardId: 11111,
    cardName: 'Lightning Bolt',
    reason: 'too_many_copies',
    details: 'Deck contains 5 copies (maximum 4 allowed)',
  },
];

const mockDeckSizeErrors: ValidationError[] = [
  {
    cardId: 0,
    cardName: '',
    reason: 'deck_size',
    details: 'Deck has 45 cards (minimum 60 required)',
  },
];

const mockWarnings: ValidationWarning[] = [
  {
    cardId: 99999,
    cardName: 'Unknown Card',
    type: 'unknown_legality',
    details: 'Card legality information not available',
  },
];

describe('LegalityBanner', () => {
  describe('Rendering', () => {
    it('does not render when deck is legal and has no warnings', () => {
      const { container } = render(
        <LegalityBanner
          isLegal={true}
          errors={[]}
          warnings={[]}
          format="standard"
        />
      );

      expect(container.firstChild).toBeNull();
    });

    it('renders when deck has banned cards', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockBannedErrors}
          warnings={[]}
          format="standard"
        />
      );

      expect(screen.getByText('Deck Contains Banned Cards')).toBeInTheDocument();
    });

    it('renders when deck has not legal cards', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockNotLegalErrors}
          warnings={[]}
          format="standard"
        />
      );

      expect(screen.getByText('Deck Not Legal in Standard')).toBeInTheDocument();
    });

    it('renders warnings even when deck is legal', () => {
      render(
        <LegalityBanner
          isLegal={true}
          errors={[]}
          warnings={mockWarnings}
          format="standard"
        />
      );

      expect(screen.getByText('Standard Legality Warnings')).toBeInTheDocument();
    });
  });

  describe('Urgency Levels', () => {
    it('shows critical urgency for banned cards', () => {
      const { container } = render(
        <LegalityBanner
          isLegal={false}
          errors={mockBannedErrors}
          warnings={[]}
          format="standard"
        />
      );

      expect(container.querySelector('.legality-banner--critical')).toBeInTheDocument();
    });

    it('shows warning urgency for not legal cards', () => {
      const { container } = render(
        <LegalityBanner
          isLegal={false}
          errors={mockNotLegalErrors}
          warnings={[]}
          format="standard"
        />
      );

      expect(container.querySelector('.legality-banner--warning')).toBeInTheDocument();
    });

    it('shows info urgency for warnings only', () => {
      const { container } = render(
        <LegalityBanner
          isLegal={true}
          errors={[]}
          warnings={mockWarnings}
          format="standard"
        />
      );

      expect(container.querySelector('.legality-banner--info')).toBeInTheDocument();
    });
  });

  describe('Error Counts', () => {
    it('shows count of banned cards', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockBannedErrors}
          warnings={[]}
          format="standard"
        />
      );

      expect(screen.getByText(/1 banned card/)).toBeInTheDocument();
    });

    it('shows count of not legal cards', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockNotLegalErrors}
          warnings={[]}
          format="standard"
        />
      );

      expect(screen.getByText(/1 card not legal/)).toBeInTheDocument();
    });

    it('shows count of copy limit violations', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockCopyErrors}
          warnings={[]}
          format="standard"
        />
      );

      expect(screen.getByText(/1 card exceed 4-copy limit/)).toBeInTheDocument();
    });

    it('shows multiple error types', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={[...mockBannedErrors, ...mockNotLegalErrors]}
          warnings={[]}
          format="standard"
        />
      );

      expect(screen.getByText(/1 banned card/)).toBeInTheDocument();
      expect(screen.getByText(/1 card not legal/)).toBeInTheDocument();
    });
  });

  describe('Expandable Details', () => {
    it('shows Details button', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockBannedErrors}
          warnings={[]}
          format="standard"
        />
      );

      expect(screen.getByText('Details')).toBeInTheDocument();
    });

    it('does not show details by default', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockBannedErrors}
          warnings={[]}
          format="standard"
        />
      );

      expect(screen.queryByText('Banned Cards')).not.toBeInTheDocument();
    });

    it('shows details when expanded', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockBannedErrors}
          warnings={[]}
          format="standard"
        />
      );

      fireEvent.click(screen.getByText('Details'));

      expect(screen.getByText('Banned Cards')).toBeInTheDocument();
      expect(screen.getByText('Omnath, Locus of Creation')).toBeInTheDocument();
    });

    it('shows banned badge for banned cards', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockBannedErrors}
          warnings={[]}
          format="standard"
        />
      );

      fireEvent.click(screen.getByText('Details'));

      expect(screen.getByText('Banned')).toBeInTheDocument();
    });

    it('shows not legal section', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockNotLegalErrors}
          warnings={[]}
          format="standard"
        />
      );

      fireEvent.click(screen.getByText('Details'));

      expect(screen.getByText('Not Legal in Standard')).toBeInTheDocument();
      expect(screen.getByText('Some Old Card')).toBeInTheDocument();
    });

    it('shows too many copies section', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockCopyErrors}
          warnings={[]}
          format="standard"
        />
      );

      fireEvent.click(screen.getByText('Details'));

      expect(screen.getByText('Too Many Copies')).toBeInTheDocument();
      expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
    });

    it('shows deck size section', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockDeckSizeErrors}
          warnings={[]}
          format="standard"
        />
      );

      fireEvent.click(screen.getByText('Details'));

      expect(screen.getByText('Deck Size')).toBeInTheDocument();
      expect(screen.getByText('Deck has 45 cards (minimum 60 required)')).toBeInTheDocument();
    });

    it('shows warnings section', () => {
      render(
        <LegalityBanner
          isLegal={true}
          errors={[]}
          warnings={mockWarnings}
          format="standard"
        />
      );

      fireEvent.click(screen.getByText('Details'));

      expect(screen.getByText('Warnings')).toBeInTheDocument();
      expect(screen.getByText('Unknown Card')).toBeInTheDocument();
    });

    it('toggles button text when expanded', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockBannedErrors}
          warnings={[]}
          format="standard"
        />
      );

      fireEvent.click(screen.getByText('Details'));
      expect(screen.getByText('Hide')).toBeInTheDocument();

      fireEvent.click(screen.getByText('Hide'));
      expect(screen.getByText('Details')).toBeInTheDocument();
    });
  });

  describe('Dismiss Button', () => {
    it('does not show dismiss button when onDismiss not provided', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockBannedErrors}
          warnings={[]}
          format="standard"
        />
      );

      expect(screen.queryByLabelText('Dismiss notification')).not.toBeInTheDocument();
    });

    it('shows dismiss button when onDismiss is provided', () => {
      const onDismiss = vi.fn();
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockBannedErrors}
          warnings={[]}
          format="standard"
          onDismiss={onDismiss}
        />
      );

      expect(screen.getByLabelText('Dismiss notification')).toBeInTheDocument();
    });

    it('calls onDismiss when clicked', () => {
      const onDismiss = vi.fn();
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockBannedErrors}
          warnings={[]}
          format="standard"
          onDismiss={onDismiss}
        />
      );

      fireEvent.click(screen.getByLabelText('Dismiss notification'));
      expect(onDismiss).toHaveBeenCalledOnce();
    });
  });

  describe('Compact Mode', () => {
    it('renders compact variant', () => {
      const { container } = render(
        <LegalityBanner
          isLegal={false}
          errors={mockBannedErrors}
          warnings={[]}
          format="standard"
          compact
        />
      );

      expect(container.querySelector('.legality-banner--compact')).toBeInTheDocument();
    });

    it('shows issue count in compact mode', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={[...mockBannedErrors, ...mockNotLegalErrors]}
          warnings={mockWarnings}
          format="standard"
          compact
        />
      );

      expect(screen.getByText(/3 legality issues in Standard/)).toBeInTheDocument();
    });

    it('shows singular form for 1 issue', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockBannedErrors}
          warnings={[]}
          format="standard"
          compact
        />
      );

      expect(screen.getByText(/1 legality issue in Standard/)).toBeInTheDocument();
    });
  });

  describe('Format Name', () => {
    it('capitalizes format name', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockNotLegalErrors}
          warnings={[]}
          format="historic"
        />
      );

      expect(screen.getByText('Deck Not Legal in Historic')).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('has accessible expand button', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockBannedErrors}
          warnings={[]}
          format="standard"
        />
      );

      const button = screen.getByText('Details');
      expect(button).toHaveAttribute('aria-label', 'Expand details');
    });

    it('updates aria-label when expanded', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={mockBannedErrors}
          warnings={[]}
          format="standard"
        />
      );

      fireEvent.click(screen.getByText('Details'));
      expect(screen.getByText('Hide')).toHaveAttribute('aria-label', 'Collapse details');
    });
  });

  describe('Edge Cases', () => {
    it('handles cards without names', () => {
      const errorWithoutName: ValidationError[] = [
        {
          cardId: 12345,
          cardName: '',
          reason: 'banned',
          details: 'Card is banned',
        },
      ];

      render(
        <LegalityBanner
          isLegal={false}
          errors={errorWithoutName}
          warnings={[]}
          format="standard"
        />
      );

      fireEvent.click(screen.getByText('Details'));
      expect(screen.getByText('Card #12345')).toBeInTheDocument();
    });

    it('handles all error types together', () => {
      render(
        <LegalityBanner
          isLegal={false}
          errors={[...mockBannedErrors, ...mockNotLegalErrors, ...mockCopyErrors, ...mockDeckSizeErrors]}
          warnings={mockWarnings}
          format="standard"
        />
      );

      fireEvent.click(screen.getByText('Details'));

      expect(screen.getByText('Banned Cards')).toBeInTheDocument();
      expect(screen.getByText('Not Legal in Standard')).toBeInTheDocument();
      expect(screen.getByText('Too Many Copies')).toBeInTheDocument();
      expect(screen.getByText('Deck Size')).toBeInTheDocument();
      expect(screen.getByText('Warnings')).toBeInTheDocument();
    });
  });
});
