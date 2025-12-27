/**
 * Legacy API Exports
 *
 * This module provides exports that match the old Wails binding function names,
 * allowing gradual migration of components to use the REST API.
 *
 * Usage:
 *   import { GetStats, GetMatches } from '@/services/api/legacy';
 */

import * as matches from './matches';
import * as drafts from './drafts';
import * as decks from './decks';
import * as cards from './cards';
import * as collection from './collection';
import * as quests from './quests';
import * as meta from './meta';
import * as settings from './settings';
import * as system from './system';
import { models, gui, grading, storage, insights, pickquality, metrics, prediction, seventeenlands } from '@/types/models';

// ============================================================================
// Match Functions
// ============================================================================

export async function GetMatches(filter: models.StatsFilter): Promise<models.Match[]> {
  return matches.getMatches(matches.statsFilterToRequest(filter));
}

export async function GetStats(filter: models.StatsFilter): Promise<models.Statistics> {
  return matches.getStats(matches.statsFilterToRequest(filter));
}

export async function GetStatsByFormat(filter: models.StatsFilter): Promise<Record<string, models.Statistics>> {
  return matches.getFormatDistribution(matches.statsFilterToRequest(filter));
}

export async function GetStatsByDeck(filter: models.StatsFilter): Promise<Record<string, models.Statistics>> {
  return matches.getMatchupMatrix(matches.statsFilterToRequest(filter));
}

export async function GetMatchGames(matchId: string): Promise<models.Game[]> {
  return matches.getMatchGames(matchId);
}

export async function GetSupportedFormats(): Promise<string[]> {
  return matches.getFormats();
}

export async function GetPerformanceMetrics(filter: models.StatsFilter): Promise<models.PerformanceMetrics> {
  return matches.getPerformanceMetrics(matches.statsFilterToRequest(filter));
}

export async function GetTrendAnalysis(
  startDate: Date,
  endDate: Date,
  periodType: string,
  formats: string[]
): Promise<storage.TrendAnalysis> {
  const result = await matches.getTrendAnalysis({
    start_date: startDate.toISOString().split('T')[0],
    end_date: endDate.toISOString().split('T')[0],
    period_type: periodType,
    formats,
  });
  return result as storage.TrendAnalysis;
}

export async function GetRankProgression(format: string): Promise<models.RankProgression> {
  return matches.getRankProgression(format);
}

export async function GetRankProgressionTimeline(
  _format: string,
  _startDate?: Date,
  _endDate?: Date,
  _periodType?: string
): Promise<storage.RankTimeline> {
  // Not implemented in REST API yet
  return { timeline: [], format: _format } as unknown as storage.RankTimeline;
}

// ============================================================================
// Draft Functions
// ============================================================================

export async function GetActiveDraftSessions(): Promise<models.DraftSession[]> {
  return drafts.getActiveDraftSessions();
}

export async function GetCompletedDraftSessions(_limit?: number): Promise<models.DraftSession[]> {
  return drafts.getCompletedDraftSessions();
}

export async function GetDraftSession(sessionId: string): Promise<models.DraftSession> {
  return drafts.getDraftSession(sessionId);
}

export async function GetDraftPicks(sessionId: string): Promise<models.DraftPickSession[]> {
  return drafts.getDraftPicks(sessionId);
}

export async function GetDraftPacks(sessionId: string): Promise<models.DraftPackSession[]> {
  return drafts.getDraftPool(sessionId) as unknown as Promise<models.DraftPackSession[]>;
}

export async function GetDraftDeckMetrics(sessionId: string): Promise<models.DeckMetrics> {
  return drafts.getDraftDeckMetrics(sessionId);
}

export async function GetDraftPerformanceMetrics(): Promise<metrics.DraftStats> {
  return drafts.getDraftPerformanceMetrics();
}

export async function AnalyzeSessionPickQuality(sessionId: string): Promise<void> {
  return drafts.analyzeSessionPickQuality(sessionId);
}

