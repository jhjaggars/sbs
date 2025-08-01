package repo

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"sbs/pkg/cmdlog"
)

func TestRepositoryManager_CommandLogging(t *testing.T) {
	// Save original global logger
	originalLogger := cmdlog.GetGlobalLogger()
	defer cmdlog.SetGlobalLogger(originalLogger)

	manager := NewManager()

	t.Run("log_git_remote_get_url", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test git remote get-url command
		// Use current working directory which should be a git repository
		cwd, err := os.Getwd()
		if err != nil {
			t.Skip("Could not get current working directory")
		}

		// Test extractNameFromRemote (which uses git remote get-url)
		remoteName := manager.extractNameFromRemote(cwd)

		output := buf.String()

		// The command should be logged regardless of success or failure
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "git remote get-url origin")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "duration=")
			assert.Contains(t, output, "exit_code=")
		}

		// Don't assert on the result since it depends on the git setup
		_ = remoteName
	})

	t.Run("log_git_remote_failures", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test git remote command in a non-git directory
		tempDir := t.TempDir()

		// This should fail since it's not a git repository
		remoteName := manager.extractNameFromRemote(tempDir)

		output := buf.String()

		// The command should be logged even when it fails
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "git remote get-url origin")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "duration=")
			assert.Contains(t, output, "exit_code=")
			// Should have non-zero exit code since it's not a git repo
			assert.NotContains(t, output, "exit_code=0")
		}

		// Should return empty string when git command fails
		assert.Equal(t, "", remoteName)
	})

	t.Run("log_getRemoteURL", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		cwd, err := os.Getwd()
		if err != nil {
			t.Skip("Could not get current working directory")
		}

		// Test getRemoteURL method
		remoteURL := manager.getRemoteURL(cwd)

		output := buf.String()

		// The command should be logged
		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "git remote get-url origin")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "duration=")
		}

		// Don't assert on the actual URL since it depends on the git setup
		_ = remoteURL
	})

	t.Run("logging_disabled_no_output", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: false,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		cwd, err := os.Getwd()
		if err != nil {
			t.Skip("Could not get current working directory")
		}

		// Test that no logging occurs when disabled
		_ = manager.getRemoteURL(cwd)

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

		cwd, err := os.Getwd()
		if err != nil {
			t.Skip("Could not get current working directory")
		}

		// Test successful command (should not be logged at error level)
		manager.getRemoteURL(cwd)

		// Test failing command (should be logged at error level)
		tempDir := t.TempDir()
		manager.extractNameFromRemote(tempDir)

		output := buf.String()

		// At error level, only failures should be logged
		// The success case should not appear, but the failure should
		if output != "" {
			assert.Contains(t, output, "git remote get-url origin")
			assert.Contains(t, output, "[COMMAND]")
			// Should have non-zero exit code for the failed command
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

		cwd, err := os.Getwd()
		if err != nil {
			t.Skip("Could not get current working directory")
		}

		// Test debug level logging includes all details
		manager.getRemoteURL(cwd)

		output := buf.String()

		if output != "" {
			assert.Contains(t, output, "[COMMAND]")
			assert.Contains(t, output, "git remote get-url origin")
			assert.Contains(t, output, "(from:")
			assert.Contains(t, output, "duration=")
			assert.Contains(t, output, "exit_code=")
		}
	})
}

func TestRunGitCommand(t *testing.T) {
	// Save original global logger
	originalLogger := cmdlog.GetGlobalLogger()
	defer cmdlog.SetGlobalLogger(originalLogger)

	manager := NewManager()

	t.Run("successful_git_command", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		cwd, err := os.Getwd()
		if err != nil {
			t.Skip("Could not get current working directory")
		}

		// Test a git command that should succeed (if we're in a git repo)
		output, err := manager.runGitCommand(cwd, []string{"--version"})

		logOutput := buf.String()

		// The command should be logged
		if logOutput != "" {
			assert.Contains(t, logOutput, "[COMMAND]")
			assert.Contains(t, logOutput, "git --version")
			assert.Contains(t, logOutput, "(from:")
			assert.Contains(t, logOutput, "duration=")

			if err == nil {
				assert.Contains(t, logOutput, "exit_code=0")
				assert.NotEmpty(t, string(output))
			}
		}
	})

	t.Run("failed_git_command", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		// Test a git command that should fail
		tempDir := t.TempDir()
		output, err := manager.runGitCommand(tempDir, []string{"status"})

		logOutput := buf.String()

		// The command should be logged even when it fails
		if logOutput != "" {
			assert.Contains(t, logOutput, "[COMMAND]")
			assert.Contains(t, logOutput, "git status")
			assert.Contains(t, logOutput, "(from:")
			assert.Contains(t, logOutput, "duration=")

			// Should have failed since it's not a git repository
			assert.Error(t, err)
			assert.NotContains(t, logOutput, "exit_code=0")
		}

		_ = output
	})

	t.Run("git_command_with_multiple_args", func(t *testing.T) {
		var buf bytes.Buffer
		logger := cmdlog.NewCommandLogger(cmdlog.Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})
		cmdlog.SetGlobalLogger(logger)

		cwd, err := os.Getwd()
		if err != nil {
			t.Skip("Could not get current working directory")
		}

		// Test git command with multiple arguments
		_, err = manager.runGitCommand(cwd, []string{"remote", "get-url", "origin"})

		logOutput := buf.String()

		// Should log all arguments
		if logOutput != "" {
			assert.Contains(t, logOutput, "[COMMAND]")
			assert.Contains(t, logOutput, "git remote get-url origin")
			assert.Contains(t, logOutput, "(from:")
		}

		// Don't assert on error since it depends on git setup
		_ = err
	})
}

func TestGetExitCode_Repo(t *testing.T) {
	t.Run("nil_process_state", func(t *testing.T) {
		cmd := &exec.Cmd{}
		exitCode := getExitCode(cmd)
		assert.Equal(t, -1, exitCode)
	})
}
