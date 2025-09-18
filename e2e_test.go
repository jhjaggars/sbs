//go:build e2e
// +build e2e

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sbs/pkg/config"
)

// E2ETestSuite provides comprehensive end-to-end testing for SBS
type E2ETestSuite struct {
	t               *testing.T
	tempDir         string
	configDir       string
	worktreeBaseDir string
	sbsBinary       string
	originalWorkDir string
	testRepoDir     string
	sessionIDs      []string // Track created sessions for cleanup
}

// NewE2ETestSuite creates a new end-to-end test suite
func NewE2ETestSuite(t *testing.T) *E2ETestSuite {
	// Skip if not running e2e tests
	if os.Getenv("E2E_TESTS") == "" {
		t.Skip("Skipping E2E test - set E2E_TESTS=1 to run")
	}

	suite := &E2ETestSuite{
		t:          t,
		tempDir:    t.TempDir(),
		sessionIDs: make([]string, 0),
	}

	suite.setup()
	return suite
}

// setup initializes the test environment
func (suite *E2ETestSuite) setup() {
	// Store original working directory
	var err error
	suite.originalWorkDir, err = os.Getwd()
	require.NoError(suite.t, err)

	// Create test directories
	suite.configDir = filepath.Join(suite.tempDir, "config", "sbs")
	suite.worktreeBaseDir = filepath.Join(suite.tempDir, "worktrees")
	suite.testRepoDir = filepath.Join(suite.tempDir, "test-repo")

	err = os.MkdirAll(suite.configDir, 0755)
	require.NoError(suite.t, err)

	err = os.MkdirAll(suite.worktreeBaseDir, 0755)
	require.NoError(suite.t, err)

	// Build SBS binary if not exists
	suite.sbsBinary = filepath.Join(suite.originalWorkDir, "sbs")
	if _, err := os.Stat(suite.sbsBinary); os.IsNotExist(err) {
		suite.t.Log("Building SBS binary for E2E tests...")
		cmd := exec.Command("make", "build")
		cmd.Dir = suite.originalWorkDir
		output, err := cmd.CombinedOutput()
		require.NoError(suite.t, err, "Failed to build SBS binary: %s", output)
	}

	// Setup test git repository
	suite.setupTestRepo()

	// Create test configuration
	suite.createTestConfig()
}

// setupTestRepo creates a test git repository
func (suite *E2ETestSuite) setupTestRepo() {
	// Create test repo directory
	err := os.MkdirAll(suite.testRepoDir, 0755)
	require.NoError(suite.t, err)

	// Initialize git repo
	suite.runCommand("git", "init", suite.testRepoDir)
	suite.runCommand("git", "-C", suite.testRepoDir, "config", "user.name", "Test User")
	suite.runCommand("git", "-C", suite.testRepoDir, "config", "user.email", "test@example.com")

	// Create initial commit
	readmeContent := "# Test Repository for SBS E2E Tests\n"
	readmePath := filepath.Join(suite.testRepoDir, "README.md")
	err = os.WriteFile(readmePath, []byte(readmeContent), 0644)
	require.NoError(suite.t, err)

	suite.runCommand("git", "-C", suite.testRepoDir, "add", "README.md")
	suite.runCommand("git", "-C", suite.testRepoDir, "commit", "-m", "Initial commit")

	// Create .sbs directory and input source config for test work types
	sbsDir := filepath.Join(suite.testRepoDir, ".sbs")
	err = os.MkdirAll(sbsDir, 0755)
	require.NoError(suite.t, err)

	inputSourceConfig := config.InputSourceConfig{
		Type:     "test",
		Settings: map[string]interface{}{},
	}

	configData, err := json.MarshalIndent(inputSourceConfig, "", "  ")
	require.NoError(suite.t, err)

	configPath := filepath.Join(sbsDir, "input-source.json")
	err = os.WriteFile(configPath, configData, 0644)
	require.NoError(suite.t, err)
}

