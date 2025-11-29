package gui

import (
	"context"
	"fmt"

	"github.com/ramonehamilton/MTGA-Companion/internal/llm"
)

// LLMFacade handles LLM-related operations for the GUI.
type LLMFacade struct {
	services *Services
}

// NewLLMFacade creates a new LLMFacade with the given services.
func NewLLMFacade(services *Services) *LLMFacade {
	return &LLMFacade{
		services: services,
	}
}

// OllamaStatus represents the status of the Ollama service.
type OllamaStatus struct {
	Available    bool     `json:"available"`
	Version      string   `json:"version,omitempty"`
	ModelReady   bool     `json:"modelReady"`
	ModelName    string   `json:"modelName"`
	ModelsLoaded []string `json:"modelsLoaded,omitempty"`
	Error        string   `json:"error,omitempty"`
}

// OllamaModel represents an available Ollama model.
type OllamaModel struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// CheckOllamaStatus checks if Ollama is available and returns its status.
func (f *LLMFacade) CheckOllamaStatus(ctx context.Context, endpoint, model string) (*OllamaStatus, error) {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	if model == "" {
		model = "qwen3:8b"
	}

	config := &llm.OllamaConfig{
		BaseURL:       endpoint,
		Model:         model,
		AutoPullModel: false, // Don't auto-pull during status check
	}

	client := llm.NewOllamaClient(config)
	llmStatus := client.CheckAvailability(ctx)

	return &OllamaStatus{
		Available:    llmStatus.Available,
		Version:      llmStatus.Version,
		ModelReady:   llmStatus.ModelReady,
		ModelName:    llmStatus.ModelName,
		ModelsLoaded: llmStatus.ModelsLoaded,
		Error:        llmStatus.Error,
	}, nil
}

// GetAvailableModels returns a list of available Ollama models.
func (f *LLMFacade) GetAvailableModels(ctx context.Context, endpoint string) ([]OllamaModel, error) {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}

	config := &llm.OllamaConfig{
		BaseURL:       endpoint,
		AutoPullModel: false,
	}

	client := llm.NewOllamaClient(config)
	status := client.CheckAvailability(ctx)

	if !status.Available {
		return nil, &AppError{Message: fmt.Sprintf("Ollama not available: %s", status.Error)}
	}

	models := make([]OllamaModel, 0, len(status.ModelsLoaded))
	for _, name := range status.ModelsLoaded {
		models = append(models, OllamaModel{Name: name})
	}

	return models, nil
}

// PullModel pulls a model from Ollama.
func (f *LLMFacade) PullModel(ctx context.Context, endpoint, model string) error {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	if model == "" {
		return &AppError{Message: "Model name is required"}
	}

	config := &llm.OllamaConfig{
		BaseURL:       endpoint,
		Model:         model,
		AutoPullModel: false,
	}

	client := llm.NewOllamaClient(config)
	if err := client.PullModel(ctx); err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to pull model: %v", err)}
	}

	return nil
}

// TestLLMGeneration tests LLM generation with a simple prompt.
func (f *LLMFacade) TestLLMGeneration(ctx context.Context, endpoint, model string) (string, error) {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	if model == "" {
		model = "qwen3:8b"
	}

	config := &llm.OllamaConfig{
		BaseURL:       endpoint,
		Model:         model,
		AutoPullModel: false,
	}

	client := llm.NewOllamaClient(config)

	// Check availability first
	status := client.CheckAvailability(ctx)
	if !status.Available || !status.ModelReady {
		return "", &AppError{Message: fmt.Sprintf("Ollama not ready: %s", status.Error)}
	}

	// Generate a simple test response
	resp, err := client.Generate(ctx, "Say 'Hello from Ollama!' and nothing else.", nil)
	if err != nil {
		return "", &AppError{Message: fmt.Sprintf("Generation failed: %v", err)}
	}

	return resp.Response, nil
}
