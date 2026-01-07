/**
 * Cards API service.
 * Replaces Wails card-related function bindings.
 */

import { get, post } from '../apiClient';
import { models, gui, seventeenlands } from '@/types/models';

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
 * Uses GET with query parameters as the backend expects.
 */
export async function searchCards(request: CardSearchRequest): Promise<SetCard[]> {
  const params = new URLSearchParams();
  params.set('q', request.query);
  if (request.set_code) {
    params.set('set', request.set_code);
  }
  if (request.limit) {
    params.set('limit', request.limit.toString());
  }
  return get<SetCard[]>(`/cards?${params.toString()}`);
}

/**
 * Get a card by Arena ID.
 */
export async function getCardByArenaId(arenaId: number): Promise<SetCard> {
  return get<SetCard>(`/cards/${arenaId}`);
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

/**
 * Get color ratings for a set and format.
 */
export async function getColorRatings(
  setCode: string,
  format: string
): Promise<seventeenlands.ColorRating[]> {
  return get<seventeenlands.ColorRating[]>(`/cards/color-ratings/${setCode}/${format}`);
}

/**
 * Card with collection quantity info.
 */
export interface CardWithCollection extends SetCard {
  quantity: number;
}

/**
 * Search cards with collection info.
 */
export async function searchCardsWithCollection(
  query: string,
  sets?: string[],
  limit?: number
): Promise<CardWithCollection[]> {
  return post<CardWithCollection[]>('/cards/search-with-collection', {
    query,
    set_codes: sets,
    limit,
  });
}

/**
 * Ratings staleness information.
 */
export interface RatingsStaleness {
  cachedAt: string;
  isStale: boolean;
  cardCount: number;
}

/**
 * Check if ratings for a set are stale.
 */
export async function getRatingsStaleness(
  setCode: string,
  format: string
): Promise<RatingsStaleness> {
  return get<RatingsStaleness>(`/cards/ratings/${setCode}/${format}/staleness`);
}

/**
 * Refresh ratings for a set (force re-download from 17Lands).
 */
export async function refreshSetRatings(
  setCode: string,
  format: string = 'PremierDraft'
): Promise<void> {
  await post(`/cards/ratings/${setCode}/refresh`, { format });
}
