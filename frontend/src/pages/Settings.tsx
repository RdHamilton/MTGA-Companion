import { useState, useMemo } from 'react';
import AboutDialog from '../components/AboutDialog';
import {
  DaemonConnectionSection,
  AppPreferencesSection,
  ImportExportSection,
  DataRecoverySection,
  ReplayToolSection,
  SeventeenLandsSection,
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
} from '../hooks';
import './Settings.css';

const Settings = () => {
  // Local UI state
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [refreshInterval, setRefreshInterval] = useState(30);
  const [showNotifications, setShowNotifications] = useState(true);
  const [theme, setTheme] = useState('dark');
  const [saved, setSaved] = useState(false);
  const [showAbout, setShowAbout] = useState(false);

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
  const handleSave = () => {
    // TODO: Implement backend settings save
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  };

  const handleReset = () => {
    setAutoRefresh(false);
    setRefreshInterval(30);
    setShowNotifications(true);
    setTheme('dark');
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

        {/* Action Buttons */}
        <div className="settings-actions">
          <button className="primary-button" onClick={handleSave}>
            Save Settings
          </button>
          <button className="secondary-button" onClick={handleReset}>
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
