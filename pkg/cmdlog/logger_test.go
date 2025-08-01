package cmdlog

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandLogger_BasicLogging(t *testing.T) {
	t.Run("log_command_execution", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewCommandLogger(Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})

		ctx := logger.LogCommand("test-command", []string{"arg1", "arg2"}, "test_function:123")
		ctx.LogCompletion(true, 0, "", 100*time.Millisecond)

		output := buf.String()
		assert.Contains(t, output, "[COMMAND]")
		assert.Contains(t, output, "test-command arg1 arg2")
		assert.Contains(t, output, "(from: test_function:123)")
		assert.Contains(t, output, "100ms")
	})

	t.Run("log_command_with_exit_code", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewCommandLogger(Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})

		ctx := logger.LogCommand("failing-command", []string{}, "test_function:456")
		ctx.LogCompletion(false, 1, "command failed", 50*time.Millisecond)

		output := buf.String()
		assert.Contains(t, output, "[COMMAND]")
		assert.Contains(t, output, "failing-command")
		assert.Contains(t, output, "exit_code=1")
		assert.Contains(t, output, "error=\"command failed\"")
	})

	t.Run("log_command_execution_time", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewCommandLogger(Config{
			Enabled: true,
			Level:   "debug",
			Output:  &buf,
		})

		ctx := logger.LogCommand("time-test", []string{}, "timer_test:789")
		time.Sleep(10 * time.Millisecond) // Simulate some execution time
		ctx.LogCompletion(true, 0, "", 25*time.Millisecond)

		output := buf.String()
		assert.Contains(t, output, "25ms")
	})
}

func TestCommandLogger_LogLevels(t *testing.T) {
	t.Run("debug_level_logging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewCommandLogger(Config{
			Enabled: true,
			Level:   "debug",
			Output:  &buf,
		})

		ctx := logger.LogCommand("debug-cmd", []string{"verbose", "args"}, "debug_test:100")
		ctx.LogCompletion(true, 0, "", 10*time.Millisecond)

		output := buf.String()
		assert.Contains(t, output, "[COMMAND]")
		assert.Contains(t, output, "debug-cmd verbose args")
		assert.Contains(t, output, "(from: debug_test:100)")
	})

	t.Run("info_level_logging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewCommandLogger(Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})

		ctx := logger.LogCommand("info-cmd", []string{"basic"}, "info_test:200")
		ctx.LogCompletion(true, 0, "", 15*time.Millisecond)

		output := buf.String()
		assert.Contains(t, output, "[COMMAND]")
		assert.Contains(t, output, "info-cmd basic")
	})

	t.Run("error_level_logging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewCommandLogger(Config{
			Enabled: true,
			Level:   "error",
			Output:  &buf,
		})

		// Success should not be logged at error level
		ctx1 := logger.LogCommand("success-cmd", []string{}, "error_test:300")
		ctx1.LogCompletion(true, 0, "", 5*time.Millisecond)

		// Failure should be logged at error level
		ctx2 := logger.LogCommand("fail-cmd", []string{}, "error_test:301")
		ctx2.LogCompletion(false, 1, "failed", 5*time.Millisecond)

		output := buf.String()
		assert.NotContains(t, output, "success-cmd")
		assert.Contains(t, output, "fail-cmd")
		assert.Contains(t, output, "exit_code=1")
	})

	t.Run("disabled_logging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewCommandLogger(Config{
			Enabled: false,
			Level:   "debug",
			Output:  &buf,
		})

		ctx := logger.LogCommand("disabled-cmd", []string{}, "disabled_test:400")
		ctx.LogCompletion(true, 0, "", 5*time.Millisecond)

		output := buf.String()
		assert.Empty(t, output)
	})
}

func TestCommandLogger_Configuration(t *testing.T) {
	t.Run("enable_disable_logging", func(t *testing.T) {
		var buf bytes.Buffer
		config := Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		}
		logger := NewCommandLogger(config)

		// Log when enabled
		ctx1 := logger.LogCommand("enabled-cmd", []string{}, "config_test:500")
		ctx1.LogCompletion(true, 0, "", 5*time.Millisecond)

		// Disable logging
		config.Enabled = false
		logger = NewCommandLogger(config)
		ctx2 := logger.LogCommand("disabled-cmd", []string{}, "config_test:501")
		ctx2.LogCompletion(true, 0, "", 5*time.Millisecond)

		output := buf.String()
		assert.Contains(t, output, "enabled-cmd")
		assert.NotContains(t, output, "disabled-cmd")
	})

	t.Run("concurrent_logging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewCommandLogger(Config{
			Enabled: true,
			Level:   "info",
			Output:  &buf,
		})

		// Start multiple concurrent logging operations
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(id int) {
				ctx := logger.LogCommand(fmt.Sprintf("concurrent-cmd-%d", id), []string{}, fmt.Sprintf("concurrent_test:%d", id))
				ctx.LogCompletion(true, 0, "", 5*time.Millisecond)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		output := buf.String()
		// Verify all commands were logged
		for i := 0; i < 10; i++ {
			assert.Contains(t, output, fmt.Sprintf("concurrent-cmd-%d", i))
		}
	})
}

