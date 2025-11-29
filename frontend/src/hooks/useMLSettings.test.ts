import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useMLSettings } from './useMLSettings';
import { mockWailsApp } from '../test/mocks/wailsApp';

// Mock showToast
vi.mock('../components/ToastContainer', () => ({
  showToast: {
    show: vi.fn(),
  },
}));

import { showToast } from '../components/ToastContainer';

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
    it('calls CheckOllamaStatus with endpoint and model', async () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleCheckOllamaStatus();
      });

      expect(mockWailsApp.CheckOllamaStatus).toHaveBeenCalledWith(
        'http://localhost:11434',
        'qwen3:8b'
      );
    });

    it('sets isCheckingOllama to true during check', async () => {
      let resolveCheck: (value: any) => void;
      mockWailsApp.CheckOllamaStatus.mockImplementationOnce(
        () => new Promise((resolve) => {
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
        resolveCheck!({ available: true, modelReady: true, version: '0.1.0', modelName: 'qwen3:8b' });
        await checkPromise;
      });

      expect(result.current.isCheckingOllama).toBe(false);
    });

    it('updates ollamaStatus on successful check', async () => {
      const mockStatus = {
        available: true,
        version: '0.2.0',
        modelReady: true,
        modelName: 'qwen3:8b',
        modelsLoaded: ['qwen3:8b', 'llama3.2:3b'],
        error: '',
      };
      mockWailsApp.CheckOllamaStatus.mockResolvedValueOnce(mockStatus);

      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleCheckOllamaStatus();
      });

      expect(result.current.ollamaStatus).toEqual(mockStatus);
    });

    it('shows success toast when Ollama is available and model is ready', async () => {
      mockWailsApp.CheckOllamaStatus.mockResolvedValueOnce({
        available: true,
        version: '0.1.0',
        modelReady: true,
        modelName: 'qwen3:8b',
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
      mockWailsApp.CheckOllamaStatus.mockResolvedValueOnce({
        available: true,
        version: '0.1.0',
        modelReady: false,
        modelName: 'qwen3:8b',
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
      mockWailsApp.CheckOllamaStatus.mockResolvedValueOnce({
        available: false,
        modelReady: false,
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

    it('shows error toast on API failure', async () => {
      mockWailsApp.CheckOllamaStatus.mockRejectedValueOnce(new Error('Network error'));

      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleCheckOllamaStatus();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Failed to check Ollama status'),
        'error'
      );
    });
  });

  describe('handleFetchModels', () => {
    it('calls GetAvailableOllamaModels with endpoint', async () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleFetchModels();
      });

      expect(mockWailsApp.GetAvailableOllamaModels).toHaveBeenCalledWith(
        'http://localhost:11434'
      );
    });

    it('updates availableModels on success', async () => {
      const mockModels = [
        { name: 'qwen3:8b', size: 1000 },
        { name: 'llama3.2:3b', size: 500 },
      ];
      mockWailsApp.GetAvailableOllamaModels.mockResolvedValueOnce(mockModels);

      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleFetchModels();
      });

      expect(result.current.availableModels).toEqual(mockModels);
    });

    it('sets empty array on failure', async () => {
      mockWailsApp.GetAvailableOllamaModels.mockRejectedValueOnce(new Error('Failed'));

      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleFetchModels();
      });

      expect(result.current.availableModels).toEqual([]);
    });
  });

  describe('handlePullModel', () => {
    it('shows warning toast when model is empty', async () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handlePullModel('');
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Please enter a model name'),
        'warning'
      );
      expect(mockWailsApp.PullOllamaModel).not.toHaveBeenCalled();
    });

    it('calls PullOllamaModel with endpoint and model', async () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handlePullModel('mistral:7b');
      });

      expect(mockWailsApp.PullOllamaModel).toHaveBeenCalledWith(
        'http://localhost:11434',
        'mistral:7b'
      );
    });

    it('sets isPullingModel to true during pull', async () => {
      let resolvePull: () => void;
      mockWailsApp.PullOllamaModel.mockImplementationOnce(
        () => new Promise<void>((resolve) => {
          resolvePull = resolve;
        })
      );

      const { result } = renderHook(() => useMLSettings(defaultProps));

      let pullPromise: Promise<void>;
      act(() => {
        pullPromise = result.current.handlePullModel('mistral:7b');
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
        await result.current.handlePullModel('mistral:7b');
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Successfully pulled'),
        'success'
      );
    });

    it('shows error toast on pull failure', async () => {
      mockWailsApp.PullOllamaModel.mockRejectedValueOnce(new Error('Pull failed'));

      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handlePullModel('mistral:7b');
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Failed to pull model'),
        'error'
      );
    });

    it('refreshes models and status after successful pull', async () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handlePullModel('mistral:7b');
      });

      expect(mockWailsApp.GetAvailableOllamaModels).toHaveBeenCalled();
      expect(mockWailsApp.CheckOllamaStatus).toHaveBeenCalled();
    });
  });

  describe('handleTestLLM', () => {
    it('calls TestLLMGeneration with endpoint and model', async () => {
      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleTestLLM();
      });

      expect(mockWailsApp.TestLLMGeneration).toHaveBeenCalledWith(
        'http://localhost:11434',
        'qwen3:8b'
      );
    });

    it('sets isTestingLLM to true during test', async () => {
      let resolveTest: (value: string) => void;
      mockWailsApp.TestLLMGeneration.mockImplementationOnce(
        () => new Promise<string>((resolve) => {
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
        resolveTest!('Hello from Ollama!');
        await testPromise;
      });

      expect(result.current.isTestingLLM).toBe(false);
    });

    it('updates llmTestResult on success', async () => {
      mockWailsApp.TestLLMGeneration.mockResolvedValueOnce('Test response');

      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleTestLLM();
      });

      expect(result.current.llmTestResult).toBe('Test response');
    });

    it('shows success toast on successful test', async () => {
      mockWailsApp.TestLLMGeneration.mockResolvedValueOnce('Test response');

      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleTestLLM();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('LLM test completed'),
        'success'
      );
    });

    it('shows error toast on test failure', async () => {
      mockWailsApp.TestLLMGeneration.mockRejectedValueOnce(new Error('Test failed'));

      const { result } = renderHook(() => useMLSettings(defaultProps));

      await act(async () => {
        await result.current.handleTestLLM();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('LLM test failed'),
        'error'
      );
      expect(result.current.llmTestResult).toBeNull();
    });
  });

  describe('with custom endpoint and model', () => {
    it('uses custom endpoint from props', async () => {
      const customProps = {
        ...defaultProps,
        ollamaEndpoint: 'http://custom-host:8080',
      };

      const { result } = renderHook(() => useMLSettings(customProps));

      await act(async () => {
        await result.current.handleCheckOllamaStatus();
      });

      expect(mockWailsApp.CheckOllamaStatus).toHaveBeenCalledWith(
        'http://custom-host:8080',
        'qwen3:8b'
      );
    });

    it('uses custom model from props', async () => {
      const customProps = {
        ...defaultProps,
        ollamaModel: 'llama3.2:3b',
      };

      const { result } = renderHook(() => useMLSettings(customProps));

      await act(async () => {
        await result.current.handleTestLLM();
      });

      expect(mockWailsApp.TestLLMGeneration).toHaveBeenCalledWith(
        'http://localhost:11434',
        'llama3.2:3b'
      );
    });
  });
});
