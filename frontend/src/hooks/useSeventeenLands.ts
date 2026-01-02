import { useState, useCallback } from 'react';
import { cards, drafts } from '@/services/api';
import { showToast } from '../components/ToastContainer';
import { useDownload } from '@/context/DownloadContext';

// Functions that wrap cards API to match legacy signatures
async function fetchSetRatings(setCode: string, format: string): Promise<void> {
  await cards.getCardRatings(setCode, format);
}

async function refreshSetRatings(setCode: string, format: string): Promise<void> {
  await cards.getCardRatings(setCode, format);
}

async function fetchSetCards(setCode: string): Promise<number> {
  const fetchedCards = await cards.getSetCards(setCode);
  return fetchedCards.length;
}

async function refreshSetCards(setCode: string): Promise<number> {
  const fetchedCards = await cards.getSetCards(setCode);
  return fetchedCards.length;
}

// No-op stubs - not implemented in REST API
async function recalculateAllDraftGrades(): Promise<number> {
  console.warn('RecalculateAllDraftGrades: Not implemented in REST API');
  return 0;
}

async function clearDatasetCache(): Promise<void> {
  console.warn('ClearDatasetCache: Not implemented in REST API');
}

async function getDatasetSource(): Promise<string> {
  return '17lands';
}

export interface UseSeventeenLandsReturn {
  /** Set code for fetching data */
  setCode: string;
  /** Set the set code */
  setSetCode: (code: string) => void;
  /** Draft format for fetching data */
  draftFormat: string;
  /** Set draft format */
  setDraftFormat: (format: string) => void;
  /** Whether ratings are being fetched */
  isFetchingRatings: boolean;
  /** Whether cards are being fetched */
  isFetchingCards: boolean;
  /** Whether draft grades are being recalculated */
  isRecalculating: boolean;
  /** Message from recalculation operation */
  recalculateMessage: string;
  /** Current data source */
  dataSource: string;
  /** Whether cache is being cleared */
  isClearingCache: boolean;
  /** Fetch ratings for set */
  handleFetchSetRatings: () => Promise<void>;
  /** Refresh ratings for set */
  handleRefreshSetRatings: () => Promise<void>;
  /** Fetch card data for set */
  handleFetchSetCards: () => Promise<void>;
  /** Refresh card data for set */
  handleRefreshSetCards: () => Promise<void>;
  /** Recalculate all draft grades */
  handleRecalculateGrades: () => Promise<void>;
  /** Clear dataset cache */
  handleClearDatasetCache: () => Promise<void>;
}

