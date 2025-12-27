import { vi } from 'vitest';

// API module mocks for the REST API service
// eslint-disable-next-line @typescript-eslint/no-explicit-any
type MockFn = (...args: any[]) => any;

export const mockCards = {
  getCardByArenaId: vi.fn((() => Promise.resolve({} as unknown)) as MockFn),
  getCardRatings: vi.fn((() => Promise.resolve([] as unknown[])) as MockFn),
  getSetCards: vi.fn((() => Promise.resolve([] as unknown[])) as MockFn),
  getSetInfo: vi.fn((() => Promise.resolve({} as unknown)) as MockFn),
  getAllSetInfo: vi.fn((() => Promise.resolve([] as unknown[])) as MockFn),
  searchCards: vi.fn((() => Promise.resolve([] as unknown[])) as MockFn),
  searchCardsWithCollection: vi.fn((() => Promise.resolve([] as unknown[])) as MockFn),
  getColorRatings: vi.fn((() => Promise.resolve([] as unknown[])) as MockFn),
};

export const mockMatches = {
  getMatches: vi.fn((() => Promise.resolve([] as unknown[])) as MockFn),
  getStats: vi.fn((() => Promise.resolve({} as unknown)) as MockFn),
  getFormats: vi.fn(() => Promise.resolve(['standard', 'historic', 'explorer', 'pioneer', 'modern'])),
  getMatchGames: vi.fn((() => Promise.resolve([] as unknown[])) as MockFn),
  getMatchesBySessionId: vi.fn((() => Promise.resolve([] as unknown[])) as MockFn),
  getStatsByDeck: vi.fn((() => Promise.resolve({} as unknown)) as MockFn),
  getStatsByFormat: vi.fn((() => Promise.resolve({} as unknown)) as MockFn),
  getTrendAnalysis: vi.fn((() => Promise.resolve({} as unknown)) as MockFn),
  getRankProgression: vi.fn((() => Promise.resolve({} as unknown)) as MockFn),
  getRankProgressionTimeline: vi.fn((() => Promise.resolve({} as unknown)) as MockFn),
  getMatchupMatrix: vi.fn((() => Promise.resolve({} as unknown)) as MockFn),
  getFormatDistribution: vi.fn((() => Promise.resolve({} as unknown)) as MockFn),
  exportMatches: vi.fn((() => Promise.resolve([] as unknown[])) as MockFn),
  statsFilterToRequest: vi.fn(((filter: unknown) => filter) as MockFn),
};

export const mockDecks = {
  getDecks: vi.fn((() => Promise.resolve([] as unknown[])) as MockFn),
  getDeck: vi.fn((() => Promise.resolve({} as unknown)) as MockFn),
  getDecksBySource: vi.fn((() => Promise.resolve([] as unknown[])) as MockFn),
  getDecksByFormat: vi.fn((() => Promise.resolve([] as unknown[])) as MockFn),
  getDeckByDraftEvent: vi.fn((() => Promise.resolve({} as unknown)) as MockFn),
  getDeckStatistics: vi.fn((() => Promise.resolve({} as unknown)) as MockFn),
  createDeck: vi.fn((() => Promise.resolve({} as unknown)) as MockFn),
  deleteDeck: vi.fn((() => Promise.resolve()) as MockFn),
  addCard: vi.fn((() => Promise.resolve()) as MockFn),
  removeCard: vi.fn((() => Promise.resolve()) as MockFn),
  exportDeck: vi.fn((() => Promise.resolve({} as unknown)) as MockFn),
  suggestDecks: vi.fn((() => Promise.resolve([] as unknown[])) as MockFn),
  applySuggestedDeck: vi.fn((() => Promise.resolve()) as MockFn),
  validateDraftDeck: vi.fn((() => Promise.resolve(true)) as MockFn),
};

