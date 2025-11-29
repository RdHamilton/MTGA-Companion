import { useState, useMemo } from 'react';
import AboutDialog from '../components/AboutDialog';
import {
  DaemonConnectionSection,
  AppPreferencesSection,
  ImportExportSection,
  DataRecoverySection,
  ReplayToolSection,
  SeventeenLandsSection,
  MLSettingsSection,
  AboutSection,
} from '../components/settings/sections';
import { SettingsAccordion } from '../components/settings/SettingsAccordion';
import type { SettingsAccordionItem } from '../components/settings/SettingsAccordion';
import {
  useDaemonConnection,
  useLogReplay,
  useReplayTool,
  useSeventeenLands,
  useDataManagement,
  useDeveloperMode,
  useSettings,
} from '../hooks';
import './Settings.css';

const Settings = () => {
  // Local UI state
  const [saved, setSaved] = useState(false);
  const [showAbout, setShowAbout] = useState(false);

  // Settings from backend
  const {
    autoRefresh,
    refreshInterval,
    showNotifications,
    theme,
    // ML/LLM Settings
    mlEnabled,
    llmEnabled,
    ollamaEndpoint,
    ollamaModel,
    metaGoldfishEnabled,
    metaTop8Enabled,
    metaWeight,
    personalWeight,
    // State
    isLoading: isLoadingSettings,
    isSaving,
    error: settingsError,
    // Setters
    setAutoRefresh,
    setRefreshInterval,
    setShowNotifications,
    setTheme,
    setMLEnabled,
    setLLMEnabled,
    setOllamaEndpoint,
    setOllamaModel,
    setMetaGoldfishEnabled,
    setMetaTop8Enabled,
    setMetaWeight,
    setPersonalWeight,
    saveSettings,
    resetToDefaults,
  } = useSettings();

  // Developer mode hook
  const {
    isDeveloperMode,
    handleVersionClick,
    toggleDeveloperMode,
  } = useDeveloperMode();

  // Custom hooks for state management
  const {
    connectionStatus,
    daemonMode,
    daemonPort,
    isReconnecting,
    handleDaemonPortChange,
    handleReconnect,
    handleModeChange,
  } = useDaemonConnection();

  const {
    clearDataBeforeReplay,
    setClearDataBeforeReplay,
    isReplaying,
    replayProgress,
    handleReplayLogs,
  } = useLogReplay();

  const {
    replayToolActive,
    replayToolPaused,
    replayToolProgress,
    replaySpeed,
    setReplaySpeed,
    replayFilter,
    setReplayFilter,
    pauseOnDraft,
    setPauseOnDraft,
    handleStartReplayTool,
    handlePauseReplayTool,
    handleResumeReplayTool,
    handleStopReplayTool,
  } = useReplayTool();

  const {
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
  } = useSeventeenLands();

  const {
    handleExportData,
    handleImportData,
    handleImportLogFile,
    handleClearAllData,
  } = useDataManagement();

  // Derived state
  const isConnected = connectionStatus.status === 'connected';

  // Local handlers
  const handleSave = async () => {
    const success = await saveSettings();
    if (success) {
      setSaved(true);
      setTimeout(() => setSaved(false), 3000);
    }
  };

  const handleReset = () => {
    resetToDefaults();
  };

  // Build accordion items
  const accordionItems: SettingsAccordionItem[] = useMemo(() => {
    const items: SettingsAccordionItem[] = [
      {
        id: 'connection',
        label: 'Connection',
        icon: '🔌',
        content: (
          <DaemonConnectionSection
            connectionStatus={connectionStatus}
            daemonMode={daemonMode}
            daemonPort={daemonPort}
            isReconnecting={isReconnecting}
            onDaemonPortChange={handleDaemonPortChange}
            onReconnect={handleReconnect}
            onModeChange={handleModeChange}
          />
        ),
      },
      {
        id: 'preferences',
        label: 'Preferences',
        icon: '⚙️',
        content: (
          <AppPreferencesSection
            autoRefresh={autoRefresh}
            onAutoRefreshChange={setAutoRefresh}
            refreshInterval={refreshInterval}
            onRefreshIntervalChange={setRefreshInterval}
            showNotifications={showNotifications}
            onShowNotificationsChange={setShowNotifications}
            theme={theme}
            onThemeChange={setTheme}
          />
        ),
      },
      {
        id: 'import-export',
        label: 'Import / Export',
        icon: '📦',
        content: (
          <ImportExportSection
            onExportData={handleExportData}
            onImportData={handleImportData}
          />
        ),
      },
      {
        id: 'data-recovery',
        label: 'Data Recovery',
        icon: '🔄',
        content: (
          <DataRecoverySection
            isConnected={isConnected}
            clearDataBeforeReplay={clearDataBeforeReplay}
            onClearDataBeforeReplayChange={setClearDataBeforeReplay}
            isReplaying={isReplaying}
            replayProgress={replayProgress}
            onImportLogFile={handleImportLogFile}
            onReplayLogs={() => handleReplayLogs(isConnected)}
            onClearAllData={handleClearAllData}
          />
        ),
      },
      {
        id: '17lands',
        label: '17Lands Integration',
        icon: '📊',
        content: (
          <SeventeenLandsSection
            setCode={setCode}
            onSetCodeChange={setSetCode}
            draftFormat={draftFormat}
            onDraftFormatChange={setDraftFormat}
            isFetchingRatings={isFetchingRatings}
            isFetchingCards={isFetchingCards}
            isRecalculating={isRecalculating}
            recalculateMessage={recalculateMessage}
            dataSource={dataSource}
            isClearingCache={isClearingCache}
            onFetchSetRatings={handleFetchSetRatings}
            onRefreshSetRatings={handleRefreshSetRatings}
            onFetchSetCards={handleFetchSetCards}
            onRefreshSetCards={handleRefreshSetCards}
            onRecalculateGrades={handleRecalculateGrades}
            onClearDatasetCache={handleClearDatasetCache}
          />
        ),
      },
      {
        id: 'ml-recommendations',
        label: 'ML / AI',
        icon: '🤖',
        content: (
          <MLSettingsSection
            mlEnabled={mlEnabled}
            onMLEnabledChange={setMLEnabled}
            llmEnabled={llmEnabled}
            onLLMEnabledChange={setLLMEnabled}
            ollamaEndpoint={ollamaEndpoint}
            onOllamaEndpointChange={setOllamaEndpoint}
            ollamaModel={ollamaModel}
            onOllamaModelChange={setOllamaModel}
            metaGoldfishEnabled={metaGoldfishEnabled}
            onMetaGoldfishEnabledChange={setMetaGoldfishEnabled}
            metaTop8Enabled={metaTop8Enabled}
            onMetaTop8EnabledChange={setMetaTop8Enabled}
            metaWeight={metaWeight}
            onMetaWeightChange={setMetaWeight}
            personalWeight={personalWeight}
            onPersonalWeightChange={setPersonalWeight}
          />
        ),
      },
    ];

    // Add Developer Tools section if developer mode is enabled
    if (isDeveloperMode) {
      items.push({
        id: 'developer-tools',
        label: 'Developer Tools',
        icon: '🛠️',
        content: (
          <ReplayToolSection
            isConnected={isConnected}
            replayToolActive={replayToolActive}
            replayToolPaused={replayToolPaused}
            replayToolProgress={replayToolProgress}
            replaySpeed={replaySpeed}
            onReplaySpeedChange={setReplaySpeed}
            replayFilter={replayFilter}
            onReplayFilterChange={setReplayFilter}
            pauseOnDraft={pauseOnDraft}
            onPauseOnDraftChange={setPauseOnDraft}
            onStartReplayTool={() => handleStartReplayTool(isConnected)}
            onPauseReplayTool={handlePauseReplayTool}
            onResumeReplayTool={handleResumeReplayTool}
            onStopReplayTool={handleStopReplayTool}
          />
        ),
      });
    }

    // About section is always last
    items.push({
      id: 'about',
      label: 'About',
      icon: 'ℹ️',
      content: (
        <AboutSection
          onShowAboutDialog={() => setShowAbout(true)}
          isDeveloperMode={isDeveloperMode}
          onVersionClick={handleVersionClick}
          onToggleDeveloperMode={toggleDeveloperMode}
        />
      ),
    });

    return items;
  }, [
    connectionStatus,
    daemonMode,
    daemonPort,
    isReconnecting,
    handleDaemonPortChange,
    handleReconnect,
    handleModeChange,
    autoRefresh,
    refreshInterval,
    showNotifications,
    theme,
    handleExportData,
    handleImportData,
    isConnected,
    clearDataBeforeReplay,
    setClearDataBeforeReplay,
    isReplaying,
    replayProgress,
    handleImportLogFile,
    handleReplayLogs,
    handleClearAllData,
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
    // ML settings
    mlEnabled,
    setMLEnabled,
    llmEnabled,
    setLLMEnabled,
    ollamaEndpoint,
    setOllamaEndpoint,
    ollamaModel,
    setOllamaModel,
    metaGoldfishEnabled,
    setMetaGoldfishEnabled,
    metaTop8Enabled,
    setMetaTop8Enabled,
    metaWeight,
    setMetaWeight,
    personalWeight,
    setPersonalWeight,
    isDeveloperMode,
    replayToolActive,
    replayToolPaused,
    replayToolProgress,
    replaySpeed,
    setReplaySpeed,
    replayFilter,
    setReplayFilter,
    pauseOnDraft,
    setPauseOnDraft,
    handleStartReplayTool,
    handlePauseReplayTool,
    handleResumeReplayTool,
    handleStopReplayTool,
    handleVersionClick,
    toggleDeveloperMode,
  ]);

  return (
    <div className="page-container">
      <div className="settings-header">
        <h1 className="page-title">Settings</h1>
        {saved && <div className="save-notification">Settings saved successfully!</div>}
      </div>

      <div className="settings-content">
        <SettingsAccordion
          items={accordionItems}
          defaultExpanded={['connection']}
          allowMultiple={true}
        />

        {/* Settings Error */}
        {settingsError && (
          <div className="settings-error">Error: {settingsError}</div>
        )}

        {/* Action Buttons */}
        <div className="settings-actions">
          <button
            className="primary-button"
            onClick={handleSave}
            disabled={isSaving || isLoadingSettings}
          >
            {isSaving ? 'Saving...' : 'Save Settings'}
          </button>
          <button
            className="secondary-button"
            onClick={handleReset}
            disabled={isSaving || isLoadingSettings}
          >
            Reset to Defaults
          </button>
        </div>
      </div>

      {/* About Dialog */}
      <AboutDialog isOpen={showAbout} onClose={() => setShowAbout(false)} />
    </div>
  );
};

export default Settings;
