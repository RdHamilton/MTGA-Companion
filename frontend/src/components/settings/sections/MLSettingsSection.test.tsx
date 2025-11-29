import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MLSettingsSection } from './MLSettingsSection';

// Mock the hooks
vi.mock('../../../hooks', () => ({
  useMLSettings: vi.fn(() => ({
    isCheckingOllama: false,
    ollamaStatus: null,
    availableModels: [],
    isPullingModel: false,
    isTestingLLM: false,
    llmTestResult: null,
    handleCheckOllamaStatus: vi.fn(),
    handleFetchModels: vi.fn(),
    handlePullModel: vi.fn(),
    handleTestLLM: vi.fn(),
  })),
}));

import { useMLSettings } from '../../../hooks';

const defaultProps = {
  mlEnabled: true,
  onMLEnabledChange: vi.fn(),
  llmEnabled: true,
  onLLMEnabledChange: vi.fn(),
  ollamaEndpoint: 'http://localhost:11434',
  onOllamaEndpointChange: vi.fn(),
  ollamaModel: 'qwen3:8b',
  onOllamaModelChange: vi.fn(),
  metaGoldfishEnabled: true,
  onMetaGoldfishEnabledChange: vi.fn(),
  metaTop8Enabled: true,
  onMetaTop8EnabledChange: vi.fn(),
  metaWeight: 0.3,
  onMetaWeightChange: vi.fn(),
  personalWeight: 0.2,
  onPersonalWeightChange: vi.fn(),
};

