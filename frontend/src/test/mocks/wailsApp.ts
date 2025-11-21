import { vi } from 'vitest';

export const mockWailsApp = {
  AnalyzeSessionPickQuality: vi.fn(() => Promise.resolve()),
  CalculateDraftGrade: vi.fn(() => Promise.resolve({})),
  ClearAllData: vi.fn(() => Promise.resolve()),
  ClearDatasetCache: vi.fn(() => Promise.resolve()),
  ExportToCSV: vi.fn(() => Promise.resolve()),
  ExportToJSON: vi.fn(() => Promise.resolve()),
  FetchSetCards: vi.fn(() => Promise.resolve(0)),
  FetchSetRatings: vi.fn(() => Promise.resolve()),
  FixDraftSessionStatuses: vi.fn(() => Promise.resolve(0)),
  GetActiveDraftSessions: vi.fn(() => Promise.resolve([])),
  GetActiveEvents: vi.fn(() => Promise.resolve([])),
  GetActiveQuests: vi.fn(() => Promise.resolve([])),
  GetArchetypeCards: vi.fn(() => Promise.resolve({})),
  GetCardByArenaID: vi.fn(() => Promise.resolve({})),
  GetCardRatingByArenaID: vi.fn(() => Promise.resolve({})),
  GetCardRatings: vi.fn(() => Promise.resolve([])),
  GetColorRatings: vi.fn(() => Promise.resolve([])),
  GetCompletedDraftSessions: vi.fn(() => Promise.resolve([])),
  GetConnectionStatus: vi.fn(() => Promise.resolve({})),
  GetDeckDetails: vi.fn(() => Promise.resolve({})),
  GetDraftGrade: vi.fn(() => Promise.resolve({})),
  GetDraftPacks: vi.fn(() => Promise.resolve([])),
  GetDraftPicks: vi.fn(() => Promise.resolve([])),
  GetEventWinDistribution: vi.fn(() => Promise.resolve([])),
  GetFormatArchetypes: vi.fn(() => Promise.resolve([])),
  GetFormatInsights: vi.fn(() => Promise.resolve({})),
  GetFormatStats: vi.fn(() => Promise.resolve({})),
  GetMatches: vi.fn(() => Promise.resolve([])),
  GetMatchesBySessionID: vi.fn(() => Promise.resolve([])),
  GetPerformanceMetrics: vi.fn(() => Promise.resolve([])),
  GetPickAlternatives: vi.fn(() => Promise.resolve([])),
  GetRankProgressionTimeline: vi.fn(() => Promise.resolve([])),
  GetSetCards: vi.fn(() => Promise.resolve([])),
  GetStats: vi.fn(() => Promise.resolve({})),
  GetStatsByDeck: vi.fn(() => Promise.resolve([])),
  GetStatsByFormat: vi.fn(() => Promise.resolve([])),
  GetTrendAnalysis: vi.fn(() => Promise.resolve({})),
  ImportLogs: vi.fn(() => Promise.resolve()),
  PauseReplay: vi.fn(() => Promise.resolve()),
  ResumeReplay: vi.fn(() => Promise.resolve()),
  StopReplay: vi.fn(() => Promise.resolve()),
};

export function resetMocks() {
  Object.values(mockWailsApp).forEach((mock) => {
    if (vi.isMockFunction(mock)) {
      mock.mockClear();
    }
  });
}
