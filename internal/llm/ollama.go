package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// OllamaConfig configures the Ollama client.
type OllamaConfig struct {
	// BaseURL is the Ollama API endpoint.
	BaseURL string

	// Model is the model name to use.
	Model string

	// RequestTimeout is the timeout for API requests.
	RequestTimeout time.Duration

	// InferenceTimeout is the timeout for inference (generation) requests.
	InferenceTimeout time.Duration

	// MaxRetries is the number of retries for failed requests.
	MaxRetries int

	// AutoPullModel automatically pulls the model if not available.
	AutoPullModel bool
}

// DefaultOllamaConfig returns sensible defaults.
func DefaultOllamaConfig() *OllamaConfig {
	return &OllamaConfig{
		BaseURL:          "http://localhost:11434",
		Model:            "qwen3:8b",
		RequestTimeout:   30 * time.Second,
		InferenceTimeout: 120 * time.Second,
		MaxRetries:       2,
		AutoPullModel:    true,
	}
}

// OllamaClient provides access to Ollama API.
type OllamaClient struct {
	config     *OllamaConfig
	httpClient *http.Client
	available  bool
	modelReady bool
	lastCheck  time.Time
	mu         sync.RWMutex
}

// OllamaStatus represents the status of Ollama.
type OllamaStatus struct {
	Available    bool     `json:"available"`
	Version      string   `json:"version,omitempty"`
	ModelReady   bool     `json:"model_ready"`
	ModelName    string   `json:"model_name"`
	ModelsLoaded []string `json:"models_loaded,omitempty"`
	Error        string   `json:"error,omitempty"`
}

// GenerateRequest is the request body for generation.
type GenerateRequest struct {
	Model   string           `json:"model"`
	Prompt  string           `json:"prompt"`
	Stream  bool             `json:"stream"`
	Options *GenerateOptions `json:"options,omitempty"`
	System  string           `json:"system,omitempty"`
	Context []int            `json:"context,omitempty"`
}

// GenerateOptions are optional parameters for generation.
type GenerateOptions struct {
	Temperature float64  `json:"temperature,omitempty"`
	TopP        float64  `json:"top_p,omitempty"`
	TopK        int      `json:"top_k,omitempty"`
	NumPredict  int      `json:"num_predict,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

// GenerateResponse is the response from generation.
type GenerateResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	Context            []int  `json:"context,omitempty"`
	TotalDuration      int64  `json:"total_duration,omitempty"`
	LoadDuration       int64  `json:"load_duration,omitempty"`
	PromptEvalCount    int    `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount          int    `json:"eval_count,omitempty"`
	EvalDuration       int64  `json:"eval_duration,omitempty"`
}

// ChatMessage represents a chat message.
type ChatMessage struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

// ChatRequest is the request body for chat completions.
type ChatRequest struct {
	Model    string           `json:"model"`
	Messages []ChatMessage    `json:"messages"`
	Stream   bool             `json:"stream"`
	Options  *GenerateOptions `json:"options,omitempty"`
}

// ChatResponse is the response from chat completions.
type ChatResponse struct {
	Model     string      `json:"model"`
	CreatedAt string      `json:"created_at"`
	Message   ChatMessage `json:"message"`
	Done      bool        `json:"done"`
}

// VersionResponse is the response from the version endpoint.
type VersionResponse struct {
	Version string `json:"version"`
}

// ListModelsResponse is the response from listing models.
type ListModelsResponse struct {
	Models []ModelInfo `json:"models"`
}