describe('MLSettingsSection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('rendering', () => {
    it('renders the section title', () => {
      render(<MLSettingsSection {...defaultProps} />);
      expect(screen.getByText('ML / AI Recommendations')).toBeInTheDocument();
    });

    it('renders the ML enabled toggle', () => {
      render(<MLSettingsSection {...defaultProps} />);
      expect(screen.getByText('Enable ML Recommendations')).toBeInTheDocument();
    });

    it('shows meta data sources section when ML is enabled', () => {
      render(<MLSettingsSection {...defaultProps} mlEnabled={true} />);
      expect(screen.getByText('Meta Data Sources')).toBeInTheDocument();
      expect(screen.getByText('MTGGoldfish Metagame Data')).toBeInTheDocument();
      expect(screen.getByText('MTGTop8 Tournament Results')).toBeInTheDocument();
    });

    it('shows recommendation weights section when ML is enabled', () => {
      render(<MLSettingsSection {...defaultProps} mlEnabled={true} />);
      expect(screen.getByText('Recommendation Weights')).toBeInTheDocument();
      expect(screen.getByText('Meta Weight')).toBeInTheDocument();
      expect(screen.getByText('Personal History Weight')).toBeInTheDocument();
    });

    it('shows LLM section when ML is enabled', () => {
      render(<MLSettingsSection {...defaultProps} mlEnabled={true} />);
      expect(screen.getByText('AI Explanations (Ollama)')).toBeInTheDocument();
      expect(screen.getByText('Enable LLM Explanations')).toBeInTheDocument();
    });

    it('hides all subsections when ML is disabled', () => {
      render(<MLSettingsSection {...defaultProps} mlEnabled={false} />);
      expect(screen.queryByText('Meta Data Sources')).not.toBeInTheDocument();
      expect(screen.queryByText('Recommendation Weights')).not.toBeInTheDocument();
      expect(screen.queryByText('AI Explanations (Ollama)')).not.toBeInTheDocument();
    });
  });

  describe('ML toggle', () => {
    it('calls onMLEnabledChange when toggled', () => {
      render(<MLSettingsSection {...defaultProps} mlEnabled={false} />);

      const checkbox = screen.getByRole('checkbox', { name: /enable ml recommendations/i });
      fireEvent.click(checkbox);

      expect(defaultProps.onMLEnabledChange).toHaveBeenCalledWith(true);
    });
  });

  describe('meta data sources', () => {
    it('calls onMetaGoldfishEnabledChange when toggled', () => {
      render(<MLSettingsSection {...defaultProps} />);

      const checkbox = screen.getByRole('checkbox', { name: /mtggoldfish metagame data/i });
      fireEvent.click(checkbox);

      expect(defaultProps.onMetaGoldfishEnabledChange).toHaveBeenCalledWith(false);
    });

    it('calls onMetaTop8EnabledChange when toggled', () => {
      render(<MLSettingsSection {...defaultProps} />);

      const checkbox = screen.getByRole('checkbox', { name: /mtgtop8 tournament results/i });
      fireEvent.click(checkbox);

      expect(defaultProps.onMetaTop8EnabledChange).toHaveBeenCalledWith(false);
    });
  });

  describe('weight sliders', () => {
    it('displays current meta weight value', () => {
      render(<MLSettingsSection {...defaultProps} metaWeight={0.5} />);
      expect(screen.getByText('0.5')).toBeInTheDocument();
    });

    it('displays current personal weight value', () => {
      render(<MLSettingsSection {...defaultProps} personalWeight={0.4} />);
      expect(screen.getByText('0.4')).toBeInTheDocument();
    });
  });

  describe('LLM settings', () => {
    it('shows Ollama endpoint input when LLM is enabled', () => {
      render(<MLSettingsSection {...defaultProps} llmEnabled={true} />);
      expect(screen.getByText('Ollama Endpoint')).toBeInTheDocument();
      expect(screen.getByDisplayValue('http://localhost:11434')).toBeInTheDocument();
    });

    it('shows model selector when LLM is enabled', () => {
      render(<MLSettingsSection {...defaultProps} llmEnabled={true} />);
      expect(screen.getByText('Model')).toBeInTheDocument();
    });

    it('shows check status button when LLM is enabled', () => {
      render(<MLSettingsSection {...defaultProps} llmEnabled={true} />);
      expect(screen.getByText('Check Status')).toBeInTheDocument();
    });

    it('shows test generation button when LLM is enabled', () => {
      render(<MLSettingsSection {...defaultProps} llmEnabled={true} />);
      expect(screen.getByText('Test Generation')).toBeInTheDocument();
    });

    it('hides Ollama settings when LLM is disabled', () => {
      render(<MLSettingsSection {...defaultProps} llmEnabled={false} />);
      expect(screen.queryByText('Ollama Endpoint')).not.toBeInTheDocument();
      expect(screen.queryByText('Check Status')).not.toBeInTheDocument();
    });

    it('calls onOllamaEndpointChange when endpoint is modified', () => {
      render(<MLSettingsSection {...defaultProps} llmEnabled={true} />);

      const input = screen.getByDisplayValue('http://localhost:11434');
      fireEvent.change(input, { target: { value: 'http://new-host:8080' } });

      expect(defaultProps.onOllamaEndpointChange).toHaveBeenCalledWith('http://new-host:8080');
    });
  });

  describe('Ollama status display', () => {
    it('shows success status when Ollama is available and model is ready', () => {
      vi.mocked(useMLSettings).mockReturnValue({
        isCheckingOllama: false,
        ollamaStatus: {
          available: true,
          version: '0.1.0',
          modelReady: true,
          modelName: 'qwen3:8b',
          modelsLoaded: ['qwen3:8b'],
          error: '',
        },
        availableModels: [],
        isPullingModel: false,
        isTestingLLM: false,
        llmTestResult: null,
        handleCheckOllamaStatus: vi.fn(),
        handleFetchModels: vi.fn(),
        handlePullModel: vi.fn(),
        handleTestLLM: vi.fn(),
      });

      render(<MLSettingsSection {...defaultProps} llmEnabled={true} />);
      expect(screen.getByText(/connected.*v0.1.0/i)).toBeInTheDocument();
    });

    it('shows warning status when Ollama is available but model is not ready', () => {
      vi.mocked(useMLSettings).mockReturnValue({
        isCheckingOllama: false,
        ollamaStatus: {
          available: true,
          version: '0.1.0',
          modelReady: false,
          modelName: 'qwen3:8b',
          modelsLoaded: [],
          error: 'Model not found',
        },
        availableModels: [],
        isPullingModel: false,
        isTestingLLM: false,
        llmTestResult: null,
        handleCheckOllamaStatus: vi.fn(),
        handleFetchModels: vi.fn(),
        handlePullModel: vi.fn(),
        handleTestLLM: vi.fn(),
      });

      render(<MLSettingsSection {...defaultProps} llmEnabled={true} />);
      expect(screen.getByText(/connected but model not available/i)).toBeInTheDocument();
    });

    it('shows error status when Ollama is not available', () => {
      vi.mocked(useMLSettings).mockReturnValue({
        isCheckingOllama: false,
        ollamaStatus: {
          available: false,
          version: '',
          modelReady: false,
          modelName: '',
          modelsLoaded: [],
          error: 'Connection refused',
        },
        availableModels: [],
        isPullingModel: false,
        isTestingLLM: false,
        llmTestResult: null,
        handleCheckOllamaStatus: vi.fn(),
        handleFetchModels: vi.fn(),
        handlePullModel: vi.fn(),
        handleTestLLM: vi.fn(),
      });

      render(<MLSettingsSection {...defaultProps} llmEnabled={true} />);
      expect(screen.getByText(/not available/i)).toBeInTheDocument();
    });

    it('shows LLM test result when available', () => {
      vi.mocked(useMLSettings).mockReturnValue({
        isCheckingOllama: false,
        ollamaStatus: null,
        availableModels: [],
        isPullingModel: false,
        isTestingLLM: false,
        llmTestResult: 'Hello from Ollama!',
        handleCheckOllamaStatus: vi.fn(),
        handleFetchModels: vi.fn(),
        handlePullModel: vi.fn(),
        handleTestLLM: vi.fn(),
      });

      render(<MLSettingsSection {...defaultProps} llmEnabled={true} />);
      expect(screen.getByText(/hello from ollama/i)).toBeInTheDocument();
    });

    it('shows pull model button when Ollama is available but model is not ready', () => {
      vi.mocked(useMLSettings).mockReturnValue({
        isCheckingOllama: false,
        ollamaStatus: {
          available: true,
          version: '0.1.0',
          modelReady: false,
          modelName: 'qwen3:8b',
          modelsLoaded: [],
          error: 'Model not found',
        },
        availableModels: [],
        isPullingModel: false,
        isTestingLLM: false,
        llmTestResult: null,
        handleCheckOllamaStatus: vi.fn(),
        handleFetchModels: vi.fn(),
        handlePullModel: vi.fn(),
        handleTestLLM: vi.fn(),
      });

      render(<MLSettingsSection {...defaultProps} llmEnabled={true} />);
      // Use getByRole to find the button specifically
      expect(screen.getByRole('button', { name: /pull model/i })).toBeInTheDocument();
    });
  });

  describe('button interactions', () => {
    it('calls handleCheckOllamaStatus when Check Status is clicked', () => {
      const mockHandleCheck = vi.fn();
      vi.mocked(useMLSettings).mockReturnValue({
        isCheckingOllama: false,
        ollamaStatus: null,
        availableModels: [],
        isPullingModel: false,
        isTestingLLM: false,
        llmTestResult: null,
        handleCheckOllamaStatus: mockHandleCheck,
        handleFetchModels: vi.fn(),
        handlePullModel: vi.fn(),
        handleTestLLM: vi.fn(),
      });

      render(<MLSettingsSection {...defaultProps} llmEnabled={true} />);

      fireEvent.click(screen.getByText('Check Status'));
      expect(mockHandleCheck).toHaveBeenCalled();
    });

    it('calls handleTestLLM when Test Generation is clicked', () => {
      const mockHandleTest = vi.fn();
      vi.mocked(useMLSettings).mockReturnValue({
        isCheckingOllama: false,
        ollamaStatus: { available: true, modelReady: true, version: '0.1.0', modelName: 'qwen3:8b', modelsLoaded: [], error: '' },
        availableModels: [],
        isPullingModel: false,
        isTestingLLM: false,
        llmTestResult: null,
        handleCheckOllamaStatus: vi.fn(),
        handleFetchModels: vi.fn(),
        handlePullModel: vi.fn(),
        handleTestLLM: mockHandleTest,
      });

      render(<MLSettingsSection {...defaultProps} llmEnabled={true} />);

      fireEvent.click(screen.getByText('Test Generation'));
      expect(mockHandleTest).toHaveBeenCalled();
    });

    it('disables Test Generation button when Ollama is not ready', () => {
      vi.mocked(useMLSettings).mockReturnValue({
        isCheckingOllama: false,
        ollamaStatus: { available: false, modelReady: false, version: '', modelName: '', modelsLoaded: [], error: '' },
        availableModels: [],
        isPullingModel: false,
        isTestingLLM: false,
        llmTestResult: null,
        handleCheckOllamaStatus: vi.fn(),
        handleFetchModels: vi.fn(),
        handlePullModel: vi.fn(),
        handleTestLLM: vi.fn(),
      });

      render(<MLSettingsSection {...defaultProps} llmEnabled={true} />);

      // Find the button by role, as the text might be wrapped in a span
      const testButton = screen.getByRole('button', { name: /test generation/i });
      expect(testButton).toBeDisabled();
    });
  });

  describe('info box', () => {
    it('shows info about ML recommendations when ML is enabled', () => {
      render(<MLSettingsSection {...defaultProps} mlEnabled={true} />);
      expect(screen.getByText(/about ml recommendations/i)).toBeInTheDocument();
      expect(screen.getByText(/all processing is done locally/i)).toBeInTheDocument();
    });
  });
});
