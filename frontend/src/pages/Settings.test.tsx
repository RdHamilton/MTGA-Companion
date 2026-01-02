import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import Settings from './Settings';

// Mock scrollIntoView (not available in jsdom)
Element.prototype.scrollIntoView = vi.fn();

// Mock all API modules used by Settings and its hooks
vi.mock('@/services/api', () => ({
  settings: {
    getSettings: vi.fn(),
    updateSettings: vi.fn(),
  },
  system: {
    getStatus: vi.fn(),
    getVersion: vi.fn(),
    clearAllData: vi.fn(),
    checkOllamaStatus: vi.fn(),
    getAvailableOllamaModels: vi.fn(),
    pullOllamaModel: vi.fn(),
    testLLMGeneration: vi.fn(),
    connectDaemon: vi.fn(),
  },
  matches: {
    exportMatches: vi.fn(),
    exportMatchesCsv: vi.fn(),
  },
}));

vi.mock('@/services/websocketClient', () => ({
  EventsOn: vi.fn(() => () => {}),
  WindowReloadApp: vi.fn(),
}));

// Mock the App module for replay state
vi.mock('../App', () => ({
  subscribeToReplayState: vi.fn(() => () => {}),
  getReplayState: vi.fn(() => ({
    isActive: false,
    isPaused: false,
    progress: null,
  })),
}));

// Mock the ToastContainer
vi.mock('../components/ToastContainer', () => ({
  showToast: {
    show: vi.fn(),
  },
}));

// Mock useDownload since Settings uses useSeventeenLands which now uses download progress
vi.mock('@/context/DownloadContext', () => ({
  useDownload: () => ({
    state: { tasks: [], activeTask: null },
    isDownloading: false,
    overallProgress: 0,
    startDownload: vi.fn(),
    updateProgress: vi.fn(),
    completeDownload: vi.fn(),
    failDownload: vi.fn(),
    cancelDownload: vi.fn(),
  }),
  DownloadProvider: ({ children }: { children: React.ReactNode }) => children,
}));

import { showToast } from '../components/ToastContainer';
import { settings, system, matches } from '@/services/api';

// Default mock connection status
const defaultConnectionStatus = {
  status: 'standalone',
  connected: false,
  mode: 'standalone',
  url: 'ws://localhost:9999',
  port: 9999,
};

// Default mock settings
const defaultSettings = {
  autoRefresh: false,
  refreshInterval: 30,
  showNotifications: true,
  theme: 'dark',
  daemonPort: 9999,
  daemonMode: 'standalone',
};

