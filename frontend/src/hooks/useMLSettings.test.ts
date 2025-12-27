import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useMLSettings } from './useMLSettings';

// Mock the API modules
vi.mock('@/services/api', () => ({
  system: {
    checkOllamaStatus: vi.fn(),
    getAvailableOllamaModels: vi.fn(),
    pullOllamaModel: vi.fn(),
    testLLMGeneration: vi.fn(),
  },
}));

// Mock showToast
vi.mock('../components/ToastContainer', () => ({
  showToast: {
    show: vi.fn(),
  },
}));

import { system } from '@/services/api';
import { showToast } from '../components/ToastContainer';

const mockCheckOllamaStatus = vi.mocked(system.checkOllamaStatus);
const mockGetAvailableOllamaModels = vi.mocked(system.getAvailableOllamaModels);
const mockPullOllamaModel = vi.mocked(system.pullOllamaModel);
const mockTestLLMGeneration = vi.mocked(system.testLLMGeneration);

const defaultProps = {
  mlEnabled: true,
  llmEnabled: true,
  ollamaEndpoint: 'http://localhost:11434',
  ollamaModel: 'qwen3:8b',
  metaGoldfishEnabled: true,
  metaTop8Enabled: true,
  metaWeight: 0.3,
  personalWeight: 0.2,
  onMLEnabledChange: vi.fn(),
  onLLMEnabledChange: vi.fn(),
  onOllamaEndpointChange: vi.fn(),
  onOllamaModelChange: vi.fn(),
  onMetaGoldfishEnabledChange: vi.fn(),
  onMetaTop8EnabledChange: vi.fn(),
  onMetaWeightChange: vi.fn(),
  onPersonalWeightChange: vi.fn(),
};

