/**
 * Drafts API service.
 * Replaces Wails draft-related function bindings.
 */

import { get, post } from '../apiClient';
import { models, gui, grading, metrics, insights, pickquality, prediction } from '@/types/models';

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

/**
 * Get deck metrics for a draft session.
 */
export async function getDraftDeckMetrics(sessionId: string): Promise<models.DeckMetrics> {
  return get<models.DeckMetrics>(`/drafts/${sessionId}/deck-metrics`);
}

/**
 * Get draft performance metrics.
 */
export async function getDraftPerformanceMetrics(): Promise<DraftStats> {
  return post<DraftStats>('/drafts/stats', {});
}

/**
 * Analyze pick quality for a session.
 */
export async function analyzeSessionPickQuality(sessionId: string): Promise<void> {
  return post(`/drafts/${sessionId}/analyze-picks`);
}

/**
 * Get pick alternatives for a specific pick.
 */
export async function getPickAlternatives(
  sessionId: string,
  packNumber: number,
  pickNumber: number
): Promise<pickquality.PickQuality> {
  return post<pickquality.PickQuality>('/drafts/grade-pick', {
    session_id: sessionId,
    pack_number: packNumber,
    pick_number: pickNumber,
  });
}

/**
 * Get draft grade for a session.
 */
export async function getDraftGrade(sessionId: string): Promise<DraftGrade> {
  return get<DraftGrade>(`/drafts/${sessionId}/analysis`);
}

/**
 * Calculate draft grade for a session.
 */
export async function calculateDraftGrade(sessionId: string): Promise<DraftGrade> {
  return post<DraftGrade>(`/drafts/${sessionId}/calculate-grade`);
}

/**
 * Get current pack with recommendation.
 */
export async function getCurrentPackWithRecommendation(
  sessionId: string
): Promise<gui.CurrentPackResponse> {
  return get<gui.CurrentPackResponse>(`/drafts/${sessionId}/current-pack`);
}

/**
 * Get win rate prediction for a draft.
 */
export async function getDraftWinRatePrediction(
  sessionId: string
): Promise<prediction.DeckPrediction> {
  return post<prediction.DeckPrediction>(`/drafts/${sessionId}/calculate-prediction`);
}

/**
 * Get recommendations for a draft.
 */
export async function getRecommendations(
  request: gui.GetRecommendationsRequest
): Promise<gui.GetRecommendationsResponse> {
  return post<gui.GetRecommendationsResponse>('/decks/recommendations', request);
}

/**
 * Record a recommendation.
 */
export async function recordRecommendation(
  request: gui.RecordRecommendationRequest
): Promise<gui.RecordRecommendationResponse> {
  return post<gui.RecordRecommendationResponse>('/feedback/recommendation', request);
}

/**
 * Record a recommendation action.
 */
export async function recordRecommendationAction(request: gui.RecordActionRequest): Promise<void> {
  return post('/feedback/action', request);
}

/**
 * Record a recommendation outcome.
 */
export async function recordRecommendationOutcome(
  request: gui.RecordOutcomeRequest
): Promise<void> {
  return post('/feedback/outcome', request);
}

/**
 * Get recommendation stats.
 */
export async function getRecommendationStats(): Promise<gui.RecommendationStatsResponse> {
  return get<gui.RecommendationStatsResponse>('/feedback/stats');
}

/**
 * Explain a recommendation.
 */
export async function explainRecommendation(
  request: gui.ExplainRecommendationRequest
): Promise<gui.ExplainRecommendationResponse> {
  return post<gui.ExplainRecommendationResponse>('/decks/explain-recommendation', request);
}

/**
 * Classify draft pool archetype.
 */
export async function classifyDraftPoolArchetype(
  sessionId: string
): Promise<gui.ArchetypeClassificationResult> {
  return post<gui.ArchetypeClassificationResult>('/decks/classify-draft-pool', {
    session_id: sessionId,
  });
}

/**
 * Recalculate draft grades for a specific set.
 * This is useful after refreshing 17Lands ratings to update grades with new data.
 *
 * Note: This is currently a stub - backend implementation pending (#734).
 */
export async function recalculateSetGrades(
  _setCode: string
): Promise<{ count: number }> {
  console.warn('[drafts.recalculateSetGrades] Not yet implemented in REST API');
  return { count: 0 };
}