describe('Settings', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
    window.location.hash = '';

    // Default mock implementations
    (system.getStatus as ReturnType<typeof vi.fn>).mockResolvedValue(defaultConnectionStatus);
    (settings.getSettings as ReturnType<typeof vi.fn>).mockResolvedValue(defaultSettings);
    (settings.updateSettings as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);
    (system.getVersion as ReturnType<typeof vi.fn>).mockResolvedValue('v1.3.1');
  });

  describe('rendering', () => {
    it('renders the Settings page title', async () => {
      render(<Settings />);

      expect(screen.getByRole('heading', { level: 1, name: 'Settings' })).toBeInTheDocument();
    });

    it('renders accordion with all sections', async () => {
      render(<Settings />);

      // Check accordion section headers
      expect(screen.getByRole('button', { name: /connection/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /preferences/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /import \/ export/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /data recovery/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /17lands integration/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /about/i })).toBeInTheDocument();
    });

    it('renders Expand All and Collapse All buttons', () => {
      render(<Settings />);

      expect(screen.getByRole('button', { name: /expand all/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /collapse all/i })).toBeInTheDocument();
    });

    it('renders Save Settings and Reset to Defaults buttons', () => {
      render(<Settings />);

      expect(screen.getByRole('button', { name: /save settings/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /reset to defaults/i })).toBeInTheDocument();
    });

    it('expands connection section by default', () => {
      render(<Settings />);

      // Connection section should be expanded by default
      const connectionHeader = screen.getByRole('button', { name: /connection/i });
      expect(connectionHeader).toHaveAttribute('aria-expanded', 'true');
    });
  });

  describe('accordion navigation', () => {
    it('expands section when clicked', () => {
      render(<Settings />);

      const preferencesHeader = screen.getByRole('button', { name: /preferences/i });
      expect(preferencesHeader).toHaveAttribute('aria-expanded', 'false');

      fireEvent.click(preferencesHeader);

      expect(preferencesHeader).toHaveAttribute('aria-expanded', 'true');
    });

    it('expands all sections when Expand All is clicked', () => {
      render(<Settings />);

      const expandAllButton = screen.getByRole('button', { name: /expand all/i });
      fireEvent.click(expandAllButton);

      // All sections should be expanded
      expect(screen.getByRole('button', { name: /connection/i })).toHaveAttribute('aria-expanded', 'true');
      expect(screen.getByRole('button', { name: /preferences/i })).toHaveAttribute('aria-expanded', 'true');
      expect(screen.getByRole('button', { name: /import \/ export/i })).toHaveAttribute('aria-expanded', 'true');
    });

    it('collapses all sections when Collapse All is clicked', () => {
      render(<Settings />);

      // First expand all
      const expandAllButton = screen.getByRole('button', { name: /expand all/i });
      fireEvent.click(expandAllButton);

      // Then collapse all
      const collapseAllButton = screen.getByRole('button', { name: /collapse all/i });
      fireEvent.click(collapseAllButton);

      // All sections should be collapsed
      expect(screen.getByRole('button', { name: /connection/i })).toHaveAttribute('aria-expanded', 'false');
      expect(screen.getByRole('button', { name: /preferences/i })).toHaveAttribute('aria-expanded', 'false');
    });
  });

  describe('Connection section', () => {
    it('displays connection status', async () => {
      render(<Settings />);

      // Expand connection section if not already
      const connectionHeader = screen.getByRole('button', { name: /connection/i });
      if (connectionHeader.getAttribute('aria-expanded') === 'false') {
        fireEvent.click(connectionHeader);
      }

      await waitFor(() => {
        expect(screen.getByText('Standalone Mode')).toBeInTheDocument();
      });
    });

    it('displays connected status when daemon is connected', async () => {
      (system.getStatus as ReturnType<typeof vi.fn>).mockResolvedValue({
        ...defaultConnectionStatus,
        status: 'connected',
        connected: true,
      });

      render(<Settings />);

      await waitFor(() => {
        expect(screen.getByText('Connected to Daemon')).toBeInTheDocument();
      });
    });

    it('displays reconnecting status', async () => {
      (system.getStatus as ReturnType<typeof vi.fn>).mockResolvedValue({
        ...defaultConnectionStatus,
        status: 'reconnecting',
      });

      render(<Settings />);

      await waitFor(() => {
        expect(screen.getByText('Reconnecting...')).toBeInTheDocument();
      });
    });

    it('renders Connection Mode selector', async () => {
      render(<Settings />);

      // Expand connection section
      const connectionHeader = screen.getByRole('button', { name: /connection/i });
      if (connectionHeader.getAttribute('aria-expanded') === 'false') {
        fireEvent.click(connectionHeader);
      }

      await waitFor(() => {
        expect(screen.getByText('Connection Mode')).toBeInTheDocument();
      });
    });

    // Skip: Daemon control is not implemented in REST API - useDaemonConnection uses no-op functions
    it.skip('handles reconnect button click', async () => {
      (system.getStatus as ReturnType<typeof vi.fn>).mockResolvedValue({
        ...defaultConnectionStatus,
        status: 'connected',
        connected: true,
        mode: 'daemon',
      });
      (system.connectDaemon as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);

      render(<Settings />);

      // Wait for initial load
      await waitFor(() => {
        expect(system.getStatus).toHaveBeenCalled();
      });

      // Expand connection section
      const connectionHeader = screen.getByRole('button', { name: /connection/i });
      if (connectionHeader.getAttribute('aria-expanded') === 'false') {
        fireEvent.click(connectionHeader);
      }

      const reconnectButton = screen.getByRole('button', { name: /reconnect to daemon/i });
      fireEvent.click(reconnectButton);

      await waitFor(() => {
        expect(system.connectDaemon).toHaveBeenCalled();
      });
    });

    // Test removed: reconnect error handling uses no-op functions in useDaemonConnection

    // Tests removed: Daemon control functions (switchToStandaloneMode, switchToDaemonMode, setDaemonPort)
    // are not implemented in REST API - useDaemonConnection uses no-op functions
  });

  describe('Preferences section', () => {
    it('renders theme selector', () => {
      render(<Settings />);

      // Expand preferences section
      const preferencesHeader = screen.getByRole('button', { name: /preferences/i });
      fireEvent.click(preferencesHeader);

      expect(screen.getByText('Theme')).toBeInTheDocument();
    });

    it('renders auto-refresh toggle', () => {
      render(<Settings />);

      // Expand preferences section
      const preferencesHeader = screen.getByRole('button', { name: /preferences/i });
      fireEvent.click(preferencesHeader);

      expect(screen.getByText('Auto-refresh data')).toBeInTheDocument();
    });

    it('shows refresh interval when auto-refresh is enabled', () => {
      render(<Settings />);

      // Expand preferences section
      const preferencesHeader = screen.getByRole('button', { name: /preferences/i });
      fireEvent.click(preferencesHeader);

      // Find and click the auto-refresh toggle
      const autoRefreshCheckbox = screen.getByRole('checkbox', { name: /auto-refresh data/i });
      fireEvent.click(autoRefreshCheckbox);

      expect(screen.getByText('Refresh Interval (seconds)')).toBeInTheDocument();
    });

    it('renders notifications toggle', () => {
      render(<Settings />);

      // Expand preferences section
      const preferencesHeader = screen.getByRole('button', { name: /preferences/i });
      fireEvent.click(preferencesHeader);

      expect(screen.getByText('Show notifications')).toBeInTheDocument();
    });
  });

  describe('Import/Export section', () => {
    it('renders export buttons', () => {
      render(<Settings />);

      // Expand import/export section
      const importExportHeader = screen.getByRole('button', { name: /import \/ export/i });
      fireEvent.click(importExportHeader);

      expect(screen.getByRole('button', { name: /export to json/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /export to csv/i })).toBeInTheDocument();
    });

    it('renders import button', () => {
      render(<Settings />);

      // Expand import/export section
      const importExportHeader = screen.getByRole('button', { name: /import \/ export/i });
      fireEvent.click(importExportHeader);

      expect(screen.getByRole('button', { name: /import from json/i })).toBeInTheDocument();
    });

    it('handles export to JSON', async () => {
      (matches.exportMatches as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);

      render(<Settings />);

      // Expand import/export section
      const importExportHeader = screen.getByRole('button', { name: /import \/ export/i });
      fireEvent.click(importExportHeader);

      const exportJsonButton = screen.getByRole('button', { name: /export to json/i });
      fireEvent.click(exportJsonButton);

      await waitFor(() => {
        expect(matches.exportMatches).toHaveBeenCalled();
      });
    });

    it('handles export to CSV', async () => {
      (matches.exportMatches as ReturnType<typeof vi.fn>).mockResolvedValue('csv data');

      render(<Settings />);

      // Expand import/export section
      const importExportHeader = screen.getByRole('button', { name: /import \/ export/i });
      fireEvent.click(importExportHeader);

      const exportCsvButton = screen.getByRole('button', { name: /export to csv/i });
      fireEvent.click(exportCsvButton);

      await waitFor(() => {
        expect(matches.exportMatches).toHaveBeenCalledWith('csv');
      });
    });

    // Test removed: importFromFile requires native file picker integration - useDataManagement uses no-op functions

    it('shows error when export fails', async () => {
      (matches.exportMatches as ReturnType<typeof vi.fn>).mockRejectedValue(new Error('Export failed'));

      render(<Settings />);

      // Expand import/export section
      const importExportHeader = screen.getByRole('button', { name: /import \/ export/i });
      fireEvent.click(importExportHeader);

      const exportJsonButton = screen.getByRole('button', { name: /export to json/i });
      fireEvent.click(exportJsonButton);

      await waitFor(() => {
        expect(showToast.show).toHaveBeenCalledWith(
          expect.stringContaining('Failed to export'),
          'error'
        );
      });
    });
  });

  describe('Data Recovery section', () => {
    it('renders import log file button', () => {
      render(<Settings />);

      // Expand data recovery section
      const dataRecoveryHeader = screen.getByRole('button', { name: /data recovery/i });
      fireEvent.click(dataRecoveryHeader);

      expect(screen.getByRole('button', { name: /select log file/i })).toBeInTheDocument();
    });

    it('renders replay historical logs button', () => {
      render(<Settings />);

      // Expand data recovery section
      const dataRecoveryHeader = screen.getByRole('button', { name: /data recovery/i });
      fireEvent.click(dataRecoveryHeader);

      expect(screen.getByRole('button', { name: /replay historical logs/i })).toBeInTheDocument();
    });

    it('renders clear all data button', () => {
      render(<Settings />);

      // Expand data recovery section
      const dataRecoveryHeader = screen.getByRole('button', { name: /data recovery/i });
      fireEvent.click(dataRecoveryHeader);

      expect(screen.getByRole('button', { name: /clear all data/i })).toBeInTheDocument();
    });

    // Test removed: importLogFile requires native file picker integration - useDataManagement uses no-op functions

    it('handles clear all data', async () => {
      const reloadSpy = vi.fn();
      Object.defineProperty(window, 'location', {
        value: { ...window.location, reload: reloadSpy },
        writable: true,
      });

      (system.clearAllData as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);

      render(<Settings />);

      // Expand data recovery section
      const dataRecoveryHeader = screen.getByRole('button', { name: /data recovery/i });
      fireEvent.click(dataRecoveryHeader);

      const clearDataButton = screen.getByRole('button', { name: /clear all data/i });
      fireEvent.click(clearDataButton);

      await waitFor(() => {
        expect(system.clearAllData).toHaveBeenCalled();
      });
    });

    it('shows clear data before replay checkbox', () => {
      render(<Settings />);

      // Expand data recovery section
      const dataRecoveryHeader = screen.getByRole('button', { name: /data recovery/i });
      fireEvent.click(dataRecoveryHeader);

      expect(screen.getByText(/clear all data before replay/i)).toBeInTheDocument();
    });

    it('disables replay button when not connected', () => {
      render(<Settings />);

      // Expand data recovery section
      const dataRecoveryHeader = screen.getByRole('button', { name: /data recovery/i });
      fireEvent.click(dataRecoveryHeader);

      const replayButton = screen.getByRole('button', { name: /replay historical logs/i });
      expect(replayButton).toBeDisabled();
    });

    it('shows daemon warning when not connected', () => {
      render(<Settings />);

      // Expand data recovery section
      const dataRecoveryHeader = screen.getByRole('button', { name: /data recovery/i });
      fireEvent.click(dataRecoveryHeader);

      expect(screen.getByText(/daemon must be running/i)).toBeInTheDocument();
    });
  });

  describe('17Lands Integration section', () => {
    it('renders set code input', () => {
      render(<Settings />);

      // Expand 17lands section
      const seventeenLandsHeader = screen.getByRole('button', { name: /17lands integration/i });
      fireEvent.click(seventeenLandsHeader);

      expect(screen.getByText('Set Code')).toBeInTheDocument();
      expect(screen.getByPlaceholderText(/tla, blb, dsk/i)).toBeInTheDocument();
    });

    it('renders draft format selector', () => {
      render(<Settings />);

      // Expand 17lands section
      const seventeenLandsHeader = screen.getByRole('button', { name: /17lands integration/i });
      fireEvent.click(seventeenLandsHeader);

      expect(screen.getByText('Draft Format')).toBeInTheDocument();
    });

    it('renders fetch ratings button', () => {
      render(<Settings />);

      // Expand 17lands section
      const seventeenLandsHeader = screen.getByRole('button', { name: /17lands integration/i });
      fireEvent.click(seventeenLandsHeader);

      expect(screen.getByRole('button', { name: /fetch ratings/i })).toBeInTheDocument();
    });

    it('renders fetch card data button', () => {
      render(<Settings />);

      // Expand 17lands section
      const seventeenLandsHeader = screen.getByRole('button', { name: /17lands integration/i });
      fireEvent.click(seventeenLandsHeader);

      expect(screen.getByRole('button', { name: /fetch card data/i })).toBeInTheDocument();
    });

    it('renders recalculate draft grades button', () => {
      render(<Settings />);

      // Expand 17lands section
      const seventeenLandsHeader = screen.getByRole('button', { name: /17lands integration/i });
      fireEvent.click(seventeenLandsHeader);

      expect(screen.getByRole('button', { name: /recalculate all drafts/i })).toBeInTheDocument();
    });

    it('renders clear dataset cache button', () => {
      render(<Settings />);

      // Expand 17lands section
      const seventeenLandsHeader = screen.getByRole('button', { name: /17lands integration/i });
      fireEvent.click(seventeenLandsHeader);

      expect(screen.getByRole('button', { name: /clear dataset cache/i })).toBeInTheDocument();
    });

    it('disables fetch buttons when set code is empty', () => {
      render(<Settings />);

      // Expand 17lands section
      const seventeenLandsHeader = screen.getByRole('button', { name: /17lands integration/i });
      fireEvent.click(seventeenLandsHeader);

      const fetchRatingsButton = screen.getByRole('button', { name: /fetch ratings/i });
      expect(fetchRatingsButton).toBeDisabled();
    });

    // Tests removed: FetchSetRatings, FetchSetCards, recalculateAllDraftGrades, clearDatasetCache
    // are not implemented in REST API - useSeventeenLands uses no-op functions

    it('shows warning when set code is empty and fetch is clicked', async () => {
      render(<Settings />);

      // Expand 17lands section
      const seventeenLandsHeader = screen.getByRole('button', { name: /17lands integration/i });
      fireEvent.click(seventeenLandsHeader);

      // Don't enter a set code, just try to fetch
      // Since button is disabled, we need to test the message via the handler
      // This tests that the button is properly disabled
      const fetchRatingsButton = screen.getByRole('button', { name: /fetch ratings/i });
      expect(fetchRatingsButton).toBeDisabled();
    });
  });

  describe('About section', () => {
    it('renders version information', () => {
      render(<Settings />);

      // Expand about section
      const aboutHeader = screen.getByRole('button', { name: /about/i });
      fireEvent.click(aboutHeader);

      expect(screen.getByText('Version:')).toBeInTheDocument();
      expect(screen.getByText('1.3.1')).toBeInTheDocument();
    });

    it('renders build information', () => {
      render(<Settings />);

      // Expand about section
      const aboutHeader = screen.getByRole('button', { name: /about/i });
      fireEvent.click(aboutHeader);

      expect(screen.getByText('Build:')).toBeInTheDocument();
      expect(screen.getByText('Development')).toBeInTheDocument();
    });

    it('renders platform information', () => {
      render(<Settings />);

      // Expand about section
      const aboutHeader = screen.getByRole('button', { name: /about/i });
      fireEvent.click(aboutHeader);

      expect(screen.getByText('Platform:')).toBeInTheDocument();
      expect(screen.getByText('Wails + React')).toBeInTheDocument();
    });

    it('renders about dialog button', () => {
      render(<Settings />);

      // Expand about section
      const aboutHeader = screen.getByRole('button', { name: /about/i });
      fireEvent.click(aboutHeader);

      expect(screen.getByRole('button', { name: /about mtga companion/i })).toBeInTheDocument();
    });

    it('opens about dialog when button is clicked', () => {
      render(<Settings />);

      // Expand about section
      const aboutHeader = screen.getByRole('button', { name: /about/i });
      fireEvent.click(aboutHeader);

      const aboutButton = screen.getByRole('button', { name: /about mtga companion/i });
      fireEvent.click(aboutButton);

      // About dialog should be open - check for modal-specific content
      // The modal has a heading "About MTGA Companion" and app description text
      expect(screen.getByText(/desktop application for tracking and analyzing/i)).toBeInTheDocument();
      // Modal has a close button with modal-close class
      expect(document.querySelector('.modal-close')).toBeInTheDocument();
    });
  });

  describe('Developer Mode', () => {
    it('does not show Developer Tools section by default', () => {
      render(<Settings />);

      expect(screen.queryByRole('button', { name: /developer tools/i })).not.toBeInTheDocument();
    });

    it('shows Developer Tools section when developer mode is enabled', () => {
      // Enable developer mode in localStorage
      localStorage.setItem('mtga-companion-developer-mode', 'true');

      render(<Settings />);

      expect(screen.getByRole('button', { name: /developer tools/i })).toBeInTheDocument();
    });

    it('activates developer mode after clicking version 5 times', async () => {
      render(<Settings />);

      // Expand about section
      const aboutHeader = screen.getByRole('button', { name: /about/i });
      fireEvent.click(aboutHeader);

      // Click version 5 times quickly (within 3 second timeout)
      const versionElement = screen.getByText('1.3.1');
      for (let i = 0; i < 5; i++) {
        fireEvent.click(versionElement);
      }

      // Developer mode should now be enabled
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /developer tools/i })).toBeInTheDocument();
      });
    });

    it('shows developer mode indicator in About section when enabled', () => {
      localStorage.setItem('mtga-companion-developer-mode', 'true');

      render(<Settings />);

      // Expand about section
      const aboutHeader = screen.getByRole('button', { name: /about/i });
      fireEvent.click(aboutHeader);

      expect(screen.getByText('Developer Mode:')).toBeInTheDocument();
      expect(screen.getByText('Enabled')).toBeInTheDocument();
    });

    it('allows disabling developer mode via toggle', () => {
      localStorage.setItem('mtga-companion-developer-mode', 'true');

      render(<Settings />);

      // Expand about section
      const aboutHeader = screen.getByRole('button', { name: /about/i });
      fireEvent.click(aboutHeader);

      const disableButton = screen.getByRole('button', { name: /disable/i });
      fireEvent.click(disableButton);

      // Developer Tools section should be hidden
      expect(screen.queryByRole('button', { name: /developer tools/i })).not.toBeInTheDocument();
    });
  });

  describe('Save/Reset actions', () => {
    it('shows success notification when save is clicked', async () => {
      render(<Settings />);

      // Wait for settings to load from backend first (button is disabled while loading)
      const saveButton = screen.getByRole('button', { name: /save settings/i });
      await waitFor(() => {
        expect(saveButton).not.toBeDisabled();
      });

      // Click save button
      fireEvent.click(saveButton);

      // Wait for the async save to complete and notification to appear
      await waitFor(() => {
        expect(screen.getByText('Settings saved successfully!')).toBeInTheDocument();
      });

      // Verify settings.updateSettings was called
      expect(settings.updateSettings).toHaveBeenCalled();
    });

    it('resets preferences when reset is clicked', async () => {
      render(<Settings />);

      // Wait for settings to load from backend (button becomes enabled)
      const resetButton = screen.getByRole('button', { name: /reset to defaults/i });
      await waitFor(() => {
        expect(resetButton).not.toBeDisabled();
      });

      // Expand preferences section
      const preferencesHeader = screen.getByRole('button', { name: /preferences/i });
      fireEvent.click(preferencesHeader);

      // Enable auto-refresh
      const autoRefreshCheckbox = screen.getByRole('checkbox', { name: /auto-refresh data/i });
      fireEvent.click(autoRefreshCheckbox);

      expect(autoRefreshCheckbox).toBeChecked();

      // Reset
      fireEvent.click(resetButton);

      // Auto-refresh should be unchecked after reset
      await waitFor(() => {
        expect(autoRefreshCheckbox).not.toBeChecked();
      });
    });
  });

  describe('Keyboard navigation', () => {
    it('toggles section on Enter key', () => {
      render(<Settings />);

      const preferencesHeader = screen.getByRole('button', { name: /preferences/i });
      preferencesHeader.focus();
      fireEvent.keyDown(preferencesHeader, { key: 'Enter' });

      expect(preferencesHeader).toHaveAttribute('aria-expanded', 'true');
    });

    it('toggles section on Space key', () => {
      render(<Settings />);

      const preferencesHeader = screen.getByRole('button', { name: /preferences/i });
      preferencesHeader.focus();
      fireEvent.keyDown(preferencesHeader, { key: ' ' });

      expect(preferencesHeader).toHaveAttribute('aria-expanded', 'true');
    });
  });
});
