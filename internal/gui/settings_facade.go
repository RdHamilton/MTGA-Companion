package gui

import (
	"context"
	"fmt"
)

// SettingsFacade handles all settings operations for the GUI.
type SettingsFacade struct {
	services *Services
}

// NewSettingsFacade creates a new SettingsFacade with the given services.
func NewSettingsFacade(services *Services) *SettingsFacade {
	return &SettingsFacade{
		services: services,
	}
}

// AppSettings represents all user-configurable settings.
type AppSettings struct {
	AutoRefresh       bool   `json:"autoRefresh"`
	RefreshInterval   int    `json:"refreshInterval"`
	ShowNotifications bool   `json:"showNotifications"`
	Theme             string `json:"theme"`
	DaemonPort        int    `json:"daemonPort"`
	DaemonMode        string `json:"daemonMode"`

	// ML/LLM Settings
	MLEnabled           bool    `json:"mlEnabled"`
	LLMEnabled          bool    `json:"llmEnabled"`
	OllamaEndpoint      string  `json:"ollamaEndpoint"`
	OllamaModel         string  `json:"ollamaModel"`
	MetaGoldfishEnabled bool    `json:"metaGoldfishEnabled"`
	MetaTop8Enabled     bool    `json:"metaTop8Enabled"`
	MetaWeight          float64 `json:"metaWeight"`
	PersonalWeight      float64 `json:"personalWeight"`

	// ML Suggestion Preferences
	SuggestionFrequency   string `json:"suggestionFrequency"` // low, medium, high
	MinimumConfidence     int    `json:"minimumConfidence"`   // 0-100
	ShowCardAdditions     bool   `json:"showCardAdditions"`
	ShowCardRemovals      bool   `json:"showCardRemovals"`
	ShowArchetypeChanges  bool   `json:"showArchetypeChanges"`
	LearnFromMatches      bool   `json:"learnFromMatches"`
	LearnFromDeckChanges  bool   `json:"learnFromDeckChanges"`
	RetentionDays         int    `json:"retentionDays"`         // 30, 90, 180, 365, -1 (forever)
	MaxSuggestionsPerView int    `json:"maxSuggestionsPerView"` // 3, 5, 10
}

// GetAllSettings retrieves all settings as an AppSettings struct.
func (s *SettingsFacade) GetAllSettings(ctx context.Context) (*AppSettings, error) {
	if s.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	repo := s.services.Storage.SettingsRepo()

	settings := &AppSettings{
		// Defaults
		AutoRefresh:         false,
		RefreshInterval:     30,
		ShowNotifications:   true,
		Theme:               "dark",
		DaemonPort:          9999,
		DaemonMode:          "standalone",
		MLEnabled:           true,
		LLMEnabled:          false,
		OllamaEndpoint:      "http://localhost:11434",
		OllamaModel:         "qwen3:8b",
		MetaGoldfishEnabled: true,
		MetaTop8Enabled:     true,
		MetaWeight:          0.3,
		PersonalWeight:      0.2,
		// ML Suggestion defaults
		SuggestionFrequency:   "medium",
		MinimumConfidence:     50,
		ShowCardAdditions:     true,
		ShowCardRemovals:      true,
		ShowArchetypeChanges:  true,
		LearnFromMatches:      true,
		LearnFromDeckChanges:  true,
		RetentionDays:         90,
		MaxSuggestionsPerView: 5,
	}

	// Get each setting, using defaults if not found (errors are intentionally ignored)
	_ = repo.GetTyped(ctx, "autoRefresh", &settings.AutoRefresh)
	_ = repo.GetTyped(ctx, "refreshInterval", &settings.RefreshInterval)
	_ = repo.GetTyped(ctx, "showNotifications", &settings.ShowNotifications)
	_ = repo.GetTyped(ctx, "theme", &settings.Theme)
	_ = repo.GetTyped(ctx, "daemonPort", &settings.DaemonPort)
	_ = repo.GetTyped(ctx, "daemonMode", &settings.DaemonMode)
	_ = repo.GetTyped(ctx, "mlEnabled", &settings.MLEnabled)
	_ = repo.GetTyped(ctx, "llmEnabled", &settings.LLMEnabled)
	_ = repo.GetTyped(ctx, "ollamaEndpoint", &settings.OllamaEndpoint)
	_ = repo.GetTyped(ctx, "ollamaModel", &settings.OllamaModel)
	_ = repo.GetTyped(ctx, "metaGoldfishEnabled", &settings.MetaGoldfishEnabled)
	_ = repo.GetTyped(ctx, "metaTop8Enabled", &settings.MetaTop8Enabled)
	_ = repo.GetTyped(ctx, "metaWeight", &settings.MetaWeight)
	_ = repo.GetTyped(ctx, "personalWeight", &settings.PersonalWeight)
	// ML Suggestion settings
	_ = repo.GetTyped(ctx, "suggestionFrequency", &settings.SuggestionFrequency)
	_ = repo.GetTyped(ctx, "minimumConfidence", &settings.MinimumConfidence)
	_ = repo.GetTyped(ctx, "showCardAdditions", &settings.ShowCardAdditions)
	_ = repo.GetTyped(ctx, "showCardRemovals", &settings.ShowCardRemovals)
	_ = repo.GetTyped(ctx, "showArchetypeChanges", &settings.ShowArchetypeChanges)
	_ = repo.GetTyped(ctx, "learnFromMatches", &settings.LearnFromMatches)
	_ = repo.GetTyped(ctx, "learnFromDeckChanges", &settings.LearnFromDeckChanges)
	_ = repo.GetTyped(ctx, "retentionDays", &settings.RetentionDays)
	_ = repo.GetTyped(ctx, "maxSuggestionsPerView", &settings.MaxSuggestionsPerView)

	return settings, nil
}

