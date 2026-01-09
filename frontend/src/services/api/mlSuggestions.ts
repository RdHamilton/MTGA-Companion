/**
 * ML Suggestions API service.
 * Handles ML-powered card suggestions based on synergy analysis.
 */

import { get, post, put } from '../apiClient';

/**
 * ML suggestion types.
 */
export type MLSuggestionType = 'add' | 'remove' | 'swap';

/**
 * A reason for an ML suggestion.
 */
export interface MLSuggestionReason {
  type: string;
  description: string;
  impact: number;
  confidence: number;
}

/**
 * An ML-powered suggestion for deck improvement.
 */
export interface MLSuggestion {
  id: number;
  deckId: string;
  suggestionType: MLSuggestionType;
  cardId?: number;
  cardName?: string;
  swapForCardId?: number;
  swapForCardName?: string;
  confidence: number;
  expectedWinRateChange: number;
  title: string;
  description?: string;
  reasoning?: string; // JSON array of reasons
  evidence?: string; // JSON object with supporting data
  isDismissed: boolean;
  wasApplied: boolean;
  outcomeWinRateChange?: number;
  createdAt: string;
  appliedAt?: string;
  outcomeRecordedAt?: string;
}

/**
 * Synergy info for a card pair.
 */
export interface CardSynergyInfo {
  cardId: number;
  cardName: string;
  synergyScore: number;
  winRateTogether: number;
  gamesTogether: number;
}

/**
 * ML suggestion result with additional context.
 */
export interface MLSuggestionResult {
  suggestion: MLSuggestion;
  synergyData?: CardSynergyInfo[];
  reasons: MLSuggestionReason[];
}

/**
 * Card combination statistics.
 */
export interface CardCombinationStats {
  id: number;
  cardId1: number;
  cardId2: number;
  deckId?: string;
  format: string;
  gamesTogether: number;
  gamesCard1Only: number;
  gamesCard2Only: number;
  winsTogether: number;
  winsCard1Only: number;
  winsCard2Only: number;
  synergyScore: number;
  confidenceScore: number;
  createdAt: string;
  updatedAt: string;
}

/**
 * Synergy between two cards.
 */
export interface CardPairSynergy {
  card1Id: number;
  card1Name?: string;
  card2Id: number;
  card2Name?: string;
  synergyScore: number;
  gamesTogether: number;
  winRate: number;
}

/**
 * Synergy report for a deck.
 */
export interface SynergyReport {
  deckId: string;
  cardCount: number;
  totalPairs: number;
  avgSynergyScore: number;
  synergies: CardPairSynergy[];
}

/**
 * User play patterns for personalization.
 */
export interface UserPlayPatterns {
  id: number;
  accountId: string;
  preferredArchetype?: string;
  aggroAffinity: number;
  midrangeAffinity: number;
  controlAffinity: number;
  comboAffinity: number;
  colorPreferences?: string; // JSON map
  avgGameLength: number;
  aggressionScore: number;
  interactionScore: number;
  totalMatches: number;
  totalDecks: number;
  createdAt: string;
  updatedAt: string;
}

// -------------------- ML Suggestions API --------------------

/**
 * Get ML-powered suggestions for a deck.
 * @param deckId - The deck ID
 * @param activeOnly - If true, only return non-dismissed suggestions (default: true)
 */
export async function getMLSuggestions(
  deckId: string,
  activeOnly: boolean = true
): Promise<MLSuggestion[]> {
  const params = activeOnly ? '' : '?active=false';
  return get<MLSuggestion[]>(`/decks/${deckId}/ml-suggestions${params}`);
}

/**
 * Generate new ML-powered suggestions for a deck.
 * Analyzes card synergies to suggest adds, removes, and swaps.
 * @param deckId - The deck ID
 */
export async function generateMLSuggestions(
  deckId: string
): Promise<MLSuggestionResult[]> {
  return post<MLSuggestionResult[]>(`/decks/${deckId}/ml-suggestions/generate`, {});
}

/**
 * Dismiss an ML suggestion.
 */
export async function dismissMLSuggestion(suggestionId: number): Promise<void> {
  return put(`/ml-suggestions/${suggestionId}/dismiss`, {});
}

/**
 * Mark an ML suggestion as applied.
 */
export async function applyMLSuggestion(suggestionId: number): Promise<void> {
  return put(`/ml-suggestions/${suggestionId}/apply`, {});
}

// -------------------- Synergy API --------------------

/**
 * Get synergy report for a deck.
 * Shows the best and worst synergies between cards in the deck.
 */
export async function getSynergyReport(deckId: string): Promise<SynergyReport> {
  return get<SynergyReport>(`/decks/${deckId}/synergy-report`);
}