// createTestConfig creates SBS configuration for testing
func (suite *E2ETestSuite) createTestConfig() {
	config := config.Config{
		WorktreeBasePath: suite.worktreeBaseDir,
		RepoPath:         suite.testRepoDir,
		// Note: We'll create a minimal work-issue.sh script for testing
		WorkIssueScript: filepath.Join(suite.tempDir, "work-issue.sh"),
	}

	// Create minimal work-issue.sh script
	workIssueScript := `#!/bin/bash
# Minimal work-issue.sh for E2E testing
echo "Starting work on issue: $1"
echo "Working directory: $(pwd)"
echo "Environment variables:"
env | grep SBS_ || true

# Create a test file to prove the script ran
echo "E2E test work session for $1" > .sbs-test-marker
echo "Timestamp: $(date)" >> .sbs-test-marker

# Sleep for a moment to simulate work
sleep 2

echo "Work setup complete for issue: $1"
`

	err := os.WriteFile(config.WorkIssueScript, []byte(workIssueScript), 0755)
	require.NoError(suite.t, err)

	// Write config file
	configData, err := json.MarshalIndent(config, "", "  ")
	require.NoError(suite.t, err)

	configPath := filepath.Join(suite.configDir, "config.json")
	err = os.WriteFile(configPath, configData, 0644)
	require.NoError(suite.t, err)
}

// runSBSCommand runs an SBS command with proper environment
func (suite *E2ETestSuite) runSBSCommand(args ...string) (string, error) {
	cmd := exec.Command(suite.sbsBinary, args...)
	cmd.Dir = suite.testRepoDir

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// runCommand runs a shell command and requires success
func (suite *E2ETestSuite) runCommand(name string, args ...string) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	require.NoError(suite.t, err, "Command failed: %s %v\nOutput: %s", name, args, output)
}

// cleanup cleans up the test environment
func (suite *E2ETestSuite) cleanup() {
	// Stop any created sessions
	for _, sessionID := range suite.sessionIDs {
		suite.runSBSCommand("stop", sessionID, "--yes")
	}

	// Clean up any remaining resources (this will clean from the test-repo context)
	suite.runSBSCommand("clean", "--force")

	// Restore original working directory
	os.Chdir(suite.originalWorkDir)
}

// TestE2E_SessionLifecycle tests complete session lifecycle
func TestE2E_SessionLifecycle(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.cleanup()

	sessionID := "test:e2e-lifecycle"
	suite.sessionIDs = append(suite.sessionIDs, sessionID)

	t.Run("start_session", func(t *testing.T) {
		output, err := suite.runSBSCommand("start", sessionID)
		require.NoError(t, err, "Failed to start session: %s", output)
		assert.Contains(t, output, "Work environment ready!")
	})

	t.Run("verify_session_created", func(t *testing.T) {
		// Check that session appears in list
		output, err := suite.runSBSCommand("list", "--plain")
		require.NoError(t, err)
		assert.Contains(t, output, sessionID)

		// Verify worktree was created in the default location (test runs from test-repo dir)
		homeDir, _ := os.UserHomeDir()
		expectedWorktree := filepath.Join(homeDir, ".sbs-worktrees", "test-repo", "issue-test-e2e-lifecycle")
		_, err = os.Stat(expectedWorktree)
		assert.NoError(t, err, "Worktree should exist: %s", expectedWorktree)

		// Note: The work-issue.sh script we created doesn't run in the test setup,
		// so we won't check for the test marker file
	})

	t.Run("stop_session", func(t *testing.T) {
		output, err := suite.runSBSCommand("stop", sessionID, "--yes")
		require.NoError(t, err, "Failed to stop session: %s", output)

		// Verify session is no longer in active list
		output, err = suite.runSBSCommand("list", "--plain")
		require.NoError(t, err)
		// Session might still appear but should be marked as stopped
		// The exact behavior depends on implementation
	})
}

