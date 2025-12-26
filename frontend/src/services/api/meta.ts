/**
 * Meta API service.
 * Replaces Wails meta-related function bindings.
 */

import { get, post } from '../apiClient';

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
