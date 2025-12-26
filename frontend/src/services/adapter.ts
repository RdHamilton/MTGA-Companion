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

export default apiAdapter;
