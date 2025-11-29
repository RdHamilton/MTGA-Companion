package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultOllamaConfig(t *testing.T) {
	config := DefaultOllamaConfig()

	if config.BaseURL != "http://localhost:11434" {
		t.Errorf("unexpected BaseURL: %s", config.BaseURL)
	}
	if config.Model != "qwen3:8b" {
		t.Errorf("unexpected Model: %s", config.Model)
	}
	if config.RequestTimeout != 30*time.Second {
		t.Errorf("unexpected RequestTimeout: %v", config.RequestTimeout)
	}
	if config.InferenceTimeout != 120*time.Second {
		t.Errorf("unexpected InferenceTimeout: %v", config.InferenceTimeout)
	}
	if !config.AutoPullModel {
		t.Error("expected AutoPullModel to be true")
	}
}

func TestNewOllamaClient(t *testing.T) {
	t.Run("with nil config uses defaults", func(t *testing.T) {
		client := NewOllamaClient(nil)
		if client == nil {
			t.Fatal("expected non-nil client")
		}
		if client.config.BaseURL != "http://localhost:11434" {
			t.Error("expected default config")
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &OllamaConfig{
			BaseURL: "http://custom:11434",
			Model:   "llama3",
		}
		client := NewOllamaClient(config)
		if client.config.BaseURL != "http://custom:11434" {
			t.Errorf("expected custom BaseURL, got %s", client.config.BaseURL)
		}
		if client.config.Model != "llama3" {
			t.Errorf("expected custom Model, got %s", client.config.Model)
		}
	})
}

func TestOllamaClient_CheckAvailability(t *testing.T) {
	t.Run("Ollama available with model", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/version":
				_ = json.NewEncoder(w).Encode(VersionResponse{Version: "0.1.0"})
			case "/api/tags":
				_ = json.NewEncoder(w).Encode(ListModelsResponse{
					Models: []ModelInfo{
						{Name: "qwen3:8b", Size: 1000000},
					},
				})
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		client := NewOllamaClient(&OllamaConfig{
			BaseURL:        server.URL,
			Model:          "qwen3:8b",
			RequestTimeout: 5 * time.Second,
			AutoPullModel:  false,
		})

		ctx := context.Background()
		status := client.CheckAvailability(ctx)

		if !status.Available {
			t.Error("expected Ollama to be available")
		}
		if !status.ModelReady {
			t.Error("expected model to be ready")
		}
		if status.Version != "0.1.0" {
			t.Errorf("unexpected version: %s", status.Version)
		}
	})

	t.Run("Ollama available without model", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/version":
				_ = json.NewEncoder(w).Encode(VersionResponse{Version: "0.1.0"})
			case "/api/tags":
				_ = json.NewEncoder(w).Encode(ListModelsResponse{
					Models: []ModelInfo{}, // No models
				})
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		client := NewOllamaClient(&OllamaConfig{
			BaseURL:        server.URL,
			Model:          "qwen3:8b",
			RequestTimeout: 5 * time.Second,
			AutoPullModel:  false, // Don't auto-pull
		})

		ctx := context.Background()
		status := client.CheckAvailability(ctx)

		if !status.Available {
			t.Error("expected Ollama to be available")
		}
		if status.ModelReady {
			t.Error("expected model to not be ready")
		}
	})

	t.Run("Ollama not running", func(t *testing.T) {
		client := NewOllamaClient(&OllamaConfig{
			BaseURL:        "http://localhost:99999", // Invalid port
			Model:          "qwen3:8b",
			RequestTimeout: 1 * time.Second,
		})

		ctx := context.Background()
		status := client.CheckAvailability(ctx)

		if status.Available {
			t.Error("expected Ollama to not be available")
		}
		if status.Error == "" {
			t.Error("expected error message")
		}
	})
}

func TestOllamaClient_IsAvailable(t *testing.T) {
	client := NewOllamaClient(nil)

	// Initially not available
	if client.IsAvailable() {
		t.Error("expected not available initially")
	}

	// Set availability
	client.setAvailability(true, true)

	if !client.IsAvailable() {
		t.Error("expected available after setting")
	}
}

func TestOllamaClient_Generate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/version":
			_ = json.NewEncoder(w).Encode(VersionResponse{Version: "0.1.0"})
		case "/api/tags":
			_ = json.NewEncoder(w).Encode(ListModelsResponse{
				Models: []ModelInfo{{Name: "qwen3:8b"}},
			})
		case "/api/generate":
			var req GenerateRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			resp := GenerateResponse{
				Model:    req.Model,
				Response: "This is a test response for: " + req.Prompt,
				Done:     true,
			}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewOllamaClient(&OllamaConfig{
		BaseURL:          server.URL,
		Model:            "qwen3:8b",
		RequestTimeout:   5 * time.Second,
		InferenceTimeout: 30 * time.Second,
		AutoPullModel:    false,
	})

	ctx := context.Background()

	// First check availability
	status := client.CheckAvailability(ctx)
	if !status.Available || !status.ModelReady {
		t.Fatalf("Ollama not available for test: %s", status.Error)
	}

	// Test generate
	resp, err := client.Generate(ctx, "Hello, world!", nil)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !resp.Done {
		t.Error("expected response to be done")
	}
	if resp.Response == "" {
		t.Error("expected non-empty response")
	}
}

