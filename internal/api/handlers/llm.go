package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
)

// LLMHandler handles LLM-related API requests.
type LLMHandler struct {
	facade *gui.LLMFacade
}

// NewLLMHandler creates a new LLMHandler.
func NewLLMHandler(facade *gui.LLMFacade) *LLMHandler {
	return &LLMHandler{facade: facade}
}

// CheckStatusRequest represents a request to check Ollama status.
type CheckStatusRequest struct {
	Endpoint string `json:"endpoint,omitempty"`
	Model    string `json:"model,omitempty"`
}

// CheckOllamaStatus checks if Ollama is available and returns its status.
func (h *LLMHandler) CheckOllamaStatus(w http.ResponseWriter, r *http.Request) {
	var req CheckStatusRequest

	// Allow empty body for GET-like behavior
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, errors.New("invalid request body"))
			return
		}
	}

	status, err := h.facade.CheckOllamaStatus(r.Context(), req.Endpoint, req.Model)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, status)
}

// GetModelsRequest represents a request to get available models.
type GetModelsRequest struct {
	Endpoint string `json:"endpoint,omitempty"`
}

// GetAvailableModels returns a list of available Ollama models.
func (h *LLMHandler) GetAvailableModels(w http.ResponseWriter, r *http.Request) {
	endpoint := r.URL.Query().Get("endpoint")

	models, err := h.facade.GetAvailableModels(r.Context(), endpoint)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, models)
}

// PullModelRequest represents a request to pull a model.
type PullModelRequest struct {
	Endpoint string `json:"endpoint,omitempty"`
	Model    string `json:"model"`
}

// PullModel pulls a model from Ollama.
func (h *LLMHandler) PullModel(w http.ResponseWriter, r *http.Request) {
	var req PullModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if req.Model == "" {
		response.BadRequest(w, errors.New("model name is required"))
		return
	}

	if err := h.facade.PullModel(r.Context(), req.Endpoint, req.Model); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{
		"status":  "success",
		"message": "Model pulled successfully",
		"model":   req.Model,
	})
}

// TestGenerationRequest represents a request to test LLM generation.
type TestGenerationRequest struct {
	Endpoint string `json:"endpoint,omitempty"`
	Model    string `json:"model,omitempty"`
}

// TestGeneration tests LLM generation with a simple prompt.
func (h *LLMHandler) TestGeneration(w http.ResponseWriter, r *http.Request) {
	var req TestGenerationRequest

	// Allow empty body for default values
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, errors.New("invalid request body"))
			return
		}
	}

	result, err := h.facade.TestLLMGeneration(r.Context(), req.Endpoint, req.Model)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{
		"status":   "success",
		"response": result,
	})
}