/**
 * Get top synergistic cards for a given card.
 * @param cardId - The card's Arena ID
 * @param format - The format to check (default: Standard)
 * @param limit - Max results (default: 10, max: 50)
 */
export async function getCardSynergies(
  cardId: number,
  format: string = 'Standard',
  limit: number = 10
): Promise<CardCombinationStats[]> {
  const params = new URLSearchParams({
    format,
    limit: String(limit),
  });
  return get<CardCombinationStats[]>(`/cards/${cardId}/synergies?${params}`);
}

/**
 * Get combination stats for a specific card pair.
 * @param card1 - First card's Arena ID
 * @param card2 - Second card's Arena ID
 * @param format - The format (default: Standard)
 */
export async function getCombinationStats(
  card1: number,
  card2: number,
  format: string = 'Standard'
): Promise<CardCombinationStats> {
  const params = new URLSearchParams({
    card1: String(card1),
    card2: String(card2),
    format,
  });
  return get<CardCombinationStats>(`/ml/combinations?${params}`);
}

// -------------------- ML Management API --------------------

/**
 * Trigger processing of match history to build synergy data.
 * @param format - Optional format filter
 * @param days - Lookback period in days (default: 90, max: 365)
 */
export async function processMatchHistory(
  format?: string,
  days: number = 90
): Promise<{ status: string; message: string }> {
  const params = new URLSearchParams({
    days: String(days),
  });
  if (format) {
    params.set('format', format);
  }
  return post(`/ml/process-history?${params}`, {});
}

/**
 * Get user's play patterns profile.
 * @param accountId - Optional account ID (default: current user)
 */
export async function getUserPlayPatterns(
  accountId?: string
): Promise<UserPlayPatterns> {
  const params = accountId ? `?account_id=${accountId}` : '';
  return get<UserPlayPatterns>(`/ml/play-patterns${params}`);
}

/**
 * Update user's play patterns profile based on recent matches.
 * @param accountId - Optional account ID (default: current user)
 */
export async function updateUserPlayPatterns(
  accountId?: string
): Promise<{ status: string; message: string }> {
  const params = accountId ? `?account_id=${accountId}` : '';
  return post(`/ml/play-patterns/update${params}`, {});
}

// -------------------- Utility Functions --------------------

/**
 * Parse the reasoning JSON string from a suggestion.
 */
export function parseReasons(reasoning: string | undefined): MLSuggestionReason[] {
  if (!reasoning) return [];
  try {
    return JSON.parse(reasoning) as MLSuggestionReason[];
  } catch {
    return [];
  }
}

/**
 * Parse color preferences from JSON.
 */
export function parseColorPreferences(
  colorPreferences: string | undefined
): Record<string, number> {
  if (!colorPreferences) return {};
  try {
    return JSON.parse(colorPreferences) as Record<string, number>;
  } catch {
    return {};
  }
}

/**
 * Get display label for suggestion type.
 */
export function getMLSuggestionTypeLabel(type: MLSuggestionType): string {
  const labels: Record<MLSuggestionType, string> = {
    add: 'Add Card',
    remove: 'Remove Card',
    swap: 'Swap Cards',
  };
  return labels[type] || type;
}

/**
 * Get icon for suggestion type.
 */
export function getMLSuggestionTypeIcon(type: MLSuggestionType): string {
  const icons: Record<MLSuggestionType, string> = {
    add: '+',
    remove: '-',
    swap: '⇄',
  };
  return icons[type] || '?';
}

/**
 * Format confidence as percentage.
 */
export function formatConfidence(confidence: number): string {
  return `${Math.round(confidence * 100)}%`;
}

/**
 * Format win rate change with sign.
 */
export function formatWinRateChange(change: number): string {
  const sign = change >= 0 ? '+' : '';
  return `${sign}${change.toFixed(1)}%`;
}

/**
 * Get confidence level label.
 */
export function getConfidenceLevel(
  confidence: number
): 'low' | 'medium' | 'high' {
  if (confidence >= 0.7) return 'high';
  if (confidence >= 0.4) return 'medium';
  return 'low';
}

/**
 * Get color for confidence level.
 */
export function getConfidenceColor(confidence: number): string {
  const level = getConfidenceLevel(confidence);
  const colors: Record<string, string> = {
    low: 'text-yellow-400',
    medium: 'text-blue-400',
    high: 'text-green-400',
  };
  return colors[level] || 'text-gray-400';
}

/**
 * Get archetype display name.
 */
export function getArchetypeLabel(archetype: string): string {
  const labels: Record<string, string> = {
    Aggro: 'Aggro',
    Midrange: 'Midrange',
    Control: 'Control',
    Combo: 'Combo',
    Balanced: 'Balanced',
  };
  return labels[archetype] || archetype;
}