export async function GetPickAlternatives(
  sessionId: string,
  packNumber: number,
  pickNumber: number
): Promise<pickquality.PickQuality> {
  return drafts.getPickAlternatives(sessionId, packNumber, pickNumber);
}

export async function GetDraftGrade(sessionId: string): Promise<grading.DraftGrade> {
  return drafts.getDraftGrade(sessionId);
}

export async function CalculateDraftGrade(sessionId: string): Promise<grading.DraftGrade> {
  return drafts.calculateDraftGrade(sessionId);
}

export async function GetCurrentPackWithRecommendation(sessionId: string): Promise<gui.CurrentPackResponse> {
  return drafts.getCurrentPackWithRecommendation(sessionId);
}

export async function GetDraftWinRatePrediction(sessionId: string): Promise<prediction.DeckPrediction> {
  return drafts.getDraftWinRatePrediction(sessionId);
}

export async function PredictDraftWinRate(sessionId: string): Promise<prediction.DeckPrediction> {
  return drafts.getDraftWinRatePrediction(sessionId);
}

// ============================================================================
// Deck Functions
// ============================================================================

export async function ListDecks(): Promise<gui.DeckListItem[]> {
  return decks.getDecks();
}

export async function GetDeck(deckId: string): Promise<gui.DeckWithCards> {
  return decks.getDeck(deckId);
}

export async function GetDecksBySource(source: string): Promise<gui.DeckListItem[]> {
  return decks.getDecksBySource(source);
}

export async function GetDecksByFormat(format: string): Promise<gui.DeckListItem[]> {
  return decks.getDecksByFormat(format);
}

export async function GetDecksByTags(tags: string[]): Promise<gui.DeckListItem[]> {
  return decks.getDecksByTags(tags);
}

export async function GetDeckLibrary(filter: gui.DeckLibraryFilter): Promise<gui.DeckListItem[]> {
  return decks.getDeckLibrary(filter);
}

export async function CreateDeck(
  name: string,
  format: string,
  source: string,
  draftEventId?: string | null
): Promise<models.Deck> {
  return decks.createDeck({ name, format, source, draft_event_id: draftEventId || undefined });
}

export async function UpdateDeck(deck: models.Deck): Promise<void> {
  await decks.updateDeck(deck.ID, {
    name: deck.Name,
    format: deck.Format,
  });
}

export async function DeleteDeck(deckId: string): Promise<void> {
  return decks.deleteDeck(deckId);
}

export async function CloneDeck(deckId: string, newName: string): Promise<models.Deck> {
  return decks.cloneDeck(deckId, newName);
}

export async function GetDeckByDraftEvent(draftEventId: string): Promise<gui.DeckWithCards> {
  return decks.getDeckByDraftEvent(draftEventId);
}

export async function GetDeckStatistics(deckId: string): Promise<gui.DeckStatistics> {
  return decks.getDeckStatistics(deckId);
}

export async function GetDeckPerformance(deckId: string): Promise<models.DeckPerformance> {
  return decks.getDeckPerformance(deckId);
}

export async function ExportDeck(request: gui.ExportDeckRequest): Promise<gui.ExportDeckResponse> {
  return decks.exportDeck(request.deckID, { format: request.format });
}

export async function ImportDeck(request: gui.ImportDeckRequest): Promise<gui.ImportDeckResponse> {
  return decks.importDeck({
    content: request.importText,
    name: request.name,
    format: request.format,
  });
}

export async function ValidateDraftDeck(deckId: string): Promise<boolean> {
  return decks.validateDraftDeck(deckId);
}

export async function SuggestDecks(sessionId: string): Promise<gui.SuggestDecksResponse> {
  const suggestions = await decks.suggestDecks({ session_id: sessionId });
  return {
    suggestions,
    totalCombos: suggestions.length,
    viableCombos: suggestions.length,
  } as gui.SuggestDecksResponse;
}

export async function ApplySuggestedDeck(deckId: string, suggestion: gui.SuggestedDeckResponse): Promise<void> {
  return decks.applySuggestedDeck(deckId, suggestion);
}

