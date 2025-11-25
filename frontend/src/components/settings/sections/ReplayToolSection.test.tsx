import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ReplayToolSection } from './ReplayToolSection';
import { gui } from '../../../../wailsjs/go/models';

describe('ReplayToolSection', () => {
  const defaultProps = {
    isConnected: false,
    replayToolActive: false,
    replayToolPaused: false,
    replayToolProgress: null,
    replaySpeed: 1,
    onReplaySpeedChange: vi.fn(),
    replayFilter: 'all',
    onReplayFilterChange: vi.fn(),
    pauseOnDraft: false,
    onPauseOnDraftChange: vi.fn(),
    onStartReplayTool: vi.fn(),
    onPauseReplayTool: vi.fn(),
    onResumeReplayTool: vi.fn(),
    onStopReplayTool: vi.fn(),
  };

  it('renders section title', () => {
    render(<ReplayToolSection {...defaultProps} />);
    expect(screen.getByText('Replay Testing Tool (Daemon Only)')).toBeInTheDocument();
  });

  it('shows daemon warning when not connected', () => {
    render(<ReplayToolSection {...defaultProps} isConnected={false} />);
    expect(screen.getByText(/Replay tool requires daemon mode/)).toBeInTheDocument();
  });

  it('does not show daemon warning when connected', () => {
    render(<ReplayToolSection {...defaultProps} isConnected={true} />);
    expect(screen.queryByText(/Replay tool requires daemon mode/)).not.toBeInTheDocument();
  });

  describe('configuration controls (not active)', () => {
    it('renders replay speed slider', () => {
      render(<ReplayToolSection {...defaultProps} />);
      expect(screen.getByText('Replay Speed')).toBeInTheDocument();
    });

    it('displays current replay speed', () => {
      render(<ReplayToolSection {...defaultProps} replaySpeed={10} />);
      expect(screen.getByText('10x')).toBeInTheDocument();
    });

    it('calls onReplaySpeedChange when slider changes', () => {
      const onReplaySpeedChange = vi.fn();
      render(<ReplayToolSection {...defaultProps} onReplaySpeedChange={onReplaySpeedChange} />);

      const slider = screen.getByRole('slider');
      fireEvent.change(slider, { target: { value: '50' } });

      expect(onReplaySpeedChange).toHaveBeenCalledWith(50);
    });

    it('renders event filter selector', () => {
      render(<ReplayToolSection {...defaultProps} />);
      expect(screen.getByText('Event Filter')).toBeInTheDocument();
    });

    it('calls onReplayFilterChange when filter changes', () => {
      const onReplayFilterChange = vi.fn();
      render(<ReplayToolSection {...defaultProps} onReplayFilterChange={onReplayFilterChange} />);

      const select = screen.getByRole('combobox');
      fireEvent.change(select, { target: { value: 'draft' } });

      expect(onReplayFilterChange).toHaveBeenCalledWith('draft');
    });

    it('renders pause on draft checkbox', () => {
      render(<ReplayToolSection {...defaultProps} />);
      expect(screen.getByText('Auto-pause on draft events')).toBeInTheDocument();
    });

    it('calls onPauseOnDraftChange when checkbox toggled', () => {
      const onPauseOnDraftChange = vi.fn();
      render(<ReplayToolSection {...defaultProps} onPauseOnDraftChange={onPauseOnDraftChange} />);

      const checkbox = screen.getByRole('checkbox');
      fireEvent.click(checkbox);

      expect(onPauseOnDraftChange).toHaveBeenCalledWith(true);
    });

    it('renders start button', () => {
      render(<ReplayToolSection {...defaultProps} />);
      expect(screen.getByText('Select Log File(s) & Start')).toBeInTheDocument();
    });

    it('disables start button when not connected', () => {
      render(<ReplayToolSection {...defaultProps} isConnected={false} />);
      expect(screen.getByText('Select Log File(s) & Start')).toBeDisabled();
    });

    it('enables start button when connected', () => {
      render(<ReplayToolSection {...defaultProps} isConnected={true} />);
      expect(screen.getByText('Select Log File(s) & Start')).not.toBeDisabled();
    });

    it('calls onStartReplayTool when start button clicked', () => {
      const onStartReplayTool = vi.fn();
      render(<ReplayToolSection {...defaultProps} isConnected={true} onStartReplayTool={onStartReplayTool} />);

      fireEvent.click(screen.getByText('Select Log File(s) & Start'));

      expect(onStartReplayTool).toHaveBeenCalled();
    });
  });

  describe('active replay controls', () => {
    const activeProps = {
      ...defaultProps,
      isConnected: true,
      replayToolActive: true,
    };

    it('does not show config controls when active', () => {
      render(<ReplayToolSection {...activeProps} />);
      expect(screen.queryByText('Replay Speed')).not.toBeInTheDocument();
    });

    it('shows active status when not paused', () => {
      render(<ReplayToolSection {...activeProps} replayToolPaused={false} />);
      expect(screen.getByText(/Replay Active/)).toBeInTheDocument();
    });

    it('shows paused status when paused', () => {
      render(<ReplayToolSection {...activeProps} replayToolPaused={true} />);
      expect(screen.getByText(/Replay Paused/)).toBeInTheDocument();
    });

    it('shows pause button when not paused', () => {
      render(<ReplayToolSection {...activeProps} replayToolPaused={false} />);
      expect(screen.getByRole('button', { name: /Pause/ })).toBeInTheDocument();
      expect(screen.queryByRole('button', { name: /Resume/ })).not.toBeInTheDocument();
    });

    it('shows resume button when paused', () => {
      render(<ReplayToolSection {...activeProps} replayToolPaused={true} />);
      expect(screen.getByRole('button', { name: /Resume/ })).toBeInTheDocument();
      expect(screen.queryByRole('button', { name: /Pause/ })).not.toBeInTheDocument();
    });

    it('shows stop button', () => {
      render(<ReplayToolSection {...activeProps} />);
      expect(screen.getByText(/Stop/)).toBeInTheDocument();
    });

    it('calls onPauseReplayTool when pause clicked', () => {
      const onPauseReplayTool = vi.fn();
      render(<ReplayToolSection {...activeProps} onPauseReplayTool={onPauseReplayTool} />);

      fireEvent.click(screen.getByText(/Pause/));

      expect(onPauseReplayTool).toHaveBeenCalled();
    });

    it('calls onResumeReplayTool when resume clicked', () => {
      const onResumeReplayTool = vi.fn();
      render(<ReplayToolSection {...activeProps} replayToolPaused={true} onResumeReplayTool={onResumeReplayTool} />);

      fireEvent.click(screen.getByText(/Resume/));

      expect(onResumeReplayTool).toHaveBeenCalled();
    });

    it('calls onStopReplayTool when stop clicked', () => {
      const onStopReplayTool = vi.fn();
      render(<ReplayToolSection {...activeProps} onStopReplayTool={onStopReplayTool} />);

      fireEvent.click(screen.getByText(/Stop/));

      expect(onStopReplayTool).toHaveBeenCalled();
    });
  });

  describe('progress display', () => {
    it('shows progress when available', () => {
      const progress = new gui.ReplayStatus({
        isActive: true,
        isPaused: false,
        currentEntry: 50,
        totalEntries: 100,
        percentComplete: 50,
        elapsed: 30,
        speed: 5,
        filter: 'all',
      });

      render(
        <ReplayToolSection
          {...defaultProps}
          isConnected={true}
          replayToolActive={true}
          replayToolProgress={progress}
        />
      );

      expect(screen.getByText(/50 \/ 100 entries/)).toBeInTheDocument();
      expect(screen.getByText(/50\.0%/)).toBeInTheDocument();
    });
  });
});