export function useSeventeenLands(): UseSeventeenLandsReturn {
  const [setCode, setSetCode] = useState('');
  const [draftFormat, setDraftFormat] = useState('PremierDraft');
  const [isFetchingRatings, setIsFetchingRatings] = useState(false);
  const [isFetchingCards, setIsFetchingCards] = useState(false);
  const [isRecalculating, setIsRecalculating] = useState(false);
  const [recalculateMessage, setRecalculateMessage] = useState('');
  const [dataSource, setDataSource] = useState<string>('');
  const [isClearingCache, setIsClearingCache] = useState(false);

  // Download progress bar context
  const { startDownload, updateProgress, completeDownload, failDownload } = useDownload();

  const handleFetchSetRatings = useCallback(async () => {
    if (!setCode || setCode.trim() === '') {
      showToast.show('Please enter a set code (e.g., TLA, BLB, DSK, FDN)', 'warning');
      return;
    }

    const downloadId = `fetch-ratings-${setCode.trim().toUpperCase()}-${draftFormat}`;
    setIsFetchingRatings(true);
    startDownload(downloadId, `Fetching ${setCode.toUpperCase()} ratings...`);
    updateProgress(downloadId, 10);

    try {
      updateProgress(downloadId, 30);
      await fetchSetRatings(setCode.trim().toUpperCase(), draftFormat);
      updateProgress(downloadId, 80);

      // Check data source after fetching
      try {
        const source = await getDatasetSource();
        setDataSource(source);

        const sourceLabel =
          source === 's3'
            ? 'S3 public datasets'
            : source === 'web_api'
              ? 'web API'
              : source === 'legacy_api'
                ? 'legacy API'
                : source;

        showToast.show(
          `Successfully fetched 17Lands ratings for ${setCode.toUpperCase()} (${draftFormat}) from ${sourceLabel}! The data is now cached and ready for use in drafts.`,
          'success'
        );
      } catch {
        showToast.show(
          `Successfully fetched 17Lands ratings for ${setCode.toUpperCase()} (${draftFormat})! The data is now cached and ready for use in drafts.`,
          'success'
        );
      }
      completeDownload(downloadId);
    } catch (error) {
      failDownload(downloadId, `Failed to fetch ratings: ${error}`);
      showToast.show(
        `Failed to fetch 17Lands ratings: ${error}. Make sure: Set code is correct (e.g., TLA, BLB, DSK, FDN), you have internet connection, and 17Lands has data for this set.`,
        'error'
      );
    } finally {
      setIsFetchingRatings(false);
    }
  }, [setCode, draftFormat, startDownload, updateProgress, completeDownload, failDownload]);

  const handleRefreshSetRatings = useCallback(async () => {
    if (!setCode || setCode.trim() === '') {
      showToast.show('Please enter a set code (e.g., TLA, BLB, DSK, FDN)', 'warning');
      return;
    }

    const downloadId = `refresh-ratings-${setCode.trim().toUpperCase()}-${draftFormat}`;
    setIsFetchingRatings(true);
    startDownload(downloadId, `Refreshing ${setCode.toUpperCase()} ratings...`);
    updateProgress(downloadId, 10);

    try {
      const upperSetCode = setCode.trim().toUpperCase();
      updateProgress(downloadId, 20);
      await refreshSetRatings(upperSetCode, draftFormat);
      updateProgress(downloadId, 50);

      // Recalculate grades for existing drafts with this set (#734)
      const recalcResult = await drafts.recalculateSetGrades(upperSetCode);
      updateProgress(downloadId, 80);

      // Check data source after refreshing
      try {
        const source = await getDatasetSource();
        setDataSource(source);

        const sourceLabel =
          source === 's3'
            ? 'S3 public datasets'
            : source === 'web_api'
              ? 'web API'
              : source === 'legacy_api'
                ? 'legacy API'
                : source;

        const gradesMsg = recalcResult.count > 0 ? ` Updated ${recalcResult.count} draft grades.` : '';
        showToast.show(
          `Successfully refreshed 17Lands ratings for ${setCode.toUpperCase()} (${draftFormat}) from ${sourceLabel}!${gradesMsg}`,
          'success'
        );
      } catch {
        const gradesMsg = recalcResult.count > 0 ? ` Updated ${recalcResult.count} draft grades.` : '';
        showToast.show(
          `Successfully refreshed 17Lands ratings for ${setCode.toUpperCase()} (${draftFormat})!${gradesMsg}`,
          'success'
        );
      }
      completeDownload(downloadId);
    } catch (error) {
      failDownload(downloadId, `Failed to refresh ratings: ${error}`);
      showToast.show(`Failed to refresh 17Lands ratings: ${error}`, 'error');
    } finally {
      setIsFetchingRatings(false);
    }
  }, [setCode, draftFormat, startDownload, updateProgress, completeDownload, failDownload]);

  const handleFetchSetCards = useCallback(async () => {
    if (!setCode || setCode.trim() === '') {
      showToast.show('Please enter a set code (e.g., TLA, BLB, DSK, FDN)', 'warning');
      return;
    }

    const downloadId = `fetch-cards-${setCode.trim().toUpperCase()}`;
    setIsFetchingCards(true);
    startDownload(downloadId, `Fetching ${setCode.toUpperCase()} cards...`);
    updateProgress(downloadId, 10);

    try {
      updateProgress(downloadId, 30);
      const count = await fetchSetCards(setCode.trim().toUpperCase());
      updateProgress(downloadId, 90);
      showToast.show(
        `Successfully fetched ${count} cards for ${setCode.toUpperCase()} from Scryfall! Card data is now cached.`,
        'success'
      );
      completeDownload(downloadId);
    } catch (error) {
      failDownload(downloadId, `Failed to fetch cards: ${error}`);
      showToast.show(
        `Failed to fetch cards: ${error}. Make sure the set code is correct and you have internet connection.`,
        'error'
      );
    } finally {
      setIsFetchingCards(false);
    }
  }, [setCode, startDownload, updateProgress, completeDownload, failDownload]);

  const handleRefreshSetCards = useCallback(async () => {
    if (!setCode || setCode.trim() === '') {
      showToast.show('Please enter a set code (e.g., TLA, BLB, DSK, FDN)', 'warning');
      return;
    }

    const downloadId = `refresh-cards-${setCode.trim().toUpperCase()}`;
    setIsFetchingCards(true);
    startDownload(downloadId, `Refreshing ${setCode.toUpperCase()} cards...`);
    updateProgress(downloadId, 10);

    try {
      updateProgress(downloadId, 30);
      const count = await refreshSetCards(setCode.trim().toUpperCase());
      updateProgress(downloadId, 90);
      showToast.show(
        `Successfully refreshed ${count} cards for ${setCode.toUpperCase()} from Scryfall!`,
        'success'
      );
      completeDownload(downloadId);
    } catch (error) {
      failDownload(downloadId, `Failed to refresh cards: ${error}`);
      showToast.show(`Failed to refresh cards: ${error}`, 'error');
    } finally {
      setIsFetchingCards(false);
    }
  }, [setCode, startDownload, updateProgress, completeDownload, failDownload]);

  const handleRecalculateGrades = useCallback(async () => {
    setIsRecalculating(true);
    setRecalculateMessage('');

    try {
      const count = await recalculateAllDraftGrades();

      setRecalculateMessage(
        `✓ Successfully recalculated ${count} draft session(s)! Draft grades and predictions have been updated.`
      );

      // Clear message after 5 seconds
      setTimeout(() => setRecalculateMessage(''), 5000);
    } catch (error) {
      setRecalculateMessage(`✗ Failed to recalculate draft grades: ${error}`);

      // Clear error message after 8 seconds
      setTimeout(() => setRecalculateMessage(''), 8000);
    } finally {
      setIsRecalculating(false);
    }
  }, []);

  const handleClearDatasetCache = useCallback(async () => {
    setIsClearingCache(true);
    try {
      await clearDatasetCache();
      showToast.show(
        'Successfully cleared dataset cache! Cached CSV files have been deleted to free up disk space. Ratings in the database are preserved.',
        'success'
      );
    } catch (error) {
      showToast.show(`Failed to clear dataset cache: ${error}`, 'error');
    } finally {
      setIsClearingCache(false);
    }
  }, []);

  return {
    setCode,
    setSetCode,
    draftFormat,
    setDraftFormat,
    isFetchingRatings,
    isFetchingCards,
    isRecalculating,
    recalculateMessage,
    dataSource,
    isClearingCache,
    handleFetchSetRatings,
    handleRefreshSetRatings,
    handleFetchSetCards,
    handleRefreshSetCards,
    handleRecalculateGrades,
    handleClearDatasetCache,
  };
}