// SaveAllSettings saves all settings at once.
func (s *SettingsFacade) SaveAllSettings(ctx context.Context, settings *AppSettings) error {
	if s.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	repo := s.services.Storage.SettingsRepo()

	settingsMap := map[string]interface{}{
		"autoRefresh":         settings.AutoRefresh,
		"refreshInterval":     settings.RefreshInterval,
		"showNotifications":   settings.ShowNotifications,
		"theme":               settings.Theme,
		"daemonPort":          settings.DaemonPort,
		"daemonMode":          settings.DaemonMode,
		"mlEnabled":           settings.MLEnabled,
		"llmEnabled":          settings.LLMEnabled,
		"ollamaEndpoint":      settings.OllamaEndpoint,
		"ollamaModel":         settings.OllamaModel,
		"metaGoldfishEnabled": settings.MetaGoldfishEnabled,
		"metaTop8Enabled":     settings.MetaTop8Enabled,
		"metaWeight":          settings.MetaWeight,
		"personalWeight":      settings.PersonalWeight,
		// ML Suggestion settings
		"suggestionFrequency":   settings.SuggestionFrequency,
		"minimumConfidence":     settings.MinimumConfidence,
		"showCardAdditions":     settings.ShowCardAdditions,
		"showCardRemovals":      settings.ShowCardRemovals,
		"showArchetypeChanges":  settings.ShowArchetypeChanges,
		"learnFromMatches":      settings.LearnFromMatches,
		"learnFromDeckChanges":  settings.LearnFromDeckChanges,
		"retentionDays":         settings.RetentionDays,
		"maxSuggestionsPerView": settings.MaxSuggestionsPerView,
	}

	if err := repo.SetMany(ctx, settingsMap); err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to save settings: %v", err)}
	}

	return nil
}

// GetSetting retrieves a single setting by key.
func (s *SettingsFacade) GetSetting(ctx context.Context, key string) (interface{}, error) {
	if s.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	repo := s.services.Storage.SettingsRepo()

	allSettings, err := repo.GetAll(ctx)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get settings: %v", err)}
	}

	if value, exists := allSettings[key]; exists {
		return value, nil
	}

	return nil, &AppError{Message: fmt.Sprintf("Setting not found: %s", key)}
}

// SetSetting saves a single setting.
func (s *SettingsFacade) SetSetting(ctx context.Context, key string, value interface{}) error {
	if s.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	repo := s.services.Storage.SettingsRepo()

	if err := repo.Set(ctx, key, value); err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to save setting: %v", err)}
	}

	return nil
}