export async function ExportSuggestedDeck(suggestion: gui.SuggestedDeckResponse, deckName: string): Promise<void> {
  const content = await decks.getSuggestedDeckExportContent(suggestion, deckName);
  downloadTextFile(content, `${deckName || 'deck'}.txt`);
}

export async function AddCard(
  deckId: string,
  arenaId: number,
  quantity: number,
  zone: string,
  isSideboard: boolean
): Promise<void> {
  return decks.addCard({ deck_id: deckId, arena_id: arenaId, quantity, zone, is_sideboard: isSideboard });
}

export async function RemoveCard(deckId: string, arenaId: number, zone: string): Promise<void> {
  return decks.removeCard({ deck_id: deckId, arena_id: arenaId, zone });
}

export async function AddTag(deckId: string, tag: string): Promise<void> {
  return decks.addTag(deckId, tag);
}

export async function RemoveTag(deckId: string, tag: string): Promise<void> {
  return decks.removeTag(deckId, tag);
}

// ============================================================================
// Card Functions
// ============================================================================

export async function GetSetCards(setCode: string): Promise<models.SetCard[]> {
  return cards.getSetCards(setCode);
}

export async function GetCardByArenaID(arenaId: string | number): Promise<models.SetCard> {
  return cards.getCardByArenaId(typeof arenaId === 'string' ? parseInt(arenaId, 10) : arenaId);
}

export async function SearchCards(query: string, _sets?: string[], _limit?: number): Promise<models.SetCard[]> {
  return cards.searchCards({ query });
}

export async function GetCardRatings(setCode: string, format: string): Promise<gui.CardRatingWithTier[]> {
  return cards.getCardRatings(setCode, format);
}

export async function GetCardRatingByArenaID(
  arenaId: string,
  setCode: string,
  format: string
): Promise<gui.CardRatingWithTier> {
  const ratings = await cards.getCardRatings(setCode, format);
  const rating = ratings.find((r) => r.mtga_id?.toString() === arenaId);
  if (!rating) {
    throw new Error(`Card rating not found for arena ID ${arenaId}`);
  }
  return rating;
}

export async function GetColorRatings(setCode: string, format: string): Promise<seventeenlands.ColorRating[]> {
  return cards.getColorRatings(setCode, format);
}

export async function GetAllSetInfo(): Promise<gui.SetInfo[]> {
  return cards.getAllSetInfo();
}

export async function GetSetInfo(setCode: string): Promise<gui.SetInfo> {
  const sets = await cards.getAllSetInfo();
  const set = sets.find((s) => s.code === setCode);
  if (!set) {
    throw new Error(`Set not found: ${setCode}`);
  }
  return set;
}

export async function FetchSetCards(setCode: string): Promise<number> {
  const fetchedCards = await cards.getSetCards(setCode);
  return fetchedCards.length;
}

export async function RefreshSetCards(setCode: string): Promise<number> {
  const fetchedCards = await cards.getSetCards(setCode);
  return fetchedCards.length;
}

export async function FetchSetRatings(setCode: string, format: string): Promise<void> {
  await cards.getCardRatings(setCode, format);
}

export async function RefreshSetRatings(setCode: string, format: string): Promise<void> {
  await cards.getCardRatings(setCode, format);
}

// ============================================================================
// Collection Functions
// ============================================================================

export async function GetCollection(filter?: gui.CollectionFilter): Promise<gui.CollectionResponse> {
  const apiFilter = filter
    ? {
        set_code: filter.setCode,
        rarity: filter.rarity,
        colors: filter.colors,
        owned_only: filter.ownedOnly,
      }
    : {};
  const collectionCards = await collection.getCollection(apiFilter);
  const response = new gui.CollectionResponse();
  response.cards = collectionCards;
  response.totalCount = collectionCards.length;
  response.filterCount = collectionCards.length;
  return response;
}

export async function GetCollectionStats(): Promise<gui.CollectionStats> {
  return collection.getCollectionStats();
}

