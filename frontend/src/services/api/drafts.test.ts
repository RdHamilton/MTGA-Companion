import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import {
  exportDraftTo17Lands,
  getExportableDrafts,
  ExportDraftTo17LandsResponse,
  SeventeenLandsDraftExport,
} from './drafts';
import type { models } from '@/types/models';

const API_BASE = 'http://localhost:8080/api/v1';

// Helper to create success response
function successResponse<T>(data: T) {
  return HttpResponse.json({ data });
}

const mockExportResponse: ExportDraftTo17LandsResponse = {
  session_id: 'test-session-123',
  file_name: 'draft_TLA_2024-01-15_14-30-00.json',
  export: {
    draft_id: 'test-session-123',
    event_type: 'QuickDraft',
    set_code: 'TLA',
    draft_time: '2024-01-15T14:30:00Z',
    picks: [
      {
        pack_number: 1,
        pick_number: 1,
        pack: [12345, 12346, 12347],
        pick: 12345,
        pick_time: '2024-01-15T14:31:00Z',
      },
    ],
    metadata: {
      exported_at: '2024-01-15T15:00:00Z',
      exported_from: 'MTGA-Companion',
      overall_grade: 'B+',
      overall_score: 78,
      predicted_win_rate: 0.55,
    },
  },
};

const mockExportableDrafts: Partial<models.DraftSession>[] = [
  {
    ID: 'test-session-123',
    SetCode: 'TLA',
    DraftType: 'QuickDraft',
    EventName: 'QuickDraft_TLA',
    Status: 'completed',
    TotalPicks: 45,
  },
  {
    ID: 'test-session-456',
    SetCode: 'DSK',
    DraftType: 'PremierDraft',
    EventName: 'PremierDraft_DSK',
    Status: 'completed',
    TotalPicks: 45,
  },
];

// MSW Server setup
const server = setupServer(
  http.get(`${API_BASE}/drafts/:sessionID/export/17lands`, ({ params }) => {
    const sessionID = params.sessionID as string;
    if (sessionID === 'not-found') {
      return HttpResponse.json(
        { error: 'Not Found', message: 'Draft session not found', code: 404 },
        { status: 404 }
      );
    }
    return successResponse({
      ...mockExportResponse,
      session_id: sessionID,
      export: { ...mockExportResponse.export, draft_id: sessionID },
    });
  }),

  http.get(`${API_BASE}/drafts/exportable`, ({ request }) => {
    const url = new URL(request.url);
    const limit = url.searchParams.get('limit');
    let drafts = mockExportableDrafts;
    if (limit) {
      drafts = drafts.slice(0, parseInt(limit, 10));
    }
    return successResponse(drafts);
  })
);

beforeEach(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => {
  server.resetHandlers();
  server.close();
});

describe('Draft Export API', () => {
  describe('exportDraftTo17Lands', () => {
    it('should export a draft session to 17Lands format', async () => {
      const result = await exportDraftTo17Lands('test-session-123');

      expect(result.session_id).toBe('test-session-123');
      expect(result.file_name).toBe('draft_TLA_2024-01-15_14-30-00.json');
      expect(result.export.draft_id).toBe('test-session-123');
      expect(result.export.event_type).toBe('QuickDraft');
      expect(result.export.set_code).toBe('TLA');
    });

    it('should include picks in the export', async () => {
      const result = await exportDraftTo17Lands('test-session-123');

      expect(result.export.picks).toBeDefined();
      expect(result.export.picks.length).toBeGreaterThan(0);
      expect(result.export.picks[0].pack_number).toBe(1);
      expect(result.export.picks[0].pick_number).toBe(1);
      expect(result.export.picks[0].pick).toBe(12345);
    });

    it('should include metadata in the export', async () => {
      const result = await exportDraftTo17Lands('test-session-123');

      expect(result.export.metadata).toBeDefined();
      expect(result.export.metadata?.exported_from).toBe('MTGA-Companion');
      expect(result.export.metadata?.overall_grade).toBe('B+');
      expect(result.export.metadata?.overall_score).toBe(78);
      expect(result.export.metadata?.predicted_win_rate).toBe(0.55);
    });

    it('should handle not found error', async () => {
      await expect(exportDraftTo17Lands('not-found')).rejects.toThrow();
    });
  });

  describe('getExportableDrafts', () => {
    it('should get list of exportable drafts', async () => {
      const result = await getExportableDrafts();

      expect(result).toHaveLength(2);
      expect(result[0].ID).toBe('test-session-123');
      expect(result[0].SetCode).toBe('TLA');
      expect(result[1].ID).toBe('test-session-456');
      expect(result[1].SetCode).toBe('DSK');
    });

    it('should respect limit parameter', async () => {
      const result = await getExportableDrafts(1);

      expect(result).toHaveLength(1);
      expect(result[0].ID).toBe('test-session-123');
    });

    it('should return drafts with status completed', async () => {
      const result = await getExportableDrafts();

      result.forEach((draft) => {
        expect(draft.Status).toBe('completed');
      });
    });
  });
});

describe('SeventeenLandsDraftExport types', () => {
  it('should have correct structure for 17Lands format', () => {
    const exportData: SeventeenLandsDraftExport = {
      draft_id: 'test-123',
      event_type: 'QuickDraft',
      set_code: 'TLA',
      draft_time: '2024-01-15T14:30:00Z',
      picks: [
        {
          pack_number: 1,
          pick_number: 1,
          pack: [12345, 12346],
          pick: 12345,
          pick_time: '2024-01-15T14:31:00Z',
        },
      ],
    };

    expect(exportData.draft_id).toBe('test-123');
    expect(exportData.picks[0].pack_number).toBe(1);
  });

  it('should support optional fields', () => {
    const exportData: SeventeenLandsDraftExport = {
      draft_id: 'test-123',
      event_type: 'QuickDraft',
      set_code: 'TLA',
      draft_time: '2024-01-15T14:30:00Z',
      picks: [],
      final_deck: [12345, 12346],
      sideboard: [12347],
      metadata: {
        exported_at: '2024-01-15T15:00:00Z',
        exported_from: 'MTGA-Companion',
        overall_grade: 'A',
        overall_score: 90,
        predicted_win_rate: 0.65,
      },
    };

    expect(exportData.final_deck).toHaveLength(2);
    expect(exportData.sideboard).toHaveLength(1);
    expect(exportData.metadata?.overall_grade).toBe('A');
  });
});
