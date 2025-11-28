import { useState, useEffect, useCallback } from 'react';
import { GetAllSettings, SaveAllSettings } from '../../wailsjs/go/main/App';
import type { gui } from '../../wailsjs/go/models';

interface SettingsState {
  autoRefresh: boolean;
  refreshInterval: number;
  showNotifications: boolean;
  theme: string;
  daemonPort: number;
  daemonMode: string;
  isLoading: boolean;
  isSaving: boolean;
  error: string | null;
}

interface UseSettingsReturn extends SettingsState {
  setAutoRefresh: (value: boolean) => void;
  setRefreshInterval: (value: number) => void;
  setShowNotifications: (value: boolean) => void;
  setTheme: (value: string) => void;
  saveSettings: () => Promise<boolean>;
  resetToDefaults: () => void;
  reloadSettings: () => Promise<void>;
}

const defaultSettings: Omit<SettingsState, 'isLoading' | 'isSaving' | 'error'> = {
  autoRefresh: false,
  refreshInterval: 30,
  showNotifications: true,
  theme: 'dark',
  daemonPort: 9999,
  daemonMode: 'standalone',
};

export function useSettings(): UseSettingsReturn {
  const [settings, setSettings] = useState<SettingsState>({
    ...defaultSettings,
    isLoading: true,
    isSaving: false,
    error: null,
  });

  // Load settings from backend on mount
  const loadSettings = useCallback(async () => {
    try {
      setSettings((prev) => ({ ...prev, isLoading: true, error: null }));
      const backendSettings = await GetAllSettings();
      if (backendSettings) {
        setSettings({
          autoRefresh: backendSettings.autoRefresh ?? defaultSettings.autoRefresh,
          refreshInterval: backendSettings.refreshInterval ?? defaultSettings.refreshInterval,
          showNotifications: backendSettings.showNotifications ?? defaultSettings.showNotifications,
          theme: backendSettings.theme ?? defaultSettings.theme,
          daemonPort: backendSettings.daemonPort ?? defaultSettings.daemonPort,
          daemonMode: backendSettings.daemonMode ?? defaultSettings.daemonMode,
          isLoading: false,
          isSaving: false,
          error: null,
        });
      }
    } catch (err) {
      console.error('Failed to load settings:', err);
      setSettings((prev) => ({
        ...prev,
        isLoading: false,
        error: err instanceof Error ? err.message : 'Failed to load settings',
      }));
    }
  }, []);

  useEffect(() => {
    loadSettings();
  }, [loadSettings]);

  // Individual setters
  const setAutoRefresh = useCallback((value: boolean) => {
    setSettings((prev) => ({ ...prev, autoRefresh: value }));
  }, []);

  const setRefreshInterval = useCallback((value: number) => {
    setSettings((prev) => ({ ...prev, refreshInterval: value }));
  }, []);

  const setShowNotifications = useCallback((value: boolean) => {
    setSettings((prev) => ({ ...prev, showNotifications: value }));
  }, []);

  const setTheme = useCallback((value: string) => {
    setSettings((prev) => ({ ...prev, theme: value }));
  }, []);

  // Save settings to backend
  const saveSettings = useCallback(async (): Promise<boolean> => {
    try {
      setSettings((prev) => ({ ...prev, isSaving: true, error: null }));
      const settingsToSave: gui.AppSettings = {
        autoRefresh: settings.autoRefresh,
        refreshInterval: settings.refreshInterval,
        showNotifications: settings.showNotifications,
        theme: settings.theme,
        daemonPort: settings.daemonPort,
        daemonMode: settings.daemonMode,
      };
      await SaveAllSettings(settingsToSave);
      setSettings((prev) => ({ ...prev, isSaving: false }));
      return true;
    } catch (err) {
      console.error('Failed to save settings:', err);
      setSettings((prev) => ({
        ...prev,
        isSaving: false,
        error: err instanceof Error ? err.message : 'Failed to save settings',
      }));
      return false;
    }
  }, [settings]);

  // Reset to default values
  const resetToDefaults = useCallback(() => {
    setSettings((prev) => ({
      ...prev,
      ...defaultSettings,
    }));
  }, []);

  return {
    ...settings,
    setAutoRefresh,
    setRefreshInterval,
    setShowNotifications,
    setTheme,
    saveSettings,
    resetToDefaults,
    reloadSettings: loadSettings,
  };
}