export async function GetSetCompletion(): Promise<models.SetCompletion[]> {
  return collection.getSetCompletion();
}

export async function GetRecentCollectionChanges(limit: number): Promise<gui.CollectionChangeEntry[]> {
  return collection.getRecentChanges(limit);
}

export async function GetMissingCards(
  sessionIdOrSetCode: string,
  packNumber?: number,
  _pickNumber?: number
): Promise<models.MissingCardsAnalysis> {
  // If packNumber is provided, this is a draft context call
  if (packNumber !== undefined) {
    // For draft context, return empty analysis (not implemented in REST API)
    return {
      TotalMissing: 0,
      ByRarity: {},
      TopMissing: [],
      SessionID: sessionIdOrSetCode,
      PackNumber: packNumber,
      PickNumber: _pickNumber || 0,
      InitialCards: [],
      MissingCards: [],
      PassedCards: [],
      CardsToLookFor: [],
      convertValues: () => ({}),
    } as unknown as models.MissingCardsAnalysis;
  }
  // Otherwise, treat as setCode for collection missing cards
  return collection.getMissingCards(sessionIdOrSetCode);
}

export async function GetMissingCardsForDeck(deckId: string): Promise<gui.MissingCardsForDeckResponse> {
  return collection.getMissingCardsForDeck(deckId);
}

export async function GetMissingCardsForSet(setCode: string): Promise<gui.MissingCardsForSetResponse> {
  const missingCards = await collection.getMissingCardsForSet(setCode);
  // getMissingCardsForSet returns CollectionCard[], we need to convert to MissingCardsForSetResponse
  return {
    setCode,
    setName: setCode,
    totalMissing: missingCards.length,
    uniqueMissing: missingCards.length,
    missingCards,
    byRarity: {},
    completionPercentage: 0,
    completionPct: 0,
    convertValues: () => ({}),
  } as unknown as gui.MissingCardsForSetResponse;
}

// ============================================================================
// Quest Functions
// ============================================================================

export async function GetActiveQuests(): Promise<models.Quest[]> {
  return quests.getActiveQuests();
}

export async function GetQuestHistory(startDate?: string, endDate?: string, limit?: number): Promise<models.Quest[]> {
  return quests.getQuestHistory(startDate, endDate, limit);
}

export async function GetQuestStats(startDate: string, endDate: string): Promise<models.QuestStats> {
  return quests.getQuestStats(startDate, endDate);
}

export async function GetCurrentAccount(): Promise<models.Account> {
  return system.getCurrentAccount();
}

// ============================================================================
// Meta Functions
// ============================================================================

export async function GetMetaDashboard(format: string): Promise<gui.MetaDashboardResponse> {
  const archetypes = await meta.getMetaArchetypes(format);
  return {
    archetypes: archetypes as unknown as gui.ArchetypeInfo[],
    format,
    totalArchetypes: archetypes.length,
    lastUpdated: new Date().toISOString(),
    sources: [],
    convertValues: () => ({}),
  } as gui.MetaDashboardResponse;
}

export async function RefreshMetaData(format: string): Promise<gui.MetaDashboardResponse> {
  const archetypes = await meta.getMetaArchetypes(format);
  return {
    archetypes: archetypes as unknown as gui.ArchetypeInfo[],
    format,
    totalArchetypes: archetypes.length,
    lastUpdated: new Date().toISOString(),
    sources: [],
    convertValues: () => ({}),
  } as gui.MetaDashboardResponse;
}

export async function GetTierArchetypes(format: string, tier: number): Promise<gui.ArchetypeInfo[]> {
  return meta.getTierArchetypes(format, tier);
}

export async function GetArchetypeCards(
  format: string,
  archetypeName: string,
  _category?: string
): Promise<insights.ArchetypeCards> {
  return meta.getArchetypeCards(format, archetypeName);
}

export async function GetFormatInsights(format: string, setCode: string): Promise<insights.FormatInsights> {
  return meta.getFormatInsights(format, setCode);
}

// ============================================================================
// Settings Functions
// ============================================================================

