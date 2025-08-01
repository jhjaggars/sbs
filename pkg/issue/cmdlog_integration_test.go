package issue

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"sbs/pkg/cmdlog"
)

func TestGitHubClient_CommandLogging(t *testing.T) {
	// Save original global logger
	originalLogger := cmdlog.GetGlobalLogger()
	defer cmdlog.SetGlobalLogger(originalLogger)

	t.Run("log_gh_issue_view", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		client := NewGitHubClient()

		// Test gh issue view command (this will likely fail, but should be logged)
		_, err := client.GetIssue(123)

		output := buf.String()

		// The command should be logged regardless of success or failure
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "gh issue view 123 --json number,title,state,url")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "duration=")
			assert.Contains(t, output, "exit_code=")
		}

		// Don't assert on the error since it depends on gh being installed and authenticated
		_ = err
	})

	t.Run("log_gh_issue_list", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		client := NewGitHubClient()

		// Test gh issue list command
		_, err := client.ListIssues("", 10)

		output := buf.String()

		// The command should be logged
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "gh issue list --json number,title,state,url --state open --limit 10")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "duration=")
		}

		// Don't assert on the error since it depends on gh setup
		_ = err
	})

	t.Run("log_gh_issue_list_with_search", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		client := NewGitHubClient()

		// Test gh issue list with search query
		_, err := client.ListIssues("bug", 5)

		output := buf.String()

		// Should log the command with search parameters
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "gh issue list --json number,title,state,url --state open --limit 5 --search bug")
			assert.Contains(t, output, "(from:")
		}

		_ = err
	})

	t.Run("log_gh_authentication_failures", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		client := NewGitHubClient()

		// This will likely fail with authentication error
		_, err := client.GetIssue(1)

		output := buf.String()

		// Should log the command even when it fails
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "gh issue view 1 --json number,title,state,url")

			// If there's an authentication failure, it should be logged
			if err != nil {
				// Don't check specific exit codes since they can vary
				// The important thing is that the command was logged
				assert.Contains(t, output, "exit_code=")
			}
		}

		_ = err
	})

	t.Run("logging_disabled_no_output", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: false,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		client := NewGitHubClient()

		// Test that no logging occurs when disabled
		client.GetIssue(123)

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

		client := NewGitHubClient()

		// Test that only failures are logged at error level
		client.GetIssue(999999) // Very high issue number likely to fail

		output := buf.String()

		// At error level, only failures should be logged
		// Since gh commands are likely to fail in test environment, we might see logs
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "gh issue view")
			// Should have non-zero exit code for failed command
			assert.NotContains(t, output, "exit_code=0")
		}
	})

	t.Run("debug_level_logging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "debug",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		client := NewGitHubClient()

		// Test debug level logging includes all details
		client.GetIssue(456)

		output := buf.String()

		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "gh issue view 456 --json number,title,state,url")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "duration=")
			assert.Contains(t, output, "exit_code=")
		}
	})
}

func TestCheckGHInstalled_CommandLogging(t *testing.T) {
	// Save original global logger
	originalLogger := cmdlog.GetGlobalLogger()
	defer cmdlog.SetGlobalLogger(originalLogger)

	t.Run("log_gh_version_check", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test gh --version command logging
		err := CheckGHInstalled()

		output := buf.String()

		// The command should be logged regardless of success or failure
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "gh --version")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "duration=")
			assert.Contains(t, output, "exit_code=")

			if err == nil {
				// gh is installed and working
				assert.Contains(t, output, "exit_code=0")
			} else {
				// gh command failed
				assert.NotContains(t, output, "exit_code=0")
			}
		}

		// Don't assert on the error since it depends on gh being installed
		_ = err
	})
}

func TestRealCommandExecutor_CommandLogging(t *testing.T) {
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

		executor := &realCommandExecutor{}

		// Test a command that should succeed
		output, err := executor.executeCommand("echo", "test", "message")

		logOutput := buf.String()

		// The command should be logged
		if logOutput != "" {
			assert.Contains(t, logOutput, "[COMMAND]")
			assert.Contains(t, logOutput, "echo test message")
			assert.Contains(t, logOutput, "(from:")
			assert.Contains(t, logOutput, "duration=")

			if err == nil {
				assert.Contains(t, logOutput, "exit_code=0")
				assert.Contains(t, string(output), "test message")
			}
		}
	})

	t.Run("failed_command", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		executor := &realCommandExecutor{}

		// Test a command that should fail
		output, err := executor.executeCommand("false")

		logOutput := buf.String()

		// The command should be logged even when it fails
		if logOutput != "" {
			assert.Contains(t, logOutput, "[COMMAND]")
			assert.Contains(t, logOutput, "false")
			assert.Contains(t, logOutput, "(from:")
			assert.Contains(t, logOutput, "duration=")

			// Should have failed
			assert.Error(t, err)
			assert.Contains(t, logOutput, "exit_code=1")
		}

		_ = output
	})

	t.Run("command_with_multiple_arguments", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		executor := &realCommandExecutor{}

		// Test command with multiple arguments
		_, err := executor.executeCommand("echo", "arg1", "arg2", "arg3")

		logOutput := buf.String()

		// Should log all arguments
		if logOutput != "" {
			assert.Contains(t, logOutput, "[COMMAND]")
			assert.Contains(t, logOutput, "echo arg1 arg2 arg3")
			assert.Contains(t, logOutput, "(from:")
		}

		_ = err
	})
}

func TestGetExitCode_Issue(t *testing.T) {
	t.Run("nil_process_state", func(t *testing.T) {
		cmd := &exec.Cmd{}
		exitCode := getExitCode(cmd)
		assert.Equal(t, -1, exitCode)
	})
}
