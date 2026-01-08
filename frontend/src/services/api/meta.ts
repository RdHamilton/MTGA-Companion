/**
 * Meta API service.
 * Replaces Wails meta-related function bindings.
 */

import { get, post } from '../apiClient';
import { gui, insights } from '@/types/models';

/**
 * Archetype info.
 */
export interface ArchetypeInfo {
  name: string;
  colors: string;
  tier: number;
  winRate: number;
  playRate: number;
  description?: string;
}

/**
 * Deck analysis result.
 */
export interface DeckAnalysisResult {
  archetype: string;
  confidence: number;
  strengths: string[];
  weaknesses: string[];
}

/**
 * Get meta archetypes for a format.
 */
export async function getMetaArchetypes(format: string): Promise<ArchetypeInfo[]> {
  const params = new URLSearchParams({ format });
  return get<ArchetypeInfo[]>(`/meta/archetypes?${params.toString()}`);
}

/**
 * Get deck analysis.
 */
export async function getDeckAnalysis(deckId: string): Promise<DeckAnalysisResult> {
  const params = new URLSearchParams({ deckId });
  return get<DeckAnalysisResult>(`/meta/deck-analysis?${params.toString()}`);
}

/**
 * Identify archetype from card list.
 */
export async function identifyArchetype(
  cardIds: number[],
  format: string
): Promise<{ archetype: string; confidence: number }> {
  return post<{ archetype: string; confidence: number }>('/meta/identify-archetype', {
    cardIds,
    format,
  });
}

/**
 * Get tier archetypes for a format.
 */
export async function getTierArchetypes(format: string, tier: number): Promise<gui.ArchetypeInfo[]> {
  const params = new URLSearchParams({ format, tier: tier.toString() });
  return get<gui.ArchetypeInfo[]>(`/meta/tier?${params.toString()}`);
}

/**
 * Get archetype cards.
 */
export async function getArchetypeCards(
  format: string,
  archetypeName: string
): Promise<insights.ArchetypeCards> {
  const params = new URLSearchParams({ format, archetype: archetypeName });
  return get<insights.ArchetypeCards>(`/meta/archetypes/cards?${params.toString()}`);
}

/**
 * Get format insights.
 */
export async function getFormatInsights(
  format: string,
  setCode: string
): Promise<insights.FormatInsights> {
  const params = new URLSearchParams({ format, setCode });
  return get<insights.FormatInsights>(`/meta/insights?${params.toString()}`);
}

/**
 * Refresh meta data from external sources.
 * Forces a fresh fetch from MTGGoldfish/MTGTop8.
 */
export async function refreshMetaData(format: string): Promise<gui.MetaDashboardResponse> {
  const params = new URLSearchParams({ format });
  return post<gui.MetaDashboardResponse>(`/meta/refresh?${params.toString()}`, {});
}
