import { useCallback } from 'react';
import {
  ExportToJSON,
  ExportToCSV,
  ImportFromFile,
  ImportLogFile,
  ClearAllData,
} from '../../wailsjs/go/main/App';
import { showToast } from '../components/ToastContainer';

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
        await ExportToJSON();
      } else {
        await ExportToCSV();
      }
      showToast.show(`Successfully exported data to ${format.toUpperCase()}!`, 'success');
    } catch (error) {
      showToast.show(`Failed to export data: ${error}`, 'error');
    }
  }, []);

  const handleImportData = useCallback(async () => {
    try {
      await ImportFromFile();
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
      const result = await ImportLogFile();

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
      await ClearAllData();
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
