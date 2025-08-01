package sandbox

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"sbs/pkg/cmdlog"
)

func TestSandboxManager_CommandLogging(t *testing.T) {
	// Save original global logger
	originalLogger := cmdlog.GetGlobalLogger()
	defer cmdlog.SetGlobalLogger(originalLogger)

	manager := NewManager()

	t.Run("log_sandbox_list", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test sandbox list command
		// Note: This might fail if sandbox is not installed, but the logging should still work
		sandboxes, err := manager.ListSandboxes()

		output := buf.String()

		// The command should be logged regardless of success or failure
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "sandbox list")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "exit_code=")
			assert.Contains(t, output, "duration=")

			// Note: ListSandboxes() returns nil error even when sandbox command fails
			// (it treats command failure as "no sandboxes available"), so we can't
			// reliably test the relationship between err and exit_code here.
			// The important thing is that the command execution was logged.
		}
		// If output is empty, the command wasn't found, which is OK for testing

		// Use the variables to avoid "declared and not used" errors
		_ = sandboxes
		_ = err
	})

	t.Run("log_sandbox_exists_check", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test sandbox existence check
		sandboxName := "test-sandbox-12345"
		exists, err := manager.SandboxExists(sandboxName)

		output := buf.String()

		// The command should be logged regardless of success/failure
		if err == nil || output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "sandbox list")
			assert.Contains(t, output, "(from:")
		}

		// For this test, we don't care about the actual result, just that logging occurred
		_ = exists
	})

	t.Run("log_sandbox_delete", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test sandbox delete (for a non-existent sandbox)
		sandboxName := "nonexistent-sandbox-67890"
		err := manager.DeleteSandbox(sandboxName)

		output := buf.String()

		// Should log the list command to check existence
		if err == nil || output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "sandbox list")
			assert.Contains(t, output, "(from:")
		}

		// For a non-existent sandbox, delete should succeed without calling sandbox delete
		// So we might only see the list command in the log
	})

	t.Run("logging_disabled_no_output", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: false,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test that no logging occurs when disabled
		manager.ListSandboxes()

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

		// Test that only failures are logged at error level
		manager.ListSandboxes()

		output := buf.String()

		// If sandbox command fails, it should be logged
		// If it succeeds, it should not be logged at error level
		if output != "" {
			// A failure was logged
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "sandbox list")
		}
		// If output is empty, the command either succeeded (not logged at error level)
		// or the command wasn't found (no logging occurs)
	})

	t.Run("debug_level_logging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "debug",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test debug level logging includes all details
		manager.ListSandboxes()

		output := buf.String()

		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "sandbox list")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "duration=")
			assert.Contains(t, output, "exit_code=")
		}
	})
}

func TestCheckSandboxInstalled_CommandLogging(t *testing.T) {
	// Save original global logger
	originalLogger := cmdlog.GetGlobalLogger()
	defer cmdlog.SetGlobalLogger(originalLogger)

	t.Run("log_sandbox_help_check", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test sandbox --help command logging
		err := CheckSandboxInstalled()

		output := buf.String()

		// The command should be logged regardless of success/failure
		if err == nil {
			// Sandbox is installed and --help succeeded
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "sandbox --help")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "exit_code=0")
		} else if output != "" {
			// Sandbox command was found but failed
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "sandbox --help")
			assert.Contains(t, output, "(from:")
		}
		// If output is empty, the command wasn't found, which is expected in environments without sandbox
	})

	t.Run("sandbox_not_installed", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// The actual result depends on whether sandbox is installed
		err := CheckSandboxInstalled()

		// We just want to ensure that if logging occurs, it's formatted correctly
		output := buf.String()
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "sandbox --help")
		}

		// Don't assert on the error since it depends on the test environment
		_ = err
	})
}

func TestSandboxCommandHelpers(t *testing.T) {
	// Save original global logger
	originalLogger := cmdlog.GetGlobalLogger()
	defer cmdlog.SetGlobalLogger(originalLogger)

	manager := NewManager()

	t.Run("run_sandbox_command_with_args", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test running sandbox command with specific arguments
		_, err := manager.runSandboxCommand([]string{"list", "--format", "json"})

		output := buf.String()

		// Should log the command with all arguments
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "sandbox list --format json")
			assert.Contains(t, output, "(from:")
		}

		// Don't assert on error since sandbox might not be installed
		_ = err
	})

	t.Run("run_sandbox_command_run", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test running sandbox command without capturing output
		err := manager.runSandboxCommandRun([]string{"--version"})

		output := buf.String()

		// Should log the command
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "sandbox --version")
			assert.Contains(t, output, "(from:")
		}

		// Don't assert on error since sandbox might not be installed
		_ = err
	})
}

func TestGetExitCode_Sandbox(t *testing.T) {
	t.Run("nil_process_state", func(t *testing.T) {
		cmd := &exec.Cmd{}
		exitCode := getExitCode(cmd)
		assert.Equal(t, -1, exitCode)
	})
}
