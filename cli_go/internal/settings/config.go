package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Model represents an AI model configuration
type Model struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MaxTokens   int    `json:"maxTokens"`
}

// ModelsConfig represents the available models
type ModelsConfig struct {
	Models []Model `json:"models"`
}

// SelectedModelConfig represents the currently selected model
type SelectedModelConfig struct {
	SelectedModel Model `json:"selectedModel"`
}

// LoadAvailableModels loads the list of available models from cli_models.json
func LoadAvailableModels() ([]Model, error) {
	configPath := filepath.Join(".", "cli_models.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config ModelsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config.Models, nil
}

// LoadSelectedModel loads the currently selected model from cli_model_set.json
func LoadSelectedModel() (*Model, error) {
	configPath := filepath.Join(".", "cli_model_set.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config SelectedModelConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config.SelectedModel, nil
}

// SaveSelectedModel saves the selected model to cli_model_set.json
func SaveSelectedModel(model Model) error {
	configPath := filepath.Join(".", "cli_model_set.json")

	config := SelectedModelConfig{
		SelectedModel: model,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
