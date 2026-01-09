/**
 * Notes and Suggestions API service.
 * Handles deck notes, match notes, and improvement suggestions.
 */

import { get, post, put, del } from '../apiClient';

/**
 * Note category types.
 */
export type NoteCategory = 'general' | 'matchup' | 'sideboard' | 'mulligan';

/**
 * Suggestion type categories.
 */
export type SuggestionType = 'curve' | 'removal' | 'mana' | 'sequencing' | 'sideboard';

/**
 * Suggestion priority levels.
 */
export type SuggestionPriority = 'low' | 'medium' | 'high';

/**
 * A deck note with timestamps and category.
 */
export interface DeckNote {
  id: number;
  deckId: string;
  content: string;
  category: NoteCategory;
  createdAt: string;
  updatedAt: string;
}

/**
 * Match notes with optional rating.
 */
export interface MatchNotes {
  matchId: string;
  notes: string;
  rating: number; // 0 = no rating, 1-5 = star rating
}

/**
 * An improvement suggestion based on play pattern analysis.
 */
export interface ImprovementSuggestion {
  id: number;
  deckId: string;
  suggestionType: SuggestionType;
  priority: SuggestionPriority;
  title: string;
  description: string;
  evidence?: string; // JSON string with supporting data
  cardReferences?: string; // JSON string with referenced cards
  isDismissed: boolean;
  createdAt: string;
}

/**
 * Request to create a deck note.
 */
export interface CreateDeckNoteRequest {
  content: string;
  category?: NoteCategory;
}

/**
 * Request to update a deck note.
 */
export interface UpdateDeckNoteRequest {
  content: string;
  category?: NoteCategory;
}

/**
 * Request to update match notes.
 */
export interface UpdateMatchNotesRequest {
  notes: string;
  rating: number;
}

// -------------------- Deck Notes API --------------------

/**
 * Get all notes for a deck.
 */
export async function getDeckNotes(
  deckId: string,
  category?: NoteCategory
): Promise<DeckNote[]> {
  const params = category ? `?category=${category}` : '';
  return get<DeckNote[]>(`/decks/${deckId}/notes${params}`);
}

/**
 * Get a single deck note by ID.
 */
export async function getDeckNote(deckId: string, noteId: number): Promise<DeckNote> {
  return get<DeckNote>(`/decks/${deckId}/notes/${noteId}`);
}

/**
 * Create a new deck note.
 */
export async function createDeckNote(
  deckId: string,
  request: CreateDeckNoteRequest
): Promise<DeckNote> {
  return post<DeckNote>(`/decks/${deckId}/notes`, request);
}

/**
 * Update an existing deck note.
 */
export async function updateDeckNote(
  deckId: string,
  noteId: number,
  request: UpdateDeckNoteRequest
): Promise<DeckNote> {
  return put<DeckNote>(`/decks/${deckId}/notes/${noteId}`, request);
}

/**
 * Delete a deck note.
 */
export async function deleteDeckNote(deckId: string, noteId: number): Promise<void> {
  return del(`/decks/${deckId}/notes/${noteId}`);
}

// -------------------- Match Notes API --------------------

/**
 * Get notes and rating for a match.
 */
export async function getMatchNotes(matchId: string): Promise<MatchNotes> {
  return get<MatchNotes>(`/matches/${matchId}/notes`);
}

/**
 * Update notes and rating for a match.
 */
export async function updateMatchNotes(
  matchId: string,
  request: UpdateMatchNotesRequest
): Promise<MatchNotes> {
  return put<MatchNotes>(`/matches/${matchId}/notes`, request);
}

// -------------------- Suggestions API --------------------

/**
 * Get improvement suggestions for a deck.
 * @param deckId - The deck ID
 * @param activeOnly - If true, only return non-dismissed suggestions (default: true)
 */
export async function getDeckSuggestions(
  deckId: string,
  activeOnly: boolean = true
): Promise<ImprovementSuggestion[]> {
  const params = activeOnly ? '' : '?active=false';
  return get<ImprovementSuggestion[]>(`/decks/${deckId}/suggestions${params}`);
}

/**
 * Generate new improvement suggestions for a deck.
 * This analyzes play patterns and creates actionable suggestions.
 * @param deckId - The deck ID
 * @param minGames - Minimum games required for analysis (default: 5)
 */
export async function generateSuggestions(
  deckId: string,
  minGames: number = 5
): Promise<ImprovementSuggestion[]> {
  const params = minGames !== 5 ? `?min_games=${minGames}` : '';
  return post<ImprovementSuggestion[]>(`/decks/${deckId}/suggestions/generate${params}`, {});
}

/**
 * Dismiss an improvement suggestion.
 */
export async function dismissSuggestion(suggestionId: number): Promise<void> {
  return put(`/suggestions/${suggestionId}/dismiss`, {});
}

// -------------------- Utility Types --------------------

/**
 * Parsed evidence data for curve suggestions.
 */
export interface CurveEvidence {
  avgFirstPlay: number;
  totalGames: number;
}

/**
 * Parsed evidence data for mana suggestions.
 */
export interface ManaEvidence {
  manaScrew?: number;
  manaFlood?: number;
  totalGames: number;
  screwRate?: number;
  floodRate?: number;
  avgLandDrops: number;
}

/**
 * Parse the evidence JSON string from a suggestion.
 */
export function parseEvidence<T>(evidence: string | undefined): T | null {
  if (!evidence) return null;
  try {
    return JSON.parse(evidence) as T;
  } catch {
    return null;
  }
}

/**
 * Get a display label for suggestion types.
 */
export function getSuggestionTypeLabel(type: SuggestionType): string {
  const labels: Record<SuggestionType, string> = {
    curve: 'Mana Curve',
    removal: 'Removal',
    mana: 'Mana Base',
    sequencing: 'Play Sequencing',
    sideboard: 'Sideboard',
  };
  return labels[type] || type;
}

/**
 * Get a display label for suggestion priorities.
 */
export function getPriorityLabel(priority: SuggestionPriority): string {
  const labels: Record<SuggestionPriority, string> = {
    low: 'Low Priority',
    medium: 'Medium Priority',
    high: 'High Priority',
  };
  return labels[priority] || priority;
}

/**
 * Get a color class for suggestion priorities.
 */
export function getPriorityColor(priority: SuggestionPriority): string {
  const colors: Record<SuggestionPriority, string> = {
    low: 'text-blue-400',
    medium: 'text-yellow-400',
    high: 'text-red-400',
  };
  return colors[priority] || 'text-gray-400';
}
