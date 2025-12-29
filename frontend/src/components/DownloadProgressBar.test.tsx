import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import DownloadProgressBar from './DownloadProgressBar';
import { DownloadProvider } from '@/context/DownloadContext';

// Mock websocketClient
vi.mock('@/services/websocketClient', () => ({
  EventsOn: vi.fn(() => vi.fn()),
}));

// Mock useDownload to control the state
const mockUseDownload = vi.fn();
vi.mock('@/context/DownloadContext', async () => {
  const actual = await vi.importActual('@/context/DownloadContext');
  return {
    ...actual,
    useDownload: () => mockUseDownload(),
  };
});

describe('DownloadProgressBar', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should not render when not downloading', () => {
    mockUseDownload.mockReturnValue({
      state: { tasks: [], activeTask: null },
      isDownloading: false,
    });

    const { container } = render(<DownloadProgressBar />);
    expect(container.firstChild).toBeNull();
  });

  it('should render progress bar when downloading', () => {
    mockUseDownload.mockReturnValue({
      state: {
        tasks: [{ id: 'test-1', description: 'Downloading meta...', progress: 50, status: 'downloading' }],
        activeTask: { id: 'test-1', description: 'Downloading meta...', progress: 50, status: 'downloading' },
      },
      isDownloading: true,
    });

    render(<DownloadProgressBar />);
    expect(screen.getByText('Downloading meta...')).toBeInTheDocument();
    expect(screen.getByText('(50%)')).toBeInTheDocument();
  });

  it('should show error message when download fails', () => {
    mockUseDownload.mockReturnValue({
      state: {
        tasks: [{ id: 'test-1', description: 'Downloading meta...', progress: 30, status: 'error', error: 'Network error' }],
        activeTask: { id: 'test-1', description: 'Downloading meta...', progress: 30, status: 'error', error: 'Network error' },
      },
      isDownloading: false,
    });

    render(<DownloadProgressBar />);
    expect(screen.getByText('Network error')).toBeInTheDocument();
  });

  it('should not show percentage at 0%', () => {
    mockUseDownload.mockReturnValue({
      state: {
        tasks: [{ id: 'test-1', description: 'Starting download...', progress: 0, status: 'downloading' }],
        activeTask: { id: 'test-1', description: 'Starting download...', progress: 0, status: 'downloading' },
      },
      isDownloading: true,
    });

    render(<DownloadProgressBar />);
    expect(screen.getByText('Starting download...')).toBeInTheDocument();
    expect(screen.queryByText('(0%)')).not.toBeInTheDocument();
  });

  it('should not show percentage at 100%', () => {
    mockUseDownload.mockReturnValue({
      state: {
        tasks: [{ id: 'test-1', description: 'Finishing...', progress: 100, status: 'downloading' }],
        activeTask: { id: 'test-1', description: 'Finishing...', progress: 100, status: 'downloading' },
      },
      isDownloading: true,
    });

    render(<DownloadProgressBar />);
    expect(screen.getByText('Finishing...')).toBeInTheDocument();
    expect(screen.queryByText('(100%)')).not.toBeInTheDocument();
  });
});
