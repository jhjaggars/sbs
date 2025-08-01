package validation

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"sbs/pkg/cmdlog"
)

func TestValidationTools_CommandLogging(t *testing.T) {
	// Save original global logger
	originalLogger := cmdlog.GetGlobalLogger()
	defer cmdlog.SetGlobalLogger(originalLogger)

	t.Run("log_tool_version_checks", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test tmux version check (this should work on most systems)
		err := checkTmux()

		// Check if tmux is available - if not, skip this part of the test
		if err != nil && buf.String() == "" {
			t.Skip("tmux not available on this system, skipping tmux logging test")
		}

		output := buf.String()
		if err == nil {
			// If tmux succeeded, verify it was logged
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "tmux -V")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "exit_code=0")
		} else {
			// If tmux failed, verify the failure was logged
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "tmux -V")
			assert.Contains(t, output, "exit_code=")
		}
	})

	t.Run("log_git_version_check", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test git version check (git should be available in development environment)
		err := checkGit()

		// Check if git is available - if not, skip this part of the test
		if err != nil && buf.String() == "" {
			t.Skip("git not available on this system, skipping git logging test")
		}

		output := buf.String()
		if err == nil {
			// If git succeeded, verify it was logged
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "git --version")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "exit_code=0")
		} else {
			// If git failed, verify the failure was logged
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "git --version")
			assert.Contains(t, output, "exit_code=")
		}
	})

	t.Run("log_tool_availability_checks", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test a command that likely doesn't exist
		err := runValidationCommand("nonexistent-command-12345", []string{"--version"}, "Command not found")

		// Should fail
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Command not found")

		output := buf.String()
		assert.Contains(t, output, "[COMMAND]")
		assert.Contains(t, output, "nonexistent-command-12345 --version")
		assert.Contains(t, output, "(from:")
		// Exit code might vary, but there should be an error logged
		assert.Contains(t, output, "exit_code=")
	})

	t.Run("logging_disabled_no_output", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: false,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Run git version check
		checkGit()

		// Should produce no log output when disabled
		output := buf.String()
		assert.Empty(t, output)
	})

	t.Run("error_level_logging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "error",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test successful command (should not be logged at error level)
		checkGit()

		// Test failing command (should be logged at error level)
		runValidationCommand("nonexistent-command-67890", []string{"--version"}, "Command not found")

		output := buf.String()

		// Should not contain successful git command
		assert.NotContains(t, output, "git --version")

		// Should contain failed command
		assert.Contains(t, output, "nonexistent-command-67890 --version")
		assert.Contains(t, output, "[COMMAND]")
	})

	t.Run("debug_level_logging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "debug",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test git version check
		checkGit()

		output := buf.String()
		assert.Contains(t, output, "[COMMAND]")
		assert.Contains(t, output, "git --version")
		assert.Contains(t, output, "(from:")
		assert.Contains(t, output, "duration=")
		assert.Contains(t, output, "exit_code=")
	})
}

func TestRunValidationCommand(t *testing.T) {
	// Save original global logger
	originalLogger := cmdlog.GetGlobalLogger()
	defer cmdlog.SetGlobalLogger(originalLogger)

	t.Run("successful_command", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Run a command that should succeed
		err := runValidationCommand("echo", []string{"test"}, "Echo failed")

		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "[COMMAND]")
		assert.Contains(t, output, "echo test")
		assert.Contains(t, output, "exit_code=0")
	})

	t.Run("failed_command", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Run a command that should fail
		err := runValidationCommand("false", []string{}, "False command failed")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "False command failed")

		output := buf.String()
		assert.Contains(t, output, "[COMMAND]")
		assert.Contains(t, output, "false")
		assert.Contains(t, output, "exit_code=1")
	})

	t.Run("command_with_arguments", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Run echo with multiple arguments
		err := runValidationCommand("echo", []string{"hello", "world"}, "Echo failed")

		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "[COMMAND]")
		assert.Contains(t, output, "echo hello world")
		assert.Contains(t, output, "exit_code=0")
	})
}

func TestGetExitCode(t *testing.T) {
	t.Run("successful_command", func(t *testing.T) {
		cmd := &exec.Cmd{}
		// Simulate a successful command
		exitCode := getExitCode(cmd)
		assert.Equal(t, -1, exitCode) // ProcessState is nil
	})
}
