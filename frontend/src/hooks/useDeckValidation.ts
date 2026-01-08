import { useState, useCallback } from 'react';
import * as standardApi from '@/services/api/standard';
import type { DeckValidationResult, ValidationError, ValidationWarning } from '@/services/api/standard';

export interface DeckValidationState {
  validation: DeckValidationResult | null;
  isValidating: boolean;
  error: string | null;
  lastValidated: Date | null;
}

export interface UseDeckValidationReturn extends DeckValidationState {
  validateDeck: (deckId: string) => Promise<void>;
  clearValidation: () => void;
  getBannedCards: () => ValidationError[];
  getNotLegalCards: () => ValidationError[];
  getWarnings: () => ValidationWarning[];
  hasLegalityIssues: () => boolean;
}

export function useDeckValidation(): UseDeckValidationReturn {
  const [state, setState] = useState<DeckValidationState>({
    validation: null,
    isValidating: false,
    error: null,
    lastValidated: null,
  });

  const validateDeck = useCallback(async (deckId: string) => {
    if (!deckId) return;

    setState((prev) => ({ ...prev, isValidating: true, error: null }));

    try {
      const result = await standardApi.validateDeckStandard(deckId);
      setState({
        validation: result,
        isValidating: false,
        error: null,
        lastValidated: new Date(),
      });
    } catch (err) {
      console.error('Failed to validate deck:', err);
      setState((prev) => ({
        ...prev,
        isValidating: false,
        error: err instanceof Error ? err.message : 'Failed to validate deck',
      }));
    }
  }, []);

  const clearValidation = useCallback(() => {
    setState({
      validation: null,
      isValidating: false,
      error: null,
      lastValidated: null,
    });
  }, []);

  const getBannedCards = useCallback((): ValidationError[] => {
    if (!state.validation) return [];
    return state.validation.errors.filter((e) => e.reason === 'banned');
  }, [state.validation]);

  const getNotLegalCards = useCallback((): ValidationError[] => {
    if (!state.validation) return [];
    return state.validation.errors.filter((e) => e.reason === 'not_legal');
  }, [state.validation]);

  const getWarnings = useCallback((): ValidationWarning[] => {
    if (!state.validation) return [];
    return state.validation.warnings;
  }, [state.validation]);

  const hasLegalityIssues = useCallback((): boolean => {
    if (!state.validation) return false;
    return !state.validation.isLegal;
  }, [state.validation]);

  return {
    ...state,
    validateDeck,
    clearValidation,
    getBannedCards,
    getNotLegalCards,
    getWarnings,
    hasLegalityIssues,
  };
}
