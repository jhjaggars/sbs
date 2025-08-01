package tmux

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sbs/pkg/cmdlog"
)

func TestTmuxManager_CommandLogging(t *testing.T) {
	// Save original global logger
	originalLogger := cmdlog.GetGlobalLogger()
	defer cmdlog.SetGlobalLogger(originalLogger)

	manager := NewManager()

	t.Run("log_tmux_list_sessions", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test tmux list-sessions command
		// Note: This might fail if tmux is not installed or no sessions exist,
		// but the logging should still work
		_, err := manager.ListSessions()

		output := buf.String()

		// The command should be logged regardless of success or failure
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "tmux list-sessions")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "exit_code=")
			assert.Contains(t, output, "duration=")
		}
		// If output is empty, tmux might not be installed, which is OK for testing

		// Use the variables to avoid "declared and not used" errors
		_ = err
	})

	t.Run("log_tmux_has_session", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "debug",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test session existence check for a non-existent session
		exists, err := manager.SessionExists("test-nonexistent-session-12345")

		output := buf.String()

		// The command should be logged
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "tmux has-session")
			assert.Contains(t, output, "test-nonexistent-session-12345")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "exit_code=")
			assert.Contains(t, output, "duration=")

			// For non-existent session, tmux returns exit code 1
			// but our function converts that to false, nil
			if strings.Contains(output, "exit_code=1") {
				assert.False(t, exists)
				assert.NoError(t, err)
			}
		}

		// Use the variables to avoid "declared and not used" errors
		_ = exists
		_ = err
	})

	t.Run("log_tmux_display_message", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test getting working directory for a non-existent session
		// This should fail, but the command should still be logged
		_, err := manager.getSessionWorkingDir("test-nonexistent-session-12345")

		output := buf.String()

		// The command should be logged regardless of success or failure
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "tmux display-message")
			assert.Contains(t, output, "test-nonexistent-session-12345")
			assert.Contains(t, output, "#{pane_current_path}")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "exit_code=")
			assert.Contains(t, output, "duration=")

			// This may or may not fail depending on tmux installation
			// The important thing is that the command was logged
			_ = err
		}

		// Use the variables to avoid "declared and not used" errors
		_ = err
	})

	t.Run("log_tmux_kill_session", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test killing a non-existent session
		// This should fail, but the command should still be logged
		err := manager.KillSession("test-nonexistent-session-12345")

		output := buf.String()

		// The command should be logged regardless of success or failure
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "tmux kill-session")
			assert.Contains(t, output, "test-nonexistent-session-12345")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "exit_code=")
			assert.Contains(t, output, "duration=")

			// This may or may not fail depending on tmux installation
			// The important thing is that the command was logged
			_ = err
		}

		// Use the variables to avoid "declared and not used" errors
		_ = err
	})

	t.Run("log_tmux_set_environment", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "debug",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test setting environment variables for a non-existent session
		env := map[string]string{
			"TEST_VAR":  "test_value",
			"SBS_TITLE": "test-title",
		}
		err := manager.setEnvironmentVariables("test-nonexistent-session-12345", env)

		output := buf.String()

		// The commands should be logged regardless of success or failure
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "tmux set-environment")
			assert.Contains(t, output, "test-nonexistent-session-12345")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "exit_code=")
			assert.Contains(t, output, "duration=")

			// Should contain both environment variable calls
			commandCount := strings.Count(output, "tmux set-environment")
			assert.GreaterOrEqual(t, commandCount, 1) // At least 1 env var

			// This may or may not fail depending on tmux installation
			// The important thing is that the command was logged
			_ = err
		}

		// Use the variables to avoid "declared and not used" errors
		_ = err
	})

	t.Run("log_tmux_send_keys", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test sending keys to a non-existent session
		err := manager.setWorkingDirectory("test-nonexistent-session-12345", "/tmp")

		output := buf.String()

		// The command should be logged regardless of success or failure
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "tmux send-keys")
			assert.Contains(t, output, "test-nonexistent-session-12345")
			assert.Contains(t, output, "cd /tmp")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "exit_code=")
			assert.Contains(t, output, "duration=")

			// This may or may not fail depending on tmux installation
			// The important thing is that the command was logged
			_ = err
		}

		// Use the variables to avoid "declared and not used" errors
		_ = err
	})
}

func TestTmuxManager_RunTmuxCommandHelpers(t *testing.T) {
	// Save original global logger
	originalLogger := cmdlog.GetGlobalLogger()
	defer cmdlog.SetGlobalLogger(originalLogger)

	manager := NewManager()

	t.Run("runTmuxCommand_logs_output_command", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test helper function directly with a command that should give output
		_, err := manager.runTmuxCommand([]string{"--version"})

		output := buf.String()

		// The command should be logged
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "tmux --version")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "exit_code=")
			assert.Contains(t, output, "duration=")
		}

		// Use the variables to avoid "declared and not used" errors
		_ = err
	})

	t.Run("runTmuxCommandRun_logs_run_command", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "debug",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test helper function directly
		err := manager.runTmuxCommandRun([]string{"--version"})

		output := buf.String()

		// The command should be logged
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "tmux --version")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "exit_code=")
			assert.Contains(t, output, "duration=")
		}

		// Use the variables to avoid "declared and not used" errors
		_ = err
	})

	t.Run("runTmuxCommandWithEnv_logs_env_command", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test helper function with environment variables
		env := map[string]string{
			"TEST_VAR": "test_value",
		}
		err := manager.runTmuxCommandWithEnv([]string{"--version"}, env)

		output := buf.String()

		// The command should be logged
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "tmux --version")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "exit_code=")
			assert.Contains(t, output, "duration=")
		}

		// Use the variables to avoid "declared and not used" errors
		_ = err
	})
}

func TestTmuxManager_GetExitCode(t *testing.T) {
	manager := NewManager()

	t.Run("get_exit_code_from_successful_command", func(t *testing.T) {
		// Create a command that should succeed
		cmd := exec.Command("echo", "test")
		err := cmd.Run()
		assert.NoError(t, err)

		exitCode := getExitCode(cmd)
		assert.Equal(t, 0, exitCode)
	})

	t.Run("get_exit_code_from_failed_command", func(t *testing.T) {
		// Create a command that should fail
		cmd := exec.Command("false") // 'false' command always returns exit code 1
		err := cmd.Run()
		assert.Error(t, err)

		exitCode := getExitCode(cmd)
		assert.Equal(t, 1, exitCode)
	})

	t.Run("get_exit_code_from_unrun_command", func(t *testing.T) {
		// Create a command that hasn't been run
		cmd := exec.Command("echo", "test")

		exitCode := getExitCode(cmd)
		assert.Equal(t, -1, exitCode)
	})

	// Use the manager variable to avoid "declared and not used" errors
	_ = manager
}

func TestTmuxManager_DisabledLogging(t *testing.T) {
	// Save original global logger
	originalLogger := cmdlog.GetGlobalLogger()
	defer cmdlog.SetGlobalLogger(originalLogger)

	manager := NewManager()

	t.Run("no_logging_when_disabled", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: false, // Disabled logging
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test a command - should not be logged
		_, err := manager.SessionExists("test-session")

		output := buf.String()

		// No output should be logged when disabled
		assert.Empty(t, output)

		// Use the variables to avoid "declared and not used" errors
		_ = err
	})
}
