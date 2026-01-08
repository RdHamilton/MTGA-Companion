import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { RotationBanner } from './RotationBanner';
import type { UpcomingRotation, RotationAffectedDeck } from '@/services/api/standard';

const mockRotation: UpcomingRotation = {
  nextRotationDate: '2025-09-01T00:00:00Z',
  daysUntilRotation: 30,
  rotatingSets: [
    {
      code: 'ONE',
      name: 'Phyrexia: All Will Be One',
      releasedAt: '2023-02-03',
      isStandardLegal: true,
      iconSvgUri: 'https://example.com/one.svg',
      cardCount: 300,
      isRotatingSoon: true,
    },
    {
      code: 'MOM',
      name: 'March of the Machine',
      releasedAt: '2023-04-21',
      isStandardLegal: true,
      iconSvgUri: 'https://example.com/mom.svg',
      cardCount: 291,
      isRotatingSoon: true,
    },
  ],
  rotatingCardCount: 150,
  affectedDecks: 3,
};

const mockAffectedDecks: RotationAffectedDeck[] = [
  {
    deckId: 'deck-1',
    deckName: 'Mono White Aggro',
    format: 'Standard',
    rotatingCardCount: 12,
    totalCards: 60,
    percentAffected: 20,
    rotatingCards: [],
  },
  {
    deckId: 'deck-2',
    deckName: 'Azorius Control',
    format: 'Standard',
    rotatingCardCount: 8,
    totalCards: 60,
    percentAffected: 13.3,
    rotatingCards: [],
  },
  {
    deckId: 'deck-3',
    deckName: 'Esper Midrange',
    format: 'Standard',
    rotatingCardCount: 15,
    totalCards: 60,
    percentAffected: 25,
    rotatingCards: [],
  },
];

const renderWithRouter = (component: React.ReactElement) => {
  return render(<MemoryRouter>{component}</MemoryRouter>);
};

