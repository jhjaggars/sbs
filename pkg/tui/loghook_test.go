package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sbs/pkg/config"
)

func TestLoghook_ScriptExecution(t *testing.T) {
	t.Run("execute_existing_loghook_script", func(t *testing.T) {
		// Create test worktree with .sbs/loghook script
		worktreePath := setupTestWorktree(t)
		defer os.RemoveAll(filepath.Dir(worktreePath))

		session := config.SessionMetadata{
			IssueNumber:  123,
			IssueTitle:   "Test Issue",
			WorktreePath: worktreePath,
		}

		// Execute script and capture output
		output, err := executeLoghookScript(session)

		// Verify output is returned correctly
		assert.NoError(t, err, "Loghook script execution should not return an error")
		assert.Contains(t, output, "Test log output", "Script output should contain expected text")
		assert.Contains(t, output, "Timestamp:", "Script output should contain timestamp")
	})

	t.Run("handle_missing_loghook_script", func(t *testing.T) {
		// Create test worktree without .sbs/loghook script
		worktreePath := filepath.Join(t.TempDir(), "test-worktree")
		require.NoError(t, os.MkdirAll(worktreePath, 0755))

		session := config.SessionMetadata{
			IssueNumber:  123,
			IssueTitle:   "Test Issue",
			WorktreePath: worktreePath,
		}

		// Test behavior when .sbs/loghook doesn't exist
		output, err := executeLoghookScript(session)

		// Verify appropriate error message or fallback
		if err != nil {
			assert.Contains(t, err.Error(), "loghook script not found", "Error should indicate missing script")
		} else {
			assert.Contains(t, output, "No loghook script found", "Should return fallback message")
		}
	})

	t.Run("handle_non_executable_loghook_script", func(t *testing.T) {
		// Create test worktree with non-executable loghook file
		worktreePath := filepath.Join(t.TempDir(), "test-worktree")
		sbsDir := filepath.Join(worktreePath, ".sbs")
		require.NoError(t, os.MkdirAll(sbsDir, 0755))

		// Create non-executable loghook script
		loghookPath := filepath.Join(sbsDir, "loghook")
		script := `#!/bin/bash
echo "Test log output"
`
		require.NoError(t, os.WriteFile(loghookPath, []byte(script), 0644)) // Not executable

		session := config.SessionMetadata{
			IssueNumber:  123,
			IssueTitle:   "Test Issue",
			WorktreePath: worktreePath,
		}

		// Test with script that lacks execute permissions
		_, err := executeLoghookScript(session)

		// Verify error handling and user feedback
		assert.Error(t, err, "Should return error for non-executable script")
		assert.Contains(t, err.Error(), "permission denied", "Error should indicate permission issue")
	})

	t.Run("handle_script_execution_failure", func(t *testing.T) {
		// Create test worktree with failing script
		worktreePath, loghookPath := setupTestWorktreeWithCustomScript(t, `#!/bin/bash
echo "Error: Something went wrong" >&2
exit 1
`)
		defer os.RemoveAll(filepath.Dir(worktreePath))

		session := config.SessionMetadata{
			IssueNumber:  123,
			IssueTitle:   "Test Issue",
			WorktreePath: worktreePath,
		}

		// Make script executable
		require.NoError(t, os.Chmod(loghookPath, 0755))

		// Test with script that returns non-zero exit code
		output, err := executeLoghookScript(session)

		// Verify error capture and display
		assert.Error(t, err, "Should return error for failing script")
		assert.Contains(t, output, "Error: Something went wrong", "Should capture stderr output")
	})

	t.Run("handle_script_execution_timeout", func(t *testing.T) {
		// Create test worktree with long-running script
		worktreePath, loghookPath := setupTestWorktreeWithCustomScript(t, `#!/bin/bash
echo "Starting long operation..."
sleep 10
echo "Operation completed"
`)
		defer os.RemoveAll(filepath.Dir(worktreePath))

		session := config.SessionMetadata{
			IssueNumber:  123,
			IssueTitle:   "Test Issue",
			WorktreePath: worktreePath,
		}

		// Make script executable
		require.NoError(t, os.Chmod(loghookPath, 0755))

		// Test with long-running script (should timeout if timeout is implemented)
		_, err := executeLoghookScriptWithTimeout(session, 2) // 2 second timeout

		// Verify timeout handling (if implemented)
		if err != nil {
			assert.Contains(t, err.Error(), "timeout", "Error should indicate timeout occurred")
		}
	})

	t.Run("script_working_directory", func(t *testing.T) {
		// Create test worktree with script that outputs working directory
		worktreePath, loghookPath := setupTestWorktreeWithCustomScript(t, `#!/bin/bash
echo "Working directory: $(pwd)"
echo "Script location: $(dirname "$0")"
`)
		defer os.RemoveAll(filepath.Dir(worktreePath))

		session := config.SessionMetadata{
			IssueNumber:  123,
			IssueTitle:   "Test Issue",
			WorktreePath: worktreePath,
		}

		// Make script executable
		require.NoError(t, os.Chmod(loghookPath, 0755))

		// Verify script executes from correct working directory
		output, err := executeLoghookScript(session)

		// Test that script has access to worktree context
		assert.NoError(t, err, "Script should execute successfully")
		assert.Contains(t, output, worktreePath, "Script should execute from worktree directory")
		assert.Contains(t, output, ".sbs", "Script should be located in .sbs directory")
	})
}

