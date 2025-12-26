/**
 * API Services Index
 *
 * This module exports all API service functions, providing a centralized
 * access point for REST API communication. These services replace the
 * direct Wails function bindings.
 *
 * Usage:
 *   import { matches, drafts, decks } from '@/services/api';
 *
 *   const stats = await matches.getStats({ format: 'Standard' });
 *   const sessions = await drafts.getActiveDraftSessions();
 *   const deckList = await decks.getDecks();
 */

// API modules
export * as matches from './matches';
export * as drafts from './drafts';
export * as decks from './decks';
export * as cards from './cards';
export * as collection from './collection';
export * as system from './system';

// Re-export commonly used types
export type {
  StatsFilterRequest,
  TrendAnalysisRequest,
} from './matches';

export type {
  DraftFilterRequest,
  GradePickRequest,
  DraftInsightsRequest,
  WinProbabilityRequest,
} from './drafts';

export type {
  CreateDeckRequest,
  UpdateDeckRequest,
  ImportDeckApiRequest,
  ExportDeckApiRequest,
  SuggestDecksRequest,
  AnalyzeDeckRequest,
} from './decks';

export type {
  CardSearchRequest,
} from './cards';

export type {
  CollectionFilter,
} from './collection';

export type {
  VersionInfo,
  DaemonStatus,
} from './system';