export async function GetAllSettings(): Promise<gui.AppSettings> {
  return settings.getSettings();
}

export async function SaveAllSettings(appSettings: gui.AppSettings): Promise<void> {
  return settings.updateSettings(appSettings);
}

export async function GetSetting(key: string): Promise<unknown> {
  return settings.getSetting(key);
}

export async function SetSetting(key: string, value: unknown): Promise<void> {
  return settings.updateSetting(key, value);
}

// ============================================================================
// System Functions
// ============================================================================

export async function GetConnectionStatus(): Promise<gui.ConnectionStatus> {
  return system.getStatus();
}

export async function GetAppVersion(): Promise<string> {
  const info = await system.getVersion();
  return info.version;
}

export async function Initialize(_logPath: string): Promise<void> {
  // No-op in REST API mode
}

export async function StartPoller(): Promise<void> {
  // No-op in REST API mode
}

export async function StopPoller(): Promise<void> {
  // No-op in REST API mode
}

export async function ReconnectToDaemon(): Promise<void> {
  // No-op in REST API mode
}

export async function SwitchToDaemonMode(): Promise<void> {
  // No-op in REST API mode
}

export async function SwitchToStandaloneMode(): Promise<void> {
  // No-op in REST API mode
}

export async function SetDaemonPort(_port: number): Promise<void> {
  // No-op in REST API mode
}

// ============================================================================
// Export/Import Functions
// ============================================================================

export async function ExportToJSON(): Promise<void> {
  const data = await matches.exportMatches('json');
  downloadTextFile(JSON.stringify(data, null, 2), 'mtga-matches.json');
}

export async function ExportToCSV(): Promise<void> {
  const data = await matches.exportMatches('csv');
  downloadTextFile(String(data), 'mtga-matches.csv');
}

export async function ImportFromFile(): Promise<void> {
  console.warn('ImportFromFile requires file picker - use browser file input');
}

export async function ImportLogFile(): Promise<gui.ImportLogFileResult> {
  console.warn('ImportLogFile requires file picker - use browser file input');
  return {
    fileName: '',
    entriesRead: 0,
    matchesStored: 0,
    gamesStored: 0,
    draftsStored: 0,
    picksStored: 0,
    collectionsStored: 0,
    inventoriesStored: 0,
    questsStored: 0,
    decksStored: 0,
    ranksStored: 0,
    errors: [],
  } as unknown as gui.ImportLogFileResult;
}

export async function ClearAllData(): Promise<void> {
  return system.clearAllData();
}

// ============================================================================
// Replay Functions
// ============================================================================

export async function GetReplayStatus(): Promise<gui.ReplayStatus> {
  return {
    isActive: false,
    isPaused: false,
    currentEntry: 0,
    totalEntries: 0,
    percentComplete: 0,
    elapsed: 0,
    estimatedRemaining: 0,
    speed: 0,
    filter: '',
  } as unknown as gui.ReplayStatus;
}

export async function StartReplayWithFileDialog(
  _speed: number,
  _filterType: string,
  _pauseOnDraft: boolean
): Promise<void> {
  console.warn('StartReplayWithFileDialog requires file picker - use browser file input');
}

export async function PauseReplay(): Promise<void> {
  // Not implemented in REST API yet
}

export async function ResumeReplay(): Promise<void> {
  // Not implemented in REST API yet
}

export async function StopReplay(): Promise<void> {
  // Not implemented in REST API yet
}

export async function GetLogReplayProgress(): Promise<gui.LogReplayProgress> {
  return {
    totalFiles: 0,
    processedFiles: 0,
    currentFile: '',
    totalEntries: 0,
    processedEntries: 0,
    matchesFound: 0,
    draftsFound: 0,
    collectionsFound: 0,
    questsFound: 0,
    inventoriesFound: 0,
    percentComplete: 0,
    matchesImported: 0,
    decksImported: 0,
    questsImported: 0,
    ranksImported: 0,
    errors: [],
  } as unknown as gui.LogReplayProgress;
}