export const mockDrafts = {
  getActiveDraftSessions: vi.fn(() => Promise.resolve([] as unknown[])),
  getCompletedDraftSessions: vi.fn(() => Promise.resolve([] as unknown[])),
  getDraftSession: vi.fn(() => Promise.resolve({} as unknown)),
  getDraftPicks: vi.fn(() => Promise.resolve([] as unknown[])),
  getDraftPool: vi.fn(() => Promise.resolve([] as unknown[])),
  getDraftGrade: vi.fn(() => Promise.resolve(null as unknown)),
  calculateDraftGrade: vi.fn(() => Promise.resolve({} as unknown)),
  predictDraftWinRate: vi.fn(() => Promise.resolve({} as unknown)),
  getWinRatePrediction: vi.fn(() => Promise.resolve(null as unknown)),
  getDraftDeckMetrics: vi.fn(() => Promise.resolve({} as unknown)),
  getDraftPerformanceMetrics: vi.fn(() => Promise.resolve({} as unknown)),
  getPickAlternatives: vi.fn(() => Promise.resolve([] as unknown[])),
  getCurrentPackWithRecommendation: vi.fn(() => Promise.resolve(null as unknown)),
  explainRecommendation: vi.fn(() => Promise.resolve({ explanation: 'This card is recommended because...', error: '' })),
  analyzeSessionPickQuality: vi.fn(() => Promise.resolve()),
  fixDraftSessionStatuses: vi.fn(() => Promise.resolve(0)),
  resetDraftPerformanceMetrics: vi.fn(() => Promise.resolve()),
  getRecommendations: vi.fn(() => Promise.resolve({} as unknown)),
};

export const mockCollection = {
  getCollection: vi.fn(() => Promise.resolve([] as unknown[])),
  getCollectionStats: vi.fn(() => Promise.resolve({} as unknown)),
  getSetCompletion: vi.fn(() => Promise.resolve([] as unknown[])),
  getMissingCards: vi.fn(() => Promise.resolve(null as unknown)),
  getRecentChanges: vi.fn(() => Promise.resolve([] as unknown[])),
};

export const mockMeta = {
  getMetaArchetypes: vi.fn(() => Promise.resolve([] as unknown[])),
  getFormatInsights: vi.fn(() => Promise.resolve({} as unknown)),
  getArchetypeCards: vi.fn(() => Promise.resolve({} as unknown)),
  refreshMetaData: vi.fn(() => Promise.resolve()),
  getFormats: vi.fn(() => Promise.resolve(['standard', 'historic', 'explorer', 'pioneer', 'modern'])),
  getTierArchetypes: vi.fn(() => Promise.resolve([] as unknown[])),
};

export const mockQuests = {
  getActiveQuests: vi.fn(),
  getQuestHistory: vi.fn(),
  getCurrentAccount: vi.fn(),
};

export const mockSettings = {
  getAllSettings: vi.fn(() => Promise.resolve({
    autoRefresh: false,
    refreshInterval: 30,
    showNotifications: true,
    theme: 'dark',
    daemonPort: 9999,
    daemonMode: 'standalone',
    mlEnabled: true,
    llmEnabled: false,
    ollamaEndpoint: 'http://localhost:11434',
    ollamaModel: 'qwen3:8b',
    metaGoldfishEnabled: true,
    metaTop8Enabled: true,
    metaWeight: 0.3,
    personalWeight: 0.2,
  } as unknown)),
  saveAllSettings: vi.fn(() => Promise.resolve()),
};

export const mockSystem = {
  getStatus: vi.fn(() => Promise.resolve({
    status: 'standalone',
    connected: false,
    mode: 'standalone',
    url: 'ws://localhost:9999',
    port: 9999,
  } as unknown)),
  getVersion: vi.fn(() => Promise.resolve({ version: '1.0.0', buildDate: '2024-01-01' } as unknown)),
  clearAllData: vi.fn(() => Promise.resolve()),
  checkOllamaStatus: vi.fn(() => Promise.resolve({
    available: true,
    version: '0.1.0',
    modelReady: true,
    modelName: 'qwen3:8b',
    modelsLoaded: ['qwen3:8b'],
    error: '',
  } as unknown)),
  getAvailableOllamaModels: vi.fn(() => Promise.resolve([{ name: 'qwen3:8b', size: 0 }] as unknown[])),
  pullOllamaModel: vi.fn(() => Promise.resolve()),
  testLLMGeneration: vi.fn(() => Promise.resolve('Hello from Ollama!')),
  getCurrentAccount: vi.fn(),
  resumeReplay: vi.fn(() => Promise.resolve()),
  stopReplay: vi.fn(() => Promise.resolve()),
  pauseReplay: vi.fn(() => Promise.resolve()),
  connectDaemon: vi.fn(() => Promise.resolve()),
};

