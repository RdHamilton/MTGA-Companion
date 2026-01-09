import { useState } from 'react';
import LoadingButton from '../../LoadingButton';
import { SettingItem, SettingToggle, SettingSelect } from '../';
import { useMLSettings } from '../../../hooks';
import { clearLearnedData } from '../../../services/api/mlSuggestions';
import type { gui } from '@/types/models';

export interface MLSettingsSectionProps {
  mlEnabled: boolean;
  onMLEnabledChange: (enabled: boolean) => void;
  llmEnabled: boolean;
  onLLMEnabledChange: (enabled: boolean) => void;
  ollamaEndpoint: string;
  onOllamaEndpointChange: (endpoint: string) => void;
  ollamaModel: string;
  onOllamaModelChange: (model: string) => void;
  metaGoldfishEnabled: boolean;
  onMetaGoldfishEnabledChange: (enabled: boolean) => void;
  metaTop8Enabled: boolean;
  onMetaTop8EnabledChange: (enabled: boolean) => void;
  metaWeight: number;
  onMetaWeightChange: (weight: number) => void;
  personalWeight: number;
  onPersonalWeightChange: (weight: number) => void;
  // ML Suggestion Preferences
  suggestionFrequency: string;
  onSuggestionFrequencyChange: (frequency: string) => void;
  minimumConfidence: number;
  onMinimumConfidenceChange: (confidence: number) => void;
  showCardAdditions: boolean;
  onShowCardAdditionsChange: (show: boolean) => void;
  showCardRemovals: boolean;
  onShowCardRemovalsChange: (show: boolean) => void;
  showArchetypeChanges: boolean;
  onShowArchetypeChangesChange: (show: boolean) => void;
  learnFromMatches: boolean;
  onLearnFromMatchesChange: (learn: boolean) => void;
  learnFromDeckChanges: boolean;
  onLearnFromDeckChangesChange: (learn: boolean) => void;
  retentionDays: number;
  onRetentionDaysChange: (days: number) => void;
  maxSuggestionsPerView: number;
  onMaxSuggestionsPerViewChange: (max: number) => void;
}

