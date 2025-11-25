import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useDataManagement } from './useDataManagement';
import { mockWailsApp } from '../test/mocks/wailsApp';

// Mock showToast
vi.mock('../components/ToastContainer', () => ({
  showToast: {
    show: vi.fn(),
  },
}));

import { showToast } from '../components/ToastContainer';

describe('useDataManagement', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('handleExportData', () => {
    it('exports to JSON when format is json', async () => {
      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleExportData('json');
      });

      expect(mockWailsApp.ExportToJSON).toHaveBeenCalled();
      expect(mockWailsApp.ExportToCSV).not.toHaveBeenCalled();
    });

    it('exports to CSV when format is csv', async () => {
      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleExportData('csv');
      });

      expect(mockWailsApp.ExportToCSV).toHaveBeenCalled();
      expect(mockWailsApp.ExportToJSON).not.toHaveBeenCalled();
    });

    it('shows success toast for JSON export', async () => {
      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleExportData('json');
      });

      expect(showToast.show).toHaveBeenCalledWith(
        'Successfully exported data to JSON!',
        'success'
      );
    });

    it('shows success toast for CSV export', async () => {
      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleExportData('csv');
      });

      expect(showToast.show).toHaveBeenCalledWith(
        'Successfully exported data to CSV!',
        'success'
      );
    });

    it('shows error toast on JSON export failure', async () => {
      mockWailsApp.ExportToJSON.mockRejectedValueOnce(new Error('Export failed'));

      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleExportData('json');
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Failed to export data'),
        'error'
      );
    });

    it('shows error toast on CSV export failure', async () => {
      mockWailsApp.ExportToCSV.mockRejectedValueOnce(new Error('Export failed'));

      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleExportData('csv');
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Failed to export data'),
        'error'
      );
    });
  });

  describe('handleImportData', () => {
    it('calls ImportFromFile API', async () => {
      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleImportData();
      });

      expect(mockWailsApp.ImportFromFile).toHaveBeenCalled();
    });

    it('shows success toast on successful import', async () => {
      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleImportData();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Successfully imported data'),
        'success'
      );
    });

    it('shows error toast on import failure', async () => {
      mockWailsApp.ImportFromFile.mockRejectedValueOnce(new Error('Import failed'));

      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleImportData();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Failed to import data'),
        'error'
      );
    });
  });

  describe('handleImportLogFile', () => {
    it('calls ImportLogFile API', async () => {
      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleImportLogFile();
      });

      expect(mockWailsApp.ImportLogFile).toHaveBeenCalled();
    });

    it('does not show toast when user cancels', async () => {
      mockWailsApp.ImportLogFile.mockResolvedValueOnce(null);

      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleImportLogFile();
      });

      expect(showToast.show).not.toHaveBeenCalled();
    });

    it('shows detailed success toast on successful import', async () => {
      mockWailsApp.ImportLogFile.mockResolvedValueOnce({
        fileName: 'Player.log',
        entriesRead: 1000,
        matchesStored: 10,
        gamesStored: 25,
        decksStored: 5,
        ranksStored: 3,
        questsStored: 2,
        draftsStored: 1,
      });

      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleImportLogFile();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Successfully imported Player.log'),
        'success'
      );
      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Entries: 1000'),
        'success'
      );
      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Matches: 10'),
        'success'
      );
    });

    it('shows error toast on import failure', async () => {
      mockWailsApp.ImportLogFile.mockRejectedValueOnce(new Error('Log import failed'));

      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleImportLogFile();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Failed to import log file'),
        'error'
      );
    });
  });

  describe('handleClearAllData', () => {
    it('calls ClearAllData API', async () => {
      // Mock window.location.reload
      const reloadMock = vi.fn();
      Object.defineProperty(window, 'location', {
        value: { reload: reloadMock },
        writable: true,
      });

      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleClearAllData();
      });

      expect(mockWailsApp.ClearAllData).toHaveBeenCalled();
    });

    it('shows success toast on successful clear', async () => {
      const reloadMock = vi.fn();
      Object.defineProperty(window, 'location', {
        value: { reload: reloadMock },
        writable: true,
      });

      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleClearAllData();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        'All data has been cleared successfully!',
        'success'
      );
    });

    it('reloads the page after clearing data', async () => {
      const reloadMock = vi.fn();
      Object.defineProperty(window, 'location', {
        value: { reload: reloadMock },
        writable: true,
      });

      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleClearAllData();
      });

      expect(reloadMock).toHaveBeenCalled();
    });

    it('shows error toast on clear failure', async () => {
      mockWailsApp.ClearAllData.mockRejectedValueOnce(new Error('Clear failed'));

      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleClearAllData();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Failed to clear data'),
        'error'
      );
    });
  });
});