// Combined API mock export
export const mockApi = {
  cards: mockCards,
  matches: mockMatches,
  decks: mockDecks,
  drafts: mockDrafts,
  collection: mockCollection,
  meta: mockMeta,
  quests: mockQuests,
  settings: mockSettings,
  system: mockSystem,
};

// Legacy mock kept for backwards compatibility with tests that haven't migrated
export const mockWailsApp = {
  AnalyzeSessionPickQuality: mockDrafts.analyzeSessionPickQuality,
  CalculateDraftGrade: mockDrafts.calculateDraftGrade,
  ClearAllData: mockSystem.clearAllData,
  ClearDatasetCache: vi.fn(() => Promise.resolve()),
  ExportDeck: mockDecks.exportDeck,
  ExportToCSV: mockMatches.exportMatches,
  ExportToJSON: mockMatches.exportMatches,
  FetchSetCards: vi.fn(() => Promise.resolve(0)),
  FetchSetRatings: mockCards.getCardRatings,
  FixDraftSessionStatuses: mockDrafts.fixDraftSessionStatuses,
  GetActiveDraftSessions: mockDrafts.getActiveDraftSessions,
  GetActiveQuests: mockQuests.getActiveQuests,
  GetCurrentAccount: mockQuests.getCurrentAccount,
  GetQuestHistory: mockQuests.getQuestHistory,
  GetArchetypeCards: mockMeta.getArchetypeCards,
  GetCardByArenaID: mockCards.getCardByArenaId,
  GetCardRatingByArenaID: vi.fn(() => Promise.resolve({} as unknown)),
  GetCardRatings: mockCards.getCardRatings,
  GetColorRatings: mockCards.getColorRatings,
  GetCompletedDraftSessions: mockDrafts.getCompletedDraftSessions,
  GetConnectionStatus: mockSystem.getStatus,
  GetDeckDetails: mockDecks.getDeck,
  GetDraftDeckMetrics: mockDrafts.getDraftDeckMetrics,
  GetDraftGrade: mockDrafts.getDraftGrade,
  GetDraftWinRatePrediction: mockDrafts.getWinRatePrediction,
  GetCurrentPackWithRecommendation: mockDrafts.getCurrentPackWithRecommendation,
  GetDraftPacks: vi.fn(() => Promise.resolve([] as unknown[])),
  GetDraftPicks: mockDrafts.getDraftPicks,
  GetFormatArchetypes: vi.fn(() => Promise.resolve([] as unknown[])),
  GetFormatInsights: mockMeta.getFormatInsights,
  GetFormatStats: vi.fn(() => Promise.resolve({} as unknown)),
  GetMatches: mockMatches.getMatches,
  GetMatchesBySessionID: mockMatches.getMatchesBySessionId,
  GetMatchGames: mockMatches.getMatchGames,
  GetMissingCards: mockCollection.getMissingCards,
  GetPerformanceMetrics: vi.fn(() => Promise.resolve({} as unknown)),
  GetDraftPerformanceMetrics: mockDrafts.getDraftPerformanceMetrics,
  ResetDraftPerformanceMetrics: mockDrafts.resetDraftPerformanceMetrics,
  GetRankProgression: mockMatches.getRankProgression,
  GetPickAlternatives: mockDrafts.getPickAlternatives,
  PredictDraftWinRate: mockDrafts.predictDraftWinRate,
  GetRankProgressionTimeline: vi.fn(() => Promise.resolve({} as unknown)),
  GetSetCards: mockCards.getSetCards,
  GetSetInfo: mockCards.getSetInfo,
  GetAllSetInfo: mockCards.getAllSetInfo,
  SearchCards: mockCards.searchCards,
  SearchCardsWithCollection: mockCards.searchCardsWithCollection,
  GetStats: mockMatches.getStats,
  GetStatsByDeck: mockMatches.getStatsByDeck,
  GetStatsByFormat: mockMatches.getStatsByFormat,
  GetTrendAnalysis: mockMatches.getTrendAnalysis,
  ImportLogs: vi.fn(() => Promise.resolve()),
  PauseReplay: vi.fn(() => Promise.resolve()),
  ResumeReplay: vi.fn(() => Promise.resolve()),
  StopReplay: vi.fn(() => Promise.resolve()),
  StartReplayWithFileDialog: vi.fn(() => Promise.resolve()),
  RefreshSetRatings: mockCards.getCardRatings,
  RefreshSetCards: vi.fn(() => Promise.resolve(0)),
  RecalculateAllDraftGrades: vi.fn(() => Promise.resolve(0)),
  GetDatasetSource: vi.fn(() => Promise.resolve('s3')),
  SetDaemonPort: vi.fn(() => Promise.resolve()),
  ReconnectToDaemon: vi.fn(() => Promise.resolve()),
  SwitchToStandaloneMode: vi.fn(() => Promise.resolve()),
  SwitchToDaemonMode: vi.fn(() => Promise.resolve()),
  ImportFromFile: vi.fn(() => Promise.resolve()),
  ImportLogFile: vi.fn(() => Promise.resolve(null)),
  TriggerReplayLogs: vi.fn(() => Promise.resolve()),
  ValidateDraftDeck: mockDecks.validateDraftDeck,
  ValidateDeckWithDialog: vi.fn(() => Promise.resolve()),
  AddCard: mockDecks.addCard,
  RemoveCard: mockDecks.removeCard,
  GetDeck: mockDecks.getDeck,
  GetDeckStatistics: mockDecks.getDeckStatistics,
  GetDeckByDraftEvent: mockDecks.getDeckByDraftEvent,
  CreateDeck: mockDecks.createDeck,
  DeleteDeck: mockDecks.deleteDeck,
  ListDecks: mockDecks.getDecks,
  GetRecommendations: vi.fn(() => Promise.resolve({} as unknown)),
  ExplainRecommendation: mockDrafts.explainRecommendation,
  ExportDeckToFile: vi.fn(() => Promise.resolve()),
  GetCollection: mockCollection.getCollection,
  GetCollectionStats: mockCollection.getCollectionStats,
  GetSetCompletion: mockCollection.getSetCompletion,
  GetRecentCollectionChanges: mockCollection.getRecentChanges,
  GetMetaDashboard: vi.fn(() => Promise.resolve({
    format: 'standard',
    archetypes: [],
    tournaments: [],
    totalArchetypes: 0,
    lastUpdated: new Date().toISOString(),
    sources: ['mtggoldfish', 'mtgtop8'],
    error: '',
  } as unknown)),
  RefreshMetaData: mockMeta.refreshMetaData,
  GetSupportedFormats: mockMeta.getFormats,
  GetTierArchetypes: mockMeta.getTierArchetypes,
  CheckOllamaStatus: mockSystem.checkOllamaStatus,
  GetAvailableOllamaModels: mockSystem.getAvailableOllamaModels,
  PullOllamaModel: mockSystem.pullOllamaModel,
  TestLLMGeneration: mockSystem.testLLMGeneration,
  GetAllSettings: mockSettings.getAllSettings,
  SaveAllSettings: mockSettings.saveAllSettings,
};