export function MLSettingsSection(props: MLSettingsSectionProps) {
  const {
    mlEnabled,
    onMLEnabledChange,
    llmEnabled,
    onLLMEnabledChange,
    ollamaEndpoint,
    onOllamaEndpointChange,
    ollamaModel,
    onOllamaModelChange,
    metaGoldfishEnabled,
    onMetaGoldfishEnabledChange,
    metaTop8Enabled,
    onMetaTop8EnabledChange,
    metaWeight,
    onMetaWeightChange,
    personalWeight,
    onPersonalWeightChange,
    // ML Suggestion Preferences
    suggestionFrequency,
    onSuggestionFrequencyChange,
    minimumConfidence,
    onMinimumConfidenceChange,
    showCardAdditions,
    onShowCardAdditionsChange,
    showCardRemovals,
    onShowCardRemovalsChange,
    showArchetypeChanges,
    onShowArchetypeChangesChange,
    learnFromMatches,
    onLearnFromMatchesChange,
    learnFromDeckChanges,
    onLearnFromDeckChangesChange,
    retentionDays,
    onRetentionDaysChange,
    maxSuggestionsPerView,
    onMaxSuggestionsPerViewChange,
  } = props;

  const [isClearingData, setIsClearingData] = useState(false);
  const [clearDataMessage, setClearDataMessage] = useState<string | null>(null);

  const {
    isCheckingOllama,
    ollamaStatus,
    availableModels,
    isPullingModel,
    isTestingLLM,
    llmTestResult,
    handleCheckOllamaStatus,
    handlePullModel,
    handleTestLLM,
  } = useMLSettings({
    mlEnabled,
    llmEnabled,
    ollamaEndpoint,
    ollamaModel,
    metaGoldfishEnabled,
    metaTop8Enabled,
    metaWeight,
    personalWeight,
    onMLEnabledChange,
    onLLMEnabledChange,
    onOllamaEndpointChange,
    onOllamaModelChange,
    onMetaGoldfishEnabledChange,
    onMetaTop8EnabledChange,
    onMetaWeightChange,
    onPersonalWeightChange,
  });

  const getStatusIcon = (status: gui.OllamaStatus | null) => {
    if (!status) return null;
    if (status.available && status.modelReady) return '✓';
    if (status.available) return '⚠';
    return '✗';
  };

  const getStatusClass = (status: gui.OllamaStatus | null) => {
    if (!status) return '';
    if (status.available && status.modelReady) return 'success';
    if (status.available) return 'warning';
    return 'error';
  };

  const handleClearLearnedData = async () => {
    if (!window.confirm('Are you sure you want to clear all learned ML data? This cannot be undone.')) {
      return;
    }
    setIsClearingData(true);
    setClearDataMessage(null);
    try {
      await clearLearnedData();
      setClearDataMessage('All learned data has been cleared successfully.');
    } catch (error) {
      setClearDataMessage(`Failed to clear data: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      setIsClearingData(false);
    }
  };

  return (
    <div className="settings-section">
      <h2 className="section-title">ML / AI Recommendations</h2>
      <div className="setting-description settings-section-description">
        Configure machine learning-powered card recommendations and natural language explanations.
        These features enhance deck building with personalized suggestions based on your play style
        and current metagame data.
      </div>

      {/* ML Recommendations Toggle */}
      <SettingToggle
        label="Enable ML Recommendations"
        description="Use machine learning to provide personalized card recommendations based on your play history and deck composition"
        checked={mlEnabled}
        onChange={onMLEnabledChange}
      />

      {mlEnabled && (
        <>
          {/* Meta Data Sources Section */}
          <div className="settings-subsection">
            <h3 className="subsection-title">Meta Data Sources</h3>
            <div className="setting-description">
              Configure which external data sources to use for metagame-aware recommendations.
            </div>

            <SettingToggle
              label="MTGGoldfish Metagame Data"
              description="Include deck archetypes and meta shares from MTGGoldfish for constructed formats"
              checked={metaGoldfishEnabled}
              onChange={onMetaGoldfishEnabledChange}
            />

            <SettingToggle
              label="MTGTop8 Tournament Results"
              description="Include tournament performance data and winning decklists from MTGTop8"
              checked={metaTop8Enabled}
              onChange={onMetaTop8EnabledChange}
            />
          </div>

          {/* Weight Configuration */}
          <div className="settings-subsection">
            <h3 className="subsection-title">Recommendation Weights</h3>
            <div className="setting-description">
              Adjust how different factors influence card recommendations. Higher weights mean more influence.
            </div>

            <SettingItem
              label="Meta Weight"
              description="How much metagame data influences recommendations (0-1)"
            >
              <input
                type="range"
                min="0"
                max="1"
                step="0.1"
                value={metaWeight}
                onChange={(e) => onMetaWeightChange(parseFloat(e.target.value))}
                className="slider-input"
              />
              <span className="slider-value">{metaWeight.toFixed(1)}</span>
            </SettingItem>

            <SettingItem
              label="Personal History Weight"
              description="How much your personal play history influences recommendations (0-1)"
            >
              <input
                type="range"
                min="0"
                max="1"
                step="0.1"
                value={personalWeight}
                onChange={(e) => onPersonalWeightChange(parseFloat(e.target.value))}
                className="slider-input"
              />
              <span className="slider-value">{personalWeight.toFixed(1)}</span>
            </SettingItem>
          </div>

          {/* LLM Explanations Section */}
          <div className="settings-subsection">
            <h3 className="subsection-title">AI Explanations (Ollama)</h3>
            <div className="setting-description">
              Enable natural language explanations for card recommendations using a local LLM.
              Requires Ollama to be installed and running on your machine.
            </div>

            <SettingToggle
              label="Enable LLM Explanations"
              description="Generate natural language explanations for why cards are recommended"
              checked={llmEnabled}
              onChange={onLLMEnabledChange}
            />

            {llmEnabled && (
              <>
                <SettingItem
                  label="Ollama Endpoint"
                  description="The URL of your Ollama server (default: http://localhost:11434)"
                >
                  <input
                    type="text"
                    value={ollamaEndpoint}
                    onChange={(e) => onOllamaEndpointChange(e.target.value)}
                    placeholder="http://localhost:11434"
                    className="text-input input-width-300"
                  />
                </SettingItem>

                <SettingSelect
                  label="Model"
                  description="The Ollama model to use for generating explanations"
                  value={ollamaModel}
                  onChange={onOllamaModelChange}
                  options={[
                    { value: 'qwen3:8b', label: 'Qwen3 8B (Recommended)' },
                    { value: 'llama3.2:3b', label: 'Llama 3.2 3B (Faster)' },
                    { value: 'llama3.2:1b', label: 'Llama 3.2 1B (Fastest)' },
                    { value: 'mistral:7b', label: 'Mistral 7B' },
                    ...(availableModels
                      .filter(m => !['qwen3:8b', 'llama3.2:3b', 'llama3.2:1b', 'mistral:7b'].includes(m.name))
                      .map(m => ({ value: m.name, label: m.name }))
                    ),
                  ]}
                />

                <SettingItem
                  label="Check Ollama Status"
                  description="Verify that Ollama is running and the selected model is available"
                >
                  <LoadingButton
                    loading={isCheckingOllama}
                    loadingText="Checking..."
                    onClick={handleCheckOllamaStatus}
                    variant="primary"
                    className="button-margin-right"
                  >
                    Check Status
                  </LoadingButton>
                  <LoadingButton
                    loading={isTestingLLM}
                    loadingText="Testing..."
                    onClick={handleTestLLM}
                    variant="default"
                    disabled={!ollamaStatus?.available || !ollamaStatus?.modelReady}
                  >
                    Test Generation
                  </LoadingButton>
                </SettingItem>

                {ollamaStatus && (
                  <div className={`setting-hint settings-${getStatusClass(ollamaStatus)}-box`}>
                    <strong>{getStatusIcon(ollamaStatus)} Ollama Status:</strong>{' '}
                    {ollamaStatus.available && ollamaStatus.modelReady ? (
                      <>
                        Connected (v{ollamaStatus.version}), Model "{ollamaStatus.modelName}" ready
                      </>
                    ) : ollamaStatus.available ? (
                      <>
                        Connected but model not available. {ollamaStatus.error}
                      </>
                    ) : (
                      <>
                        Not available. {ollamaStatus.error}
                      </>
                    )}
                  </div>
                )}

                {llmTestResult && (
                  <div className="setting-hint settings-success-box">
                    <strong>LLM Response:</strong> {llmTestResult}
                  </div>
                )}

                {ollamaStatus?.available && !ollamaStatus?.modelReady && (
                  <SettingItem
                    label="Pull Model"
                    description={`Download the "${ollamaModel}" model to enable LLM explanations`}
                  >
                    <LoadingButton
                      loading={isPullingModel}
                      loadingText="Pulling..."
                      onClick={() => handlePullModel(ollamaModel)}
                      variant="primary"
                    >
                      Pull Model
                    </LoadingButton>
                  </SettingItem>
                )}
              </>
            )}
          </div>

          {/* Suggestion Preferences Section */}
          <div className="settings-subsection">
            <h3 className="subsection-title">Suggestion Preferences</h3>
            <div className="setting-description">
              Customize how ML suggestions are generated and displayed.
            </div>

            <SettingSelect
              label="Suggestion Frequency"
              description="How often suggestions are generated when viewing decks"
              value={suggestionFrequency}
              onChange={onSuggestionFrequencyChange}
              options={[
                { value: 'low', label: 'Low (less aggressive)' },
                { value: 'medium', label: 'Medium (balanced)' },
                { value: 'high', label: 'High (more suggestions)' },
              ]}
            />

            <SettingItem
              label="Minimum Confidence"
              description="Only show suggestions with confidence above this threshold"
            >
              <input
                type="range"
                min="0"
                max="100"
                step="5"
                value={minimumConfidence}
                onChange={(e) => onMinimumConfidenceChange(parseInt(e.target.value))}
                className="slider-input"
              />
              <span className="slider-value">{minimumConfidence}%</span>
            </SettingItem>

            <SettingSelect
              label="Max Suggestions Per View"
              description="Maximum number of suggestions to show at once"
              value={String(maxSuggestionsPerView)}
              onChange={(v) => onMaxSuggestionsPerViewChange(parseInt(v))}
              options={[
                { value: '3', label: '3 suggestions' },
                { value: '5', label: '5 suggestions' },
                { value: '10', label: '10 suggestions' },
              ]}
            />
          </div>

          {/* Suggestion Types Section */}
          <div className="settings-subsection">
            <h3 className="subsection-title">Suggestion Types</h3>
            <div className="setting-description">
              Choose which types of suggestions to show.
            </div>

            <SettingToggle
              label="Card Additions"
              description="Suggest cards to add to your deck"
              checked={showCardAdditions}
              onChange={onShowCardAdditionsChange}
            />

            <SettingToggle
              label="Card Removals"
              description="Suggest underperforming cards to remove"
              checked={showCardRemovals}
              onChange={onShowCardRemovalsChange}
            />

            <SettingToggle
              label="Archetype Changes"
              description="Suggest strategic shifts in deck direction"
              checked={showArchetypeChanges}
              onChange={onShowArchetypeChangesChange}
            />
          </div>

          {/* Learning Options Section */}
          <div className="settings-subsection">
            <h3 className="subsection-title">Learning Options</h3>
            <div className="setting-description">
              Control how the ML engine learns from your activity.
            </div>

            <SettingToggle
              label="Learn from Match Results"
              description="Improve suggestions based on your win/loss outcomes"
              checked={learnFromMatches}
              onChange={onLearnFromMatchesChange}
            />

            <SettingToggle
              label="Learn from Deck Changes"
              description="Improve suggestions based on card swaps you make"
              checked={learnFromDeckChanges}
              onChange={onLearnFromDeckChangesChange}
            />

            <SettingSelect
              label="Data Retention"
              description="How long to keep learned data before clearing old entries"
              value={String(retentionDays)}
              onChange={(v) => onRetentionDaysChange(parseInt(v))}
              options={[
                { value: '30', label: '30 days' },
                { value: '90', label: '90 days' },
                { value: '180', label: '6 months' },
                { value: '365', label: '1 year' },
                { value: '-1', label: 'Forever' },
              ]}
            />

            <SettingItem
              label="Clear Learned Data"
              description="Remove all ML learned data (synergies, patterns, suggestions)"
            >
              <LoadingButton
                loading={isClearingData}
                loadingText="Clearing..."
                onClick={handleClearLearnedData}
                variant="danger"
              >
                Clear All Data
              </LoadingButton>
            </SettingItem>

            {clearDataMessage && (
              <div className={`setting-hint ${clearDataMessage.includes('Failed') ? 'settings-error-box' : 'settings-success-box'}`}>
                {clearDataMessage}
              </div>
            )}
          </div>

          <div className="setting-hint settings-info-box">
            <strong>About ML Recommendations:</strong>
            <ul className="info-list">
              <li>Recommendations are based on your personal play history, deck composition, and current metagame</li>
              <li>The ML model learns from your match results to improve suggestions over time</li>
              <li>LLM explanations require Ollama to be installed locally (visit ollama.ai)</li>
              <li>All processing is done locally - no data is sent to external servers</li>
            </ul>
          </div>
        </>
      )}
    </div>
  );
}
