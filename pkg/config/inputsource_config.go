package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// InputSourceConfig defines the configuration for different input sources
type InputSourceConfig struct {
	Type     string                 `json:"type"`     // github, test, jira, etc.
	Settings map[string]interface{} `json:"settings"` // source-specific settings
}

// DefaultInputSourceConfig returns the default configuration (GitHub)
func DefaultInputSourceConfig() *InputSourceConfig {
	return &InputSourceConfig{
		Type: "github",
		Settings: map[string]interface{}{
			"repository":         "auto-detect", // Auto-detect from git remote
			"allow_cross_source": false,         // By default, don't allow cross-source usage
		},
	}
}

// LoadInputSourceConfig loads input source configuration from project root
// Returns default configuration if no config file exists
func LoadInputSourceConfig(projectRoot string) (*InputSourceConfig, error) {
	configPath := filepath.Join(projectRoot, ".sbs", "input-source.json")

	// If config file doesn't exist, return default
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultInputSourceConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read input source config: %w", err)
	}

	// Parse JSON
	var config InputSourceConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse input source config: %w", err)
	}

	// Validate configuration
	if err := validateInputSourceConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid input source config: %w", err)
	}

	// Ensure settings is not nil
	if config.Settings == nil {
		config.Settings = make(map[string]interface{})
	}

	return &config, nil
}

// SaveInputSourceConfig saves input source configuration to project root
func SaveInputSourceConfig(projectRoot string, config *InputSourceConfig) error {
	// Validate configuration
	if err := validateInputSourceConfig(config); err != nil {
		return fmt.Errorf("invalid input source config: %w", err)
	}

	// Ensure .sbs directory exists
	sbsDir := filepath.Join(projectRoot, ".sbs")
	if err := os.MkdirAll(sbsDir, 0755); err != nil {
		return fmt.Errorf("failed to create .sbs directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	// Write config file
	configPath := filepath.Join(sbsDir, "input-source.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write input source config: %w", err)
	}

	return nil
}

// validateInputSourceConfig validates the input source configuration
func validateInputSourceConfig(config *InputSourceConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Type is required and cannot be empty
	config.Type = strings.TrimSpace(config.Type)
	if config.Type == "" {
		return fmt.Errorf("type cannot be empty")
	}

	// Ensure settings exists
	if config.Settings == nil {
		config.Settings = make(map[string]interface{})
	}

	return nil
}

// AllowCrossSource returns whether cross-source usage is allowed based on configuration
func (config *InputSourceConfig) AllowCrossSource() bool {
	if config.Settings == nil {
		return false
	}

	if value, exists := config.Settings["allow_cross_source"]; exists {
		if boolValue, ok := value.(bool); ok {
			return boolValue
		}
	}

	return false
}
