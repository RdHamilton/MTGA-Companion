import { useState, useCallback } from 'react';
import {
  CheckOllamaStatus,
  GetAvailableOllamaModels,
  PullOllamaModel,
  TestLLMGeneration,
} from '@/services/api/legacy';
import { showToast } from '../components/ToastContainer';
import type { gui } from '@/types/models';

export interface UseMLSettingsProps {
  /** ML enabled state from parent */
  mlEnabled: boolean;
  /** LLM enabled state from parent */
  llmEnabled: boolean;
  /** Ollama endpoint from parent */
  ollamaEndpoint: string;
  /** Ollama model from parent */
  ollamaModel: string;
  /** MTGGoldfish enabled state from parent */
  metaGoldfishEnabled: boolean;
  /** MTGTop8 enabled state from parent */
  metaTop8Enabled: boolean;
  /** Meta weight from parent */
  metaWeight: number;
  /** Personal weight from parent */
  personalWeight: number;
  /** Handler for ML enabled change */
  onMLEnabledChange: (enabled: boolean) => void;
  /** Handler for LLM enabled change */
  onLLMEnabledChange: (enabled: boolean) => void;
  /** Handler for Ollama endpoint change */
  onOllamaEndpointChange: (endpoint: string) => void;
  /** Handler for Ollama model change */
  onOllamaModelChange: (model: string) => void;
  /** Handler for MTGGoldfish enabled change */
  onMetaGoldfishEnabledChange: (enabled: boolean) => void;
  /** Handler for MTGTop8 enabled change */
  onMetaTop8EnabledChange: (enabled: boolean) => void;
  /** Handler for meta weight change */
  onMetaWeightChange: (weight: number) => void;
  /** Handler for personal weight change */
  onPersonalWeightChange: (weight: number) => void;
}

export interface UseMLSettingsReturn {
  /** Whether Ollama status is being checked */
  isCheckingOllama: boolean;
  /** Current Ollama status */
  ollamaStatus: gui.OllamaStatus | null;
  /** Available Ollama models */
  availableModels: gui.OllamaModel[];
  /** Whether a model is being pulled */
  isPullingModel: boolean;
  /** Whether LLM test is running */
  isTestingLLM: boolean;
  /** LLM test result */
  llmTestResult: string | null;
  /** Check Ollama status */
  handleCheckOllamaStatus: () => Promise<void>;
  /** Fetch available models */
  handleFetchModels: () => Promise<void>;
  /** Pull a model */
  handlePullModel: (model: string) => Promise<void>;
  /** Test LLM generation */
  handleTestLLM: () => Promise<void>;
}

export function useMLSettings(props: UseMLSettingsProps): UseMLSettingsReturn {
  const {
    ollamaEndpoint,
    ollamaModel,
  } = props;

  const [isCheckingOllama, setIsCheckingOllama] = useState(false);
  const [ollamaStatus, setOllamaStatus] = useState<gui.OllamaStatus | null>(null);
  const [availableModels, setAvailableModels] = useState<gui.OllamaModel[]>([]);
  const [isPullingModel, setIsPullingModel] = useState(false);
  const [isTestingLLM, setIsTestingLLM] = useState(false);
  const [llmTestResult, setLlmTestResult] = useState<string | null>(null);

  const handleCheckOllamaStatus = useCallback(async () => {
    setIsCheckingOllama(true);
    setOllamaStatus(null);

    try {
      const status = await CheckOllamaStatus(ollamaEndpoint, ollamaModel);
      setOllamaStatus(status);

      if (status.available && status.modelReady) {
        showToast.show(
          `Ollama is ready! Version: ${status.version}, Model: ${status.modelName}`,
          'success'
        );
      } else if (status.available && !status.modelReady) {
        showToast.show(
          `Ollama is running but model "${ollamaModel}" is not available. ${status.error || ''}`,
          'warning'
        );
      } else {
        showToast.show(
          `Ollama is not available: ${status.error || 'Unknown error'}`,
          'error'
        );
      }
    } catch (error) {
      showToast.show(`Failed to check Ollama status: ${error}`, 'error');
    } finally {
      setIsCheckingOllama(false);
    }
  }, [ollamaEndpoint, ollamaModel]);

  const handleFetchModels = useCallback(async () => {
    try {
      const models = await GetAvailableOllamaModels(ollamaEndpoint);
      setAvailableModels(models);
    } catch (error) {
      console.error('Failed to fetch available models:', error);
      setAvailableModels([]);
    }
  }, [ollamaEndpoint]);

  const handlePullModel = useCallback(async (model: string) => {
    if (!model) {
      showToast.show('Please enter a model name', 'warning');
      return;
    }

    setIsPullingModel(true);
    showToast.show(
      `Pulling model "${model}"... This may take several minutes.`,
      'info'
    );

    try {
      await PullOllamaModel(ollamaEndpoint, model);
      showToast.show(`Successfully pulled model "${model}"!`, 'success');
      // Refresh models list after pulling
      await handleFetchModels();
      // Re-check status with new model
      const status = await CheckOllamaStatus(ollamaEndpoint, model);
      setOllamaStatus(status);
    } catch (error) {
      showToast.show(`Failed to pull model: ${error}`, 'error');
    } finally {
      setIsPullingModel(false);
    }
  }, [ollamaEndpoint, handleFetchModels]);

  const handleTestLLM = useCallback(async () => {
    setIsTestingLLM(true);
    setLlmTestResult(null);

    try {
      const result = await TestLLMGeneration(ollamaEndpoint, ollamaModel);
      setLlmTestResult(result);
      showToast.show('LLM test completed successfully!', 'success');
    } catch (error) {
      showToast.show(`LLM test failed: ${error}`, 'error');
      setLlmTestResult(null);
    } finally {
      setIsTestingLLM(false);
    }
  }, [ollamaEndpoint, ollamaModel]);

  return {
    isCheckingOllama,
    ollamaStatus,
    availableModels,
    isPullingModel,
    isTestingLLM,
    llmTestResult,
    handleCheckOllamaStatus,
    handleFetchModels,
    handlePullModel,
    handleTestLLM,
  };
}