describe('RotationBanner', () => {
  describe('Basic Rendering', () => {
    it('renders the banner with correct title', () => {
      renderWithRouter(
        <RotationBanner rotation={mockRotation} affectedDecks={mockAffectedDecks} />
      );

      expect(screen.getByText('Standard Rotation in 30 Days')).toBeInTheDocument();
    });

    it('renders singular day correctly', () => {
      const singleDayRotation = { ...mockRotation, daysUntilRotation: 1 };
      renderWithRouter(
        <RotationBanner rotation={singleDayRotation} affectedDecks={mockAffectedDecks} />
      );

      expect(screen.getByText('Standard Rotation in 1 Day')).toBeInTheDocument();
    });

    it('renders affected decks count correctly', () => {
      renderWithRouter(
        <RotationBanner rotation={mockRotation} affectedDecks={mockAffectedDecks} />
      );

      expect(screen.getByText(/3 of your decks will lose cards/)).toBeInTheDocument();
    });

    it('renders singular deck correctly', () => {
      const singleDeck = [mockAffectedDecks[0]];
      renderWithRouter(
        <RotationBanner rotation={mockRotation} affectedDecks={singleDeck} />
      );

      expect(screen.getByText(/1 of your deck will lose cards/)).toBeInTheDocument();
    });
  });

  describe('Urgency Levels', () => {
    it('renders critical urgency for <= 7 days', () => {
      const criticalRotation = { ...mockRotation, daysUntilRotation: 5 };
      const { container } = renderWithRouter(
        <RotationBanner rotation={criticalRotation} affectedDecks={mockAffectedDecks} />
      );

      expect(container.querySelector('.rotation-banner--critical')).toBeInTheDocument();
    });

    it('renders warning urgency for <= 30 days', () => {
      const warningRotation = { ...mockRotation, daysUntilRotation: 20 };
      const { container } = renderWithRouter(
        <RotationBanner rotation={warningRotation} affectedDecks={mockAffectedDecks} />
      );

      expect(container.querySelector('.rotation-banner--warning')).toBeInTheDocument();
    });

    it('renders info urgency for > 30 days', () => {
      const infoRotation = { ...mockRotation, daysUntilRotation: 60 };
      const { container } = renderWithRouter(
        <RotationBanner rotation={infoRotation} affectedDecks={mockAffectedDecks} />
      );

      expect(container.querySelector('.rotation-banner--info')).toBeInTheDocument();
    });

    it('shows correct icon for critical urgency', () => {
      const criticalRotation = { ...mockRotation, daysUntilRotation: 3 };
      renderWithRouter(
        <RotationBanner rotation={criticalRotation} affectedDecks={mockAffectedDecks} />
      );

      const icons = screen.getAllByText('!');
      expect(icons.length).toBeGreaterThan(0);
    });

    it('shows correct icon for info urgency', () => {
      const infoRotation = { ...mockRotation, daysUntilRotation: 60 };
      renderWithRouter(
        <RotationBanner rotation={infoRotation} affectedDecks={mockAffectedDecks} />
      );

      const icons = screen.getAllByText('i');
      expect(icons.length).toBeGreaterThan(0);
    });
  });

  describe('Expandable Details', () => {
    it('shows Details button', () => {
      renderWithRouter(
        <RotationBanner rotation={mockRotation} affectedDecks={mockAffectedDecks} />
      );

      expect(screen.getByText('Details')).toBeInTheDocument();
    });

    it('does not show details section by default', () => {
      renderWithRouter(
        <RotationBanner rotation={mockRotation} affectedDecks={mockAffectedDecks} />
      );

      expect(screen.queryByText('Rotating Sets')).not.toBeInTheDocument();
    });

    it('shows details section when expanded', () => {
      renderWithRouter(
        <RotationBanner rotation={mockRotation} affectedDecks={mockAffectedDecks} />
      );

      fireEvent.click(screen.getByText('Details'));

      expect(screen.getByText('Rotating Sets')).toBeInTheDocument();
      expect(screen.getByText('Affected Decks')).toBeInTheDocument();
    });

    it('toggles button text when expanded', () => {
      renderWithRouter(
        <RotationBanner rotation={mockRotation} affectedDecks={mockAffectedDecks} />
      );

      const button = screen.getByText('Details');
      fireEvent.click(button);

      expect(screen.getByText('Hide')).toBeInTheDocument();
      expect(screen.queryByText('Details')).not.toBeInTheDocument();
    });

    it('shows rotating sets in details', () => {
      renderWithRouter(
        <RotationBanner rotation={mockRotation} affectedDecks={mockAffectedDecks} />
      );

      fireEvent.click(screen.getByText('Details'));

      expect(screen.getByText('ONE')).toBeInTheDocument();
      expect(screen.getByText('Phyrexia: All Will Be One')).toBeInTheDocument();
      expect(screen.getByText('MOM')).toBeInTheDocument();
      expect(screen.getByText('March of the Machine')).toBeInTheDocument();
    });

    it('shows affected decks in details', () => {
      renderWithRouter(
        <RotationBanner rotation={mockRotation} affectedDecks={mockAffectedDecks} />
      );

      fireEvent.click(screen.getByText('Details'));

      expect(screen.getByText('Mono White Aggro')).toBeInTheDocument();
      expect(screen.getByText('Azorius Control')).toBeInTheDocument();
    });

    it('shows card count and percentage for affected decks', () => {
      renderWithRouter(
        <RotationBanner rotation={mockRotation} affectedDecks={mockAffectedDecks} />
      );

      fireEvent.click(screen.getByText('Details'));

      expect(screen.getByText('12 cards (20%)')).toBeInTheDocument();
      expect(screen.getByText('8 cards (13%)')).toBeInTheDocument();
    });
  });

  describe('More Decks Link', () => {
    it('does not show more link when <= 5 decks', () => {
      renderWithRouter(
        <RotationBanner rotation={mockRotation} affectedDecks={mockAffectedDecks} />
      );

      fireEvent.click(screen.getByText('Details'));

      expect(screen.queryByText(/\+\d+ more deck/)).not.toBeInTheDocument();
    });

    it('shows more link when > 5 decks', () => {
      const manyDecks: RotationAffectedDeck[] = Array.from({ length: 8 }, (_, i) => ({
        deckId: `deck-${i}`,
        deckName: `Deck ${i}`,
        format: 'Standard',
        rotatingCardCount: 10,
        totalCards: 60,
        percentAffected: 16.7,
        rotatingCards: [],
      }));

      renderWithRouter(
        <RotationBanner rotation={mockRotation} affectedDecks={manyDecks} />
      );

      fireEvent.click(screen.getByText('Details'));

      expect(screen.getByText('+3 more decks')).toBeInTheDocument();
    });

    it('shows correct singular form for 1 more deck', () => {
      const sixDecks: RotationAffectedDeck[] = Array.from({ length: 6 }, (_, i) => ({
        deckId: `deck-${i}`,
        deckName: `Deck ${i}`,
        format: 'Standard',
        rotatingCardCount: 10,
        totalCards: 60,
        percentAffected: 16.7,
        rotatingCards: [],
      }));

      renderWithRouter(
        <RotationBanner rotation={mockRotation} affectedDecks={sixDecks} />
      );

      fireEvent.click(screen.getByText('Details'));

      expect(screen.getByText('+1 more deck')).toBeInTheDocument();
    });
  });

  describe('Dismiss Button', () => {
    it('does not show dismiss button when onDismiss not provided', () => {
      renderWithRouter(
        <RotationBanner rotation={mockRotation} affectedDecks={mockAffectedDecks} />
      );

      expect(screen.queryByLabelText('Dismiss notification')).not.toBeInTheDocument();
    });

    it('shows dismiss button when onDismiss is provided', () => {
      const onDismiss = vi.fn();
      renderWithRouter(
        <RotationBanner
          rotation={mockRotation}
          affectedDecks={mockAffectedDecks}
          onDismiss={onDismiss}
        />
      );

      expect(screen.getByLabelText('Dismiss notification')).toBeInTheDocument();
    });

    it('calls onDismiss when dismiss button clicked', () => {
      const onDismiss = vi.fn();
      renderWithRouter(
        <RotationBanner
          rotation={mockRotation}
          affectedDecks={mockAffectedDecks}
          onDismiss={onDismiss}
        />
      );

      fireEvent.click(screen.getByLabelText('Dismiss notification'));

      expect(onDismiss).toHaveBeenCalledOnce();
    });
  });

  describe('Compact Mode', () => {
    it('renders compact variant', () => {
      const { container } = renderWithRouter(
        <RotationBanner
          rotation={mockRotation}
          affectedDecks={mockAffectedDecks}
          compact
        />
      );

      expect(container.querySelector('.rotation-banner--compact')).toBeInTheDocument();
    });

    it('shows simplified text in compact mode', () => {
      renderWithRouter(
        <RotationBanner
          rotation={mockRotation}
          affectedDecks={mockAffectedDecks}
          compact
        />
      );

      expect(screen.getByText(/3 decks affected by rotation in 30 days/)).toBeInTheDocument();
    });

    it('shows singular forms in compact mode', () => {
      const singleDeckSingleDay = { ...mockRotation, daysUntilRotation: 1 };
      const singleDeck = [mockAffectedDecks[0]];

      renderWithRouter(
        <RotationBanner
          rotation={singleDeckSingleDay}
          affectedDecks={singleDeck}
          compact
        />
      );

      expect(screen.getByText(/1 deck affected by rotation in 1 day/)).toBeInTheDocument();
    });

    it('shows View link in compact mode', () => {
      renderWithRouter(
        <RotationBanner
          rotation={mockRotation}
          affectedDecks={mockAffectedDecks}
          compact
        />
      );

      expect(screen.getByText('View')).toBeInTheDocument();
    });

    it('does not show Details button in compact mode', () => {
      renderWithRouter(
        <RotationBanner
          rotation={mockRotation}
          affectedDecks={mockAffectedDecks}
          compact
        />
      );

      expect(screen.queryByText('Details')).not.toBeInTheDocument();
    });
  });

  describe('Links', () => {
    it('renders deck links with correct href', () => {
      renderWithRouter(
        <RotationBanner rotation={mockRotation} affectedDecks={mockAffectedDecks} />
      );

      fireEvent.click(screen.getByText('Details'));

      const deckLink = screen.getByText('Mono White Aggro').closest('a');
      expect(deckLink).toHaveAttribute('href', '/decks/deck-1');
    });

    it('renders View link with filter in compact mode', () => {
      renderWithRouter(
        <RotationBanner
          rotation={mockRotation}
          affectedDecks={mockAffectedDecks}
          compact
        />
      );

      const viewLink = screen.getByText('View').closest('a');
      expect(viewLink).toHaveAttribute('href', '/decks?filter=rotating');
    });
  });

  describe('Date Formatting', () => {
    it('formats rotation date correctly', () => {
      renderWithRouter(
        <RotationBanner rotation={mockRotation} affectedDecks={mockAffectedDecks} />
      );

      // The date should be formatted as "Month Day, Year"
      // Note: exact date may vary based on locale/timezone (UTC to local conversion)
      // The date 2025-09-01T00:00:00Z may appear as Aug 31 or Sep 1 depending on timezone
      expect(screen.getByText(/(August 31|September 1), 2025/)).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('has accessible expand button', () => {
      renderWithRouter(
        <RotationBanner rotation={mockRotation} affectedDecks={mockAffectedDecks} />
      );

      const button = screen.getByText('Details');
      expect(button).toHaveAttribute('aria-label', 'Expand details');
    });

    it('updates aria-label when expanded', () => {
      renderWithRouter(
        <RotationBanner rotation={mockRotation} affectedDecks={mockAffectedDecks} />
      );

      const button = screen.getByText('Details');
      fireEvent.click(button);

      expect(screen.getByText('Hide')).toHaveAttribute('aria-label', 'Collapse details');
    });

    it('has accessible dismiss button', () => {
      const onDismiss = vi.fn();
      renderWithRouter(
        <RotationBanner
          rotation={mockRotation}
          affectedDecks={mockAffectedDecks}
          onDismiss={onDismiss}
        />
      );

      expect(screen.getByLabelText('Dismiss notification')).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('handles empty rotating sets', () => {
      const noSetsRotation = { ...mockRotation, rotatingSets: [] };
      renderWithRouter(
        <RotationBanner rotation={noSetsRotation} affectedDecks={mockAffectedDecks} />
      );

      fireEvent.click(screen.getByText('Details'));

      expect(screen.getByText('Rotating Sets')).toBeInTheDocument();
      // Should not crash, just show empty list
    });

    it('handles zero days until rotation', () => {
      const todayRotation = { ...mockRotation, daysUntilRotation: 0 };
      renderWithRouter(
        <RotationBanner rotation={todayRotation} affectedDecks={mockAffectedDecks} />
      );

      expect(screen.getByText('Standard Rotation in 0 Days')).toBeInTheDocument();
    });

    it('handles single card rotating', () => {
      const singleCardDeck: RotationAffectedDeck[] = [
        {
          deckId: 'deck-1',
          deckName: 'Deck',
          format: 'Standard',
          rotatingCardCount: 1,
          totalCards: 60,
          percentAffected: 1.7,
          rotatingCards: [],
        },
      ];

      renderWithRouter(
        <RotationBanner rotation={mockRotation} affectedDecks={singleCardDeck} />
      );

      fireEvent.click(screen.getByText('Details'));

      expect(screen.getByText('1 card (2%)')).toBeInTheDocument();
    });
  });
});
