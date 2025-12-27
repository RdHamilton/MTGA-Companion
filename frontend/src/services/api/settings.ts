/**
 * Settings API service.
 * Replaces Wails settings-related function bindings.
 */

import { get, put } from '../apiClient';
import { gui } from '@/types/models';

// Re-export types for convenience
export type AppSettings = gui.AppSettings;

/**
 * Get all settings.
 */
export async function getSettings(): Promise<AppSettings> {
  return get<AppSettings>('/settings');
}

/**
 * Update all settings.
 */
export async function updateSettings(settings: AppSettings): Promise<void> {
  await put('/settings', settings);
}

/**
 * Get a single setting by key.
 */
export async function getSetting(key: string): Promise<unknown> {
  return get<unknown>(`/settings/${encodeURIComponent(key)}`);
}

/**
 * Update a single setting.
 */
export async function updateSetting(key: string, value: unknown): Promise<void> {
  await put(`/settings/${encodeURIComponent(key)}`, { value });
}
