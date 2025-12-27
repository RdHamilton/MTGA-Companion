import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useDataManagement } from './useDataManagement';

// Mock the API modules
vi.mock('@/services/api', () => ({
  matches: {
    exportMatches: vi.fn(),
  },
  system: {
    clearAllData: vi.fn(),
  },
}));

// Mock download utility
vi.mock('@/utils/download', () => ({
  downloadTextFile: vi.fn(),
}));

// Mock showToast
vi.mock('../components/ToastContainer', () => ({
  showToast: {
    show: vi.fn(),
  },
}));

import { matches, system } from '@/services/api';
import { downloadTextFile } from '@/utils/download';
import { showToast } from '../components/ToastContainer';

const mockExportMatches = vi.mocked(matches.exportMatches);
const mockClearAllData = vi.mocked(system.clearAllData);
const mockDownloadTextFile = vi.mocked(downloadTextFile);

describe('useDataManagement', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExportMatches.mockResolvedValue([]);
    mockClearAllData.mockResolvedValue(undefined);
  });

  describe('handleExportData', () => {
    it('exports to JSON when format is json', async () => {
      mockExportMatches.mockResolvedValueOnce([{ id: 1 }]);

      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleExportData('json');
      });

      expect(mockExportMatches).toHaveBeenCalledWith('json');
      expect(mockDownloadTextFile).toHaveBeenCalledWith(
        expect.any(String),
        'mtga-matches.json'
      );
    });

    it('exports to CSV when format is csv', async () => {
      mockExportMatches.mockResolvedValueOnce('csv,data');

      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleExportData('csv');
      });

      expect(mockExportMatches).toHaveBeenCalledWith('csv');
      expect(mockDownloadTextFile).toHaveBeenCalledWith(
        expect.any(String),
        'mtga-matches.csv'
      );
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

    it('shows error toast on export failure', async () => {
      mockExportMatches.mockRejectedValueOnce(new Error('Export failed'));

      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleExportData('json');
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Failed to export data'),
        'error'
      );
    });
  });

  describe('handleImportData', () => {
    // Note: Import functions are no-ops in REST API mode (require native file picker)
    it('shows success toast even for no-op import', async () => {
      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleImportData();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Successfully imported data'),
        'success'
      );
    });
  });

  describe('handleImportLogFile', () => {
    // Note: ImportLogFile is a no-op in REST API mode (requires native file picker)
    // The no-op stub returns an empty result with fileName='' but still triggers the success path
    it('shows success toast even for no-op import with empty result', async () => {
      const { result } = renderHook(() => useDataManagement());

      await act(async () => {
        await result.current.handleImportLogFile();
      });

      // The no-op returns a truthy object (even with empty fileName)
      // so a success toast is shown with zeros
      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Successfully imported'),
        'success'
      );
    });
  });

  describe('handleClearAllData', () => {
    it('calls system.clearAllData API', async () => {
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

      expect(mockClearAllData).toHaveBeenCalled();
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
      mockClearAllData.mockRejectedValueOnce(new Error('Clear failed'));

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
