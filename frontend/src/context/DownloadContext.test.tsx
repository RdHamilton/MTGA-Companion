import { describe, it, expect, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { DownloadProvider, useDownload } from './DownloadContext';

// Mock websocketClient
vi.mock('@/services/websocketClient', () => ({
  EventsOn: vi.fn(() => vi.fn()),
}));

describe('DownloadContext', () => {
  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <DownloadProvider>{children}</DownloadProvider>
  );

  describe('startDownload', () => {
    it('should add a new download task', () => {
      const { result } = renderHook(() => useDownload(), { wrapper });

      act(() => {
        result.current.startDownload('test-1', 'Downloading test file...');
      });

      expect(result.current.state.tasks).toHaveLength(1);
      expect(result.current.state.tasks[0]).toMatchObject({
        id: 'test-1',
        description: 'Downloading test file...',
        progress: 0,
        status: 'downloading',
      });
      expect(result.current.isDownloading).toBe(true);
    });

    it('should update existing task with same id', () => {
      const { result } = renderHook(() => useDownload(), { wrapper });

      act(() => {
        result.current.startDownload('test-1', 'First download');
      });

      act(() => {
        result.current.startDownload('test-1', 'Updated download');
      });

      expect(result.current.state.tasks).toHaveLength(1);
      expect(result.current.state.tasks[0].description).toBe('Updated download');
    });
  });

  describe('updateProgress', () => {
    it('should update task progress', () => {
      const { result } = renderHook(() => useDownload(), { wrapper });

      act(() => {
        result.current.startDownload('test-1', 'Downloading...');
      });

      act(() => {
        result.current.updateProgress('test-1', 50);
      });

      expect(result.current.state.tasks[0].progress).toBe(50);
    });

    it('should clamp progress between 0 and 100', () => {
      const { result } = renderHook(() => useDownload(), { wrapper });

      act(() => {
        result.current.startDownload('test-1', 'Downloading...');
      });

      act(() => {
        result.current.updateProgress('test-1', 150);
      });

      expect(result.current.state.tasks[0].progress).toBe(100);

      act(() => {
        result.current.updateProgress('test-1', -10);
      });

      expect(result.current.state.tasks[0].progress).toBe(0);
    });
  });

  describe('completeDownload', () => {
    it('should remove task when complete', () => {
      const { result } = renderHook(() => useDownload(), { wrapper });

      act(() => {
        result.current.startDownload('test-1', 'Downloading...');
      });

      expect(result.current.state.tasks).toHaveLength(1);

      act(() => {
        result.current.completeDownload('test-1');
      });

      expect(result.current.state.tasks).toHaveLength(0);
      expect(result.current.isDownloading).toBe(false);
    });
  });

  describe('failDownload', () => {
    it('should mark task as error', () => {
      const { result } = renderHook(() => useDownload(), { wrapper });

      act(() => {
        result.current.startDownload('test-1', 'Downloading...');
      });

      act(() => {
        result.current.failDownload('test-1', 'Network error');
      });

      expect(result.current.state.tasks[0].status).toBe('error');
      expect(result.current.state.tasks[0].error).toBe('Network error');
    });
  });

  describe('cancelDownload', () => {
    it('should remove task', () => {
      const { result } = renderHook(() => useDownload(), { wrapper });

      act(() => {
        result.current.startDownload('test-1', 'Downloading...');
      });

      act(() => {
        result.current.cancelDownload('test-1');
      });

      expect(result.current.state.tasks).toHaveLength(0);
    });
  });

  describe('overallProgress', () => {
    it('should calculate average progress across tasks', () => {
      const { result } = renderHook(() => useDownload(), { wrapper });

      act(() => {
        result.current.startDownload('test-1', 'Download 1');
        result.current.startDownload('test-2', 'Download 2');
      });

      act(() => {
        result.current.updateProgress('test-1', 50);
        result.current.updateProgress('test-2', 100);
      });

      expect(result.current.overallProgress).toBe(75);
    });
  });
});
