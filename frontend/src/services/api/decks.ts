/**
 * Decks API service.
 * Replaces Wails deck-related function bindings.
 */

import { get, post, put, del } from '../apiClient';
import { models, gui } from '@/types/models';

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

/**
 * Get decks by tags.
 */
export async function getDecksByTags(tags: string[]): Promise<DeckListItem[]> {
  return post<DeckListItem[]>('/decks/by-tags', { tags });
}

/**
 * Get deck library with filters.
 */
export async function getDeckLibrary(filter: gui.DeckLibraryFilter): Promise<DeckListItem[]> {
  return post<DeckListItem[]>('/decks/library', filter);
}

/**
 * Clone a deck.
 */
export async function cloneDeck(deckId: string, newName: string): Promise<Deck> {
  return post<Deck>(`/decks/${deckId}/clone`, { name: newName });
}

/**
 * Get deck by draft event ID.
 */
export async function getDeckByDraftEvent(draftEventId: string): Promise<DeckWithCards> {
  return get<DeckWithCards>(`/decks/by-draft/${draftEventId}`);
}

/**
 * Get deck statistics.
 */
export async function getDeckStatistics(deckId: string): Promise<DeckStatistics> {
  return get<DeckStatistics>(`/decks/${deckId}/statistics`);
}

/**
 * Get deck performance.
 */
export async function getDeckPerformance(deckId: string): Promise<DeckPerformance> {
  return get<DeckPerformance>(`/decks/${deckId}/performance`);
}

/**
 * Validate a draft deck.
 */
export async function validateDraftDeck(deckId: string): Promise<boolean> {
  const result = await get<{ valid: boolean }>(`/decks/${deckId}/validate-draft`);
  return result.valid;
}

/**
 * Apply a suggested deck to an existing deck.
 */
export async function applySuggestedDeck(deckId: string, suggestion: SuggestedDeckResponse): Promise<void> {
  await post(`/decks/apply-suggestion`, { deck_id: deckId, suggestion });
}

/**
 * Get suggested deck export content.
 */
export async function getSuggestedDeckExportContent(suggestion: SuggestedDeckResponse, deckName: string): Promise<string> {
  const result = await post<{ content: string }>('/decks/suggested/export-content', {
    suggestion,
    deck_name: deckName,
  });
  return result.content;
}

/**
 * Classify deck archetype.
 */
export async function classifyDeckArchetype(deckId: string): Promise<ArchetypeClassificationResult> {
  return get<ArchetypeClassificationResult>(`/decks/${deckId}/classify`);
}

/**
 * Add a tag to a deck.
 */
export async function addTag(deckId: string, tag: string): Promise<void> {
  await post(`/decks/${deckId}/tags`, { tag });
}

/**
 * Remove a tag from a deck.
 */
export async function removeTag(deckId: string, tag: string): Promise<void> {
  await del(`/decks/${deckId}/tags/${encodeURIComponent(tag)}`);
}

/**
 * Add a card to a deck.
 */
export async function addCard(request: {
  deck_id: string;
  arena_id: number;
  quantity: number;
  zone: string;
  is_sideboard: boolean;
  from_draft?: boolean;
}): Promise<void> {
  // Map frontend field names to backend expected field names
  const body = {
    card_id: request.arena_id,
    quantity: request.quantity,
    board: request.zone,
    from_draft: request.from_draft ?? false,
  };
  await post(`/decks/${request.deck_id}/cards`, body);
}

/**
 * Remove one copy of a card from a deck (decrements quantity by 1).
 */
export async function removeCard(request: {
  deck_id: string;
  arena_id: number;
  zone: string;
}): Promise<void> {
  await del(`/decks/${request.deck_id}/cards/${request.arena_id}?zone=${request.zone}`);
}

/**
 * Remove all copies of a card from a deck.
 */
export async function removeAllCopies(request: {
  deck_id: string;
  arena_id: number;
  zone: string;
}): Promise<void> {
  await del(`/decks/${request.deck_id}/cards/${request.arena_id}/all?zone=${request.zone}`);
}

/**
 * Request to build a deck around a seed card.
 */
export interface BuildAroundSeedRequest {
  seed_card_id: number;
  max_results?: number;
  budget_mode?: boolean;
  set_restriction?: 'single' | 'multiple' | 'all';
  allowed_sets?: string[];
}

/**
 * Score breakdown for detailed reasoning about card suggestions.
 */
export interface ScoreBreakdown {
  colorFit: number;  // 0.0-1.0, weight: 25%
  curveFit: number;  // 0.0-1.0, weight: 20%
  synergy: number;   // 0.0-1.0, weight: 30%
  quality: number;   // 0.0-1.0, weight: 15%
  overall: number;   // Final weighted score
}

/**
 * Synergy detail describing a specific synergy between a card and the deck.
 */
export interface SynergyDetail {
  type: 'keyword' | 'theme' | 'creature_type';
  name: string;        // e.g., "flying", "tokens", "Elf"
  description: string; // e.g., "Matches 3 other flying creatures"
}

/**
 * Card suggestion with ownership information.
 */
export interface CardWithOwnership {
  cardID: number;
  name: string;
  manaCost?: string;
  cmc: number;
  colors: string[];
  typeLine: string;
  rarity?: string;
  imageURI?: string;
  score: number;
  reasoning: string;
  inCollection: boolean;
  ownedCount: number;
  neededCount: number;
  currentCopies: number;     // Copies currently in deck
  recommendedCopies: number; // Recommended total copies (1-4)
  scoreBreakdown?: ScoreBreakdown;
  synergyDetails?: SynergyDetail[];
}

/**
 * Suggested land for a deck.
 */