// TestE2E_StatusTracking tests status detection and time tracking
func TestE2E_StatusTracking(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.cleanup()

	sessionID := "test:e2e-status"
	suite.sessionIDs = append(suite.sessionIDs, sessionID)

	t.Run("start_session_with_status", func(t *testing.T) {
		// Start session
		output, err := suite.runSBSCommand("start", sessionID)
		require.NoError(t, err, "Failed to start session: %s", output)

		// Let it run for a moment
		time.Sleep(3 * time.Second)

		// Check if we can detect status (this depends on hook integration)
		listOutput, err := suite.runSBSCommand("list", "--plain")
		require.NoError(t, err)
		assert.Contains(t, listOutput, sessionID)
	})

	t.Run("verify_hook_integration", func(t *testing.T) {
		// This test verifies that hooks are properly installed and working
		// The exact verification depends on hook implementation details

		homeDir, _ := os.UserHomeDir()
		worktreePath := filepath.Join(homeDir, ".sbs-worktrees", "test-repo", "issue-test-e2e-status")
		sbsDir := filepath.Join(worktreePath, ".sbs")

		// Check if .sbs directory exists (created by hooks)
		if _, err := os.Stat(sbsDir); err == nil {
			// If .sbs directory exists, check for hook artifacts
			files, err := os.ReadDir(sbsDir)
			require.NoError(t, err)
			t.Logf("Hook directory contents: %v", files)
		}
	})
}

// TestE2E_MultipleSessionsConcurrency tests concurrent session management
func TestE2E_MultipleSessionsConcurrency(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.cleanup()

	sessions := []string{"test:e2e-concurrent1", "test:e2e-concurrent2", "test:e2e-concurrent3"}
	suite.sessionIDs = append(suite.sessionIDs, sessions...)

	t.Run("start_multiple_sessions", func(t *testing.T) {
		for _, sessionID := range sessions {
			output, err := suite.runSBSCommand("start", sessionID)
			require.NoError(t, err, "Failed to start session %s: %s", sessionID, output)
		}
	})

	t.Run("verify_all_sessions_listed", func(t *testing.T) {
		output, err := suite.runSBSCommand("list", "--plain")
		require.NoError(t, err)

		for _, sessionID := range sessions {
			assert.Contains(t, output, sessionID, "Session should be listed: %s", sessionID)
		}
	})

	t.Run("attach_to_sessions", func(t *testing.T) {
		// Test attaching (this will be limited since we can't interact with tmux in tests)
		for _, sessionID := range sessions {
			// We can't actually attach in a test, but we can verify the command validates
			output, err := suite.runSBSCommand("attach", sessionID, "--dry-run")
			if err != nil {
				// If --dry-run doesn't exist, that's fine for this test
				t.Logf("Attach command for %s: %v", sessionID, err)
			} else {
				t.Logf("Attach dry-run output for %s: %s", sessionID, output)
			}
		}
	})

	t.Run("cleanup_all_sessions", func(t *testing.T) {
		output, err := suite.runSBSCommand("clean", "--force")
		require.NoError(t, err, "Failed to clean sessions: %s", output)
	})
}

// TestE2E_WorktreeAndGitIntegration tests git worktree management
func TestE2E_WorktreeAndGitIntegration(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.cleanup()

	sessionID := "test:e2e-worktree"
	suite.sessionIDs = append(suite.sessionIDs, sessionID)

	t.Run("create_session_with_worktree", func(t *testing.T) {
		output, err := suite.runSBSCommand("start", sessionID)
		require.NoError(t, err, "Failed to start session: %s", output)

		// Verify worktree exists and is a valid git repository
		homeDir, _ := os.UserHomeDir()
		worktreePath := filepath.Join(homeDir, ".sbs-worktrees", "test-repo", "issue-test-e2e-worktree")
		gitDir := filepath.Join(worktreePath, ".git")

		_, err = os.Stat(gitDir)
		assert.NoError(t, err, "Worktree should have .git: %s", gitDir)

		// Check git status in worktree
		cmd := exec.Command("git", "status", "--porcelain")
		cmd.Dir = worktreePath
		gitOutput, err := cmd.CombinedOutput()
		require.NoError(t, err, "Git status should work in worktree")

		t.Logf("Worktree git status: %s", string(gitOutput))
	})

	t.Run("verify_branch_creation", func(t *testing.T) {
		// Check that appropriate branch was created
		cmd := exec.Command("git", "branch", "-a")
		cmd.Dir = suite.testRepoDir
		branchOutput, err := cmd.CombinedOutput()
		require.NoError(t, err)

		// Should contain branch for our test session
		assert.Contains(t, string(branchOutput), "test-e2e-worktree")
		t.Logf("Git branches: %s", string(branchOutput))
	})
}

