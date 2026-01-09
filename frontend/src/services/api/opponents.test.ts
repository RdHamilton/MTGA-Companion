import { describe, it, expect } from 'vitest';
import {
  parseCardIds,
  getDeckStyleDisplayName,
  getPriorityColorClass,
  getCategoryDisplayName,
  formatConfidence,
  getConfidenceColorClass,
} from './opponents';

describe('opponents API utilities', () => {
  describe('parseCardIds', () => {
    it('parses valid JSON array of numbers', () => {
      const result = parseCardIds('[1, 2, 3, 4]');
      expect(result).toEqual([1, 2, 3, 4]);
    });

    it('returns empty array for null input', () => {
      const result = parseCardIds(null);
      expect(result).toEqual([]);
    });

    it('returns empty array for invalid JSON', () => {
      const result = parseCardIds('not valid json');
      expect(result).toEqual([]);
    });

    it('returns empty array for empty string', () => {
      const result = parseCardIds('');
      expect(result).toEqual([]);
    });

    it('handles empty JSON array', () => {
      const result = parseCardIds('[]');
      expect(result).toEqual([]);
    });
  });

  describe('getDeckStyleDisplayName', () => {
    it('returns Aggro for aggro style', () => {
      expect(getDeckStyleDisplayName('aggro')).toBe('Aggro');
    });

    it('returns Midrange for midrange style', () => {
      expect(getDeckStyleDisplayName('midrange')).toBe('Midrange');
    });

    it('returns Control for control style', () => {
      expect(getDeckStyleDisplayName('control')).toBe('Control');
    });

    it('returns Combo for combo style', () => {
      expect(getDeckStyleDisplayName('combo')).toBe('Combo');
    });

    it('returns Tempo for tempo style', () => {
      expect(getDeckStyleDisplayName('tempo')).toBe('Tempo');
    });

    it('returns Unknown for null style', () => {
      expect(getDeckStyleDisplayName(null)).toBe('Unknown');
    });

    it('handles uppercase input', () => {
      expect(getDeckStyleDisplayName('AGGRO')).toBe('Aggro');
    });

    it('returns original for unknown style', () => {
      expect(getDeckStyleDisplayName('unknown_style')).toBe('unknown_style');
    });
  });

  describe('getPriorityColorClass', () => {
    it('returns red for high priority', () => {
      expect(getPriorityColorClass('high')).toBe('text-red-400');
    });

    it('returns yellow for medium priority', () => {
      expect(getPriorityColorClass('medium')).toBe('text-yellow-400');
    });

    it('returns blue for low priority', () => {
      expect(getPriorityColorClass('low')).toBe('text-blue-400');
    });

    it('returns gray for unknown priority', () => {
      // Cast to test default case
      expect(getPriorityColorClass('unknown' as 'high' | 'medium' | 'low')).toBe('text-gray-400');
    });
  });

  describe('getCategoryDisplayName', () => {
    it('returns Removal for removal category', () => {
      expect(getCategoryDisplayName('removal')).toBe('Removal');
    });

    it('returns Threat for threat category', () => {
      expect(getCategoryDisplayName('threat')).toBe('Threat');
    });

    it('returns Interaction for interaction category', () => {
      expect(getCategoryDisplayName('interaction')).toBe('Interaction');
    });

    it('returns Win Condition for wincon category', () => {
      expect(getCategoryDisplayName('wincon')).toBe('Win Condition');
    });

    it('returns Utility for utility category', () => {
      expect(getCategoryDisplayName('utility')).toBe('Utility');
    });

    it('returns Ramp for ramp category', () => {
      expect(getCategoryDisplayName('ramp')).toBe('Ramp');
    });

    it('returns Card Draw for card_draw category', () => {
      expect(getCategoryDisplayName('card_draw')).toBe('Card Draw');
    });

    it('returns Unknown for null category', () => {
      expect(getCategoryDisplayName(null)).toBe('Unknown');
    });

    it('handles uppercase input', () => {
      expect(getCategoryDisplayName('REMOVAL')).toBe('Removal');
    });

    it('returns original for unknown category', () => {
      expect(getCategoryDisplayName('custom_category')).toBe('custom_category');
    });
  });

  describe('formatConfidence', () => {
    it('formats 0.85 as 85%', () => {
      expect(formatConfidence(0.85)).toBe('85%');
    });

    it('formats 1.0 as 100%', () => {
      expect(formatConfidence(1.0)).toBe('100%');
    });

    it('formats 0 as 0%', () => {
      expect(formatConfidence(0)).toBe('0%');
    });

    it('rounds to nearest integer', () => {
      expect(formatConfidence(0.456)).toBe('46%');
    });

    it('handles small decimals', () => {
      expect(formatConfidence(0.001)).toBe('0%');
    });
  });

  describe('getConfidenceColorClass', () => {
    it('returns green for confidence >= 0.7', () => {
      expect(getConfidenceColorClass(0.7)).toBe('text-green-400');
      expect(getConfidenceColorClass(0.85)).toBe('text-green-400');
      expect(getConfidenceColorClass(1.0)).toBe('text-green-400');
    });

    it('returns yellow for confidence >= 0.5 and < 0.7', () => {
      expect(getConfidenceColorClass(0.5)).toBe('text-yellow-400');
      expect(getConfidenceColorClass(0.6)).toBe('text-yellow-400');
      expect(getConfidenceColorClass(0.69)).toBe('text-yellow-400');
    });

    it('returns gray for confidence < 0.5', () => {
      expect(getConfidenceColorClass(0.49)).toBe('text-gray-400');
      expect(getConfidenceColorClass(0.3)).toBe('text-gray-400');
      expect(getConfidenceColorClass(0)).toBe('text-gray-400');
    });
  });
});