export async function TriggerReplayLogs(_forceRefresh: boolean): Promise<void> {
  // Not implemented in REST API yet
}

// ============================================================================
// LLM/ML Functions
// ============================================================================

export async function CheckOllamaStatus(endpoint: string, model: string): Promise<gui.OllamaStatus> {
  return system.checkOllamaStatus(endpoint, model);
}

export async function GetAvailableOllamaModels(endpoint: string): Promise<gui.OllamaModel[]> {
  return system.getAvailableOllamaModels(endpoint);
}

export async function PullOllamaModel(endpoint: string, model: string): Promise<void> {
  return system.pullOllamaModel(endpoint, model);
}

export async function TestLLMGeneration(endpoint: string, model: string): Promise<string> {
  return system.testLLMGeneration(endpoint, model);
}

export async function ExportMLTrainingData(limit: number): Promise<gui.MLTrainingDataExport> {
  return system.exportMLTrainingData(limit);
}

// ============================================================================
// Recommendation Functions
// ============================================================================

export async function GetRecommendations(request: gui.GetRecommendationsRequest): Promise<gui.GetRecommendationsResponse> {
  return drafts.getRecommendations(request);
}

export async function RecordRecommendation(
  request: gui.RecordRecommendationRequest
): Promise<gui.RecordRecommendationResponse> {
  return drafts.recordRecommendation(request);
}

export async function RecordRecommendationAction(request: gui.RecordActionRequest): Promise<void> {
  return drafts.recordRecommendationAction(request);
}

export async function RecordRecommendationOutcome(request: gui.RecordOutcomeRequest): Promise<void> {
  return drafts.recordRecommendationOutcome(request);
}

export async function GetRecommendationStats(): Promise<gui.RecommendationStatsResponse> {
  return drafts.getRecommendationStats();
}

export async function ExplainRecommendation(
  request: gui.ExplainRecommendationRequest
): Promise<gui.ExplainRecommendationResponse> {
  return drafts.explainRecommendation(request);
}

// ============================================================================
// Classification Functions
// ============================================================================

export async function ClassifyDeckArchetype(deckId: string): Promise<gui.ArchetypeClassificationResult> {
  return decks.classifyDeckArchetype(deckId);
}

export async function ClassifyDraftPoolArchetype(sessionId: string): Promise<gui.ArchetypeClassificationResult> {
  return drafts.classifyDraftPoolArchetype(sessionId);
}

// ============================================================================
// Helper Functions
// ============================================================================

function downloadTextFile(content: string, filename: string): void {
  const blob = new Blob([content], { type: 'text/plain' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

// ============================================================================
// Additional Missing Exports
// ============================================================================

export async function SearchCardsWithCollection(
  query: string,
  sets?: string[],
  limit?: number,
  _collectionOnly?: boolean
): Promise<cards.CardWithCollection[]> {
  return cards.searchCardsWithCollection(query, sets, limit);
}

export async function ResetDraftPerformanceMetrics(): Promise<void> {
  // Not implemented in REST API - stats are computed on demand
  console.warn('ResetDraftPerformanceMetrics: Not implemented in REST API');
}

export async function RecalculateAllDraftGrades(_setCode?: string, _format?: string): Promise<number> {
  // Not implemented in REST API
  console.warn('RecalculateAllDraftGrades: Not implemented in REST API');
  return 0;
}

export async function ClearDatasetCache(): Promise<void> {
  // Not implemented in REST API
  console.warn('ClearDatasetCache: Not implemented in REST API');
}

export async function GetDatasetSource(_setCode?: string, _format?: string): Promise<string> {
  // Not implemented in REST API - return default
  return '17lands';
}

export async function ExportDeckToFile(deckId: string, format: string = 'txt'): Promise<void> {
  const response = await decks.exportDeck(deckId, { format });
  downloadTextFile(response.content, response.filename || `deck.${format}`);
}

export async function ValidateDeckWithDialog(deckId: string): Promise<boolean> {
  return decks.validateDraftDeck(deckId);
}
