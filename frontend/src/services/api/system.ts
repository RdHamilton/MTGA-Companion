/**
 * System API service.
 * Replaces Wails system-related function bindings.
 */

import { get, post } from '../apiClient';
import { gui, models } from '@/types/models';

// Re-export types for convenience
export type ConnectionStatus = gui.ConnectionStatus;

/**
 * Version information.
 */
export interface VersionInfo {
  version: string;
  service: string;
}

/**
 * Daemon status response.
 */
export interface DaemonStatus {
  status: string;
  connected: boolean;
}

/**
 * Get the current connection status.
 */
export async function getStatus(): Promise<ConnectionStatus> {
  return get<ConnectionStatus>('/system/status');
}

/**
 * Get the daemon connection status.
 */
export async function getDaemonStatus(): Promise<DaemonStatus> {
  return get<DaemonStatus>('/system/daemon/status');
}

/**
 * Connect to the daemon.
 */
export async function connectDaemon(): Promise<{ status: string }> {
  return post<{ status: string }>('/system/daemon/connect');
}

/**
 * Disconnect from the daemon.
 */
export async function disconnectDaemon(): Promise<{ status: string }> {
  return post<{ status: string }>('/system/daemon/disconnect');
}

/**
 * Get the application version.
 */
export async function getVersion(): Promise<VersionInfo> {
  return get<VersionInfo>('/system/version');
}

/**
 * Get the database path.
 */
export async function getDatabasePath(): Promise<{ path: string }> {
  return get<{ path: string }>('/system/database/path');
}

/**
 * Set the database path.
 */
export async function setDatabasePath(path: string): Promise<{ status: string }> {
  return post<{ status: string }>('/system/database/path', { path });
}

/**
 * Get current account.
 */
export async function getCurrentAccount(): Promise<models.Account> {
  return get<models.Account>('/system/account');
}

/**
 * Clear all data.
 */
export async function clearAllData(): Promise<void> {
  return post('/export/clear');
}

/**
 * Check Ollama status.
 */
export async function checkOllamaStatus(
  endpoint: string,
  model: string
): Promise<gui.OllamaStatus> {
  return post<gui.OllamaStatus>('/llm/status', { endpoint, model });
}

/**
 * Get available Ollama models.
 */
export async function getAvailableOllamaModels(endpoint: string): Promise<gui.OllamaModel[]> {
  const params = endpoint ? `?endpoint=${encodeURIComponent(endpoint)}` : '';
  return get<gui.OllamaModel[]>(`/llm/models${params}`);
}

/**
 * Pull an Ollama model.
 */
export async function pullOllamaModel(endpoint: string, model: string): Promise<void> {
  return post('/llm/models/pull', { endpoint, model });
}

/**
 * Test LLM generation.
 */
export async function testLLMGeneration(endpoint: string, model: string): Promise<string> {
  const result = await post<{ response: string }>('/llm/test', { endpoint, model });
  return result.response;
}

/**
 * Export ML training data.
 */
export async function exportMLTrainingData(limit: number): Promise<gui.MLTrainingDataExport> {
  return get<gui.MLTrainingDataExport>(`/feedback/ml-training?limit=${limit}`);
}

/**
 * Get feedback dashboard metrics.
 */
export async function getFeedbackDashboardMetrics(): Promise<gui.DashboardMetricsResponse> {
  return get<gui.DashboardMetricsResponse>('/feedback/dashboard');
}
