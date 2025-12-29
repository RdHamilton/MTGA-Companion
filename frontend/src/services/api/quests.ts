/**
 * Quests API service.
 * Replaces Wails quest-related function bindings.
 */

import { get } from '../apiClient';
import { models } from '@/types/models';

/**
 * Get active quests.
 */
export async function getActiveQuests(): Promise<models.Quest[]> {
  return get<models.Quest[]>('/quests/active');
}

/**
 * Get quest history.
 */
export async function getQuestHistory(
  startDate?: string,
  endDate?: string,
  limit?: number
): Promise<models.Quest[]> {
  const params = new URLSearchParams();
  if (startDate) params.append('startDate', startDate);
  if (endDate) params.append('endDate', endDate);
  if (limit) params.append('limit', limit.toString());

  const queryString = params.toString();
  const url = queryString ? `/quests/history?${queryString}` : '/quests/history';
  return get<models.Quest[]>(url);
}

/**
 * Get daily wins progress.
 */
export async function getDailyWins(): Promise<{ wins: number; goal: number }> {
  const response = await get<{ dailyWins: number; goal: number }>('/quests/wins/daily');
  return { wins: response.dailyWins, goal: response.goal };
}

/**
 * Get weekly wins progress.
 */
export async function getWeeklyWins(): Promise<{ wins: number; goal: number }> {
  const response = await get<{ weeklyWins: number; goal: number }>('/quests/wins/weekly');
  return { wins: response.weeklyWins, goal: response.goal };
}

/**
 * Get quest stats for a date range.
 */
export async function getQuestStats(
  startDate: string,
  endDate: string
): Promise<models.QuestStats> {
  const params = new URLSearchParams({ startDate, endDate });
  return get<models.QuestStats>(`/quests/stats?${params.toString()}`);
}
