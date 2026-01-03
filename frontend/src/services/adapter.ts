/**
 * REST API Adapter
 *
 * This adapter provides API access using REST endpoints and WebSocket for events.
 * It exposes a unified interface for accessing all backend services.
 *
 * Usage:
 *   import { apiAdapter } from '@/services/adapter';
 *   const matches = await apiAdapter.matches.getMatches(filter);
 */

import * as api from './api';
import {
  connect as wsConnect,
  disconnect as wsDisconnect,
  EventsOn as WsEventsOn,
  EventsOff as WsEventsOff,
} from './websocketClient';
import { configureApi, healthCheck } from './apiClient';
import { models, gui } from '@/types/models';

// Configuration state
let isInitialized = false;

/**
 * Check if REST API mode is enabled (always true now).
 */
export function isRestApiEnabled(): boolean {
  return true;
}

/**
 * Enable or disable REST API mode (no-op, always REST now).
 */
// eslint-disable-next-line @typescript-eslint/no-unused-vars
export function setUseRestApi(_enabled: boolean): void {
  // No-op - always using REST API
}

/**
 * Initialize the API services.
 * Call this once at app startup.
 */
export async function initializeServices(options?: {
  useRest?: boolean;
  apiBaseUrl?: string;
  wsUrl?: string;
}): Promise<void> {
  if (isInitialized) {
    return;
  }

  // Configure REST API
  if (options?.apiBaseUrl) {
    configureApi({ baseUrl: options.apiBaseUrl });
  }

  // Check if API is available
  const isHealthy = await healthCheck();
  if (!isHealthy) {
    console.error('[Adapter] REST API not available');
    throw new Error('REST API not available');
  }

  // Connect WebSocket
  try {
    await wsConnect();
    console.log('[Adapter] REST API mode enabled');
  } catch (error) {
    console.error('[Adapter] WebSocket connection failed:', error);
  }

  isInitialized = true;
}

/**
 * Cleanup services on app shutdown.
 */
export function cleanupServices(): void {
  wsDisconnect();
  isInitialized = false;
}

// ============================================================================
// Matches Adapter
// ============================================================================

export const matchesAdapter = {
  async getMatches(filter: models.StatsFilter): Promise<models.Match[]> {
    return api.matches.getMatches(api.matches.statsFilterToRequest(filter));
  },

  async getStats(filter: models.StatsFilter): Promise<models.Statistics> {
    return api.matches.getStats(api.matches.statsFilterToRequest(filter));
  },

  async getFormats(): Promise<string[]> {
    return api.matches.getFormats();
  },

  async getMatchGames(matchId: string): Promise<models.Game[]> {
    return api.matches.getMatchGames(matchId);
  },
};

// ============================================================================
// Drafts Adapter
// ============================================================================

export const draftsAdapter = {
  async getActiveDraftSessions(): Promise<models.DraftSession[]> {
    return api.drafts.getActiveDraftSessions();
  },

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  async getCompletedDraftSessions(_limit = 100): Promise<models.DraftSession[]> {
    return api.drafts.getCompletedDraftSessions();
  },

  async getDraftSession(sessionId: string): Promise<models.DraftSession> {
    return api.drafts.getDraftSession(sessionId);
  },

  async getDraftPicks(sessionId: string): Promise<models.DraftPickSession[]> {
    return api.drafts.getDraftPicks(sessionId);
  },

  async getCardRatings(setCode: string, format: string): Promise<gui.CardRatingWithTier[]> {
    return api.cards.getCardRatings(setCode, format);
  },
};

// ============================================================================
// Decks Adapter
// ============================================================================

