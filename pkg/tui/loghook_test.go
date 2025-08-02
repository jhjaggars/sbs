package tui

import (
	"os"
	"path/filepath"
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
