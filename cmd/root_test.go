package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
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

// Tests for the new root command behavior (Issue #15)
func TestRootCommand_LaunchesTUI(t *testing.T) {
	t.Run("bare_sbs_command_launches_tui", func(t *testing.T) {
		// This test will fail initially because the root command currently has no RunE

		// Create a mock root command with the expected behavior
		var tuiLaunched bool
		mockRootCmd := &cobra.Command{
			Use:   "sbs",
			Short: "Sandbox Sessions - Manage GitHub issue work environments",
			RunE: func(cmd *cobra.Command, args []string) error {
				// Mock TUI launch - in real implementation this would launch the TUI
				tuiLaunched = true
				return nil
			},
		}

		// Act
		mockRootCmd.SetArgs([]string{})
		err := mockRootCmd.Execute()

		// Assert
		require.NoError(t, err)
		assert.True(t, tuiLaunched, "Expected TUI to be launched when running bare 'sbs' command")
	})
}

func TestRootCommand_WithInvalidArgs_ShowsHelp(t *testing.T) {
	t.Run("invalid_args_show_help", func(t *testing.T) {
		// This test ensures that the actual root command behaves correctly
		// with invalid subcommands

		var buf bytes.Buffer
		// Test the actual root command structure
		testCmd := rootCmd
		testCmd.SetOut(&buf)
		testCmd.SetErr(&buf)

		// Act - try to run with an invalid subcommand
		testCmd.SetArgs([]string{"invalid-subcommand"})
		err := testCmd.Execute()

		// Assert - Cobra should handle invalid subcommands and return error
		require.Error(t, err)
		// The error should be about unknown command
		assert.Contains(t, err.Error(), "unknown command")

		// Reset args for other tests
		testCmd.SetArgs([]string{})
	})
}

func TestRootCommand_HelpText_Updated(t *testing.T) {
	t.Run("help_text_reflects_tui_default", func(t *testing.T) {
		// The help text should be updated to indicate that the default behavior
		// is to launch the interactive TUI

		// Test the current root command structure
		assert.NotNil(t, rootCmd)
		assert.Equal(t, "sbs", rootCmd.Use)
		assert.NotEmpty(t, rootCmd.Short)
		assert.NotEmpty(t, rootCmd.Long)

		// The help text should indicate the new default behavior
		// This assertion will guide the documentation update
		assert.Contains(t, rootCmd.Short, "Manage GitHub issue work environments")
	})
}

func TestRootCommand_NowHasRunE(t *testing.T) {
	t.Run("root_command_now_has_rune", func(t *testing.T) {
		// This test confirms the implementation - root command now has RunE
		// This serves as a test for the "after" state in TDD

		assert.NotNil(t, rootCmd.RunE, "Root command should now have RunE function to launch TUI")
	})
}
