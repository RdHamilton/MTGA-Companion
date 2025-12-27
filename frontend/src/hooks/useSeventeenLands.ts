import { useState, useCallback } from 'react';
import {
  FetchSetRatings,
  RefreshSetRatings,
  FetchSetCards,
  RefreshSetCards,
  RecalculateAllDraftGrades,
  ClearDatasetCache,
  GetDatasetSource,
} from '@/services/api/legacy';
import { showToast } from '../components/ToastContainer';

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

  const handleFetchSetRatings = useCallback(async () => {
    if (!setCode || setCode.trim() === '') {
      showToast.show('Please enter a set code (e.g., TLA, BLB, DSK, FDN)', 'warning');
      return;
    }

    setIsFetchingRatings(true);
    try {
      await FetchSetRatings(setCode.trim().toUpperCase(), draftFormat);

      // Check data source after fetching
      try {
        const source = await GetDatasetSource(setCode.trim().toUpperCase(), draftFormat);
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
    } catch (error) {
      showToast.show(
        `Failed to fetch 17Lands ratings: ${error}. Make sure: Set code is correct (e.g., TLA, BLB, DSK, FDN), you have internet connection, and 17Lands has data for this set.`,
        'error'
      );
    } finally {
      setIsFetchingRatings(false);
    }
  }, [setCode, draftFormat]);

  const handleRefreshSetRatings = useCallback(async () => {
    if (!setCode || setCode.trim() === '') {
      showToast.show('Please enter a set code (e.g., TLA, BLB, DSK, FDN)', 'warning');
      return;
    }

    setIsFetchingRatings(true);
    try {
      await RefreshSetRatings(setCode.trim().toUpperCase(), draftFormat);

      // Check data source after refreshing
      try {
        const source = await GetDatasetSource(setCode.trim().toUpperCase(), draftFormat);
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
          `Successfully refreshed 17Lands ratings for ${setCode.toUpperCase()} (${draftFormat}) from ${sourceLabel}!`,
          'success'
        );
      } catch {
        showToast.show(
          `Successfully refreshed 17Lands ratings for ${setCode.toUpperCase()} (${draftFormat})!`,
          'success'
        );
      }
    } catch (error) {
      showToast.show(`Failed to refresh 17Lands ratings: ${error}`, 'error');
    } finally {
      setIsFetchingRatings(false);
    }
  }, [setCode, draftFormat]);

  const handleFetchSetCards = useCallback(async () => {
    if (!setCode || setCode.trim() === '') {
      showToast.show('Please enter a set code (e.g., TLA, BLB, DSK, FDN)', 'warning');
      return;
    }

    setIsFetchingCards(true);
    try {
      const count = await FetchSetCards(setCode.trim().toUpperCase());
      showToast.show(
        `Successfully fetched ${count} cards for ${setCode.toUpperCase()} from Scryfall! Card data is now cached.`,
        'success'
      );
    } catch (error) {
      showToast.show(
        `Failed to fetch cards: ${error}. Make sure the set code is correct and you have internet connection.`,
        'error'
      );
    } finally {
      setIsFetchingCards(false);
    }
  }, [setCode]);

  const handleRefreshSetCards = useCallback(async () => {
    if (!setCode || setCode.trim() === '') {
      showToast.show('Please enter a set code (e.g., TLA, BLB, DSK, FDN)', 'warning');
      return;
    }

    setIsFetchingCards(true);
    try {
      const count = await RefreshSetCards(setCode.trim().toUpperCase());
      showToast.show(
        `Successfully refreshed ${count} cards for ${setCode.toUpperCase()} from Scryfall!`,
        'success'
      );
    } catch (error) {
      showToast.show(`Failed to refresh cards: ${error}`, 'error');
    } finally {
      setIsFetchingCards(false);
    }
  }, [setCode]);

  const handleRecalculateGrades = useCallback(async () => {
    setIsRecalculating(true);
    setRecalculateMessage('');

    try {
      const count = await RecalculateAllDraftGrades();

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
      await ClearDatasetCache();
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
