/**
 * Decks API service.
 * Replaces Wails deck-related function bindings.
 */

import { get, post, put, del } from '../apiClient';
import { models, gui } from 'wailsjs/go/models';

// Re-export types for convenience
export type Deck = models.Deck;
export type DeckWithCards = gui.DeckWithCards;
export type DeckListItem = gui.DeckListItem;
export type DeckStatistics = gui.DeckStatistics;
export type DeckPerformance = models.DeckPerformance;
export type ExportDeckRequest = gui.ExportDeckRequest;
export type ExportDeckResponse = gui.ExportDeckResponse;
export type ImportDeckRequest = gui.ImportDeckRequest;
export type ImportDeckResponse = gui.ImportDeckResponse;
export type SuggestedDeckResponse = gui.SuggestedDeckResponse;
export type ArchetypeClassificationResult = gui.ArchetypeClassificationResult;

/**
 * Request to create a deck.
 */
export interface CreateDeckRequest {
  name: string;
  format: string;
  source: string;
  draft_event_id?: string;
}

/**
 * Request to update a deck.
 */
export interface UpdateDeckRequest {
  name?: string;
  format?: string;
}

/**
 * Request to import a deck.
 */
export interface ImportDeckApiRequest {
  content: string;
  name: string;
  format: string;
}

/**
 * Request to export a deck.
 */
export interface ExportDeckApiRequest {
  format: string;
}

/**
 * Request to suggest decks.
 */
export interface SuggestDecksRequest {
  session_id: string;
}

/**
 * Request to analyze a deck.
 */
export interface AnalyzeDeckRequest {
  deck_id: string;
}

/**
 * Get all decks with optional filtering.
 */
export async function getDecks(options?: {
  format?: string;
  source?: string;
}): Promise<DeckListItem[]> {
  const params = new URLSearchParams();
  if (options?.format) params.set('format', options.format);
  if (options?.source) params.set('source', options.source);

  const query = params.toString();
  return get<DeckListItem[]>(`/decks${query ? `?${query}` : ''}`);
}

/**
 * Get a single deck by ID with cards.
 */
export async function getDeck(deckId: string): Promise<DeckWithCards> {
  return get<DeckWithCards>(`/decks/${deckId}`);
}

/**
 * Create a new deck.
 */
export async function createDeck(request: CreateDeckRequest): Promise<Deck> {
  return post<Deck>('/decks', request);
}

/**
 * Update a deck.
 */
export async function updateDeck(deckId: string, request: UpdateDeckRequest): Promise<DeckWithCards> {
  return put<DeckWithCards>(`/decks/${deckId}`, request);
}

/**
 * Delete a deck.
 */
export async function deleteDeck(deckId: string): Promise<void> {
  return del<void>(`/decks/${deckId}`);
}

/**
 * Get deck statistics.
 */
export async function getDeckStats(deckId: string): Promise<DeckStatistics> {
  return get<DeckStatistics>(`/decks/${deckId}/stats`);
}

/**
 * Get deck performance/matches.
 */
export async function getDeckMatches(deckId: string): Promise<DeckPerformance> {
  return get<DeckPerformance>(`/decks/${deckId}/matches`);
}

/**
 * Get deck mana curve.
 */
export async function getDeckCurve(deckId: string): Promise<DeckStatistics> {
  return get<DeckStatistics>(`/decks/${deckId}/curve`);
}

/**
 * Get deck color distribution.
 */
export async function getDeckColors(deckId: string): Promise<DeckStatistics> {
  return get<DeckStatistics>(`/decks/${deckId}/colors`);
}

/**
 * Export a deck.
 */
export async function exportDeck(
  deckId: string,
  request: ExportDeckApiRequest
): Promise<ExportDeckResponse> {
  return post<ExportDeckResponse>(`/decks/${deckId}/export`, request);
}

/**
 * Import a deck from text.
 */
export async function importDeck(request: ImportDeckApiRequest): Promise<ImportDeckResponse> {
  return post<ImportDeckResponse>('/decks/import', request);
}

/**
 * Parse a deck list without saving.
 */
export async function parseDeckList(content: string): Promise<ImportDeckResponse> {
  return post<ImportDeckResponse>('/decks/parse', { content });
}

/**
 * Get deck suggestions for a draft.
 */
export async function suggestDecks(request: SuggestDecksRequest): Promise<SuggestedDeckResponse[]> {
  return post<SuggestedDeckResponse[]>('/decks/suggest', request);
}

/**
 * Analyze a deck (classify archetype).
 */
export async function analyzeDeck(
  request: AnalyzeDeckRequest
): Promise<ArchetypeClassificationResult> {
  return post<ArchetypeClassificationResult>('/decks/analyze', request);
}

/**
 * Get decks by source (draft, constructed, imported).
 */
export async function getDecksBySource(source: string): Promise<DeckListItem[]> {
  return getDecks({ source });
}

/**
 * Get decks by format.
 */
export async function getDecksByFormat(format: string): Promise<DeckListItem[]> {
  return getDecks({ format });
}