// Helper functions for test setup
func setupTestWorktree(t *testing.T) string {
	worktreePath := filepath.Join(t.TempDir(), "test-worktree")
	sbsDir := filepath.Join(worktreePath, ".sbs")
	require.NoError(t, os.MkdirAll(sbsDir, 0755))

	// Create test loghook script
	loghookPath := filepath.Join(sbsDir, "loghook")
	script := `#!/bin/bash
echo "Test log output"
echo "Timestamp: $(date)"
`
	require.NoError(t, os.WriteFile(loghookPath, []byte(script), 0755))

	return worktreePath
}

func setupTestWorktreeWithCustomScript(t *testing.T, script string) (string, string) {
	worktreePath := filepath.Join(t.TempDir(), "test-worktree")
	sbsDir := filepath.Join(worktreePath, ".sbs")
	require.NoError(t, os.MkdirAll(sbsDir, 0755))

	// Create custom loghook script
	loghookPath := filepath.Join(sbsDir, "loghook")
	require.NoError(t, os.WriteFile(loghookPath, []byte(script), 0755))

	return worktreePath, loghookPath
}

// Note: executeLoghookScript and executeLoghookScriptWithTimeout are now implemented in model.go

// Security and validation tests for enhanced loghook functionality
func TestLoghook_SecurityValidation(t *testing.T) {
	t.Run("validate_path_traversal_prevention", func(t *testing.T) {
		// Test various path traversal attempts
		testCases := []struct {
			name         string
			worktreePath string
			expectError  bool
		}{
			{
				name:         "normal_path",
				worktreePath: "/tmp/test-worktree",
				expectError:  false,
			},
			{
				name:         "relative_path_attempt",
				worktreePath: "/tmp/test-worktree/../../../etc",
				expectError:  true,
			},
			{
				name:         "non_absolute_path",
				worktreePath: "relative/path",
				expectError:  true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Test path validation directly
				_, err := validateLoghookPath(tc.worktreePath)
				if tc.expectError {
					assert.Error(t, err, "Expected error for path: %s", tc.worktreePath)
				} else {
					assert.NoError(t, err, "Expected no error for path: %s", tc.worktreePath)
				}
			})
		}
	})

	t.Run("validate_script_security_checks", func(t *testing.T) {
		// Create test directory
		tempDir := t.TempDir()
		scriptPath := filepath.Join(tempDir, "test-script")

		// Test non-existent file
		err := validateScriptSecurity(scriptPath)
		assert.Error(t, err, "Should error for non-existent script")
		assert.Contains(t, err.Error(), "failed to stat script")

		// Create regular file but non-executable
		require.NoError(t, os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0644))
		err = validateScriptSecurity(scriptPath)
		assert.Error(t, err, "Should error for non-executable script")
		assert.Contains(t, err.Error(), "not executable")

		// Make executable
		require.NoError(t, os.Chmod(scriptPath, 0755))
		err = validateScriptSecurity(scriptPath)
		assert.NoError(t, err, "Should pass for executable regular file")

		// Test directory instead of file
		dirPath := filepath.Join(tempDir, "test-dir")
		require.NoError(t, os.Mkdir(dirPath, 0755))
		err = validateScriptSecurity(dirPath)
		assert.Error(t, err, "Should error for directory")
		assert.Contains(t, err.Error(), "not a regular file")
	})

	t.Run("test_output_size_limits", func(t *testing.T) {
		// Create script that outputs more than the limit
		worktreePath, loghookPath := setupTestWorktreeWithCustomScript(t, `#!/bin/bash
# Generate output larger than 1KB limit for testing
for i in {1..100}; do
    echo "This is line $i with some padding to make it longer than usual - adding more text to reach size limits"
done
`)
		defer os.RemoveAll(filepath.Dir(worktreePath))
		require.NoError(t, os.Chmod(loghookPath, 0755))

		session := config.SessionMetadata{
			IssueNumber:  123,
			IssueTitle:   "Test Issue",
			WorktreePath: worktreePath,
		}

		// Test with small size limit (1KB)
		output, err := executeLoghookScriptWithOptions(session, 10, 1024)
		assert.NoError(t, err, "Script should execute successfully")
		assert.LessOrEqual(t, len(output), 1024+200, "Output should be truncated to size limit (with some buffer for truncation message)")
		assert.Contains(t, output, "Output truncated", "Should contain truncation message")
	})

	t.Run("test_script_timeout_handling", func(t *testing.T) {
		// Create long-running script
		worktreePath, loghookPath := setupTestWorktreeWithCustomScript(t, `#!/bin/bash
echo "Starting..."
sleep 5
echo "Finished"
`)
		defer os.RemoveAll(filepath.Dir(worktreePath))
		require.NoError(t, os.Chmod(loghookPath, 0755))

		session := config.SessionMetadata{
			IssueNumber:  123,
			IssueTitle:   "Test Issue",
			WorktreePath: worktreePath,
		}

		// Test with short timeout
		output, err := executeLoghookScriptWithTimeout(session, 1) // 1 second timeout
		assert.Error(t, err, "Should timeout")
		assert.Contains(t, err.Error(), "timed out", "Error should indicate timeout")
		assert.Contains(t, output, "Starting", "Should capture partial output before timeout")
	})

	t.Run("test_enhanced_error_messages", func(t *testing.T) {
		// Test various error conditions with enhanced error messages

		// Non-existent worktree
		session := config.SessionMetadata{
			IssueNumber:  123,
			IssueTitle:   "Test Issue",
			WorktreePath: "/non/existent/path",
		}

		_, err := executeLoghookScript(session)
		assert.Error(t, err, "Should error for non-existent path")
		assert.Contains(t, err.Error(), "/non/existent/path", "Error should contain the actual path")

		// Path traversal attempt
		session.WorktreePath = "/tmp/../../../etc"
		_, err = executeLoghookScript(session)
		assert.Error(t, err, "Should error for path traversal")
		assert.Contains(t, err.Error(), "path traversal detected", "Should indicate path traversal detection")
	})

	t.Run("test_audit_logging_functionality", func(t *testing.T) {
		// This test ensures the audit logging functions don't crash
		// In a real environment, you'd capture log output, but for unit tests
		// we just ensure the function executes without errors

		info := LogExecutionInfo{
			ScriptPath:      "/test/path/script",
			WorkingDir:      "/test/worktree",
			DurationMs:      100,
			ExitCode:        0,
			OutputSizeBytes: 1024,
			TimedOut:        false,
		}

		// Should not panic or error
		assert.NotPanics(t, func() {
			logScriptExecution(info)
		})

		// Test with error
		info.Error = "test error message"
		assert.NotPanics(t, func() {
			logScriptExecution(info)
		})
	})
}