export interface SuggestedLandResponse {
  cardID: number;
  name: string;
  quantity: number;
  color: string;
}

/**
 * Analysis of seed deck suggestions.
 */
export interface SeedDeckAnalysis {
  colorIdentity: string[];
  keywords: string[];
  themes: string[];
  idealCurve: Record<number, number>;
  suggestedLandCount: number;
  totalCards: number;
  inCollectionCount: number;
  missingCount: number;
  missingWildcardCost: Record<string, number>;
}

/**
 * Response from build around seed endpoint.
 */
export interface BuildAroundSeedResponse {
  seedCard: CardWithOwnership;
  suggestions: CardWithOwnership[];
  lands: SuggestedLandResponse[];
  analysis: SeedDeckAnalysis;
}

/**
 * Build a deck around a seed card.
 */
export async function buildAroundSeed(
  request: BuildAroundSeedRequest
): Promise<BuildAroundSeedResponse> {
  return post<BuildAroundSeedResponse>('/decks/build-around', request);
}

/**
 * Request for iterative deck building suggestions.
 * The API analyzes ALL deck cards collectively to find commonalities
 * (colors, themes, keywords) and suggests cards that complement the deck.
 */
export interface IterativeBuildAroundRequest {
  seed_card_id?: number;      // Optional - API analyzes all deck cards collectively
  deck_card_ids: number[];    // Required - all cards currently in the deck
  max_results?: number;
  budget_mode?: boolean;
  set_restriction?: 'single' | 'multiple' | 'all';
  allowed_sets?: string[];
}

/**
 * Live analysis of the deck being built.
 */
export interface LiveDeckAnalysis {
  colorIdentity: string[];
  keywords: string[];
  themes: string[];
  currentCurve: Record<number, number>;
  recommendedLandCount: number;
  totalCards: number;
  inCollectionCount: number;
}

/**
 * Response from iterative build-around endpoint.
 */
export interface IterativeBuildAroundResponse {
  suggestions: CardWithOwnership[];
  deckAnalysis: LiveDeckAnalysis;
  slotsRemaining: number;
  landSuggestions: SuggestedLandResponse[];
}

/**
 * Get next card suggestions for iterative deck building.
 * Called as user picks cards one-by-one.
 */
export async function suggestNextCards(
  request: IterativeBuildAroundRequest
): Promise<IterativeBuildAroundResponse> {
  return post<IterativeBuildAroundResponse>('/decks/build-around/suggest-next', request);
}

// ==========================================
// Complete Deck Generation (Issue #774)
// ==========================================

/**
 * Archetype profile for deck building.
 * Matches backend ArchetypeProfileResponse.
 */
export interface ArchetypeProfile {
  name: string;
  landCount: number;
  curveTargets: Record<number, number>;
  creatureRatio: number;
  removalCount: number;
  cardAdvantage: number;
  description: string;
}

/**
 * Request to generate a complete 60-card deck.
 */
export interface GenerateCompleteDeckRequest {
  seed_card_id: number;
  archetype: 'aggro' | 'midrange' | 'control';
  budget_mode?: boolean;
  set_restriction?: string;
  allowed_sets?: string[];
}

/**
 * Score breakdown for card scoring.
 */
export interface ScoreBreakdown {
  colorFit: number;
  curveFit: number;
  synergy: number;
  quality: number;
  overall: number;
}

/**
 * Synergy detail explaining card synergy.
 */
export interface SynergyDetail {
  type: string;
  name: string;
  description: string;
}

/**
 * Card with quantity for generated deck.
 */
export interface CardWithQuantity {
  cardID: number;
  name: string;
  manaCost?: string;
  cmc: number;
  colors: string[];
  typeLine: string;
  rarity?: string;
  imageURI?: string;
  quantity: number;
  score: number;
  reasoning: string;
  inCollection: boolean;
  ownedCount: number;
  neededCount: number;
  scoreBreakdown?: ScoreBreakdown;
  synergyDetails?: SynergyDetail[];
}

/**
 * Land with quantity for generated deck.
 */
export interface LandWithQuantity {
  cardID: number;
  name: string;
  quantity: number;
  colors: string[];
  isBasic: boolean;
  entersTapped: boolean;
}

/**
 * Strategy and game plan for a generated deck.
 */
export interface DeckStrategy {
  summary: string;
  gamePlan: string;
  keyCards: string[];
  mulligan: string;
  strengths: string[];
  weaknesses: string[];
}

/**
 * Analysis of a generated deck.
 * Matches backend GeneratedDeckAnalysisResponse.
 */
export interface GeneratedDeckAnalysis {
  totalCards: number;
  spellCount: number;
  landCount: number;
  creatureCount: number;
  nonCreatureCount: number;
  averageCMC: number;
  manaCurve: Record<number, number>;
  colorDistribution: Record<string, number>;
  inCollectionCount: number;
  missingCount: number;
  missingWildcardCost: Record<string, number>;
  archetypeMatch: number;
}

/**
 * Response from complete deck generation.
 */
export interface GenerateCompleteDeckResponse {
  seedCard: CardWithOwnership;
  spells: CardWithQuantity[];
  lands: LandWithQuantity[];
  strategy: DeckStrategy;
  analysis: GeneratedDeckAnalysis;
}

/**
 * Generate a complete 60-card deck from a seed card.
 */
export async function generateCompleteDeck(
  request: GenerateCompleteDeckRequest
): Promise<GenerateCompleteDeckResponse> {
  return post<GenerateCompleteDeckResponse>('/decks/generate', request);
}

/**
 * Get available archetype profiles.
 */
export async function getArchetypeProfiles(): Promise<ArchetypeProfile[]> {
  return get<ArchetypeProfile[]>('/decks/archetypes');
}
