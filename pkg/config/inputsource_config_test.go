package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInputSourceConfig_LoadFromProject(t *testing.T) {
	// Test loading .sbs/input-source.json from project root
	tempDir := t.TempDir()

	// Create test configuration
	sbsDir := filepath.Join(tempDir, ".sbs")
	require.NoError(t, os.MkdirAll(sbsDir, 0755))

	configData := `{
        "type": "test",
        "settings": {
            "description": "Test project configuration"
        }
    }`

	configPath := filepath.Join(sbsDir, "input-source.json")
	require.NoError(t, os.WriteFile(configPath, []byte(configData), 0644))

	// Test loading
	config, err := LoadInputSourceConfig(tempDir)
	require.NoError(t, err)
	assert.Equal(t, "test", config.Type)
	assert.NotNil(t, config.Settings)

	description, exists := config.Settings["description"]
	assert.True(t, exists)
	assert.Equal(t, "Test project configuration", description)
}

func TestInputSourceConfig_FallbackToDefault(t *testing.T) {
	// Test that missing config falls back to GitHub
	tempDir := t.TempDir()

	config, err := LoadInputSourceConfig(tempDir)
	require.NoError(t, err)
	assert.Equal(t, "github", config.Type)
	assert.NotNil(t, config.Settings)
}

func TestInputSourceConfig_ValidationError(t *testing.T) {
	// Test invalid configuration
	tempDir := t.TempDir()
	sbsDir := filepath.Join(tempDir, ".sbs")
	require.NoError(t, os.MkdirAll(sbsDir, 0755))

	// Invalid JSON
	configPath := filepath.Join(sbsDir, "input-source.json")
	require.NoError(t, os.WriteFile(configPath, []byte(`{invalid json}`), 0644))

	config, err := LoadInputSourceConfig(tempDir)
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestInputSourceConfig_ConfigDirectoryCreation(t *testing.T) {
	// Test that .sbs directory is created if it doesn't exist
	tempDir := t.TempDir()

	config := &InputSourceConfig{
		Type: "test",
		Settings: map[string]interface{}{
			"description": "Test config",
		},
	}
	err := SaveInputSourceConfig(tempDir, config)
	require.NoError(t, err)

	// Verify directory and file were created
	configPath := filepath.Join(tempDir, ".sbs", "input-source.json")
	assert.FileExists(t, configPath)

	// Verify content is correct
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var savedConfig InputSourceConfig
	err = json.Unmarshal(data, &savedConfig)
	require.NoError(t, err)
	assert.Equal(t, "test", savedConfig.Type)
	assert.Equal(t, "Test config", savedConfig.Settings["description"])
}

func TestInputSourceConfig_ErrorScenarios(t *testing.T) {
	t.Run("corrupted_config_file", func(t *testing.T) {
		tempDir := t.TempDir()
		sbsDir := filepath.Join(tempDir, ".sbs")
		require.NoError(t, os.MkdirAll(sbsDir, 0755))

		// Create corrupted JSON
		configPath := filepath.Join(sbsDir, "input-source.json")
		require.NoError(t, os.WriteFile(configPath, []byte(`{corrupted json`), 0644))

		config, err := LoadInputSourceConfig(tempDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse")
		assert.Nil(t, config)
	})

	t.Run("permission_denied", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Permission tests not reliable on Windows")
		}

		tempDir := t.TempDir()
		sbsDir := filepath.Join(tempDir, ".sbs")
		require.NoError(t, os.MkdirAll(sbsDir, 0755))

		configPath := filepath.Join(sbsDir, "input-source.json")
		require.NoError(t, os.WriteFile(configPath, []byte(`{"type": "test"}`), 0000)) // No read permissions

		config, err := LoadInputSourceConfig(tempDir)
		assert.Error(t, err)
		assert.Nil(t, config)
	})

	t.Run("empty_type", func(t *testing.T) {
		tempDir := t.TempDir()
		sbsDir := filepath.Join(tempDir, ".sbs")
		require.NoError(t, os.MkdirAll(sbsDir, 0755))

		configData := `{"type": "", "settings": {}}`
		configPath := filepath.Join(sbsDir, "input-source.json")
		require.NoError(t, os.WriteFile(configPath, []byte(configData), 0644))

		config, err := LoadInputSourceConfig(tempDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "type cannot be empty")
		assert.Nil(t, config)
	})

	t.Run("missing_type", func(t *testing.T) {
		tempDir := t.TempDir()
		sbsDir := filepath.Join(tempDir, ".sbs")
		require.NoError(t, os.MkdirAll(sbsDir, 0755))

		configData := `{"settings": {}}`
		configPath := filepath.Join(sbsDir, "input-source.json")
		require.NoError(t, os.WriteFile(configPath, []byte(configData), 0644))

		config, err := LoadInputSourceConfig(tempDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "type cannot be empty")
		assert.Nil(t, config)
	})
}

