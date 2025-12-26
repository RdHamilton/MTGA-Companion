/**
 * Wails-to-REST Adapter
 *
 * This adapter provides a compatibility layer that allows gradual migration
 * from Wails bindings to REST API calls. Components can use this adapter
 * to transparently switch between backends without code changes.
 *
 * Usage:
 *   import { useApiAdapter } from '@/services/adapter';
 *
 *   // In component
 *   const api = useApiAdapter();
 *   const matches = await api.matches.getMatches(filter);
 *
 * Configuration:
 *   Set USE_REST_API environment variable or call setUseRestApi(true)
 *   to switch from Wails to REST API.
 */

import * as WailsApp from 'wailsjs/go/main/App';
import { EventsOn as WailsEventsOn, EventsOff as WailsEventsOff } from 'wailsjs/runtime/runtime';
import * as api from './api';
import {
  connect as wsConnect,
  disconnect as wsDisconnect,
  EventsOn as WsEventsOn,
  EventsOff as WsEventsOff,
} from './websocketClient';
import { configureApi, healthCheck } from './apiClient';
import { models, gui } from 'wailsjs/go/models';

// Configuration state
let useRestApi = false;
let isInitialized = false;

/**
 * Check if REST API mode is enabled.
 */
export function isRestApiEnabled(): boolean {
  return useRestApi;
}

/**
 * Enable or disable REST API mode.
 */
export function setUseRestApi(enabled: boolean): void {
  useRestApi = enabled;
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

  // Check environment or options
  useRestApi = options?.useRest ?? import.meta.env.VITE_USE_REST_API === 'true';

  if (useRestApi) {
    // Configure REST API
    if (options?.apiBaseUrl) {
      configureApi({ baseUrl: options.apiBaseUrl });
    }

    // Check if API is available
    const isHealthy = await healthCheck();
    if (!isHealthy) {
      console.warn('[Adapter] REST API not available, falling back to Wails');
      useRestApi = false;
    } else {
      // Connect WebSocket
      try {
        await wsConnect();
        console.log('[Adapter] REST API mode enabled');
      } catch (error) {
        console.error('[Adapter] WebSocket connection failed:', error);
      }
    }
  }

  isInitialized = true;
}

/**
 * Cleanup services on app shutdown.
 */
export function cleanupServices(): void {
  if (useRestApi) {
    wsDisconnect();
  }
  isInitialized = false;
}

// ============================================================================
// Matches Adapter
// ============================================================================

export const matchesAdapter = {
  async getMatches(filter: models.StatsFilter): Promise<models.Match[]> {
    if (useRestApi) {
      return api.matches.getMatches(api.matches.statsFilterToRequest(filter));
    }
    return WailsApp.GetMatches(filter);
  },

  async getStats(filter: models.StatsFilter): Promise<models.Statistics> {
    if (useRestApi) {
      return api.matches.getStats(api.matches.statsFilterToRequest(filter));
    }
    return WailsApp.GetStats(filter);
  },

  async getFormats(): Promise<string[]> {
    if (useRestApi) {
      return api.matches.getFormats();
    }
    return WailsApp.GetSupportedFormats();
  },

  async getMatchGames(matchId: string): Promise<models.Game[]> {
    if (useRestApi) {
      return api.matches.getMatchGames(matchId);
    }
    return WailsApp.GetMatchGames(matchId);
  },
};

// ============================================================================
// Drafts Adapter
// ============================================================================

export const draftsAdapter = {
  async getActiveDraftSessions(): Promise<models.DraftSession[]> {
    if (useRestApi) {
      return api.drafts.getActiveDraftSessions();
    }
    return WailsApp.GetActiveDraftSessions();
  },

  async getCompletedDraftSessions(limit = 100): Promise<models.DraftSession[]> {
    if (useRestApi) {
      return api.drafts.getCompletedDraftSessions();
    }
    return WailsApp.GetCompletedDraftSessions(limit);
  },

  async getDraftSession(sessionId: string): Promise<models.DraftSession> {
    if (useRestApi) {
      return api.drafts.getDraftSession(sessionId);
    }
    return WailsApp.GetDraftSession(sessionId);
  },

  async getDraftPicks(sessionId: string): Promise<models.DraftPickSession[]> {
    if (useRestApi) {
      return api.drafts.getDraftPicks(sessionId);
    }
    return WailsApp.GetDraftPicks(sessionId);
  },

  async getCardRatings(setCode: string, format: string): Promise<gui.CardRatingWithTier[]> {
    if (useRestApi) {
      return api.cards.getCardRatings(setCode, format);
    }
    return WailsApp.GetCardRatings(setCode, format);
  },
};

// ============================================================================
// Decks Adapter
// ============================================================================

