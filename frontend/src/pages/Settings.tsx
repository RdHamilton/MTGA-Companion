import { useState } from 'react';
import AboutDialog from '../components/AboutDialog';
import {
  DatabaseSection,
  DaemonConnectionSection,
  AppPreferencesSection,
  DataManagementSection,
  ReplayToolSection,
  SeventeenLandsSection,
  AppearanceSection,
  AboutSection,
} from '../components/settings/sections';
import {
  useDaemonConnection,
  useLogReplay,
  useReplayTool,
  useSeventeenLands,
  useDataManagement,
} from '../hooks';
import './Settings.css';

const Settings = () => {
  // Local UI state
  const [dbPath, setDbPath] = useState('');
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [refreshInterval, setRefreshInterval] = useState(30);
  const [showNotifications, setShowNotifications] = useState(true);
  const [saved, setSaved] = useState(false);
  const [showAbout, setShowAbout] = useState(false);

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
    setDbPath('');
    setAutoRefresh(false);
    setRefreshInterval(30);
    setShowNotifications(true);
  };

  return (
    <div className="page-container">
      <div className="settings-header">
        <h1 className="page-title">Settings</h1>
        {saved && <div className="save-notification">Settings saved successfully!</div>}
      </div>

      <div className="settings-content">
        <DatabaseSection
          dbPath={dbPath}
          onDbPathChange={setDbPath}
        />

        <DaemonConnectionSection
          connectionStatus={connectionStatus}
          daemonMode={daemonMode}
          daemonPort={daemonPort}
          isReconnecting={isReconnecting}
          onDaemonPortChange={handleDaemonPortChange}
          onReconnect={handleReconnect}
          onModeChange={handleModeChange}
        />

        <AppPreferencesSection
          autoRefresh={autoRefresh}
          onAutoRefreshChange={setAutoRefresh}
          refreshInterval={refreshInterval}
          onRefreshIntervalChange={setRefreshInterval}
          showNotifications={showNotifications}
          onShowNotificationsChange={setShowNotifications}
        />

        <DataManagementSection
          isConnected={isConnected}
          clearDataBeforeReplay={clearDataBeforeReplay}
          onClearDataBeforeReplayChange={setClearDataBeforeReplay}
          isReplaying={isReplaying}
          replayProgress={replayProgress}
          onExportData={handleExportData}
          onImportData={handleImportData}
          onImportLogFile={handleImportLogFile}
          onReplayLogs={() => handleReplayLogs(isConnected)}
          onClearAllData={handleClearAllData}
        />

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

        <AppearanceSection />

        <AboutSection onShowAboutDialog={() => setShowAbout(true)} />

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
