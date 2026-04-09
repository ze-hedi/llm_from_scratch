package extensions

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Extension represents an available extension
type Extension struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Command     string `json:"command"`
	Icon        string `json:"icon"`
}

// ExtensionsConfig represents the list of available extensions
type ExtensionsConfig struct {
	Extensions []Extension `json:"extensions"`
}

// LoadExtensions loads the list of available extensions from extensions.json
func LoadExtensions() ([]Extension, error) {
	configPath := filepath.Join(".", "extensions.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config ExtensionsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config.Extensions, nil
}
