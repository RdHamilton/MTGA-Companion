import { useCallback } from 'react';
import { matches, system } from '@/services/api';
import { downloadTextFile } from '@/utils/download';
import { showToast } from '../components/ToastContainer';
import { gui } from '@/types/models';

// Export functions
async function exportToJSON(): Promise<void> {
  const data = await matches.exportMatches('json');
  downloadTextFile(JSON.stringify(data, null, 2), 'mtga-matches.json');
}

async function exportToCSV(): Promise<void> {
  const data = await matches.exportMatches('csv');
  downloadTextFile(String(data), 'mtga-matches.csv');
}

// No-op stubs - file picker requires native integration
async function importFromFile(): Promise<void> {
  console.warn('ImportFromFile requires file picker - use browser file input');
}

async function importLogFile(): Promise<gui.ImportLogFileResult> {
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

export interface UseDataManagementReturn {
  /** Export data to JSON or CSV */
  handleExportData: (format: 'json' | 'csv') => Promise<void>;
  /** Import data from JSON file */
  handleImportData: () => Promise<void>;
  /** Import single log file */
  handleImportLogFile: () => Promise<void>;
  /** Clear all data */
  handleClearAllData: () => Promise<void>;
}

export function useDataManagement(): UseDataManagementReturn {
  const handleExportData = useCallback(async (format: 'json' | 'csv') => {
    try {
      if (format === 'json') {
        await exportToJSON();
      } else {
        await exportToCSV();
      }
      showToast.show(`Successfully exported data to ${format.toUpperCase()}!`, 'success');
    } catch (error) {
      showToast.show(`Failed to export data: ${error}`, 'error');
    }
  }, []);

  const handleImportData = useCallback(async () => {
    try {
      await importFromFile();
      showToast.show(
        'Successfully imported data! Refresh the page to see updated statistics.',
        'success'
      );
    } catch (error) {
      showToast.show(`Failed to import data: ${error}`, 'error');
    }
  }, []);

  const handleImportLogFile = useCallback(async () => {
    try {
      const result = await importLogFile();

      // User cancelled
      if (!result) {
        return;
      }

      // Show success message with detailed results
      showToast.show(
        `Successfully imported ${result.fileName}! ` +
          `Entries: ${result.entriesRead}, ` +
          `Matches: ${result.matchesStored}, ` +
          `Games: ${result.gamesStored}, ` +
          `Decks: ${result.decksStored}, ` +
          `Ranks: ${result.ranksStored}, ` +
          `Quests: ${result.questsStored}, ` +
          `Drafts: ${result.draftsStored}. ` +
          `Refresh to see updated statistics.`,
        'success'
      );
    } catch (error) {
      showToast.show(`Failed to import log file: ${error}`, 'error');
    }
  }, []);

  const handleClearAllData = useCallback(async () => {
    try {
      await system.clearAllData();
      showToast.show('All data has been cleared successfully!', 'success');
      window.location.reload(); // Refresh to show empty state
    } catch (error) {
      showToast.show(`Failed to clear data: ${error}`, 'error');
    }
  }, []);

  return {
    handleExportData,
    handleImportData,
    handleImportLogFile,
    handleClearAllData,
  };
}