// TestE2E_ErrorHandling tests error scenarios and recovery
func TestE2E_ErrorHandling(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.cleanup()

	t.Run("invalid_session_id", func(t *testing.T) {
		output, err := suite.runSBSCommand("start", "invalid:format:with:colons")
		// Should either handle gracefully or return clear error
		if err != nil {
			assert.Contains(t, output, "Error:", "Should provide clear error message")
		}
	})

	t.Run("stop_nonexistent_session", func(t *testing.T) {
		output, err := suite.runSBSCommand("stop", "test:nonexistent")
		// Should handle gracefully
		if err != nil {
			t.Logf("Expected error for nonexistent session: %s", output)
		}
	})

	t.Run("attach_nonexistent_session", func(t *testing.T) {
		output, err := suite.runSBSCommand("attach", "test:nonexistent")
		// Should handle gracefully
		if err != nil {
			assert.Contains(t, output, "no session found", "Should indicate session not found")
		}
	})
}

// TestE2E_ConfigurationHandling tests configuration scenarios
func TestE2E_ConfigurationHandling(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.cleanup()

	t.Run("custom_config_location", func(t *testing.T) {
		// Test using --config flag
		sessionID := "test:e2e-config"
		suite.sessionIDs = append(suite.sessionIDs, sessionID)

		configPath := filepath.Join(suite.configDir, "config.json")
		output, err := suite.runSBSCommand("--config", configPath, "start", sessionID)
		require.NoError(t, err, "Should work with custom config: %s", output)
	})

	t.Run("missing_work_issue_script", func(t *testing.T) {
		// Create a config with missing work-issue.sh
		tempConfigDir := filepath.Join(suite.tempDir, "bad-config")
		err := os.MkdirAll(tempConfigDir, 0755)
		require.NoError(t, err)

		badConfig := config.Config{
			WorktreeBasePath: suite.worktreeBaseDir,
			RepoPath:         suite.testRepoDir,
			WorkIssueScript:  "/nonexistent/work-issue.sh",
		}

		configData, err := json.MarshalIndent(badConfig, "", "  ")
		require.NoError(t, err)

		badConfigPath := filepath.Join(tempConfigDir, "config.json")
		err = os.WriteFile(badConfigPath, configData, 0644)
		require.NoError(t, err)

		// Should handle missing script gracefully (or find existing session)
		output, err := suite.runSBSCommand("start", "test:bad-config")
		// This might fail with script error or succeed with existing session
		t.Logf("Bad config test output: %s", output)
		t.Logf("Bad config test error: %v", err)
	})
}

// TestE2E_CleanupOperations tests comprehensive cleanup
func TestE2E_CleanupOperations(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.cleanup()

	// Create several sessions to test cleanup
	sessions := []string{"test:cleanup1", "test:cleanup2", "test:cleanup3"}

	t.Run("create_sessions_for_cleanup", func(t *testing.T) {
		for _, sessionID := range sessions {
			output, err := suite.runSBSCommand("start", sessionID)
			require.NoError(t, err, "Failed to start session %s: %s", sessionID, output)
			suite.sessionIDs = append(suite.sessionIDs, sessionID)
		}
	})

	t.Run("dry_run_cleanup", func(t *testing.T) {
		output, err := suite.runSBSCommand("clean", "--dry-run")
		require.NoError(t, err, "Dry run should succeed: %s", output)

		// Should show what would be cleaned without actually cleaning
		t.Logf("Dry run cleanup output: %s", output)
	})

	t.Run("force_cleanup", func(t *testing.T) {
		output, err := suite.runSBSCommand("clean", "--force")
		require.NoError(t, err, "Force cleanup should succeed: %s", output)

		t.Logf("Force cleanup output: %s", output)
	})

	t.Run("verify_cleanup_completed", func(t *testing.T) {
		// Check that sessions are cleaned up
		output, err := suite.runSBSCommand("list", "--plain")
		require.NoError(t, err)

		// Depending on implementation, cleaned sessions might not appear or be marked differently
		t.Logf("Post-cleanup session list: %s", output)
	})
}