export const decksAdapter = {
  async getDecks(): Promise<gui.DeckListItem[]> {
    if (useRestApi) {
      return api.decks.getDecks();
    }
    return WailsApp.ListDecks();
  },

  async getDeck(deckId: string): Promise<gui.DeckWithCards> {
    if (useRestApi) {
      return api.decks.getDeck(deckId);
    }
    return WailsApp.GetDeck(deckId);
  },

  async getDecksBySource(source: string): Promise<gui.DeckListItem[]> {
    if (useRestApi) {
      return api.decks.getDecksBySource(source);
    }
    return WailsApp.GetDecksBySource(source);
  },

  async getDecksByFormat(format: string): Promise<gui.DeckListItem[]> {
    if (useRestApi) {
      return api.decks.getDecksByFormat(format);
    }
    return WailsApp.GetDecksByFormat(format);
  },

  async createDeck(
    name: string,
    format: string,
    source: string,
    draftEventId?: string
  ): Promise<models.Deck> {
    if (useRestApi) {
      return api.decks.createDeck({ name, format, source, draft_event_id: draftEventId });
    }
    return WailsApp.CreateDeck(name, format, source, draftEventId || null);
  },

  async deleteDeck(deckId: string): Promise<void> {
    if (useRestApi) {
      return api.decks.deleteDeck(deckId);
    }
    return WailsApp.DeleteDeck(deckId);
  },

  async exportDeck(request: gui.ExportDeckRequest): Promise<gui.ExportDeckResponse> {
    if (useRestApi) {
      return api.decks.exportDeck(request.deckID, { format: request.format });
    }
    return WailsApp.ExportDeck(request);
  },

  async importDeck(request: gui.ImportDeckRequest): Promise<gui.ImportDeckResponse> {
    if (useRestApi) {
      return api.decks.importDeck({
        content: request.importText,
        name: request.name,
        format: request.format,
      });
    }
    return WailsApp.ImportDeck(request);
  },

  async suggestDecks(sessionId: string): Promise<gui.SuggestDecksResponse> {
    if (useRestApi) {
      const suggestions = await api.decks.suggestDecks({ session_id: sessionId });
      // Wrap in response format
      return {
        suggestions: suggestions,
        totalCombos: suggestions.length,
        viableCombos: suggestions.length,
      } as gui.SuggestDecksResponse;
    }
    return WailsApp.SuggestDecks(sessionId);
  },
};

// ============================================================================
// Collection Adapter
// ============================================================================

export const collectionAdapter = {
  async getCollection(filter?: gui.CollectionFilter): Promise<gui.CollectionResponse> {
    if (useRestApi) {
      const apiFilter: api.CollectionFilter = filter
        ? {
            set_code: filter.setCode,
            rarity: filter.rarity,
            colors: filter.colors,
            owned_only: filter.ownedOnly,
          }
        : {};
      const cards = await api.collection.getCollection(apiFilter);
      // Create a proper CollectionResponse object
      const response = new gui.CollectionResponse();
      response.cards = cards;
      response.totalCount = cards.length;
      response.filterCount = cards.length;
      return response;
    }
    return WailsApp.GetCollection(filter || new gui.CollectionFilter());
  },

  async getCollectionStats(): Promise<gui.CollectionStats> {
    if (useRestApi) {
      return api.collection.getCollectionStats();
    }
    return WailsApp.GetCollectionStats();
  },

  async getSetCompletion(): Promise<models.SetCompletion[]> {
    if (useRestApi) {
      return api.collection.getSetCompletion();
    }
    return WailsApp.GetSetCompletion();
  },
};

// ============================================================================
// System Adapter
// ============================================================================

export const systemAdapter = {
  async getConnectionStatus(): Promise<gui.ConnectionStatus> {
    if (useRestApi) {
      return api.system.getStatus();
    }
    return WailsApp.GetConnectionStatus();
  },

  async getVersion(): Promise<{ version: string; service: string }> {
    if (useRestApi) {
      return api.system.getVersion();
    }
    // Wails doesn't have a version endpoint, return app version
    return { version: '1.4.0', service: 'mtga-companion' };
  },
};

// ============================================================================
// Events Adapter
// ============================================================================

/**
 * Subscribe to an event.
 * Uses WebSocket in REST mode, Wails EventsOn otherwise.
 */
export function EventsOn(eventName: string, callback: (...data: unknown[]) => void): () => void {
  if (useRestApi) {
    return WsEventsOn(eventName, callback);
  }
  return WailsEventsOn(eventName, callback);
}

/**
 * Unsubscribe from events.
 * Uses WebSocket in REST mode, Wails EventsOff otherwise.
 */