export const decksAdapter = {
  async getDecks(): Promise<gui.DeckListItem[]> {
    return api.decks.getDecks();
  },

  async getDeck(deckId: string): Promise<gui.DeckWithCards> {
    return api.decks.getDeck(deckId);
  },

  async getDecksBySource(source: string): Promise<gui.DeckListItem[]> {
    return api.decks.getDecksBySource(source);
  },

  async getDecksByFormat(format: string): Promise<gui.DeckListItem[]> {
    return api.decks.getDecksByFormat(format);
  },

  async createDeck(
    name: string,
    format: string,
    source: string,
    draftEventId?: string
  ): Promise<models.Deck> {
    return api.decks.createDeck({ name, format, source, draft_event_id: draftEventId });
  },

  async deleteDeck(deckId: string): Promise<void> {
    return api.decks.deleteDeck(deckId);
  },

  async exportDeck(request: gui.ExportDeckRequest): Promise<gui.ExportDeckResponse> {
    return api.decks.exportDeck(request.deckID, { format: request.format });
  },

  async importDeck(request: gui.ImportDeckRequest): Promise<gui.ImportDeckResponse> {
    return api.decks.importDeck({
      content: request.importText,
      name: request.name,
      format: request.format,
    });
  },

  async suggestDecks(sessionId: string): Promise<gui.SuggestDecksResponse> {
    const suggestions = await api.decks.suggestDecks({ session_id: sessionId });
    return {
      suggestions: suggestions,
      totalCombos: suggestions.length,
      viableCombos: suggestions.length,
    } as gui.SuggestDecksResponse;
  },
};

// ============================================================================
// Collection Adapter
// ============================================================================

export const collectionAdapter = {
  async getCollection(filter?: gui.CollectionFilter): Promise<gui.CollectionResponse> {
    const apiFilter: api.CollectionFilter = filter
      ? {
          set_code: filter.setCode,
          rarity: filter.rarity,
          colors: filter.colors,
          owned_only: filter.ownedOnly,
        }
      : {};
    const apiResponse = await api.collection.getCollectionWithMetadata(apiFilter);
    // Create a proper CollectionResponse object
    const response = new gui.CollectionResponse();
    response.cards = apiResponse.cards;
    response.totalCount = apiResponse.totalCount;
    response.filterCount = apiResponse.filterCount;
    response.unknownCardsRemaining = apiResponse.unknownCardsRemaining;
    response.unknownCardsFetched = apiResponse.unknownCardsFetched;
    return response;
  },

  async getCollectionStats(): Promise<gui.CollectionStats> {
    return api.collection.getCollectionStats();
  },

  async getSetCompletion(): Promise<models.SetCompletion[]> {
    return api.collection.getSetCompletion();
  },
};

// ============================================================================
// System Adapter
// ============================================================================

export const systemAdapter = {
  async getConnectionStatus(): Promise<gui.ConnectionStatus> {
    return api.system.getStatus();
  },

  async getVersion(): Promise<{ version: string; service: string }> {
    return api.system.getVersion();
  },
};

// ============================================================================
// Events Adapter (WebSocket-based)
// ============================================================================

/**
 * Subscribe to an event.
 * Uses WebSocket for real-time events.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function EventsOn(eventName: string, callback: (...data: any[]) => void): () => void {
  return WsEventsOn(eventName, callback);
}

/**
 * Unsubscribe from events.
 */
export function EventsOff(eventName: string, ...additionalEventNames: string[]): void {
  WsEventsOff(eventName, ...additionalEventNames);
}

// ============================================================================
// Combined API Object
// ============================================================================

/**
 * Combined API adapter object for easy access to all services.
 */
export const apiAdapter = {
  matches: matchesAdapter,
  drafts: draftsAdapter,
  decks: decksAdapter,
  collection: collectionAdapter,
  system: systemAdapter,
  EventsOn,
  EventsOff,
  isRestApiEnabled,
  setUseRestApi,
  initialize: initializeServices,
  cleanup: cleanupServices,
};

// ============================================================================
// REST API Client Factory
// ============================================================================

