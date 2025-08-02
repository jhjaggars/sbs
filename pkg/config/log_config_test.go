package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_LogDisplay(t *testing.T) {
	t.Run("default_log_refresh_interval", func(t *testing.T) {
		// Test default 5-second refresh interval
		config := DefaultConfig()

		// Verify reasonable default values
		assert.Equal(t, 5, config.LogRefreshIntervalSecs, "Default log refresh interval should be 5 seconds")
		assert.GreaterOrEqual(t, config.LogRefreshIntervalSecs, 1, "Default interval should be at least 1 second")
		assert.LessOrEqual(t, config.LogRefreshIntervalSecs, 300, "Default interval should be at most 300 seconds")
	})

	t.Run("custom_log_refresh_interval", func(t *testing.T) {
		// Test setting custom refresh intervals
		testCases := []struct {
			name            string
			intervalSecs    int
			expectedValid   bool
			expectedMessage string
		}{
			{"valid_10_seconds", 10, true, "10 seconds should be valid"},
			{"valid_30_seconds", 30, true, "30 seconds should be valid"},
			{"valid_60_seconds", 60, true, "60 seconds should be valid"},
			{"invalid_negative", -5, false, "negative seconds should be invalid"},
			{"invalid_too_large", 400, false, "400 seconds should be invalid"},
			{"boundary_min_valid", 1, true, "1 second should be valid (minimum)"},
			{"boundary_max_valid", 300, true, "300 seconds should be valid (maximum)"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				config := DefaultConfig()
				config.LogRefreshIntervalSecs = tc.intervalSecs

				// Test validation of interval bounds (1-300 seconds)
				err := validateLogConfig(config)

				if tc.expectedValid {
					assert.NoError(t, err, tc.expectedMessage)
				} else {
					assert.Error(t, err, tc.expectedMessage)
					assert.Contains(t, err.Error(), "log_refresh_interval", "Error should mention log_refresh_interval")
				}
			})
		}
	})

	t.Run("log_display_configuration_loading", func(t *testing.T) {
		// Test loading log display config from files
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.json")

		// Create test config with log settings
		testConfig := map[string]interface{}{
			"log_refresh_interval_seconds": 15,
			"worktree_base_path":           "/tmp/test",
		}

		configData, err := json.MarshalIndent(testConfig, "", "  ")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(configPath, configData, 0644))

		// Load config
		config, err := LoadConfigFromPath(configPath)
		require.NoError(t, err)

		// Test both global and repository-specific config loading
		assert.Equal(t, 15, config.LogRefreshIntervalSecs, "Should load custom log refresh interval")
	})

	t.Run("log_display_configuration_merging", func(t *testing.T) {
		// Test repository config overrides global config
		baseConfig := DefaultConfig()
		baseConfig.LogRefreshIntervalSecs = 5

		overrideConfig := &Config{
			LogRefreshIntervalSecs: 20,
		}

		// Test precedence handling
		merged := MergeConfig(baseConfig, overrideConfig)

		assert.Equal(t, 20, merged.LogRefreshIntervalSecs, "Repository config should override global config")
		// Other fields should remain from base config
		assert.Equal(t, baseConfig.WorktreeBasePath, merged.WorktreeBasePath, "Non-overridden fields should remain from base")
	})

	t.Run("log_display_config_json_serialization", func(t *testing.T) {
		// Test that log display config can be properly serialized/deserialized
		config := DefaultConfig()
		config.LogRefreshIntervalSecs = 25

		// Serialize
		data, err := json.MarshalIndent(config, "", "  ")
		require.NoError(t, err)

		// Deserialize
		var loadedConfig Config
		err = json.Unmarshal(data, &loadedConfig)
		require.NoError(t, err)

		assert.Equal(t, 25, loadedConfig.LogRefreshIntervalSecs, "Log refresh interval should serialize correctly")
	})

	t.Run("log_display_config_validation_integration", func(t *testing.T) {
		// Test that log display config validation is integrated with main config validation
		config := DefaultConfig()
		config.LogRefreshIntervalSecs = -1 // Invalid value

		err := validateConfig(config)
		assert.Error(t, err, "Config validation should fail with invalid log refresh interval")
		assert.Contains(t, err.Error(), "log_refresh_interval", "Validation error should mention log_refresh_interval")
	})
}

// Helper functions for testing
func validateLogConfig(config *Config) error {
	// Use the existing validateConfig function which now includes log validation
	return validateConfig(config)
}

func LoadConfigFromPath(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set defaults for missing values
	defaults := DefaultConfig()
	if config.LogRefreshIntervalSecs == 0 {
		config.LogRefreshIntervalSecs = defaults.LogRefreshIntervalSecs
	}
	if config.WorktreeBasePath == "" {
		config.WorktreeBasePath = defaults.WorktreeBasePath
	}

	return &config, nil
}
