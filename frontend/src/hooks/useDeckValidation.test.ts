import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useDeckValidation } from './useDeckValidation';
import type { DeckValidationResult } from '@/services/api/standard';

// Mock the standard API module
vi.mock('@/services/api/standard', () => ({
  validateDeckStandard: vi.fn(),
}));

import * as standardApi from '@/services/api/standard';

const mockValidateDeckStandard = vi.mocked(standardApi.validateDeckStandard);

const mockValidationResult: DeckValidationResult = {
  isLegal: false,
  errors: [
    {
      cardId: 12345,
      cardName: 'Omnath, Locus of Creation',
      reason: 'banned',
      details: 'Card is banned in Standard',
    },
    {
      cardId: 67890,
      cardName: 'Some Old Card',
      reason: 'not_legal',
      details: 'Card is not legal in Standard',
    },
    {
      cardId: 11111,
      cardName: 'Lightning Bolt',
      reason: 'too_many_copies',
      details: 'Deck contains 5 copies (maximum 4 allowed)',
    },
  ],
  warnings: [
    {
      cardId: 99999,
      cardName: 'Unknown Card',
      type: 'unknown_legality',
      details: 'Card legality information not available',
    },
  ],
  rotatingCards: [],
  setBreakdown: [],
};

const mockLegalResult: DeckValidationResult = {
  isLegal: true,
  errors: [],
  warnings: [],
  rotatingCards: [],
  setBreakdown: [],
};

describe('useDeckValidation', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockValidateDeckStandard.mockResolvedValue(mockValidationResult);
  });

  describe('initial state', () => {
    it('starts with null validation', () => {
      const { result } = renderHook(() => useDeckValidation());

      expect(result.current.validation).toBeNull();
      expect(result.current.isValidating).toBe(false);
      expect(result.current.error).toBeNull();
      expect(result.current.lastValidated).toBeNull();
    });
  });

  describe('validateDeck', () => {
    it('calls API and sets validation result', async () => {
      const { result } = renderHook(() => useDeckValidation());

      await act(async () => {
        await result.current.validateDeck('deck-123');
      });

      expect(mockValidateDeckStandard).toHaveBeenCalledWith('deck-123');
      expect(result.current.validation).toEqual(mockValidationResult);
      expect(result.current.isValidating).toBe(false);
      expect(result.current.lastValidated).toBeInstanceOf(Date);
    });

    it('sets isValidating to true while loading', async () => {
      mockValidateDeckStandard.mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve(mockValidationResult), 100))
      );

      const { result } = renderHook(() => useDeckValidation());

      act(() => {
        result.current.validateDeck('deck-123');
      });

      expect(result.current.isValidating).toBe(true);

      await waitFor(() => {
        expect(result.current.isValidating).toBe(false);
      });
    });

    it('handles API errors', async () => {
      mockValidateDeckStandard.mockRejectedValue(new Error('Network error'));

      const { result } = renderHook(() => useDeckValidation());

      await act(async () => {
        await result.current.validateDeck('deck-123');
      });

      expect(result.current.error).toBe('Network error');
      expect(result.current.validation).toBeNull();
    });

    it('does nothing for empty deck ID', async () => {
      const { result } = renderHook(() => useDeckValidation());

      await act(async () => {
        await result.current.validateDeck('');
      });

      expect(mockValidateDeckStandard).not.toHaveBeenCalled();
    });
  });

  describe('clearValidation', () => {
    it('clears validation state', async () => {
      const { result } = renderHook(() => useDeckValidation());

      await act(async () => {
        await result.current.validateDeck('deck-123');
      });

      expect(result.current.validation).not.toBeNull();

      act(() => {
        result.current.clearValidation();
      });

      expect(result.current.validation).toBeNull();
      expect(result.current.error).toBeNull();
      expect(result.current.lastValidated).toBeNull();
    });
  });

  describe('getBannedCards', () => {
    it('returns empty array when no validation', () => {
      const { result } = renderHook(() => useDeckValidation());

      expect(result.current.getBannedCards()).toEqual([]);
    });

    it('returns only banned cards', async () => {
      const { result } = renderHook(() => useDeckValidation());

      await act(async () => {
        await result.current.validateDeck('deck-123');
      });

      const bannedCards = result.current.getBannedCards();
      expect(bannedCards).toHaveLength(1);
      expect(bannedCards[0].reason).toBe('banned');
      expect(bannedCards[0].cardName).toBe('Omnath, Locus of Creation');
    });
  });

  describe('getNotLegalCards', () => {
    it('returns empty array when no validation', () => {
      const { result } = renderHook(() => useDeckValidation());

      expect(result.current.getNotLegalCards()).toEqual([]);
    });

    it('returns only not_legal cards', async () => {
      const { result } = renderHook(() => useDeckValidation());

      await act(async () => {
        await result.current.validateDeck('deck-123');
      });

      const notLegalCards = result.current.getNotLegalCards();
      expect(notLegalCards).toHaveLength(1);
      expect(notLegalCards[0].reason).toBe('not_legal');
      expect(notLegalCards[0].cardName).toBe('Some Old Card');
    });
  });

  describe('getWarnings', () => {
    it('returns empty array when no validation', () => {
      const { result } = renderHook(() => useDeckValidation());

      expect(result.current.getWarnings()).toEqual([]);
    });

    it('returns warnings', async () => {
      const { result } = renderHook(() => useDeckValidation());

      await act(async () => {
        await result.current.validateDeck('deck-123');
      });

      const warnings = result.current.getWarnings();
      expect(warnings).toHaveLength(1);
      expect(warnings[0].type).toBe('unknown_legality');
    });
  });

  describe('hasLegalityIssues', () => {
    it('returns false when no validation', () => {
      const { result } = renderHook(() => useDeckValidation());

      expect(result.current.hasLegalityIssues()).toBe(false);
    });

    it('returns true when deck is not legal', async () => {
      const { result } = renderHook(() => useDeckValidation());

      await act(async () => {
        await result.current.validateDeck('deck-123');
      });

      expect(result.current.hasLegalityIssues()).toBe(true);
    });

    it('returns false when deck is legal', async () => {
      mockValidateDeckStandard.mockResolvedValue(mockLegalResult);

      const { result } = renderHook(() => useDeckValidation());

      await act(async () => {
        await result.current.validateDeck('deck-123');
      });

      expect(result.current.hasLegalityIssues()).toBe(false);
    });
  });
});
