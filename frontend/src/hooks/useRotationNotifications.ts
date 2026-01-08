import { useState, useEffect, useCallback } from 'react';
import * as standardApi from '@/services/api/standard';
import type { RotationAffectedDeck, UpcomingRotation } from '@/services/api/standard';

export interface RotationNotificationState {
  rotation: UpcomingRotation | null;
  affectedDecks: RotationAffectedDeck[];
  isLoading: boolean;
  error: string | null;
  lastChecked: Date | null;
  hasNotified: boolean;
}

export interface UseRotationNotificationsReturn extends RotationNotificationState {
  checkRotation: () => Promise<void>;
  markAsNotified: () => void;
  shouldShowNotification: (thresholdDays: number) => boolean;
  getUrgencyLevel: () => 'critical' | 'warning' | 'info' | null;
}

const STORAGE_KEY = 'rotation_notification_last_shown';

export function useRotationNotifications(): UseRotationNotificationsReturn {
  const [state, setState] = useState<RotationNotificationState>({
    rotation: null,
    affectedDecks: [],
    isLoading: false,
    error: null,
    lastChecked: null,
    hasNotified: false,
  });

  // Check if we've already shown notification today
  const checkIfAlreadyNotified = useCallback((): boolean => {
    const lastShown = localStorage.getItem(STORAGE_KEY);
    if (!lastShown) return false;

    const lastShownDate = new Date(lastShown);
    const today = new Date();
    return lastShownDate.toDateString() === today.toDateString();
  }, []);

  // Fetch rotation data from API
  const checkRotation = useCallback(async () => {
    setState((prev) => ({ ...prev, isLoading: true, error: null }));

    try {
      const [rotation, affectedDecks] = await Promise.all([
        standardApi.getUpcomingRotation(),
        standardApi.getRotationAffectedDecks(),
      ]);

      setState({
        rotation,
        affectedDecks,
        isLoading: false,
        error: null,
        lastChecked: new Date(),
        hasNotified: checkIfAlreadyNotified(),
      });
    } catch (err) {
      console.error('Failed to check rotation:', err);
      setState((prev) => ({
        ...prev,
        isLoading: false,
        error: err instanceof Error ? err.message : 'Failed to check rotation',
      }));
    }
  }, [checkIfAlreadyNotified]);

  // Mark notification as shown (won't show again today)
  const markAsNotified = useCallback(() => {
    localStorage.setItem(STORAGE_KEY, new Date().toISOString());
    setState((prev) => ({ ...prev, hasNotified: true }));
  }, []);

  // Check if we should show a notification based on threshold
  const shouldShowNotification = useCallback(
    (thresholdDays: number): boolean => {
      if (!state.rotation) return false;
      if (state.hasNotified) return false;
      if (state.affectedDecks.length === 0) return false;

      return state.rotation.daysUntilRotation <= thresholdDays;
    },
    [state.rotation, state.hasNotified, state.affectedDecks]
  );

  // Get urgency level based on days until rotation
  const getUrgencyLevel = useCallback((): 'critical' | 'warning' | 'info' | null => {
    if (!state.rotation || state.affectedDecks.length === 0) return null;

    const days = state.rotation.daysUntilRotation;

    if (days <= 7) return 'critical';
    if (days <= 30) return 'warning';
    if (days <= 90) return 'info';

    return null;
  }, [state.rotation, state.affectedDecks]);

  // Check rotation on mount
  useEffect(() => {
    checkRotation();
  }, [checkRotation]);

  return {
    ...state,
    checkRotation,
    markAsNotified,
    shouldShowNotification,
    getUrgencyLevel,
  };
}
