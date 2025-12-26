/**
 * Cards API service.
 * Replaces Wails card-related function bindings.
 */

import { get, post } from '../apiClient';
import { models, gui } from 'wailsjs/go/models';

// Re-export types for convenience
export type SetCard = models.SetCard;
export type CardRatingWithTier = gui.CardRatingWithTier;
export type SetInfo = gui.SetInfo;

/**
 * Search request for cards.
 */
export interface CardSearchRequest {
  query: string;
  set_code?: string;
  colors?: string[];
  types?: string[];
  rarity?: string;
  limit?: number;
}

/**
 * Search for cards.
 */
export async function searchCards(request: CardSearchRequest): Promise<SetCard[]> {
  return post<SetCard[]>('/cards/search', request);
}

/**
 * Get a card by Arena ID.
 */
export async function getCardByArenaId(arenaId: number): Promise<SetCard> {
  return get<SetCard>(`/cards/arena/${arenaId}`);
}

/**
 * Get all set information.
 */
export async function getAllSetInfo(): Promise<SetInfo[]> {
  return get<SetInfo[]>('/cards/sets');
}

/**
 * Get cards for a specific set.
 */
export async function getSetCards(setCode: string): Promise<SetCard[]> {
  return get<SetCard[]>(`/cards/sets/${setCode}/cards`);
}

/**
 * Get card ratings for a set and format.
 */
export async function getCardRatings(
  setCode: string,
  format: string
): Promise<CardRatingWithTier[]> {
  return get<CardRatingWithTier[]>(`/cards/ratings/${setCode}/${format}`);
}

/**
 * Get collection quantities for cards.
 */
export async function getCollectionQuantities(
  arenaIds: number[]
): Promise<Record<number, number>> {
  return post<Record<number, number>>('/cards/collection-quantities', { arena_ids: arenaIds });
}
