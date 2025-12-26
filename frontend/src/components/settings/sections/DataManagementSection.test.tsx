import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { DataManagementSection } from './DataManagementSection';
import { gui } from '@/types/models';

describe('DataManagementSection', () => {
  const defaultProps = {
    isConnected: false,
    clearDataBeforeReplay: false,
    onClearDataBeforeReplayChange: vi.fn(),
    isReplaying: false,
    replayProgress: null,
    onExportData: vi.fn(),
    onImportData: vi.fn(),
    onImportLogFile: vi.fn(),
    onReplayLogs: vi.fn(),
    onClearAllData: vi.fn(),
  };

  it('renders section title', () => {
    render(<DataManagementSection {...defaultProps} />);
    expect(screen.getByText('Data Management')).toBeInTheDocument();
  });

  describe('export buttons', () => {
    it('renders export to JSON button', () => {
      render(<DataManagementSection {...defaultProps} />);
      expect(screen.getByText('Export to JSON')).toBeInTheDocument();
    });

    it('renders export to CSV button', () => {
      render(<DataManagementSection {...defaultProps} />);
      expect(screen.getByText('Export to CSV')).toBeInTheDocument();
    });

    it('calls onExportData with json when JSON button clicked', () => {
      const onExportData = vi.fn();
      render(<DataManagementSection {...defaultProps} onExportData={onExportData} />);

      fireEvent.click(screen.getByText('Export to JSON'));

      expect(onExportData).toHaveBeenCalledWith('json');
    });

    it('calls onExportData with csv when CSV button clicked', () => {
      const onExportData = vi.fn();
      render(<DataManagementSection {...defaultProps} onExportData={onExportData} />);

      fireEvent.click(screen.getByText('Export to CSV'));

      expect(onExportData).toHaveBeenCalledWith('csv');
    });
  });

  describe('import buttons', () => {
    it('renders import from JSON button', () => {
      render(<DataManagementSection {...defaultProps} />);
      expect(screen.getByText('Import from JSON')).toBeInTheDocument();
    });

    it('renders select log file button', () => {
      render(<DataManagementSection {...defaultProps} />);
      expect(screen.getByText('Select Log File...')).toBeInTheDocument();
    });

    it('calls onImportData when import JSON clicked', () => {
      const onImportData = vi.fn();
      render(<DataManagementSection {...defaultProps} onImportData={onImportData} />);

      fireEvent.click(screen.getByText('Import from JSON'));

      expect(onImportData).toHaveBeenCalled();
    });

    it('calls onImportLogFile when select log file clicked', () => {
      const onImportLogFile = vi.fn();
      render(<DataManagementSection {...defaultProps} onImportLogFile={onImportLogFile} />);

      fireEvent.click(screen.getByText('Select Log File...'));

      expect(onImportLogFile).toHaveBeenCalled();
    });
  });

  describe('replay logs', () => {
    it('renders replay logs checkbox', () => {
      render(<DataManagementSection {...defaultProps} />);
      expect(screen.getByText(/Clear all data before replay/)).toBeInTheDocument();
    });

    it('calls onClearDataBeforeReplayChange when checkbox toggled', () => {
      const onClearDataBeforeReplayChange = vi.fn();
      render(
        <DataManagementSection
          {...defaultProps}
          onClearDataBeforeReplayChange={onClearDataBeforeReplayChange}
        />
      );

      const checkbox = screen.getByRole('checkbox');
      fireEvent.click(checkbox);

      expect(onClearDataBeforeReplayChange).toHaveBeenCalledWith(true);
    });

    it('disables checkbox when replaying', () => {
      render(<DataManagementSection {...defaultProps} isReplaying={true} />);
      expect(screen.getByRole('checkbox')).toBeDisabled();
    });

    it('disables replay button when not connected', () => {
      render(<DataManagementSection {...defaultProps} isConnected={false} />);
      expect(screen.getByRole('button', { name: 'Replay Historical Logs' })).toBeDisabled();
    });

    it('enables replay button when connected', () => {
      render(<DataManagementSection {...defaultProps} isConnected={true} />);
      expect(screen.getByRole('button', { name: 'Replay Historical Logs' })).not.toBeDisabled();
    });

    it('calls onReplayLogs when replay button clicked', () => {
      const onReplayLogs = vi.fn();
      render(<DataManagementSection {...defaultProps} isConnected={true} onReplayLogs={onReplayLogs} />);

      fireEvent.click(screen.getByRole('button', { name: 'Replay Historical Logs' }));

      expect(onReplayLogs).toHaveBeenCalled();
    });

    it('shows daemon hint when not connected', () => {
      render(<DataManagementSection {...defaultProps} isConnected={false} />);
      expect(screen.getByText('Daemon must be running to replay logs')).toBeInTheDocument();
    });

    it('does not show daemon hint when connected', () => {
      render(<DataManagementSection {...defaultProps} isConnected={true} />);
      expect(screen.queryByText('Daemon must be running to replay logs')).not.toBeInTheDocument();
    });
  });

  describe('replay progress', () => {
    it('does not show progress when not replaying and no progress', () => {
      render(<DataManagementSection {...defaultProps} isReplaying={false} replayProgress={null} />);
      expect(screen.queryByText('Replaying Historical Logs...')).not.toBeInTheDocument();
    });

    it('shows progress when replaying', () => {
      const progress = new gui.LogReplayProgress({
        processedFiles: 5,
        totalFiles: 10,
        totalEntries: 1000,
        matchesImported: 50,
        decksImported: 10,
        questsImported: 5,
      });

      render(<DataManagementSection {...defaultProps} isReplaying={true} replayProgress={progress} />);
      expect(screen.getByText('Replaying Historical Logs...')).toBeInTheDocument();
      expect(screen.getByText(/Files: 5 \/ 10/)).toBeInTheDocument();
    });

    it('shows completion message when done', () => {
      const progress = new gui.LogReplayProgress({
        processedFiles: 10,
        totalFiles: 10,
      });

      render(<DataManagementSection {...defaultProps} isReplaying={false} replayProgress={progress} />);
      expect(screen.getByText('✓ Replay Complete')).toBeInTheDocument();
    });
  });

  describe('clear data', () => {
    it('renders clear all data button', () => {
      render(<DataManagementSection {...defaultProps} />);
      expect(screen.getByRole('button', { name: 'Clear All Data' })).toBeInTheDocument();
    });

    it('calls onClearAllData when clicked', () => {
      const onClearAllData = vi.fn();
      render(<DataManagementSection {...defaultProps} onClearAllData={onClearAllData} />);

      fireEvent.click(screen.getByRole('button', { name: 'Clear All Data' }));

      expect(onClearAllData).toHaveBeenCalled();
    });
  });
});