export function resetMocks() {
  // Reset all API module mocks
  Object.values(mockCards).forEach((mock) => {
    if (vi.isMockFunction(mock)) mock.mockClear();
  });
  Object.values(mockMatches).forEach((mock) => {
    if (vi.isMockFunction(mock)) mock.mockClear();
  });
  Object.values(mockDecks).forEach((mock) => {
    if (vi.isMockFunction(mock)) mock.mockClear();
  });
  Object.values(mockDrafts).forEach((mock) => {
    if (vi.isMockFunction(mock)) mock.mockClear();
  });
  Object.values(mockCollection).forEach((mock) => {
    if (vi.isMockFunction(mock)) mock.mockClear();
  });
  Object.values(mockMeta).forEach((mock) => {
    if (vi.isMockFunction(mock)) mock.mockClear();
  });
  Object.values(mockQuests).forEach((mock) => {
    if (vi.isMockFunction(mock)) mock.mockClear();
  });
  Object.values(mockSettings).forEach((mock) => {
    if (vi.isMockFunction(mock)) mock.mockClear();
  });
  Object.values(mockSystem).forEach((mock) => {
    if (vi.isMockFunction(mock)) mock.mockClear();
  });
  Object.values(mockWailsApp).forEach((mock) => {
    if (vi.isMockFunction(mock)) mock.mockClear();
  });
}