export function EventsOff(eventName: string, ...additionalEventNames: string[]): void {
  if (useRestApi) {
    WsEventsOff(eventName, ...additionalEventNames);
  } else {
    WailsEventsOff(eventName, ...additionalEventNames);
  }
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
 * Create a REST API client that maps Wails method names to REST API calls.
 * This is used by the Go App polyfill to redirect window.go.main.App calls.
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
    GetCompletedDraftSessions: (limit: number) => draftsAdapter.getCompletedDraftSessions(limit),
    GetDraftSession: (sessionId: string) => draftsAdapter.getDraftSession(sessionId),
    GetDraftPicks: (sessionId: string) => draftsAdapter.getDraftPicks(sessionId),
    GetDraftPacks: (sessionId: string) => draftsAdapter.getDraftPacks(sessionId),

    // Deck methods
    ListDecks: () => decksAdapter.listDecks(),
    GetDeck: (deckId: string) => decksAdapter.getDeck(deckId),
    CreateDeck: (name: string, format: string, source: string, draftEventId?: string) =>
      decksAdapter.createDeck({ name, format, source, draftEventID: draftEventId }),
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

    // Card methods (use REST API directly)
    GetSetCards: async (setCode: string) => {
      const response = await api.cards.getSetCards(setCode);
      return response;
    },
    GetCardByArenaID: async (arenaId: string) => {
      const response = await api.cards.getCard(arenaId);
      return response;
    },
    GetAllSetInfo: async () => {
      const response = await api.cards.getSets();
      return response;
    },
    GetCardRatings: async (setCode: string, draftFormat: string) => {
      const response = await api.cards.getRatings(setCode, draftFormat);
      return response;
    },
    SearchCards: async (query: string, setCodes?: string[], limit?: number) => {
      const response = await api.cards.searchCards(query, setCodes, limit);
      return response;
    },

    // Quest methods (use REST API directly)
    GetActiveQuests: async () => {
      const response = await api.quests.getActiveQuests();
      return response;
    },
    GetQuestHistory: async (startDate?: string, endDate?: string, limit?: number) => {
      const response = await api.quests.getQuestHistory(startDate, endDate, limit);
      return response;
    },
    GetCurrentAccount: async () => {
      // Return a default account since this isn't implemented yet
      return { displayName: 'Player', accountID: '' };
    },

    // Stats methods
    GetTrendAnalysis: async (startDate: Date, endDate: Date, periodType: string, formats: string[]) => {
      const response = await api.matches.getTrendAnalysis({
        startDate: startDate.toISOString().split('T')[0],
        endDate: endDate.toISOString().split('T')[0],
        periodType,
        formats,
      });
      return response;
    },
    GetStatsByDeck: async (filter: models.StatsFilter) => {
      const response = await api.matches.getStats(api.matches.statsFilterToRequest(filter));
      // The REST API returns aggregate stats, not by-deck - return empty for now
      return {};
    },
    GetStatsByFormat: async (filter: models.StatsFilter) => {
      const response = await api.matches.getFormatDistribution(api.matches.statsFilterToRequest(filter));
      return response;
    },
    GetRankProgressionTimeline: async (format: string, startDate?: Date, endDate?: Date, periodType?: string) => {
      // Not implemented in REST API yet
      return { timeline: [], format };
    },

    // Meta methods
    GetMetaDashboard: async (format: string) => {
      const response = await api.meta.getMetaArchetypes(format);
      return { archetypes: response, format };
    },
    RefreshMetaData: async (format: string) => {
      const response = await api.meta.getMetaArchetypes(format);
      return { archetypes: response, format };
    },

    // Draft analysis methods (stubs - not fully implemented in REST API)
    AnalyzeSessionPickQuality: async () => Promise.resolve(),
    GetPickAlternatives: async () => null,
    GetDraftGrade: async () => null,
    CalculateDraftGrade: async () => null,
    GetCurrentPackWithRecommendation: async () => null,

    // Replay methods (stubs - handled differently in REST API mode)
    PauseReplay: async () => ({ status: 'ok' }),
    ResumeReplay: async () => ({ status: 'ok' }),
    StopReplay: async () => ({ status: 'ok' }),
    GetReplayStatus: async () => ({ isActive: false, isPaused: false }),

    // Settings methods
    GetAllSettings: async () => {
      const response = await api.settings.getSettings();
      return response;
    },
    SaveAllSettings: async (settings: gui.AppSettings) => {
      await api.settings.updateSettings(settings);
    },
    GetSetting: async (key: string) => {
      const response = await api.settings.getSetting(key);
      return response;
    },
    SetSetting: async (key: string, value: unknown) => {
      await api.settings.updateSetting(key, value);
    },

    // Format methods
    GetFormats: async () => {
      return matchesAdapter.getFormats();
    },
  };
}

export default apiAdapter;
