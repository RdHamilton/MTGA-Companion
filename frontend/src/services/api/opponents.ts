import { get } from '../apiClient';

// Types for opponent analysis

export interface OpponentDeckProfile {
  id: number;
  matchId: string;
  detectedArchetype: string | null;
  archetypeConfidence: number;
  colorIdentity: string;
  deckStyle: string | null; // aggro, control, midrange, combo
  cardsObserved: number;
  estimatedDeckSize: number;
  observedCardIds: string | null; // JSON array
  inferredCardIds: string | null; // JSON array
  signatureCards: string | null; // JSON array
  format: string | null;
  metaArchetypeId: number | null;
  createdAt: string;
  updatedAt: string;
}

export interface ObservedCard {
  cardId: number;
  cardName: string;
  zone: string;
  turnFirstSeen: number;
  timesSeen: number;
  isSignature: boolean;
  category: string | null;
}

export interface ExpectedCard {
  cardId: number;
  cardName: string;
  inclusionRate: number;
  avgCopies: number;
  wasSeen: boolean;
  category: string;
  playAround: string;
}

export interface StrategicInsight {
  type: string;
  description: string;
  priority: 'high' | 'medium' | 'low';
  cards: number[];
}

export interface MetaArchetypeMatch {
  archetypeId: number;
  archetypeName: string;
  metaShare: number;
  tier: number;
  confidence: number;
  source: string;
}

export interface MatchupStatistic {
  id: number;
  accountId: number;
  playerArchetype: string;
  opponentArchetype: string;
  format: string;
  totalMatches: number;
  wins: number;
  losses: number;
  winRate: number;
  avgGameDuration: number | null;
  lastMatchAt: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface OpponentAnalysis {
  profile: OpponentDeckProfile | null;
  observedCards: ObservedCard[];
  expectedCards: ExpectedCard[];
  strategicInsights: StrategicInsight[];
  matchupStats: MatchupStatistic | null;
  metaArchetype: MetaArchetypeMatch | null;
}

export interface ArchetypeBreakdownEntry {
  archetype: string;
  count: number;
  percentage: number;
  winRate: number;
}

export interface ColorIdentityStatsEntry {
  colorIdentity: string;
  count: number;
  percentage: number;
  winRate: number;
}

export interface OpponentHistorySummary {
  totalOpponents: number;
  uniqueArchetypes: number;
  mostCommonArchetype: string;
  mostCommonCount: number;
  archetypeBreakdown: ArchetypeBreakdownEntry[];
  colorIdentityStats: ColorIdentityStatsEntry[];
}

export interface ArchetypeExpectedCard {
  id: number;
  archetypeName: string;
  format: string;
  cardId: number;
  cardName: string;
  inclusionRate: number;
  avgCopies: number;
  isSignature: boolean;
  category: string | null;
  createdAt: string;
}

// API Functions

/**
 * Get opponent analysis for a specific match
 */
export async function getOpponentAnalysis(matchId: string): Promise<OpponentAnalysis> {
  return get<OpponentAnalysis>(`/matches/${matchId}/opponent-analysis`);
}

/**
 * List reconstructed opponent deck profiles
 */
export async function listOpponentDecks(params?: {
  archetype?: string;
  format?: string;
  minConfidence?: number;
  limit?: number;
}): Promise<{ profiles: OpponentDeckProfile[]; total: number }> {
  const searchParams = new URLSearchParams();
  if (params?.archetype) searchParams.set('archetype', params.archetype);
  if (params?.format) searchParams.set('format', params.format);
  if (params?.minConfidence !== undefined) {
    searchParams.set('min_confidence', params.minConfidence.toString());
  }
  if (params?.limit !== undefined) {
    searchParams.set('limit', params.limit.toString());
  }

  const query = searchParams.toString();
  const url = `/opponents/decks${query ? `?${query}` : ''}`;
  return get<{ profiles: OpponentDeckProfile[]; total: number }>(url);
}

/**
 * Get matchup statistics
 */
export async function getMatchupStats(format?: string): Promise<{
  matchups: MatchupStatistic[];
  total: number;
}> {
  const params = format ? `?format=${encodeURIComponent(format)}` : '';
  return get<{ matchups: MatchupStatistic[]; total: number }>(
    `/analytics/matchups${params}`
  );
}

/**
 * Get opponent history summary
 */
export async function getOpponentHistory(
  format?: string
): Promise<OpponentHistorySummary> {
  const params = format ? `?format=${encodeURIComponent(format)}` : '';
  return get<OpponentHistorySummary>(`/analytics/opponent-history${params}`);
}

/**
 * Get expected cards for an archetype
 */
export async function getExpectedCards(
  archetypeName: string,
  format?: string
): Promise<{
  archetype: string;
  format: string;
  expectedCards: ArchetypeExpectedCard[];
  total: number;
}> {
  const params = format ? `?format=${encodeURIComponent(format)}` : '';
  return get<{
    archetype: string;
    format: string;
    expectedCards: ArchetypeExpectedCard[];
    total: number;
  }>(`/archetypes/${encodeURIComponent(archetypeName)}/expected-cards${params}`);
}

// Utility functions

/**
 * Parse card IDs from JSON string
 */
export function parseCardIds(jsonStr: string | null): number[] {
  if (!jsonStr) return [];
  try {
    return JSON.parse(jsonStr);
  } catch {
    return [];
  }
}

/**
 * Get deck style display name
 */
export function getDeckStyleDisplayName(style: string | null): string {
  if (!style) return 'Unknown';
  const styles: Record<string, string> = {
    aggro: 'Aggro',
    midrange: 'Midrange',
    control: 'Control',
    combo: 'Combo',
    tempo: 'Tempo',
  };
  return styles[style.toLowerCase()] || style;
}

/**
 * Get priority color class
 */
export function getPriorityColorClass(
  priority: 'high' | 'medium' | 'low'
): string {
  switch (priority) {
    case 'high':
      return 'text-red-400';
    case 'medium':
      return 'text-yellow-400';
    case 'low':
      return 'text-blue-400';
    default:
      return 'text-gray-400';
  }
}

/**
 * Get card category display name
 */
export function getCategoryDisplayName(category: string | null): string {
  if (!category) return 'Unknown';
  const categories: Record<string, string> = {
    removal: 'Removal',
    threat: 'Threat',
    interaction: 'Interaction',
    wincon: 'Win Condition',
    utility: 'Utility',
    ramp: 'Ramp',
    card_draw: 'Card Draw',
  };
  return categories[category.toLowerCase()] || category;
}

/**
 * Format confidence as percentage
 */
export function formatConfidence(confidence: number): string {
  return `${Math.round(confidence * 100)}%`;
}

/**
 * Get confidence color class based on value
 */
export function getConfidenceColorClass(confidence: number): string {
  if (confidence >= 0.7) return 'text-green-400';
  if (confidence >= 0.5) return 'text-yellow-400';
  return 'text-gray-400';
}
