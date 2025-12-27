/**
 * Matches API service.
 * Replaces Wails match-related function bindings.
 */

import { get, post } from '../apiClient';
import { models, storage } from '@/types/models';

// Re-export types for convenience
export type Match = models.Match;
export type StatsFilter = models.StatsFilter;
export type Statistics = models.Statistics;
export type PerformanceMetrics = models.PerformanceMetrics;

/**
 * Filter request for API calls.
 */
export interface StatsFilterRequest {
  account_id?: number;
  start_date?: string;
  end_date?: string;
  format?: string;
  formats?: string[];
  deck_format?: string;
  deck_id?: string;
  event_name?: string;
  event_names?: string[];
  opponent_name?: string;
  opponent_id?: string;
  result?: string;
  rank_class?: string;
  rank_min_class?: string;
  rank_max_class?: string;
  result_reason?: string;
}

/**
 * Trend analysis request.
 */
export interface TrendAnalysisRequest {
  start_date: string;
  end_date: string;
  period_type: string;
  formats?: string[];
}

/**
 * Get matches with optional filters.
 */
export async function getMatches(filter: StatsFilterRequest = {}): Promise<Match[]> {
  return post<Match[]>('/matches', filter);
}

/**
 * Get a single match by ID.
 */
export async function getMatch(matchId: string): Promise<Match> {
  return get<Match>(`/matches/${matchId}`);
}

/**
 * Get games for a specific match.
 */
export async function getMatchGames(matchId: string): Promise<models.Game[]> {
  return get<models.Game[]>(`/matches/${matchId}/games`);
}

/**
 * Get statistics with optional filters.
 */
export async function getStats(filter: StatsFilterRequest = {}): Promise<Statistics> {
  return post<Statistics>('/matches/stats', filter);
}

/**
 * Get trend analysis over time.
 */
export async function getTrendAnalysis(request: TrendAnalysisRequest): Promise<unknown> {
  return post('/matches/trends', request);
}

/**
 * Get all available match formats.
 */
export async function getFormats(): Promise<string[]> {
  return get<string[]>('/matches/formats');
}

/**
 * Get all available archetypes.
 */
export async function getArchetypes(): Promise<string[]> {
  return get<string[]>('/matches/archetypes');
}

/**
 * Get match distribution by format.
 */
export async function getFormatDistribution(
  filter: StatsFilterRequest = {}
): Promise<Record<string, Statistics>> {
  return post<Record<string, Statistics>>('/matches/format-distribution', filter);
}

/**
 * Get win rate trends over time.
 */
export async function getWinRateOverTime(request: TrendAnalysisRequest): Promise<unknown> {
  return post('/matches/win-rate-over-time', request);
}

/**
 * Get performance metrics by hour.
 */
export async function getPerformanceByHour(
  filter: StatsFilterRequest = {}
): Promise<PerformanceMetrics> {
  return post<PerformanceMetrics>('/matches/performance-by-hour', filter);
}

/**
 * Get matchup matrix (win rates against different decks).
 */
export async function getMatchupMatrix(
  filter: StatsFilterRequest = {}
): Promise<Record<string, Statistics>> {
  return post<Record<string, Statistics>>('/matches/matchup-matrix', filter);
}

/**
 * Get performance metrics with optional filters.
 */
export async function getPerformanceMetrics(
  filter: StatsFilterRequest = {}
): Promise<PerformanceMetrics> {
  return post<PerformanceMetrics>('/matches/performance', filter);
}

/**
 * Get rank progression for a format.
 */
export async function getRankProgression(format: string): Promise<models.RankProgression> {
  return get<models.RankProgression>(`/matches/rank-progression/${encodeURIComponent(format)}`);
}

/**
 * Get rank progression timeline for a format.
 */
export async function getRankProgressionTimeline(
  format: string,
  startDate: Date,
  endDate: Date,
  period: string
): Promise<storage.RankTimeline> {
  const params = new URLSearchParams({
    format,
    start_date: startDate.toISOString(),
    end_date: endDate.toISOString(),
    period,
  });
  return get<storage.RankTimeline>(`/matches/rank-progression-timeline?${params.toString()}`);
}

/**
 * Export matches in specified format.
 */
export async function exportMatches(format: 'json' | 'csv'): Promise<unknown> {
  return get(`/matches/export?format=${format}`);
}

/**
 * Helper to convert a time value to a date string (YYYY-MM-DD).
 * Handles both Date objects and time.Time (which serializes to ISO string).
 */
function formatDateParam(date: unknown): string | undefined {
  if (!date) return undefined;
  if (typeof date === 'string') {
    return date.split('T')[0];
  }
  if (date instanceof Date) {
    return date.toISOString().split('T')[0];
  }
  // Handle time.Time which may have been serialized
  const dateObj = date as { toString?: () => string };
  if (dateObj.toString) {
    const str = dateObj.toString();
    if (str.includes('T')) {
      return str.split('T')[0];
    }
  }
  return undefined;
}

/**
 * Helper to convert StatsFilter model to API request format.
 */
export function statsFilterToRequest(filter: StatsFilter): StatsFilterRequest {
  return {
    account_id: filter.AccountID,
    start_date: formatDateParam(filter.StartDate),
    end_date: formatDateParam(filter.EndDate),
    format: filter.Format,
    formats: filter.Formats,
    deck_format: filter.DeckFormat,
    deck_id: filter.DeckID,
    event_name: filter.EventName,
    event_names: filter.EventNames,
    opponent_name: filter.OpponentName,
    opponent_id: filter.OpponentID,
    result: filter.Result,
    rank_class: filter.RankClass,
    rank_min_class: filter.RankMinClass,
    rank_max_class: filter.RankMaxClass,
    result_reason: filter.ResultReason,
  };
}
