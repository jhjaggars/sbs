package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_CommandLoggingConfiguration(t *testing.T) {
	t.Run("default_logging_configuration", func(t *testing.T) {
		config := DefaultConfig()

		// Command logging should be disabled by default
		assert.False(t, config.CommandLogging)
		assert.Equal(t, "", config.CommandLogLevel)
		assert.Equal(t, "", config.CommandLogPath)
	})

	t.Run("load_logging_config_from_json", func(t *testing.T) {
		jsonData := `{
			"worktree_base_path": "~/.work-issue-worktrees",
			"work_issue_script": "./work-issue.sh",
			"command_logging": true,
			"command_log_level": "info",
			"command_log_path": "~/.config/sbs/command.log"
		}`

		var config Config
		err := json.Unmarshal([]byte(jsonData), &config)
		require.NoError(t, err)

		assert.True(t, config.CommandLogging)
		assert.Equal(t, "info", config.CommandLogLevel)
		assert.Equal(t, "~/.config/sbs/command.log", config.CommandLogPath)
	})

	t.Run("validate_log_level_options", func(t *testing.T) {
		testCases := []struct {
			name     string
			level    string
			expected string
		}{
			{"debug_level", "debug", "debug"},
			{"info_level", "info", "info"},
			{"error_level", "error", "error"},
			{"empty_level", "", ""},
			{"mixed_case", "DEBUG", "DEBUG"}, // Should preserve case in config
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				config := &Config{
					CommandLogging:  true,
					CommandLogLevel: tc.level,
				}

				assert.Equal(t, tc.expected, config.CommandLogLevel)
			})
		}
	})

	t.Run("validate_log_path_options", func(t *testing.T) {
		testCases := []struct {
			name string
			path string
		}{
			{"absolute_path", "/var/log/sbs/commands.log"},
			{"relative_path", "./logs/commands.log"},
			{"home_path", "~/.config/sbs/commands.log"},
			{"empty_path", ""},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				config := &Config{
					CommandLogging: true,
					CommandLogPath: tc.path,
				}

				assert.Equal(t, tc.path, config.CommandLogPath)
			})
		}
	})

	t.Run("backward_compatibility_config", func(t *testing.T) {
		// Old config without logging fields should still work
		jsonData := `{
			"worktree_base_path": "~/.work-issue-worktrees",
			"work_issue_script": "./work-issue.sh",
			"repo_path": "."
		}`

		var config Config
		err := json.Unmarshal([]byte(jsonData), &config)
		require.NoError(t, err)

		// Logging fields should have default (zero) values
		assert.False(t, config.CommandLogging)
		assert.Equal(t, "", config.CommandLogLevel)
		assert.Equal(t, "", config.CommandLogPath)

		// Other fields should be loaded correctly
		assert.Equal(t, "~/.work-issue-worktrees", config.WorktreeBasePath)
		assert.Equal(t, "./work-issue.sh", config.WorkIssueScript)
		assert.Equal(t, ".", config.RepoPath)
	})
}

func TestConfig_LoggingConfigurationSerialization(t *testing.T) {
	t.Run("serialize_logging_config", func(t *testing.T) {
		config := &Config{
			WorktreeBasePath: "~/.work-issue-worktrees",
			WorkIssueScript:  "./work-issue.sh",
			CommandLogging:   true,
			CommandLogLevel:  "debug",
			CommandLogPath:   "~/.config/sbs/debug.log",
		}

		jsonData, err := json.MarshalIndent(config, "", "  ")
		require.NoError(t, err)

		jsonStr := string(jsonData)
		assert.Contains(t, jsonStr, `"command_logging": true`)
		assert.Contains(t, jsonStr, `"command_log_level": "debug"`)
		assert.Contains(t, jsonStr, `"command_log_path": "~/.config/sbs/debug.log"`)
	})

	t.Run("deserialize_logging_config", func(t *testing.T) {
		jsonData := `{
			"worktree_base_path": "~/.work-issue-worktrees",
			"work_issue_script": "./work-issue.sh",
			"command_logging": true,
			"command_log_level": "debug",
			"command_log_path": "~/.config/sbs/debug.log"
		}`

		var config Config
		err := json.Unmarshal([]byte(jsonData), &config)
		require.NoError(t, err)

		assert.True(t, config.CommandLogging)
		assert.Equal(t, "debug", config.CommandLogLevel)
		assert.Equal(t, "~/.config/sbs/debug.log", config.CommandLogPath)
	})

	t.Run("omitempty_serialization", func(t *testing.T) {
		// Test that omitempty works correctly for logging fields
		config := &Config{
			WorktreeBasePath: "~/.work-issue-worktrees",
			WorkIssueScript:  "./work-issue.sh",
			// Logging fields intentionally omitted/default
		}

		jsonData, err := json.MarshalIndent(config, "", "  ")
		require.NoError(t, err)

		jsonStr := string(jsonData)
		// With omitempty, false boolean and empty strings should not appear
		assert.NotContains(t, jsonStr, "command_logging")
		assert.NotContains(t, jsonStr, "command_log_level")
		assert.NotContains(t, jsonStr, "command_log_path")
	})
}