func TestCommandLogger_ErrorHandling(t *testing.T) {
	t.Run("log_file_creation_failure", func(t *testing.T) {
		// Try to create logger with invalid file path
		invalidPath := "/invalid/path/that/does/not/exist/log.txt"
		config := Config{
			Enabled:  true,
			Level:    "info",
			FilePath: invalidPath,
		}

		// Should not panic and should gracefully fall back
		logger := NewCommandLogger(config)
		require.NotNil(t, logger)

		// Should still work (probably falling back to stderr or no-op)
		ctx := logger.LogCommand("test-cmd", []string{}, "error_test:600")
		ctx.LogCompletion(true, 0, "", 5*time.Millisecond)
	})

	t.Run("log_write_failure", func(t *testing.T) {
		// Create a closed writer to simulate write failure
		var buf bytes.Buffer
		closedWriter := &failingWriter{buf: &buf, failAfter: 1}

		logger := NewCommandLogger(Config{
			Enabled: true,
			Level:   "info",
			Output:  closedWriter,
		})

		// Should not panic even if write fails
		ctx := logger.LogCommand("write-fail-cmd", []string{}, "error_test:700")
		ctx.LogCompletion(true, 0, "", 5*time.Millisecond)
	})
}

func TestCommandLogger_Integration_ExecCommand(t *testing.T) {
	t.Run("real_command_execution_logging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewCommandLogger(Config{
			Enabled: true,
			Level:   "debug",
			Output:  &buf,
		})

		// Execute a real command with logging
		cmd := exec.Command("echo", "hello", "world")

		ctx := logger.LogCommand("echo", []string{"hello", "world"}, "integration_test:800")
		start := time.Now()
		err := cmd.Run()
		duration := time.Since(start)

		ctx.LogCompletion(err == nil, cmd.ProcessState.ExitCode(), "", duration)

		output := buf.String()
		assert.Contains(t, output, "[COMMAND]")
		assert.Contains(t, output, "echo hello world")
		assert.Contains(t, output, "(from: integration_test:800)")
		assert.Contains(t, output, "exit_code=0")
	})

	t.Run("failed_command_execution_logging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewCommandLogger(Config{
			Enabled: true,
			Level:   "error",
			Output:  &buf,
		})

		// Execute a command that will fail
		cmd := exec.Command("false") // 'false' command always exits with code 1

		ctx := logger.LogCommand("false", []string{}, "integration_test:801")
		start := time.Now()
		err := cmd.Run()
		duration := time.Since(start)

		var errorMsg string
		if err != nil {
			errorMsg = err.Error()
		}

		ctx.LogCompletion(err == nil, cmd.ProcessState.ExitCode(), errorMsg, duration)

		output := buf.String()
		assert.Contains(t, output, "[COMMAND]")
		assert.Contains(t, output, "false")
		assert.Contains(t, output, "exit_code=1")
	})
}

func TestCommandLogger_FileLogging(t *testing.T) {
	t.Run("log_to_file", func(t *testing.T) {
		// Create temporary file for testing
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "test.log")

		logger := NewCommandLogger(Config{
			Enabled:  true,
			Level:    "info",
			FilePath: logFile,
		})

		// Log a command
		ctx := logger.LogCommand("file-test-cmd", []string{"arg1"}, "file_test:900")
		ctx.LogCompletion(true, 0, "", 10*time.Millisecond)

		// Read the log file and verify content
		content, err := os.ReadFile(logFile)
		require.NoError(t, err)

		output := string(content)
		assert.Contains(t, output, "[COMMAND]")
		assert.Contains(t, output, "file-test-cmd arg1")
		assert.Contains(t, output, "(from: file_test:900)")
	})
}

// Helper types for testing
type failingWriter struct {
	buf       *bytes.Buffer
	failAfter int
	count     int
}

func (fw *failingWriter) Write(p []byte) (n int, err error) {
	fw.count++
	if fw.count > fw.failAfter {
		return 0, fmt.Errorf("simulated write failure")
	}
	return fw.buf.Write(p)
}

// Test utilities for creating test scenarios
func TestCreateTestCommandScenario(t *testing.T) {
	scenario := CreateTestCommandScenario("test_scenario", "git", []string{"status"}, 0, "clean working tree")

	assert.Equal(t, "test_scenario", scenario.Name)
	assert.Equal(t, "git", scenario.Command)
	assert.Equal(t, []string{"status"}, scenario.Args)
	assert.Equal(t, 0, scenario.ExitCode)
	assert.Equal(t, "clean working tree", scenario.Output)
}
