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
 * Response from collection API.
 */
export interface CollectionResponse {
  cards: CollectionCard[];
  totalCount: number;
  filterCount: number;
  unknownCardsRemaining: number;   // Cards without metadata that need Scryfall lookup
  unknownCardsFetched: number;     // Cards fetched from Scryfall in this request
}

/**
 * Get collection with optional filters.
 * Returns full response including metadata counts.
 */
export async function getCollectionWithMetadata(filter: CollectionFilter = {}): Promise<CollectionResponse> {
  const response = await post<CollectionResponse>('/collection', filter);
  // Handle null/undefined response or missing fields
  return {
    cards: response?.cards ?? [],
    totalCount: response?.totalCount ?? 0,
    filterCount: response?.filterCount ?? 0,
    unknownCardsRemaining: response?.unknownCardsRemaining ?? 0,
    unknownCardsFetched: response?.unknownCardsFetched ?? 0,
  };
}

/**
 * Get collection with optional filters.
 * Returns just the cards array for backward compatibility.
 */
export async function getCollection(filter: CollectionFilter = {}): Promise<CollectionCard[]> {
  const response = await getCollectionWithMetadata(filter);
  return response.cards;
}

/**
 * Get collection statistics.
 */
export async function getCollectionStats(): Promise<CollectionStats> {
  return get<CollectionStats>('/collection/stats');
}

/**
 * Get set completion progress.
 * Returns completion statistics for all sets.
 */
export async function getSetCompletion(): Promise<models.SetCompletion[]> {
  return get<models.SetCompletion[]>('/collection/sets');
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
  return get<CollectionCard[]>(`/collection/missing/${setCode}`);
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