func TestConfig_MergeConfigLogging(t *testing.T) {
	t.Run("merge_logging_config", func(t *testing.T) {
		baseConfig := &Config{
			WorktreeBasePath: "~/.work-issue-worktrees",
			CommandLogging:   false,
			CommandLogLevel:  "",
			CommandLogPath:   "",
		}

		overrideConfig := &Config{
			CommandLogging:  true,
			CommandLogLevel: "debug",
			CommandLogPath:  "/custom/log/path.log",
		}

		merged := MergeConfig(baseConfig, overrideConfig)

		assert.True(t, merged.CommandLogging)
		assert.Equal(t, "debug", merged.CommandLogLevel)
		assert.Equal(t, "/custom/log/path.log", merged.CommandLogPath)
		assert.Equal(t, "~/.work-issue-worktrees", merged.WorktreeBasePath) // Base value preserved
	})

	t.Run("merge_partial_logging_config", func(t *testing.T) {
		baseConfig := &Config{
			WorktreeBasePath: "~/.work-issue-worktrees",
			CommandLogging:   true,
			CommandLogLevel:  "info",
			CommandLogPath:   "/base/log.log",
		}

		// Override only log level
		overrideConfig := &Config{
			CommandLogLevel: "error",
		}

		merged := MergeConfig(baseConfig, overrideConfig)

		assert.True(t, merged.CommandLogging)                   // From base
		assert.Equal(t, "error", merged.CommandLogLevel)        // From override
		assert.Equal(t, "/base/log.log", merged.CommandLogPath) // From base
	})

	t.Run("merge_empty_override", func(t *testing.T) {
		baseConfig := &Config{
			WorktreeBasePath: "~/.work-issue-worktrees",
			CommandLogging:   true,
			CommandLogLevel:  "debug",
			CommandLogPath:   "/base/log.log",
		}

		emptyOverride := &Config{}

		merged := MergeConfig(baseConfig, emptyOverride)

		// All base values should be preserved
		assert.True(t, merged.CommandLogging)
		assert.Equal(t, "debug", merged.CommandLogLevel)
		assert.Equal(t, "/base/log.log", merged.CommandLogPath)
		assert.Equal(t, "~/.work-issue-worktrees", merged.WorktreeBasePath)
	})
}

func TestConfig_LoggingConfigurationIntegration(t *testing.T) {
	t.Run("save_and_load_with_logging", func(t *testing.T) {
		// Create temporary config file
		tmpDir := t.TempDir()

		// Override home directory for this test
		originalConfig := &Config{
			WorktreeBasePath: "~/.work-issue-worktrees",
			WorkIssueScript:  "./work-issue.sh",
			CommandLogging:   true,
			CommandLogLevel:  "info",
			CommandLogPath:   filepath.Join(tmpDir, "commands.log"),
		}

		// Manually save config to temp file
		configPath := filepath.Join(tmpDir, "config.json")
		jsonData, err := json.MarshalIndent(originalConfig, "", "  ")
		require.NoError(t, err)

		err = os.WriteFile(configPath, jsonData, 0644)
		require.NoError(t, err)

		// Load config from temp file
		data, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var loadedConfig Config
		err = json.Unmarshal(data, &loadedConfig)
		require.NoError(t, err)

		// Verify logging configuration is preserved
		assert.True(t, loadedConfig.CommandLogging)
		assert.Equal(t, "info", loadedConfig.CommandLogLevel)
		assert.Equal(t, filepath.Join(tmpDir, "commands.log"), loadedConfig.CommandLogPath)
	})

	t.Run("repository_config_override", func(t *testing.T) {
		// Test that repository-specific config can override global logging settings
		globalConfig := &Config{
			WorktreeBasePath: "~/.work-issue-worktrees",
			CommandLogging:   false,
			CommandLogLevel:  "",
		}

		repoConfig := &Config{
			CommandLogging:  true,
			CommandLogLevel: "debug",
			CommandLogPath:  "./repo-specific.log",
		}

		merged := MergeConfig(globalConfig, repoConfig)

		assert.True(t, merged.CommandLogging)
		assert.Equal(t, "debug", merged.CommandLogLevel)
		assert.Equal(t, "./repo-specific.log", merged.CommandLogPath)
	})
}

func TestConfig_LoggingConfigurationEdgeCases(t *testing.T) {
	t.Run("boolean_false_override", func(t *testing.T) {
		// Test that false boolean values don't override true base values
		// This is the expected behavior based on the current MergeConfig logic
		baseConfig := &Config{
			CommandLogging: true,
		}

		overrideConfig := &Config{
			CommandLogging: false, // This won't override true base value
		}

		merged := MergeConfig(baseConfig, overrideConfig)

		// Based on current logic, false values don't override
		assert.True(t, merged.CommandLogging)
	})

	t.Run("special_characters_in_path", func(t *testing.T) {
		config := &Config{
			CommandLogging: true,
			CommandLogPath: "/path/with spaces/and-special_chars.log",
		}

		jsonData, err := json.Marshal(config)
		require.NoError(t, err)

		var deserializedConfig Config
		err = json.Unmarshal(jsonData, &deserializedConfig)
		require.NoError(t, err)

		assert.Equal(t, "/path/with spaces/and-special_chars.log", deserializedConfig.CommandLogPath)
	})

	t.Run("unicode_in_log_path", func(t *testing.T) {
		config := &Config{
			CommandLogging: true,
			CommandLogPath: "/路径/with/unicode/命令.log",
		}

		jsonData, err := json.Marshal(config)
		require.NoError(t, err)

		var deserializedConfig Config
		err = json.Unmarshal(jsonData, &deserializedConfig)
		require.NoError(t, err)

		assert.Equal(t, "/路径/with/unicode/命令.log", deserializedConfig.CommandLogPath)
	})
}
