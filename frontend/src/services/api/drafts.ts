/**
 * Drafts API service.
 * Replaces Wails draft-related function bindings.
 */

import { get, post } from '../apiClient';
import { models, gui, grading, metrics, insights } from '@/types/models';

// Re-export types for convenience
export type DraftSession = models.DraftSession;
export type DraftPickSession = models.DraftPickSession;
export type DraftGrade = grading.DraftGrade;
export type DraftStats = metrics.DraftStats;
export type FormatInsights = insights.FormatInsights;
export type CardRatingWithTier = gui.CardRatingWithTier;

/**
 * Filter for draft sessions.
 */
export interface DraftFilterRequest {
  format?: string;
  set_code?: string;
  start_date?: string;
  end_date?: string;
  status?: string;
}

/**
 * Request for grading a pick.
 */
export interface GradePickRequest {
  session_id: string;
  pick_number: number;
  picked_card_id: number;
  available_card_ids: number[];
}

/**
 * Request for draft insights.
 */
export interface DraftInsightsRequest {
  format: string;
  set_code?: string;
}

/**
 * Request for win probability prediction.
 */
export interface WinProbabilityRequest {
  session_id: string;
}

/**
 * Get draft sessions with optional filters.
 */
export async function getDraftSessions(
  filter: DraftFilterRequest = {}
): Promise<DraftSession[]> {
  return post<DraftSession[]>('/drafts', filter);
}

/**
 * Get a single draft session by ID.
 */
export async function getDraftSession(sessionId: string): Promise<DraftSession> {
  return get<DraftSession>(`/drafts/${sessionId}`);
}

/**
 * Get picks for a draft session.
 */
export async function getDraftPicks(sessionId: string): Promise<DraftPickSession[]> {
  return get<DraftPickSession[]>(`/drafts/${sessionId}/picks`);
}

/**
 * Get the card pool for a draft session.
 */
export async function getDraftPool(sessionId: string): Promise<models.SetCard[]> {
  return get<models.SetCard[]>(`/drafts/${sessionId}/pool`);
}

/**
 * Get analysis for a draft session.
 */
export async function getDraftAnalysis(sessionId: string): Promise<unknown> {
  return get(`/drafts/${sessionId}/analysis`);
}

/**
 * Get mana curve for a draft session.
 */
export async function getDraftCurve(sessionId: string): Promise<Record<number, number>> {
  return get<Record<number, number>>(`/drafts/${sessionId}/curve`);
}

/**
 * Get color distribution for a draft session.
 */
export async function getDraftColors(sessionId: string): Promise<Record<string, number>> {
  return get<Record<string, number>>(`/drafts/${sessionId}/colors`);
}

/**
 * Get draft statistics.
 */
export async function getDraftStats(filter: DraftFilterRequest = {}): Promise<DraftStats> {
  return post<DraftStats>('/drafts/stats', filter);
}

/**
 * Get available draft formats.
 */
export async function getDraftFormats(): Promise<string[]> {
  return get<string[]>('/drafts/formats');
}

/**
 * Get recent drafts.
 */
export async function getRecentDrafts(limit?: number): Promise<DraftSession[]> {
  const params = limit ? `?limit=${limit}` : '';
  return get<DraftSession[]>(`/drafts/recent${params}`);
}

/**
 * Grade a draft pick.
 */
export async function gradePick(request: GradePickRequest): Promise<DraftGrade> {
  return post<DraftGrade>('/drafts/grade-pick', request);
}

/**
 * Get draft insights for a format.
 */
export async function getDraftInsights(request: DraftInsightsRequest): Promise<FormatInsights> {
  return post<FormatInsights>('/drafts/insights', request);
}

/**
 * Predict win probability for a draft.
 */
export async function predictWinProbability(
  request: WinProbabilityRequest
): Promise<{ probability: number }> {
  return post<{ probability: number }>('/drafts/win-probability', request);
}

/**
 * Get active draft sessions (in progress).
 */
export async function getActiveDraftSessions(): Promise<DraftSession[]> {
  return getDraftSessions({ status: 'active' });
}

/**
 * Get completed draft sessions.
 */
export async function getCompletedDraftSessions(): Promise<DraftSession[]> {
  return getDraftSessions({ status: 'completed' });
}
