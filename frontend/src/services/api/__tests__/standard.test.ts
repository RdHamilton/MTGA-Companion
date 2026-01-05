import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as standard from '../standard';

// Mock the apiClient
vi.mock('../../apiClient', () => ({
  get: vi.fn(),
  post: vi.fn(),
}));

import { get, post } from '../../apiClient';

describe('standard API', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('getStandardSets', () => {
    it('should call get with correct path', async () => {
      const mockSets: standard.StandardSet[] = [
        {
          code: 'dsk',
          name: 'Duskmourn',
          releasedAt: '2024-09-27',
          isStandardLegal: true,
          iconSvgUri: 'https://example.com/dsk.svg',
          cardCount: 291,
          daysUntilRotation: 365,
          isRotatingSoon: false,
        },
        {
          code: 'fdn',
          name: 'Foundations',
          releasedAt: '2024-11-15',
          isStandardLegal: true,
          iconSvgUri: 'https://example.com/fdn.svg',
          cardCount: 271,
          isRotatingSoon: false,
        },
      ];
      vi.mocked(get).mockResolvedValue(mockSets);

      const result = await standard.getStandardSets();

      expect(get).toHaveBeenCalledWith('/standard/sets');
      expect(result).toEqual(mockSets);
      expect(result[0].code).toBe('dsk');
      expect(result[1].isStandardLegal).toBe(true);
    });

    it('should return empty array when no sets', async () => {
      vi.mocked(get).mockResolvedValue([]);

      const result = await standard.getStandardSets();

      expect(result).toEqual([]);
    });
  });

  describe('getUpcomingRotation', () => {
    it('should call get with correct path', async () => {
      const mockRotation: standard.UpcomingRotation = {
        nextRotationDate: '2027-01-01',
        daysUntilRotation: 365,
        rotatingSets: [
          {
            code: 'mkm',
            name: 'Murders at Karlov Manor',
            releasedAt: '2024-02-09',
            isStandardLegal: true,
            iconSvgUri: 'https://example.com/mkm.svg',
            cardCount: 286,
            rotationDate: '2027-01-01',
            daysUntilRotation: 365,
            isRotatingSoon: true,
          },
        ],
        rotatingCardCount: 286,
        affectedDecks: 5,
      };
      vi.mocked(get).mockResolvedValue(mockRotation);

      const result = await standard.getUpcomingRotation();

      expect(get).toHaveBeenCalledWith('/standard/rotation');
      expect(result).toEqual(mockRotation);
      expect(result.daysUntilRotation).toBe(365);
      expect(result.rotatingSets[0].code).toBe('mkm');
    });
  });

  describe('getRotationAffectedDecks', () => {
    it('should call get with correct path', async () => {
      const mockDecks: standard.RotationAffectedDeck[] = [
        {
          deckId: 'deck-1',
          deckName: 'Mono Red Aggro',
          format: 'Standard',
          rotatingCardCount: 12,
          totalCards: 60,
          percentAffected: 20,
          rotatingCards: [],
        },
        {
          deckId: 'deck-2',
          deckName: 'Dimir Control',
          format: 'Standard',
          rotatingCardCount: 8,
          totalCards: 60,
          percentAffected: 13.3,
          rotatingCards: [],
        },
      ];
      vi.mocked(get).mockResolvedValue(mockDecks);

      const result = await standard.getRotationAffectedDecks();

      expect(get).toHaveBeenCalledWith('/standard/rotation/affected-decks');
      expect(result).toEqual(mockDecks);
      expect(result.length).toBe(2);
      expect(result[0].percentAffected).toBe(20);
    });

    it('should return empty array when no affected decks', async () => {
      vi.mocked(get).mockResolvedValue([]);

      const result = await standard.getRotationAffectedDecks();

      expect(result).toEqual([]);
    });
  });

  describe('getStandardConfig', () => {
    it('should call get with correct path', async () => {
      const mockConfig: standard.StandardConfig = {
        id: 1,
        nextRotationDate: '2027-01-01',
        rotationEnabled: true,
        updatedAt: '2024-01-01T00:00:00Z',
      };
      vi.mocked(get).mockResolvedValue(mockConfig);

      const result = await standard.getStandardConfig();

      expect(get).toHaveBeenCalledWith('/standard/config');
      expect(result).toEqual(mockConfig);
      expect(result.rotationEnabled).toBe(true);
    });

    it('should handle disabled rotation', async () => {
      const mockConfig: standard.StandardConfig = {
        id: 1,
        nextRotationDate: '',
        rotationEnabled: false,
        updatedAt: '2024-01-01T00:00:00Z',
      };
      vi.mocked(get).mockResolvedValue(mockConfig);

      const result = await standard.getStandardConfig();

      expect(result.rotationEnabled).toBe(false);
    });
  });

  describe('validateDeckStandard', () => {
    it('should call post with deck ID', async () => {
      const mockResult: standard.DeckValidationResult = {
        isLegal: true,
        errors: [],
        warnings: [],
        rotatingCards: [],
        setBreakdown: [],
      };
      vi.mocked(post).mockResolvedValue(mockResult);

      const result = await standard.validateDeckStandard('deck-123');

      expect(post).toHaveBeenCalledWith('/standard/validate/deck-123');
      expect(result).toEqual(mockResult);
      expect(result.isLegal).toBe(true);
    });

    it('should return validation errors for illegal deck', async () => {
      const mockResult: standard.DeckValidationResult = {
        isLegal: false,
        errors: [
          {
            cardId: 12345,
            cardName: 'Banned Card',
            reason: 'banned',
            details: 'Card is banned in Standard',
          },
          {
            cardId: 0,
            cardName: '',
            reason: 'deck_size',
            details: 'Deck has 40 cards (minimum 60 required)',
          },
        ],
        warnings: [
          {
            cardId: 67890,
            cardName: 'Unknown Card',
            type: 'unknown_legality',
            details: 'Card legality information not available',
          },
        ],
        rotatingCards: [],
        setBreakdown: [],
      };
      vi.mocked(post).mockResolvedValue(mockResult);

      const result = await standard.validateDeckStandard('deck-456');

      expect(result.isLegal).toBe(false);
      expect(result.errors.length).toBe(2);
      expect(result.errors[0].reason).toBe('banned');
      expect(result.warnings.length).toBe(1);
    });

    it('should include rotating cards in result', async () => {
      const mockResult: standard.DeckValidationResult = {
        isLegal: true,
        errors: [],
        warnings: [],
        rotatingCards: [
          {
            cardId: 11111,
            cardName: 'Rotating Card',
            setCode: 'mkm',
            setName: 'Murders at Karlov Manor',
            rotationDate: '2027-01-01',
            daysUntilRotation: 365,
          },
        ],
        setBreakdown: [
          {
            setCode: 'mkm',
            setName: 'Murders at Karlov Manor',
            cardCount: 15,
            iconSvgUri: 'https://example.com/mkm.svg',
            isRotating: true,
          },
        ],
      };
      vi.mocked(post).mockResolvedValue(mockResult);

      const result = await standard.validateDeckStandard('deck-789');

      expect(result.rotatingCards.length).toBe(1);
      expect(result.rotatingCards[0].daysUntilRotation).toBe(365);
      expect(result.setBreakdown[0].isRotating).toBe(true);
    });
  });

  describe('getCardLegality', () => {
    it('should call get with arena ID', async () => {
      const mockLegality: standard.CardLegality = {
        standard: 'legal',
        historic: 'legal',
        explorer: 'legal',
        pioneer: 'legal',
        modern: 'legal',
        alchemy: 'legal',
        brawl: 'legal',
        commander: 'legal',
      };
      vi.mocked(get).mockResolvedValue(mockLegality);

      const result = await standard.getCardLegality('12345');

      expect(get).toHaveBeenCalledWith('/standard/cards/12345/legality');
      expect(result).toEqual(mockLegality);
      expect(result.standard).toBe('legal');
    });

    it('should return banned legality', async () => {
      const mockLegality: standard.CardLegality = {
        standard: 'banned',
        historic: 'legal',
        explorer: 'legal',
        pioneer: 'legal',
        modern: 'legal',
        alchemy: 'legal',
        brawl: 'legal',
        commander: 'legal',
      };
      vi.mocked(get).mockResolvedValue(mockLegality);

      const result = await standard.getCardLegality('67890');

      expect(result.standard).toBe('banned');
    });

    it('should return not_legal for non-Standard cards', async () => {
      const mockLegality: standard.CardLegality = {
        standard: 'not_legal',
        historic: 'legal',
        explorer: 'not_legal',
        pioneer: 'not_legal',
        modern: 'legal',
        alchemy: 'legal',
        brawl: 'not_legal',
        commander: 'legal',
      };
      vi.mocked(get).mockResolvedValue(mockLegality);

      const result = await standard.getCardLegality('11111');

      expect(result.standard).toBe('not_legal');
      expect(result.historic).toBe('legal');
    });
  });
});