func TestOllamaClient_GenerateWithSystem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/version":
			_ = json.NewEncoder(w).Encode(VersionResponse{Version: "0.1.0"})
		case "/api/tags":
			_ = json.NewEncoder(w).Encode(ListModelsResponse{
				Models: []ModelInfo{{Name: "qwen3:8b"}},
			})
		case "/api/generate":
			var req GenerateRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// Verify system prompt was included
			resp := GenerateResponse{
				Model:    req.Model,
				Response: "System: " + req.System + ", Prompt: " + req.Prompt,
				Done:     true,
			}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewOllamaClient(&OllamaConfig{
		BaseURL:          server.URL,
		Model:            "qwen3:8b",
		RequestTimeout:   5 * time.Second,
		InferenceTimeout: 30 * time.Second,
		AutoPullModel:    false,
	})

	ctx := context.Background()
	_ = client.CheckAvailability(ctx)

	resp, err := client.GenerateWithSystem(ctx, "You are a helpful assistant", "Help me", nil)
	if err != nil {
		t.Fatalf("GenerateWithSystem failed: %v", err)
	}

	if !resp.Done {
		t.Error("expected response to be done")
	}
}

func TestOllamaClient_Chat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/version":
			_ = json.NewEncoder(w).Encode(VersionResponse{Version: "0.1.0"})
		case "/api/tags":
			_ = json.NewEncoder(w).Encode(ListModelsResponse{
				Models: []ModelInfo{{Name: "qwen3:8b"}},
			})
		case "/api/chat":
			var req ChatRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			resp := ChatResponse{
				Model: req.Model,
				Message: ChatMessage{
					Role:    "assistant",
					Content: "Hello! How can I help you?",
				},
				Done: true,
			}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewOllamaClient(&OllamaConfig{
		BaseURL:          server.URL,
		Model:            "qwen3:8b",
		RequestTimeout:   5 * time.Second,
		InferenceTimeout: 30 * time.Second,
		AutoPullModel:    false,
	})

	ctx := context.Background()
	_ = client.CheckAvailability(ctx)

	messages := []ChatMessage{
		{Role: "user", Content: "Hello!"},
	}

	resp, err := client.Chat(ctx, messages, nil)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if resp.Message.Role != "assistant" {
		t.Errorf("expected assistant role, got %s", resp.Message.Role)
	}
	if resp.Message.Content == "" {
		t.Error("expected non-empty content")
	}
}

func TestOllamaClient_GenerateNotAvailable(t *testing.T) {
	client := NewOllamaClient(&OllamaConfig{
		BaseURL:        "http://localhost:99999",
		Model:          "qwen3:8b",
		RequestTimeout: 1 * time.Second,
	})

	ctx := context.Background()
	_, err := client.Generate(ctx, "Hello", nil)
	if err == nil {
		t.Error("expected error when Ollama not available")
	}
}

func TestOllamaClient_GetConfig(t *testing.T) {
	config := &OllamaConfig{
		BaseURL: "http://custom:11434",
		Model:   "llama3",
	}
	client := NewOllamaClient(config)

	retrieved := client.GetConfig()
	if retrieved.BaseURL != "http://custom:11434" {
		t.Errorf("unexpected BaseURL: %s", retrieved.BaseURL)
	}
}

func TestOllamaClient_UpdateConfig(t *testing.T) {
	client := NewOllamaClient(nil)
	client.setAvailability(true, true)

	newConfig := &OllamaConfig{
		BaseURL: "http://new:11434",
		Model:   "llama3",
	}
	client.UpdateConfig(newConfig)

	if client.config.BaseURL != "http://new:11434" {
		t.Error("config not updated")
	}
	// Should reset availability
	if client.IsAvailable() {
		t.Error("expected availability to be reset")
	}
}

func TestOllamaClient_GetModel(t *testing.T) {
	config := &OllamaConfig{
		Model: "qwen3:8b",
	}
	client := NewOllamaClient(config)

	if client.GetModel() != "qwen3:8b" {
		t.Errorf("unexpected model: %s", client.GetModel())
	}
}

func TestGenerateOptions(t *testing.T) {
	options := &GenerateOptions{
		Temperature: 0.7,
		TopP:        0.9,
		TopK:        40,
		NumPredict:  100,
		Stop:        []string{"\n"},
	}

	if options.Temperature != 0.7 {
		t.Errorf("unexpected Temperature: %f", options.Temperature)
	}
	if options.TopP != 0.9 {
		t.Errorf("unexpected TopP: %f", options.TopP)
	}
	if options.TopK != 40 {
		t.Errorf("unexpected TopK: %d", options.TopK)
	}
}

func TestOllamaStatus_Fields(t *testing.T) {
	status := &OllamaStatus{
		Available:    true,
		Version:      "0.1.0",
		ModelReady:   true,
		ModelName:    "qwen3:8b",
		ModelsLoaded: []string{"qwen3:8b", "llama3"},
		Error:        "",
	}

	if !status.Available {
		t.Error("expected Available to be true")
	}
	if !status.ModelReady {
		t.Error("expected ModelReady to be true")
	}
	if len(status.ModelsLoaded) != 2 {
		t.Errorf("unexpected ModelsLoaded length: %d", len(status.ModelsLoaded))
	}
}

func TestChatMessage_Fields(t *testing.T) {
	msg := ChatMessage{
		Role:    "user",
		Content: "Hello!",
	}

	if msg.Role != "user" {
		t.Errorf("unexpected Role: %s", msg.Role)
	}
	if msg.Content != "Hello!" {
		t.Errorf("unexpected Content: %s", msg.Content)
	}
}

func TestGenerateResponse_Fields(t *testing.T) {
	resp := &GenerateResponse{
		Model:         "qwen3:8b",
		Response:      "Hello!",
		Done:          true,
		TotalDuration: 1000000000,
		EvalCount:     10,
		EvalDuration:  500000000,
	}

	if resp.Model != "qwen3:8b" {
		t.Errorf("unexpected Model: %s", resp.Model)
	}
	if !resp.Done {
		t.Error("expected Done to be true")
	}
}
