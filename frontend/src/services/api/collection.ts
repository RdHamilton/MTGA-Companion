/**
 * Collection API service.
 * Replaces Wails collection-related function bindings.
 */

import { get, post } from '../apiClient';
import { gui, models } from '@/types/models';

// Re-export types for convenience
export type CollectionCard = gui.CollectionCard;
export type CollectionStats = gui.CollectionStats;
export type CollectionChangeEntry = gui.CollectionChangeEntry;

/**
 * Filter for collection queries.
 */
export interface CollectionFilter {
  set_code?: string;
  rarity?: string;
  colors?: string[];
  owned_only?: boolean;
  missing_only?: boolean;
}

/**
 * Get collection with optional filters.
 */
export async function getCollection(filter: CollectionFilter = {}): Promise<CollectionCard[]> {
  return post<CollectionCard[]>('/collection', filter);
}

/**
 * Get collection statistics.
 */
export async function getCollectionStats(): Promise<CollectionStats> {
  return get<CollectionStats>('/collection/stats');
}

/**
 * Get set completion progress.
 */
export async function getSetCompletion(): Promise<models.SetCompletion[]> {
  return get<models.SetCompletion[]>('/collection/sets/completion');
}

/**
 * Get recent collection changes.
 */
export async function getRecentChanges(limit?: number): Promise<CollectionChangeEntry[]> {
  const params = limit ? `?limit=${limit}` : '';
  return get<CollectionChangeEntry[]>(`/collection/recent${params}`);
}

/**
 * Get missing cards for a set.
 */
export async function getMissingCardsForSet(setCode: string): Promise<CollectionCard[]> {
  return get<CollectionCard[]>(`/collection/sets/${setCode}/missing`);
}

/**
 * Get collection for a specific set.
 */
export async function getCollectionBySet(setCode: string): Promise<CollectionCard[]> {
  return getCollection({ set_code: setCode });
}

/**
 * Get collection by rarity.
 */
export async function getCollectionByRarity(rarity: string): Promise<CollectionCard[]> {
  return getCollection({ rarity });
}

/**
 * Get missing cards analysis for a set.
 */
export async function getMissingCards(setCode: string): Promise<models.MissingCardsAnalysis> {
  return get<models.MissingCardsAnalysis>(`/collection/missing/${setCode}`);
}

/**
 * Get missing cards for a deck.
 */
export async function getMissingCardsForDeck(deckId: string): Promise<gui.MissingCardsForDeckResponse> {
  return get<gui.MissingCardsForDeckResponse>(`/collection/decks/${deckId}/missing`);
}
