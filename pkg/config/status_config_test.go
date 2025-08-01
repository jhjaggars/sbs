package config

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusConfig_DefaultConfiguration(t *testing.T) {
	config := DefaultConfig()

	assert.True(t, config.StatusTracking, "Status tracking should be enabled by default")
	assert.Equal(t, 60, config.StatusRefreshIntervalSecs, "Default refresh interval should be 60 seconds")
	assert.Equal(t, 1048576, config.StatusMaxFileSizeBytes, "Default max file size should be 1MB")
	assert.Equal(t, 5, config.StatusTimeoutSeconds, "Default timeout should be 5 seconds")
}

func TestStatusConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid status config",
			config: Config{
				WorktreeBasePath:          "/tmp/worktrees",
				StatusTracking:            true,
				StatusRefreshIntervalSecs: 30,
				StatusMaxFileSizeBytes:    2048,
				StatusTimeoutSeconds:      10,
			},
			expectError: false,
		},
		{
			name: "refresh interval too low",
			config: Config{
				WorktreeBasePath:          "/tmp/worktrees",
				StatusTracking:            true,
				StatusRefreshIntervalSecs: 3,
				StatusMaxFileSizeBytes:    1048576,
				StatusTimeoutSeconds:      5,
			},
			expectError: true,
			errorMsg:    "status_refresh_interval_seconds must be between 5 and 600",
		},
		{
			name: "refresh interval too high",
			config: Config{
				WorktreeBasePath:          "/tmp/worktrees",
				StatusTracking:            true,
				StatusRefreshIntervalSecs: 700,
				StatusMaxFileSizeBytes:    1048576,
				StatusTimeoutSeconds:      5,
			},
			expectError: true,
			errorMsg:    "status_refresh_interval_seconds must be between 5 and 600",
		},
		{
			name: "max file size too small",
			config: Config{
				WorktreeBasePath:          "/tmp/worktrees",
				StatusTracking:            true,
				StatusRefreshIntervalSecs: 60,
				StatusMaxFileSizeBytes:    500,
				StatusTimeoutSeconds:      5,
			},
			expectError: true,
			errorMsg:    "status_max_file_size_bytes must be between 1KB and 10MB",
		},
		{
			name: "max file size too large",
			config: Config{
				WorktreeBasePath:          "/tmp/worktrees",
				StatusTracking:            true,
				StatusRefreshIntervalSecs: 60,
				StatusMaxFileSizeBytes:    11 * 1024 * 1024,
				StatusTimeoutSeconds:      5,
			},
			expectError: true,
			errorMsg:    "status_max_file_size_bytes must be between 1KB and 10MB",
		},
		{
			name: "timeout too low",
			config: Config{
				WorktreeBasePath:          "/tmp/worktrees",
				StatusTracking:            true,
				StatusRefreshIntervalSecs: 60,
				StatusMaxFileSizeBytes:    1048576,
				StatusTimeoutSeconds:      0,
			},
			expectError: true,
			errorMsg:    "status_timeout_seconds must be between 1 and 30",
		},
		{
			name: "timeout too high",
			config: Config{
				WorktreeBasePath:          "/tmp/worktrees",
				StatusTracking:            true,
				StatusRefreshIntervalSecs: 60,
				StatusMaxFileSizeBytes:    1048576,
				StatusTimeoutSeconds:      35,
			},
			expectError: true,
			errorMsg:    "status_timeout_seconds must be between 1 and 30",
		},
		{
			name: "status tracking disabled - no validation",
			config: Config{
				WorktreeBasePath:          "/tmp/worktrees",
				StatusTracking:            false,
				StatusRefreshIntervalSecs: 1,   // Would normally be invalid
				StatusMaxFileSizeBytes:    100, // Would normally be invalid
				StatusTimeoutSeconds:      0,   // Would normally be invalid
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(&tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStatusConfig_JSONSerialization(t *testing.T) {
	config := Config{
		WorktreeBasePath:          "/tmp/worktrees",
		StatusTracking:            true,
		StatusRefreshIntervalSecs: 45,
		StatusMaxFileSizeBytes:    2048000,
		StatusTimeoutSeconds:      8,
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)

	var deserialized Config
	err = json.Unmarshal(data, &deserialized)
	require.NoError(t, err)

	assert.Equal(t, config.StatusTracking, deserialized.StatusTracking)
	assert.Equal(t, config.StatusRefreshIntervalSecs, deserialized.StatusRefreshIntervalSecs)
	assert.Equal(t, config.StatusMaxFileSizeBytes, deserialized.StatusMaxFileSizeBytes)
	assert.Equal(t, config.StatusTimeoutSeconds, deserialized.StatusTimeoutSeconds)
}

func TestStatusConfig_OmitEmptyFields(t *testing.T) {
	config := Config{
		WorktreeBasePath: "/tmp/worktrees",
		// Status tracking fields are all zero values - should be omitted
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)

	// Check that status tracking fields are not present in JSON
	var jsonObj map[string]interface{}
	err = json.Unmarshal(data, &jsonObj)
	require.NoError(t, err)

	assert.NotContains(t, jsonObj, "status_tracking")
	assert.NotContains(t, jsonObj, "status_refresh_interval_seconds")
	assert.NotContains(t, jsonObj, "status_max_file_size_bytes")
	assert.NotContains(t, jsonObj, "status_timeout_seconds")
}

func TestStatusConfig_MergeConfiguration(t *testing.T) {
	baseConfig := &Config{
		WorktreeBasePath:          "/base/worktrees",
		StatusTracking:            true,
		StatusRefreshIntervalSecs: 60,
		StatusMaxFileSizeBytes:    1048576,
		StatusTimeoutSeconds:      5,
	}

	overrideConfig := &Config{
		StatusTracking:            false,
		StatusRefreshIntervalSecs: 30,
		StatusMaxFileSizeBytes:    2048000,
		// StatusTimeoutSeconds not specified - should keep base value
	}

	merged := MergeConfig(baseConfig, overrideConfig)

	assert.Equal(t, "/base/worktrees", merged.WorktreeBasePath, "Base path should be preserved")
	assert.False(t, merged.StatusTracking, "Status tracking should be overridden")
	assert.Equal(t, 30, merged.StatusRefreshIntervalSecs, "Refresh interval should be overridden")
	assert.Equal(t, 2048000, merged.StatusMaxFileSizeBytes, "Max file size should be overridden")
	assert.Equal(t, 5, merged.StatusTimeoutSeconds, "Timeout should keep base value")
}

func TestStatusConfig_LoadConfigurationFromJSON(t *testing.T) {
	jsonConfig := `{
		"worktree_base_path": "/tmp/test-worktrees",
		"status_tracking": true,
		"status_refresh_interval_seconds": 45,
		"status_max_file_size_bytes": 2097152,
		"status_timeout_seconds": 10
	}`

	var config Config
	err := json.Unmarshal([]byte(jsonConfig), &config)
	require.NoError(t, err)

	assert.Equal(t, "/tmp/test-worktrees", config.WorktreeBasePath)
	assert.True(t, config.StatusTracking)
	assert.Equal(t, 45, config.StatusRefreshIntervalSecs)
	assert.Equal(t, 2097152, config.StatusMaxFileSizeBytes)
	assert.Equal(t, 10, config.StatusTimeoutSeconds)

	// Validate the loaded configuration
	err = validateConfig(&config)
	assert.NoError(t, err)
}

func TestStatusConfig_BackwardCompatibility(t *testing.T) {
	// Test that configurations without status tracking fields still work
	jsonConfig := `{
		"worktree_base_path": "/tmp/test-worktrees",
		"github_token": "test-token"
	}`

	var config Config
	err := json.Unmarshal([]byte(jsonConfig), &config)
	require.NoError(t, err)

	// Status tracking fields should have zero values
	assert.False(t, config.StatusTracking)
	assert.Equal(t, 0, config.StatusRefreshIntervalSecs)
	assert.Equal(t, 0, config.StatusMaxFileSizeBytes)
	assert.Equal(t, 0, config.StatusTimeoutSeconds)

	// Configuration should still be valid (status tracking is disabled)
	err = validateConfig(&config)
	assert.NoError(t, err)
}

func TestStatusConfig_RepositoryOverride(t *testing.T) {
	// Test that repository-specific config can override status tracking settings
	tmpDir := t.TempDir()

	// Create base config
	baseConfig := &Config{
		WorktreeBasePath:          filepath.Join(tmpDir, "worktrees"),
		StatusTracking:            true,
		StatusRefreshIntervalSecs: 60,
		StatusMaxFileSizeBytes:    1048576,
		StatusTimeoutSeconds:      5,
	}

	// Create override config (like from repository .sbs/config.json)
	overrideConfig := &Config{
		StatusTracking:            true,
		StatusRefreshIntervalSecs: 30, // Faster refresh for this repo
		StatusTimeoutSeconds:      3,  // Shorter timeout for this repo
	}

	merged := MergeConfig(baseConfig, overrideConfig)

	assert.True(t, merged.StatusTracking)
	assert.Equal(t, 30, merged.StatusRefreshIntervalSecs, "Should use repository-specific refresh interval")
	assert.Equal(t, 1048576, merged.StatusMaxFileSizeBytes, "Should keep global max file size")
	assert.Equal(t, 3, merged.StatusTimeoutSeconds, "Should use repository-specific timeout")
}