/**
 * Create a REST API client that maps legacy method names to REST API calls.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function createRestApiClient(): Record<string, (...args: any[]) => Promise<any>> {
  return {
    // Match methods
    GetMatches: (filter: models.StatsFilter) => matchesAdapter.getMatches(filter),
    GetStats: (filter: models.StatsFilter) => matchesAdapter.getStats(filter),
    GetSupportedFormats: () => matchesAdapter.getFormats(),
    GetMatchGames: (matchId: string) => matchesAdapter.getMatchGames(matchId),

    // Draft methods
    GetActiveDraftSessions: () => draftsAdapter.getActiveDraftSessions(),
    GetCompletedDraftSessions: () => draftsAdapter.getCompletedDraftSessions(),
    GetDraftSession: (sessionId: string) => draftsAdapter.getDraftSession(sessionId),
    GetDraftPicks: (sessionId: string) => draftsAdapter.getDraftPicks(sessionId),
    GetDraftPacks: (sessionId: string) => api.drafts.getDraftPool(sessionId),

    // Deck methods
    ListDecks: () => decksAdapter.getDecks(),
    GetDecks: () => decksAdapter.getDecks(),
    GetDeck: (deckId: string) => decksAdapter.getDeck(deckId),
    CreateDeck: (name: string, format: string, source: string, draftEventId?: string) =>
      api.decks.createDeck({ name, format, source, draft_event_id: draftEventId }),
    DeleteDeck: (deckId: string) => decksAdapter.deleteDeck(deckId),
    ImportDeck: (req: gui.ImportDeckRequest) => decksAdapter.importDeck(req),
    ExportDeck: (req: gui.ExportDeckRequest) => decksAdapter.exportDeck(req),
    SuggestDecks: (draftEventId: string) => decksAdapter.suggestDecks(draftEventId),

    // Collection methods
    GetCollection: (filter?: gui.CollectionFilter) => collectionAdapter.getCollection(filter),
    GetCollectionStats: () => collectionAdapter.getCollectionStats(),
    GetSetCompletion: () => collectionAdapter.getSetCompletion(),

    // System methods
    GetConnectionStatus: () => systemAdapter.getConnectionStatus(),

    // Card methods
    GetSetCards: (setCode: string) => api.cards.getSetCards(setCode),
    GetCardByArenaID: (arenaId: number) => api.cards.getCardByArenaId(arenaId),
    GetAllSetInfo: () => api.cards.getAllSetInfo(),
    GetCardRatings: (setCode: string, draftFormat: string) =>
      api.cards.getCardRatings(setCode, draftFormat),
    SearchCards: (query: string) => api.cards.searchCards({ query }),

    // Quest methods
    GetActiveQuests: () => api.quests.getActiveQuests(),
    GetQuestHistory: (startDate?: string, endDate?: string, limit?: number) =>
      api.quests.getQuestHistory(startDate, endDate, limit),
    GetCurrentAccount: async () => ({ displayName: 'Player', accountID: '' }),

    // Stats methods
    GetTrendAnalysis: (startDate: Date, endDate: Date, periodType: string, formats: string[]) =>
      api.matches.getTrendAnalysis({
        start_date: startDate.toISOString().split('T')[0],
        end_date: endDate.toISOString().split('T')[0],
        period_type: periodType,
        formats,
      }),
    GetStatsByDeck: async () => ({}),
    GetStatsByFormat: (filter: models.StatsFilter) =>
      api.matches.getFormatDistribution(api.matches.statsFilterToRequest(filter)),
    GetRankProgressionTimeline: async (format: string) => ({ timeline: [], format }),

    // Meta methods
    GetMetaDashboard: async (format: string) => {
      const archetypes = await api.meta.getMetaArchetypes(format);
      return { archetypes, format };
    },
    RefreshMetaData: async (format: string) => {
      const archetypes = await api.meta.getMetaArchetypes(format);
      return { archetypes, format };
    },

    // Draft analysis methods
    AnalyzeSessionPickQuality: async () => Promise.resolve(),
    GetPickAlternatives: async () => null,
    GetDraftGrade: async () => null,
    CalculateDraftGrade: async () => null,
    GetCurrentPackWithRecommendation: async () => null,

    // Replay methods
    PauseReplay: async () => ({ status: 'ok' }),
    ResumeReplay: async () => ({ status: 'ok' }),
    StopReplay: async () => ({ status: 'ok' }),
    GetReplayStatus: async () => ({ isActive: false, isPaused: false }),

    // Settings methods
    GetAllSettings: () => api.settings.getSettings(),
    SaveAllSettings: (settings: gui.AppSettings) => api.settings.updateSettings(settings),
    GetSetting: (key: string) => api.settings.getSetting(key),
    SetSetting: (key: string, value: unknown) => api.settings.updateSetting(key, value),

    // Format methods
    GetFormats: () => matchesAdapter.getFormats(),
  };
}

export default apiAdapter;
