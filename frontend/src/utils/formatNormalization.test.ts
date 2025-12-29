import { describe, it, expect } from 'vitest';
import { models } from '@/types/models';
import {
  normalizeQueueType,
  getDisplayFormat,
  getDisplayEventName,
} from './formatNormalization';

describe('formatNormalization', () => {
  describe('normalizeQueueType', () => {
    describe('basic queue types', () => {
      it('should normalize "Play" to "Play Queue"', () => {
        expect(normalizeQueueType('Play')).toBe('Play Queue');
      });

      it('should normalize "Ladder" to "Ranked"', () => {
        expect(normalizeQueueType('Ladder')).toBe('Ranked');
      });

      it('should normalize "Traditional_Ladder" to "Traditional Ranked"', () => {
        expect(normalizeQueueType('Traditional_Ladder')).toBe('Traditional Ranked');
      });

      it('should normalize "Traditional_Play" to "Traditional Play"', () => {
        expect(normalizeQueueType('Traditional_Play')).toBe('Traditional Play');
      });
    });

    describe('draft formats', () => {
      it('should normalize QuickDraft with set code and date', () => {
        expect(normalizeQueueType('QuickDraft_TLA_20251127')).toBe('QuickDraft');
      });

      it('should normalize PremierDraft with set code', () => {
        expect(normalizeQueueType('PremierDraft_MKM_20241120')).toBe('PremierDraft');
      });

      it('should normalize TradDraft to Traditional Draft', () => {
        expect(normalizeQueueType('TradDraft_DSK')).toBe('Traditional Draft');
      });

      it('should normalize SealedDeck to Sealed', () => {
        expect(normalizeQueueType('SealedDeck_BLB')).toBe('Sealed');
      });
    });

    describe('edge cases', () => {
      it('should return empty string for empty input', () => {
        expect(normalizeQueueType('')).toBe('');
      });

      it('should return unknown format as-is', () => {
        expect(normalizeQueueType('UnknownFormat')).toBe('UnknownFormat');
      });

      it('should handle unknown format with underscore by returning prefix', () => {
        expect(normalizeQueueType('CustomEvent_ABC_123')).toBe('CustomEvent');
      });
    });
  });

  describe('getDisplayFormat', () => {
    const createMatch = (overrides: Partial<models.Match> = {}): models.Match => {
      return new models.Match({
        ID: 'test-match',
        AccountID: 1,
        EventID: 'Ladder',
        EventName: 'Ladder',
        Format: 'Ladder',
        Result: 'Win',
        PlayerWins: 2,
        OpponentWins: 1,
        PlayerTeamID: 1,
        ...overrides,
      });
    };

    it('should return DeckFormat when available', () => {
      const match = createMatch({ DeckFormat: 'Standard', Format: 'Ladder' });
      expect(getDisplayFormat(match)).toBe('Standard');
    });

    it('should return DeckFormat for Historic', () => {
      const match = createMatch({ DeckFormat: 'Historic', Format: 'Play' });
      expect(getDisplayFormat(match)).toBe('Historic');
    });

    it('should fall back to normalized queue type when no DeckFormat', () => {
      const match = createMatch({ DeckFormat: undefined, Format: 'Ladder' });
      expect(getDisplayFormat(match)).toBe('Ranked');
    });

    it('should fall back to normalized queue type for Play', () => {
      const match = createMatch({ DeckFormat: undefined, Format: 'Play' });
      expect(getDisplayFormat(match)).toBe('Play Queue');
    });

    it('should handle draft formats without DeckFormat', () => {
      const match = createMatch({ DeckFormat: undefined, Format: 'QuickDraft_TLA_20251127' });
      expect(getDisplayFormat(match)).toBe('QuickDraft');
    });
  });

  describe('getDisplayEventName', () => {
    const createMatch = (overrides: Partial<models.Match> = {}): models.Match => {
      return new models.Match({
        ID: 'test-match',
        AccountID: 1,
        EventID: 'Ladder',
        EventName: 'Ladder',
        Format: 'Ladder',
        Result: 'Win',
        PlayerWins: 2,
        OpponentWins: 1,
        PlayerTeamID: 1,
        ...overrides,
      });
    };

    describe('constructed matches with DeckFormat', () => {
      it('should combine Standard + Ladder to "Standard Ranked"', () => {
        const match = createMatch({ DeckFormat: 'Standard', EventName: 'Ladder' });
        expect(getDisplayEventName(match)).toBe('Standard Ranked');
      });

      it('should combine Standard + Play to "Standard Play Queue"', () => {
        const match = createMatch({ DeckFormat: 'Standard', EventName: 'Play' });
        expect(getDisplayEventName(match)).toBe('Standard Play Queue');
      });

      it('should combine Historic + Ladder to "Historic Ranked"', () => {
        const match = createMatch({ DeckFormat: 'Historic', EventName: 'Ladder' });
        expect(getDisplayEventName(match)).toBe('Historic Ranked');
      });

      it('should combine Explorer + Traditional_Ladder', () => {
        const match = createMatch({ DeckFormat: 'Explorer', EventName: 'Traditional_Ladder' });
        expect(getDisplayEventName(match)).toBe('Explorer Traditional Ranked');
      });
    });

    describe('matches without DeckFormat', () => {
      it('should return normalized queue type for Ladder', () => {
        const match = createMatch({ DeckFormat: undefined, EventName: 'Ladder' });
        expect(getDisplayEventName(match)).toBe('Ranked');
      });

      it('should return normalized queue type for Play', () => {
        const match = createMatch({ DeckFormat: undefined, EventName: 'Play' });
        expect(getDisplayEventName(match)).toBe('Play Queue');
      });
    });

    describe('draft matches', () => {
      it('should return QuickDraft for draft events', () => {
        const match = createMatch({
          DeckFormat: undefined,
          EventName: 'QuickDraft_TLA_20251127',
        });
        expect(getDisplayEventName(match)).toBe('QuickDraft');
      });

      it('should return PremierDraft for premier draft events', () => {
        const match = createMatch({
          DeckFormat: undefined,
          EventName: 'PremierDraft_MKM_20241120',
        });
        expect(getDisplayEventName(match)).toBe('PremierDraft');
      });

      it('should not combine DeckFormat with draft queue types', () => {
        // Even if a draft deck has a format, the event name should just be the draft type
        const match = createMatch({
          DeckFormat: 'Limited',
          EventName: 'QuickDraft_TLA_20251127',
        });
        expect(getDisplayEventName(match)).toBe('QuickDraft');
      });
    });

    describe('fallback behavior', () => {
      it('should use Format when EventName is missing', () => {
        const match = createMatch({
          DeckFormat: 'Standard',
          EventName: '',
          Format: 'Ladder',
        });
        expect(getDisplayEventName(match)).toBe('Standard Ranked');
      });
    });
  });
});
