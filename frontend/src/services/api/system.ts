/**
 * System API service.
 * Replaces Wails system-related function bindings.
 */

import { get, post } from '../apiClient';
import { gui } from 'wailsjs/go/models';

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