describe('useMLSettings', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCheckOllamaStatus.mockResolvedValue({
      available: true,
      modelReady: true,
      version: '0.1.0',
      modelName: 'qwen3:8b',
      error: '',
    });
    mockGetAvailableOllamaModels.mockResolvedValue([]);
    mockPullOllamaModel.mockResolvedValue(undefined);
    mockTestLLMGeneration.mockResolvedValue('Test response');
  });

  describe('initial state', () => {
    it('returns isCheckingOllama as false', () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));
      expect(result.current.isCheckingOllama).toBe(false);
    });

    it('returns ollamaStatus as null', () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));
      expect(result.current.ollamaStatus).toBeNull();
    });

    it('returns empty availableModels', () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));
      expect(result.current.availableModels).toEqual([]);
    });

    it('returns isPullingModel as false', () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));
      expect(result.current.isPullingModel).toBe(false);
    });

    it('returns isTestingLLM as false', () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));
      expect(result.current.isTestingLLM).toBe(false);
    });

    it('returns llmTestResult as null', () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));
      expect(result.current.llmTestResult).toBeNull();
    });
  });

  describe('handleCheckOllamaStatus', () => {
    it('calls checkOllamaStatus with endpoint and model', async () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleCheckOllamaStatus();
      });

      expect(mockCheckOllamaStatus).toHaveBeenCalledWith(
        'http://localhost:11434',
        'qwen3:8b'
      );
    });

    it('sets isCheckingOllama to true during check', async () => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      let resolveCheck: (value: any) => void;
      mockCheckOllamaStatus.mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveCheck = resolve;
          })
      );

      const { result } = renderHook(() => useMLSettings(defaultProps));

      let checkPromise: Promise<void>;
      act(() => {
        checkPromise = result.current.handleCheckOllamaStatus();
      });

      expect(result.current.isCheckingOllama).toBe(true);

      await act(async () => {
        resolveCheck!({
          available: true,
          modelReady: true,
          version: '0.1.0',
          modelName: 'qwen3:8b',
          error: '',
        });
        await checkPromise;
      });

      expect(result.current.isCheckingOllama).toBe(false);
    });

    it('updates ollamaStatus on successful check', async () => {
      const mockStatus = {
        available: true,
        modelReady: true,
        version: '0.2.0',
        modelName: 'qwen3:8b',
        error: '',
      };
      mockCheckOllamaStatus.mockResolvedValueOnce(mockStatus);

      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleCheckOllamaStatus();
      });

      expect(result.current.ollamaStatus).toEqual(mockStatus);
    });

    it('shows success toast when Ollama is available and model is ready', async () => {
      mockCheckOllamaStatus.mockResolvedValueOnce({
        available: true,
        modelReady: true,
        version: '0.2.0',
        modelName: 'qwen3:8b',
        error: '',
      });

      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleCheckOllamaStatus();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Ollama is ready'),
        'success'
      );
    });

    it('shows warning toast when Ollama is available but model is not ready', async () => {
      mockCheckOllamaStatus.mockResolvedValueOnce({
        available: true,
        modelReady: false,
        version: '0.2.0',
        modelName: '',
        error: 'Model not found',
      });

      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleCheckOllamaStatus();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('model'),
        'warning'
      );
    });

    it('shows error toast when Ollama is not available', async () => {
      mockCheckOllamaStatus.mockResolvedValueOnce({
        available: false,
        modelReady: false,
        version: '',
        modelName: '',
        error: 'Connection refused',
      });

      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleCheckOllamaStatus();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('not available'),
        'error'
      );
    });
  });

  describe('handleFetchModels', () => {
    it('calls getAvailableOllamaModels with endpoint', async () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleFetchModels();
      });

      expect(mockGetAvailableOllamaModels).toHaveBeenCalledWith('http://localhost:11434');
    });

    it('updates availableModels on success', async () => {
      const mockModels = [
        { name: 'llama2', size: 1000, modified: new Date().toISOString() },
        { name: 'qwen3:8b', size: 2000, modified: new Date().toISOString() },
      ];
      mockGetAvailableOllamaModels.mockResolvedValueOnce(mockModels);

      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleFetchModels();
      });

      expect(result.current.availableModels).toEqual(mockModels);
    });
  });

  describe('handlePullModel', () => {
    it('calls pullOllamaModel with endpoint and model', async () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handlePullModel('llama2');
      });

      expect(mockPullOllamaModel).toHaveBeenCalledWith('http://localhost:11434', 'llama2');
    });

    it('sets isPullingModel to true during pull', async () => {
      let resolvePull: () => void;
      mockPullOllamaModel.mockImplementationOnce(
        () =>
          new Promise<void>((resolve) => {
            resolvePull = resolve;
          })
      );

      const { result } = renderHook(() => useMLSettings(defaultProps));

      let pullPromise: Promise<void>;
      act(() => {
        pullPromise = result.current.handlePullModel('llama2');
      });

      expect(result.current.isPullingModel).toBe(true);

      await act(async () => {
        resolvePull!();
        await pullPromise;
      });

      expect(result.current.isPullingModel).toBe(false);
    });

    it('shows success toast on successful pull', async () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handlePullModel('llama2');
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Successfully pulled'),
        'success'
      );
    });

    it('refreshes models and status after successful pull', async () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handlePullModel('llama2');
      });

      // Should fetch models after pull
      expect(mockGetAvailableOllamaModels).toHaveBeenCalled();
      // Should re-check status with new model
      expect(mockCheckOllamaStatus).toHaveBeenCalledWith('http://localhost:11434', 'llama2');
    });

    it('shows error toast on pull failure', async () => {
      mockPullOllamaModel.mockRejectedValueOnce(new Error('Pull failed'));

      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handlePullModel('llama2');
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Failed to pull'),
        'error'
      );
    });

    it('shows warning toast for empty model name', async () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handlePullModel('');
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('enter a model name'),
        'warning'
      );
      expect(mockPullOllamaModel).not.toHaveBeenCalled();
    });
  });

  describe('handleTestLLM', () => {
    it('calls testLLMGeneration with endpoint and model', async () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleTestLLM();
      });

      expect(mockTestLLMGeneration).toHaveBeenCalledWith('http://localhost:11434', 'qwen3:8b');
    });

    it('sets isTestingLLM to true during test', async () => {
      let resolveTest: (value: string) => void;
      mockTestLLMGeneration.mockImplementationOnce(
        () =>
          new Promise<string>((resolve) => {
            resolveTest = resolve;
          })
      );

      const { result } = renderHook(() => useMLSettings(defaultProps));

      let testPromise: Promise<void>;
      act(() => {
        testPromise = result.current.handleTestLLM();
      });

      expect(result.current.isTestingLLM).toBe(true);

      await act(async () => {
        resolveTest!('Test result');
        await testPromise;
      });

      expect(result.current.isTestingLLM).toBe(false);
    });

    it('updates llmTestResult on success', async () => {
      mockTestLLMGeneration.mockResolvedValueOnce('Generated text here');

      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleTestLLM();
      });

      expect(result.current.llmTestResult).toBe('Generated text here');
    });

    it('shows success toast on successful test', async () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleTestLLM();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('test completed'),
        'success'
      );
    });

    it('shows error toast on test failure', async () => {
      mockTestLLMGeneration.mockRejectedValueOnce(new Error('Generation failed'));

      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleTestLLM();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('test failed'),
        'error'
      );
    });

    it('clears llmTestResult on failure', async () => {
      mockTestLLMGeneration.mockRejectedValueOnce(new Error('Generation failed'));

      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleTestLLM();
      });

      expect(result.current.llmTestResult).toBeNull();
    });
  });
});