// ModelInfo describes a model.
type ModelInfo struct {
	Name       string    `json:"name"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
}

// PullResponse is a streaming response from model pull.
type PullResponse struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

// NewOllamaClient creates a new Ollama client.
func NewOllamaClient(config *OllamaConfig) *OllamaClient {
	if config == nil {
		config = DefaultOllamaConfig()
	}

	return &OllamaClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
		},
	}
}

// CheckAvailability checks if Ollama is available and the model is ready.
func (c *OllamaClient) CheckAvailability(ctx context.Context) *OllamaStatus {
	status := &OllamaStatus{
		ModelName: c.config.Model,
	}

	// Check if Ollama is running
	version, err := c.getVersion(ctx)
	if err != nil {
		status.Error = fmt.Sprintf("Ollama not available: %v", err)
		c.setAvailability(false, false)
		return status
	}

	status.Available = true
	status.Version = version

	// List loaded models
	models, err := c.listModels(ctx)
	if err != nil {
		status.Error = fmt.Sprintf("Failed to list models: %v", err)
		c.setAvailability(true, false)
		return status
	}

	status.ModelsLoaded = make([]string, 0, len(models))
	for _, m := range models {
		status.ModelsLoaded = append(status.ModelsLoaded, m.Name)
		// Check if our target model is available
		if strings.HasPrefix(m.Name, strings.Split(c.config.Model, ":")[0]) {
			status.ModelReady = true
		}
	}

	// Auto-pull model if configured and not available
	if !status.ModelReady && c.config.AutoPullModel {
		status.Error = fmt.Sprintf("Model %s not found, pulling...", c.config.Model)
		if pullErr := c.PullModel(ctx); pullErr != nil {
			status.Error = fmt.Sprintf("Failed to pull model: %v", pullErr)
		} else {
			status.ModelReady = true
			status.Error = ""
		}
	}

	c.setAvailability(status.Available, status.ModelReady)
	return status
}

// IsAvailable returns whether Ollama is currently available.
func (c *OllamaClient) IsAvailable() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.available && c.modelReady
}

// Generate generates text using the model.
func (c *OllamaClient) Generate(ctx context.Context, prompt string, options *GenerateOptions) (*GenerateResponse, error) {
	if !c.IsAvailable() {
		// Quick re-check
		status := c.CheckAvailability(ctx)
		if !status.Available || !status.ModelReady {
			return nil, fmt.Errorf("ollama not available: %s", status.Error)
		}
	}

	req := &GenerateRequest{
		Model:   c.config.Model,
		Prompt:  prompt,
		Stream:  false,
		Options: options,
	}

	return c.doGenerate(ctx, req)
}

// GenerateWithSystem generates text with a system prompt.
func (c *OllamaClient) GenerateWithSystem(ctx context.Context, system, prompt string, options *GenerateOptions) (*GenerateResponse, error) {
	if !c.IsAvailable() {
		status := c.CheckAvailability(ctx)
		if !status.Available || !status.ModelReady {
			return nil, fmt.Errorf("ollama not available: %s", status.Error)
		}
	}

	req := &GenerateRequest{
		Model:   c.config.Model,
		System:  system,
		Prompt:  prompt,
		Stream:  false,
		Options: options,
	}

	return c.doGenerate(ctx, req)
}

// Chat sends a chat completion request.
func (c *OllamaClient) Chat(ctx context.Context, messages []ChatMessage, options *GenerateOptions) (*ChatResponse, error) {
	if !c.IsAvailable() {
		status := c.CheckAvailability(ctx)
		if !status.Available || !status.ModelReady {
			return nil, fmt.Errorf("ollama not available: %s", status.Error)
		}
	}

	req := &ChatRequest{
		Model:    c.config.Model,
		Messages: messages,
		Stream:   false,
		Options:  options,
	}

	return c.doChat(ctx, req)
}

// PullModel pulls the configured model.
func (c *OllamaClient) PullModel(ctx context.Context) error {
	url := c.config.BaseURL + "/api/pull"

	body, err := json.Marshal(map[string]interface{}{
		"name":   c.config.Model,
		"stream": false,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Use a longer timeout for model pull
	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("pull request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pull failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// doGenerate performs the generate API call.
func (c *OllamaClient) doGenerate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	url := c.config.BaseURL + "/api/generate"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Use inference timeout for generation
	client := &http.Client{Timeout: c.config.InferenceTimeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("generate request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("generate failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var genResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &genResp, nil
}

// doChat performs the chat API call.
func (c *OllamaClient) doChat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	url := c.config.BaseURL + "/api/chat"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Use inference timeout for chat
	client := &http.Client{Timeout: c.config.InferenceTimeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("chat request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("chat failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &chatResp, nil
}

// getVersion gets the Ollama version.
func (c *OllamaClient) getVersion(ctx context.Context) (string, error) {
	url := c.config.BaseURL + "/api/version"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("version check failed with status %d", resp.StatusCode)
	}

	var version VersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&version); err != nil {
		return "", err
	}

	return version.Version, nil
}

// listModels lists available models.
func (c *OllamaClient) listModels(ctx context.Context) ([]ModelInfo, error) {
	url := c.config.BaseURL + "/api/tags"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list models failed with status %d", resp.StatusCode)
	}

	var models ListModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		return nil, err
	}

	return models.Models, nil
}

// setAvailability updates the availability status.
func (c *OllamaClient) setAvailability(available, modelReady bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.available = available
	c.modelReady = modelReady
	c.lastCheck = time.Now()
}

// GetConfig returns the current configuration.
func (c *OllamaClient) GetConfig() *OllamaConfig {
	return c.config
}

// UpdateConfig updates the configuration.
func (c *OllamaClient) UpdateConfig(config *OllamaConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config = config
	c.available = false
	c.modelReady = false
}

// GetModel returns the configured model name.
func (c *OllamaClient) GetModel() string {
	return c.config.Model
}
