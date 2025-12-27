import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { DaemonConnectionSection } from './DaemonConnectionSection';
import { gui } from '@/types/models';

describe('DaemonConnectionSection', () => {
  const createConnectionStatus = (status: string) => new gui.ConnectionStatus({
    status,
    connected: status === 'connected',
    mode: status === 'connected' ? 'daemon' : 'standalone',
    url: 'ws://localhost:9999',
    port: 9999,
  });

  const defaultProps = {
    connectionStatus: createConnectionStatus('standalone'),
    daemonMode: 'auto',
    daemonPort: 9999,
    isReconnecting: false,
    onDaemonPortChange: vi.fn(),
    onReconnect: vi.fn(),
    onModeChange: vi.fn(),
  };

  it('renders section title', () => {
    render(<DaemonConnectionSection {...defaultProps} />);
    expect(screen.getByText('Daemon Connection')).toBeInTheDocument();
  });

  describe('connection status', () => {
    it('shows connected status', () => {
      render(
        <DaemonConnectionSection
          {...defaultProps}
          connectionStatus={createConnectionStatus('connected')}
        />
      );
      expect(screen.getByText('Connected to Daemon')).toBeInTheDocument();
    });

    it('shows standalone status', () => {
      render(
        <DaemonConnectionSection
          {...defaultProps}
          connectionStatus={createConnectionStatus('standalone')}
        />
      );
      expect(screen.getByText('Standalone Mode')).toBeInTheDocument();
    });

    it('shows reconnecting status', () => {
      render(
        <DaemonConnectionSection
          {...defaultProps}
          connectionStatus={createConnectionStatus('reconnecting')}
        />
      );
      expect(screen.getByText('Reconnecting...')).toBeInTheDocument();
    });
  });

  describe('connection mode selector', () => {
    it('renders connection mode select', () => {
      render(<DaemonConnectionSection {...defaultProps} />);
      expect(screen.getByText('Connection Mode')).toBeInTheDocument();
    });

    it('calls onModeChange when mode changes', () => {
      const onModeChange = vi.fn();
      render(<DaemonConnectionSection {...defaultProps} onModeChange={onModeChange} />);

      const select = screen.getByRole('combobox');
      fireEvent.change(select, { target: { value: 'standalone' } });

      expect(onModeChange).toHaveBeenCalledWith('standalone');
    });
  });

  describe('daemon port', () => {
    it('renders port input with current value', () => {
      render(<DaemonConnectionSection {...defaultProps} daemonPort={8888} />);
      expect(screen.getByDisplayValue('8888')).toBeInTheDocument();
    });

    it('shows port hint', () => {
      render(<DaemonConnectionSection {...defaultProps} daemonPort={8888} />);
      expect(screen.getByText('ws://localhost:8888')).toBeInTheDocument();
    });

    it('calls onDaemonPortChange when port changes on blur', () => {
      const onDaemonPortChange = vi.fn();
      render(<DaemonConnectionSection {...defaultProps} onDaemonPortChange={onDaemonPortChange} />);

      const input = screen.getByDisplayValue('9999');
      fireEvent.change(input, { target: { value: '8080' } });
      fireEvent.blur(input);

      expect(onDaemonPortChange).toHaveBeenCalledWith(8080);
    });

    it('allows typing any digits before blur', () => {
      const onDaemonPortChange = vi.fn();
      render(<DaemonConnectionSection {...defaultProps} onDaemonPortChange={onDaemonPortChange} />);

      const input = screen.getByDisplayValue('9999');
      fireEvent.change(input, { target: { value: '68' } });

      // Should show intermediate value while typing
      expect(screen.getByDisplayValue('68')).toBeInTheDocument();
      // Should not call handler until blur
      expect(onDaemonPortChange).not.toHaveBeenCalled();
    });

    it('resets to valid port if invalid on blur', () => {
      const onDaemonPortChange = vi.fn();
      render(<DaemonConnectionSection {...defaultProps} daemonPort={9999} onDaemonPortChange={onDaemonPortChange} />);

      const input = screen.getByDisplayValue('9999');
      fireEvent.change(input, { target: { value: '500' } }); // Invalid - below 1024
      fireEvent.blur(input);

      // Should reset to original valid port
      expect(screen.getByDisplayValue('9999')).toBeInTheDocument();
      expect(onDaemonPortChange).not.toHaveBeenCalled();
    });

    it('disables port input in standalone mode', () => {
      render(<DaemonConnectionSection {...defaultProps} daemonMode="standalone" />);
      expect(screen.getByDisplayValue('9999')).toBeDisabled();
    });
  });

  describe('reconnect button', () => {
    it('renders reconnect button', () => {
      render(<DaemonConnectionSection {...defaultProps} />);
      expect(screen.getByRole('button', { name: 'Reconnect to Daemon' })).toBeInTheDocument();
    });

    it('calls onReconnect when clicked', () => {
      const onReconnect = vi.fn();
      render(<DaemonConnectionSection {...defaultProps} onReconnect={onReconnect} />);

      fireEvent.click(screen.getByRole('button', { name: 'Reconnect to Daemon' }));

      expect(onReconnect).toHaveBeenCalled();
    });

    it('shows loading state when reconnecting', () => {
      render(<DaemonConnectionSection {...defaultProps} isReconnecting={true} />);
      expect(screen.getByRole('button', { name: 'Reconnecting...' })).toBeInTheDocument();
    });

    it('disables button in standalone mode', () => {
      render(<DaemonConnectionSection {...defaultProps} daemonMode="standalone" />);
      expect(screen.getByRole('button', { name: 'Reconnect to Daemon' })).toBeDisabled();
    });
  });
});
