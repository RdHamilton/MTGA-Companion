import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { DataRecoverySection } from './DataRecoverySection';
import { gui } from '@/types/models';

describe('DataRecoverySection', () => {
  const defaultProps = {
    isConnected: false,
    clearDataBeforeReplay: false,
    onClearDataBeforeReplayChange: vi.fn(),
    isReplaying: false,
    replayProgress: null,
    onImportLogFile: vi.fn(),
    onReplayLogs: vi.fn(),
    onClearAllData: vi.fn(),
  };

  it('renders section title', () => {
    render(<DataRecoverySection {...defaultProps} />);
    expect(screen.getByText('Data Recovery')).toBeInTheDocument();
  });

  it('renders section description', () => {
    render(<DataRecoverySection {...defaultProps} />);
    expect(screen.getByText(/Recover historical data/)).toBeInTheDocument();
  });

  describe('import log file', () => {
    it('renders select log file button', () => {
      render(<DataRecoverySection {...defaultProps} />);
      expect(screen.getByRole('button', { name: 'Select Log File...' })).toBeInTheDocument();
    });

    it('calls onImportLogFile when button clicked', () => {
      const onImportLogFile = vi.fn();
      render(<DataRecoverySection {...defaultProps} onImportLogFile={onImportLogFile} />);

      fireEvent.click(screen.getByRole('button', { name: 'Select Log File...' }));

      expect(onImportLogFile).toHaveBeenCalled();
    });
  });

  describe('replay logs', () => {
    it('renders replay logs checkbox', () => {
      render(<DataRecoverySection {...defaultProps} />);
      expect(screen.getByText(/Clear all data before replay/)).toBeInTheDocument();
    });

    it('calls onClearDataBeforeReplayChange when checkbox toggled', () => {
      const onClearDataBeforeReplayChange = vi.fn();
      render(
        <DataRecoverySection
          {...defaultProps}
          onClearDataBeforeReplayChange={onClearDataBeforeReplayChange}
        />
      );

      const checkbox = screen.getByRole('checkbox');
      fireEvent.click(checkbox);

      expect(onClearDataBeforeReplayChange).toHaveBeenCalledWith(true);
    });

    it('disables checkbox when replaying', () => {
      render(<DataRecoverySection {...defaultProps} isReplaying={true} />);
      expect(screen.getByRole('checkbox')).toBeDisabled();
    });

    it('disables replay button when not connected', () => {
      render(<DataRecoverySection {...defaultProps} isConnected={false} />);
      expect(screen.getByRole('button', { name: 'Replay Historical Logs' })).toBeDisabled();
    });

    it('enables replay button when connected', () => {
      render(<DataRecoverySection {...defaultProps} isConnected={true} />);
      expect(screen.getByRole('button', { name: 'Replay Historical Logs' })).not.toBeDisabled();
    });

    it('calls onReplayLogs when replay button clicked', () => {
      const onReplayLogs = vi.fn();
      render(<DataRecoverySection {...defaultProps} isConnected={true} onReplayLogs={onReplayLogs} />);

      fireEvent.click(screen.getByRole('button', { name: 'Replay Historical Logs' }));

      expect(onReplayLogs).toHaveBeenCalled();
    });

    it('shows daemon hint when not connected', () => {
      render(<DataRecoverySection {...defaultProps} isConnected={false} />);
      expect(screen.getByText('Daemon must be running to replay logs')).toBeInTheDocument();
    });

    it('does not show daemon hint when connected', () => {
      render(<DataRecoverySection {...defaultProps} isConnected={true} />);
      expect(screen.queryByText('Daemon must be running to replay logs')).not.toBeInTheDocument();
    });
  });

  describe('replay progress', () => {
    it('does not show progress when not replaying and no progress', () => {
      render(<DataRecoverySection {...defaultProps} isReplaying={false} replayProgress={null} />);
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

      render(<DataRecoverySection {...defaultProps} isReplaying={true} replayProgress={progress} />);
      expect(screen.getByText('Replaying Historical Logs...')).toBeInTheDocument();
      expect(screen.getByText(/Files: 5 \/ 10/)).toBeInTheDocument();
    });

    it('shows completion message when done', () => {
      const progress = new gui.LogReplayProgress({
        processedFiles: 10,
        totalFiles: 10,
      });

      render(<DataRecoverySection {...defaultProps} isReplaying={false} replayProgress={progress} />);
      expect(screen.getByText('✓ Replay Complete')).toBeInTheDocument();
    });
  });

  describe('clear data', () => {
    it('renders clear all data button', () => {
      render(<DataRecoverySection {...defaultProps} />);
      expect(screen.getByRole('button', { name: 'Clear All Data' })).toBeInTheDocument();
    });

    it('calls onClearAllData when clicked', () => {
      const onClearAllData = vi.fn();
      render(<DataRecoverySection {...defaultProps} onClearAllData={onClearAllData} />);

      fireEvent.click(screen.getByRole('button', { name: 'Clear All Data' }));

      expect(onClearAllData).toHaveBeenCalled();
    });
  });
});