func TestInputSourceConfig_DefaultConfiguration(t *testing.T) {
	config := DefaultInputSourceConfig()

	assert.Equal(t, "github", config.Type)
	assert.NotNil(t, config.Settings)
	assert.Contains(t, config.Settings, "repository")
	assert.Equal(t, "auto-detect", config.Settings["repository"])
}

func TestInputSourceConfig_DifferentTypes(t *testing.T) {
	tests := []struct {
		name       string
		configData string
		expected   string
	}{
		{
			name:       "github_config",
			configData: `{"type": "github", "settings": {"repository": "owner/repo"}}`,
			expected:   "github",
		},
		{
			name:       "test_config",
			configData: `{"type": "test", "settings": {"description": "Test project"}}`,
			expected:   "test",
		},
		{
			name:       "jira_config",
			configData: `{"type": "jira", "settings": {"url": "https://company.atlassian.net", "project": "PROJ"}}`,
			expected:   "jira",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			sbsDir := filepath.Join(tempDir, ".sbs")
			require.NoError(t, os.MkdirAll(sbsDir, 0755))

			configPath := filepath.Join(sbsDir, "input-source.json")
			require.NoError(t, os.WriteFile(configPath, []byte(tt.configData), 0644))

			config, err := LoadInputSourceConfig(tempDir)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, config.Type)
		})
	}
}

func TestInputSourceConfig_SaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()

	originalConfig := &InputSourceConfig{
		Type: "test",
		Settings: map[string]interface{}{
			"description": "Test save and load",
			"number":      42,
			"boolean":     true,
		},
	}

	// Save configuration
	err := SaveInputSourceConfig(tempDir, originalConfig)
	require.NoError(t, err)

	// Load configuration
	loadedConfig, err := LoadInputSourceConfig(tempDir)
	require.NoError(t, err)

	// Verify they match
	assert.Equal(t, originalConfig.Type, loadedConfig.Type)
	assert.Equal(t, originalConfig.Settings["description"], loadedConfig.Settings["description"])
	assert.Equal(t, float64(42), loadedConfig.Settings["number"]) // JSON unmarshals numbers as float64
	assert.Equal(t, true, loadedConfig.Settings["boolean"])
}

func TestInputSourceConfig_AllowCrossSource(t *testing.T) {
	t.Run("allow_cross_source_true", func(t *testing.T) {
		config := &InputSourceConfig{
			Type: "github",
			Settings: map[string]interface{}{
				"allow_cross_source": true,
			},
		}

		assert.True(t, config.AllowCrossSource())
	})

	t.Run("allow_cross_source_false", func(t *testing.T) {
		config := &InputSourceConfig{
			Type: "github",
			Settings: map[string]interface{}{
				"allow_cross_source": false,
			},
		}

		assert.False(t, config.AllowCrossSource())
	})

	t.Run("allow_cross_source_missing", func(t *testing.T) {
		config := &InputSourceConfig{
			Type: "github",
			Settings: map[string]interface{}{
				"repository": "auto-detect",
			},
		}

		assert.False(t, config.AllowCrossSource())
	})

	t.Run("allow_cross_source_wrong_type", func(t *testing.T) {
		config := &InputSourceConfig{
			Type: "github",
			Settings: map[string]interface{}{
				"allow_cross_source": "yes", // string instead of bool
			},
		}

		assert.False(t, config.AllowCrossSource())
	})

	t.Run("nil_settings", func(t *testing.T) {
		config := &InputSourceConfig{
			Type:     "github",
			Settings: nil,
		}

		assert.False(t, config.AllowCrossSource())
	})

	t.Run("default_config_behavior", func(t *testing.T) {
		// Test that the default config has the expected cross-source setting
		config := DefaultInputSourceConfig()

		assert.False(t, config.AllowCrossSource(), "Default config should deny cross-source usage")

		// Verify the setting is explicitly set
		assert.Contains(t, config.Settings, "allow_cross_source")
		assert.Equal(t, false, config.Settings["allow_cross_source"])
	})
}
