import { useState, useEffect, useCallback } from 'react';
import { settings as settingsApi } from '@/services/api';
import type { gui } from '@/types/models';

interface SettingsState {
  autoRefresh: boolean;
  refreshInterval: number;
  showNotifications: boolean;
  theme: string;
  daemonPort: number;
  daemonMode: string;
  // ML/LLM Settings
  mlEnabled: boolean;
  llmEnabled: boolean;
  ollamaEndpoint: string;
  ollamaModel: string;
  metaGoldfishEnabled: boolean;
  metaTop8Enabled: boolean;
  metaWeight: number;
  personalWeight: number;
  // Rotation Settings
  rotationNotificationsEnabled: boolean;
  rotationNotificationThreshold: number; // Days before rotation to notify
  // State
  isLoading: boolean;
  isSaving: boolean;
  error: string | null;
}

interface UseSettingsReturn extends SettingsState {
  setAutoRefresh: (value: boolean) => void;
  setRefreshInterval: (value: number) => void;
  setShowNotifications: (value: boolean) => void;
  setTheme: (value: string) => void;
  // ML/LLM Settings setters
  setMLEnabled: (value: boolean) => void;
  setLLMEnabled: (value: boolean) => void;
  setOllamaEndpoint: (value: string) => void;
  setOllamaModel: (value: string) => void;
  setMetaGoldfishEnabled: (value: boolean) => void;
  setMetaTop8Enabled: (value: boolean) => void;
  setMetaWeight: (value: number) => void;
  setPersonalWeight: (value: number) => void;
  // Rotation Settings setters
  setRotationNotificationsEnabled: (value: boolean) => void;
  setRotationNotificationThreshold: (value: number) => void;
  // Actions
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
  // ML/LLM defaults
  mlEnabled: true,
  llmEnabled: false,
  ollamaEndpoint: 'http://localhost:11434',
  ollamaModel: 'qwen3:8b',
  metaGoldfishEnabled: true,
  metaTop8Enabled: true,
  metaWeight: 0.3,
  personalWeight: 0.2,
  // Rotation defaults
  rotationNotificationsEnabled: true,
  rotationNotificationThreshold: 30, // Notify 30 days before rotation
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
      const backendSettings = await settingsApi.getSettings();
      if (backendSettings) {
        setSettings({
          autoRefresh: backendSettings.autoRefresh ?? defaultSettings.autoRefresh,
          refreshInterval: backendSettings.refreshInterval ?? defaultSettings.refreshInterval,
          showNotifications: backendSettings.showNotifications ?? defaultSettings.showNotifications,
          theme: backendSettings.theme ?? defaultSettings.theme,
          daemonPort: backendSettings.daemonPort ?? defaultSettings.daemonPort,
          daemonMode: backendSettings.daemonMode ?? defaultSettings.daemonMode,
          // ML/LLM settings
          mlEnabled: backendSettings.mlEnabled ?? defaultSettings.mlEnabled,
          llmEnabled: backendSettings.llmEnabled ?? defaultSettings.llmEnabled,
          ollamaEndpoint: backendSettings.ollamaEndpoint ?? defaultSettings.ollamaEndpoint,
          ollamaModel: backendSettings.ollamaModel ?? defaultSettings.ollamaModel,
          metaGoldfishEnabled: backendSettings.metaGoldfishEnabled ?? defaultSettings.metaGoldfishEnabled,
          metaTop8Enabled: backendSettings.metaTop8Enabled ?? defaultSettings.metaTop8Enabled,
          metaWeight: backendSettings.metaWeight ?? defaultSettings.metaWeight,
          personalWeight: backendSettings.personalWeight ?? defaultSettings.personalWeight,
          // Rotation settings
          rotationNotificationsEnabled:
            backendSettings.rotationNotificationsEnabled ?? defaultSettings.rotationNotificationsEnabled,
          rotationNotificationThreshold:
            backendSettings.rotationNotificationThreshold ?? defaultSettings.rotationNotificationThreshold,
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

  // ML/LLM setters
  const setMLEnabled = useCallback((value: boolean) => {
    setSettings((prev) => ({ ...prev, mlEnabled: value }));
  }, []);

  const setLLMEnabled = useCallback((value: boolean) => {
    setSettings((prev) => ({ ...prev, llmEnabled: value }));
  }, []);

  const setOllamaEndpoint = useCallback((value: string) => {
    setSettings((prev) => ({ ...prev, ollamaEndpoint: value }));
  }, []);

  const setOllamaModel = useCallback((value: string) => {
    setSettings((prev) => ({ ...prev, ollamaModel: value }));
  }, []);

  const setMetaGoldfishEnabled = useCallback((value: boolean) => {
    setSettings((prev) => ({ ...prev, metaGoldfishEnabled: value }));
  }, []);

  const setMetaTop8Enabled = useCallback((value: boolean) => {
    setSettings((prev) => ({ ...prev, metaTop8Enabled: value }));
  }, []);

  const setMetaWeight = useCallback((value: number) => {
    setSettings((prev) => ({ ...prev, metaWeight: value }));
  }, []);

  const setPersonalWeight = useCallback((value: number) => {
    setSettings((prev) => ({ ...prev, personalWeight: value }));
  }, []);

  // Rotation setters
  const setRotationNotificationsEnabled = useCallback((value: boolean) => {
    setSettings((prev) => ({ ...prev, rotationNotificationsEnabled: value }));
  }, []);

  const setRotationNotificationThreshold = useCallback((value: number) => {
    setSettings((prev) => ({ ...prev, rotationNotificationThreshold: value }));
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
        // ML/LLM settings
        mlEnabled: settings.mlEnabled,
        llmEnabled: settings.llmEnabled,
        ollamaEndpoint: settings.ollamaEndpoint,
        ollamaModel: settings.ollamaModel,
        metaGoldfishEnabled: settings.metaGoldfishEnabled,
        metaTop8Enabled: settings.metaTop8Enabled,
        metaWeight: settings.metaWeight,
        personalWeight: settings.personalWeight,
        // Rotation settings
        rotationNotificationsEnabled: settings.rotationNotificationsEnabled,
        rotationNotificationThreshold: settings.rotationNotificationThreshold,
      };
      await settingsApi.updateSettings(settingsToSave);
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
    // ML/LLM setters
    setMLEnabled,
    setLLMEnabled,
    setOllamaEndpoint,
    setOllamaModel,
    setMetaGoldfishEnabled,
    setMetaTop8Enabled,
    setMetaWeight,
    setPersonalWeight,
    // Rotation setters
    setRotationNotificationsEnabled,
    setRotationNotificationThreshold,
    // Actions
    saveSettings,
    resetToDefaults,
    reloadSettings: loadSettings,
  };
}
