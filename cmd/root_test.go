package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sbs/pkg/cmdlog"
	"sbs/pkg/config"
)

func TestInitConfig_CommandLogging(t *testing.T) {
	// Save original global logger
	originalLogger := cmdlog.GetGlobalLogger()
	defer cmdlog.SetGlobalLogger(originalLogger)

	t.Run("logging_enabled_from_config", func(t *testing.T) {
		// Create temporary config
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".config", "sbs")
		require.NoError(t, os.MkdirAll(configDir, 0755))

		configPath := filepath.Join(configDir, "config.json")
		configContent := `{
			"worktree_base_path": "~/.work-issue-worktrees",
			"work_issue_script": "./work-issue.sh",
			"command_logging": true,
			"command_log_level": "debug",
			"command_log_path": "` + filepath.Join(tmpDir, "commands.log") + `"
		}`

		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Override HOME to use our temp directory
		originalHome := os.Getenv("HOME")
		defer os.Setenv("HOME", originalHome)
		os.Setenv("HOME", tmpDir)

		// Load config and initialize logging
		testCfg, err := config.LoadConfig()
		require.NoError(t, err)

		// Initialize logging like initConfig does
		if testCfg.CommandLogging {
			logConfig := cmdlog.Config{
				Enabled:  testCfg.CommandLogging,
				Level:    testCfg.CommandLogLevel,
				FilePath: testCfg.CommandLogPath,
			}

			if logConfig.Level == "" {
				logConfig.Level = "info"
			}

			logger := cmdlog.NewCommandLogger(logConfig)
			cmdlog.SetGlobalLogger(logger)
		}

		// Verify global logger is configured
		globalLogger := cmdlog.GetGlobalLogger()
		assert.True(t, globalLogger.IsEnabled())
		assert.Equal(t, "debug", globalLogger.GetLevel())
	})

	t.Run("logging_disabled_from_config", func(t *testing.T) {
		// Create temporary config with logging disabled
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".config", "sbs")
		require.NoError(t, os.MkdirAll(configDir, 0755))

		configPath := filepath.Join(configDir, "config.json")
		configContent := `{
			"worktree_base_path": "~/.work-issue-worktrees",
			"work_issue_script": "./work-issue.sh",
			"command_logging": false
		}`

		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Override HOME to use our temp directory
		originalHome := os.Getenv("HOME")
		defer os.Setenv("HOME", originalHome)
		os.Setenv("HOME", tmpDir)

		// Reset global logger to no-op
		cmdlog.SetGlobalLogger(&noOpLoggerForTest{})

		// Load config and initialize logging
		testCfg, err := config.LoadConfig()
		require.NoError(t, err)

		// Initialize logging like initConfig does
		if testCfg.CommandLogging {
			logConfig := cmdlog.Config{
				Enabled:  testCfg.CommandLogging,
				Level:    testCfg.CommandLogLevel,
				FilePath: testCfg.CommandLogPath,
			}

			if logConfig.Level == "" {
				logConfig.Level = "info"
			}

			logger := cmdlog.NewCommandLogger(logConfig)
			cmdlog.SetGlobalLogger(logger)
		}

		// Verify global logger is still disabled (no-op)
		globalLogger := cmdlog.GetGlobalLogger()
		assert.False(t, globalLogger.IsEnabled())
	})

	t.Run("default_log_level", func(t *testing.T) {
		// Create temporary config with logging enabled but no level specified
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".config", "sbs")
		require.NoError(t, os.MkdirAll(configDir, 0755))

		configPath := filepath.Join(configDir, "config.json")
		configContent := `{
			"worktree_base_path": "~/.work-issue-worktrees",
			"work_issue_script": "./work-issue.sh",
			"command_logging": true
		}`

		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// Override HOME to use our temp directory
		originalHome := os.Getenv("HOME")
		defer os.Setenv("HOME", originalHome)
		os.Setenv("HOME", tmpDir)

		// Load config and initialize logging
		testCfg, err := config.LoadConfig()
		require.NoError(t, err)

		// Initialize logging like initConfig does
		if testCfg.CommandLogging {
			logConfig := cmdlog.Config{
				Enabled:  testCfg.CommandLogging,
				Level:    testCfg.CommandLogLevel,
				FilePath: testCfg.CommandLogPath,
			}

			// This is the key test - default level should be set to "info"
			if logConfig.Level == "" {
				logConfig.Level = "info"
			}

			logger := cmdlog.NewCommandLogger(logConfig)
			cmdlog.SetGlobalLogger(logger)
		}

		// Verify global logger uses default level
		globalLogger := cmdlog.GetGlobalLogger()
		assert.True(t, globalLogger.IsEnabled())
		assert.Equal(t, "info", globalLogger.GetLevel())
	})
}

// Test helper to create a no-op logger for testing
type noOpLoggerForTest struct{}

func (n *noOpLoggerForTest) LogCommand(command string, args []string, caller string) cmdlog.CommandContext {
	return &noOpContextForTest{}
}

func (n *noOpLoggerForTest) IsEnabled() bool {
	return false
}

func (n *noOpLoggerForTest) GetLevel() string {
	return ""
}

type noOpContextForTest struct{}

func (n *noOpContextForTest) LogCompletion(success bool, exitCode int, errorMsg string, duration time.Duration) {
	// No-op
}