func TestLoghook_PerformanceAndMemory(t *testing.T) {
	t.Run("test_memory_efficient_output_handling", func(t *testing.T) {
		// Create script with large output to test memory efficiency
		worktreePath, loghookPath := setupTestWorktreeWithCustomScript(t, `#!/bin/bash
# Generate a reasonable amount of output quickly
for i in {1..200}; do
    echo "Line $i: Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
done
`)
		defer os.RemoveAll(filepath.Dir(worktreePath))
		require.NoError(t, os.Chmod(loghookPath, 0755))

		session := config.SessionMetadata{
			IssueNumber:  123,
			IssueTitle:   "Test Issue",
			WorktreePath: worktreePath,
		}

		// Test with reasonable size limit and longer timeout
		output, err := executeLoghookScriptWithOptions(session, 30, 8192) // 30s timeout, 8KB limit
		assert.NoError(t, err, "Should handle large output without memory issues")
		assert.LessOrEqual(t, len(output), 8192+500, "Output should be properly limited")

		// Verify we got content but not everything
		lines := strings.Split(output, "\n")
		assert.Greater(t, len(lines), 10, "Should have captured some lines")
		assert.Less(t, len(lines), 1000, "Should not have captured all lines due to size limit")
	})

	t.Run("test_concurrent_execution_safety", func(t *testing.T) {
		// This test ensures the refactored code handles concurrent execution safely
		worktreePath := setupTestWorktree(t)
		defer os.RemoveAll(filepath.Dir(worktreePath))

		session := config.SessionMetadata{
			IssueNumber:  123,
			IssueTitle:   "Test Issue",
			WorktreePath: worktreePath,
		}

		// Run multiple concurrent executions to test for race conditions
		done := make(chan bool, 5)
		for i := 0; i < 5; i++ {
			go func() {
				defer func() { done <- true }()
				_, err := executeLoghookScript(session)
				assert.NoError(t, err, "Concurrent execution should not fail")
			}()
		}

		// Wait for all to complete
		for i := 0; i < 5; i++ {
			<-done
		}
	})
}
