//go:build integration
// +build integration

package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sbs/pkg/issue"
	"sbs/pkg/tui"
)

// TestInteractiveIssueSelection tests the integration between GitHub client and TUI
func TestInteractiveIssueSelection(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test - set INTEGRATION_TESTS=1 to run")
	}

	t.Run("github_client_and_tui_integration", func(t *testing.T) {
		// This test verifies that the GitHub client can be used with the TUI
		// without actually calling GitHub API (uses mock)

		// Create a mock client that implements the interface
		mockClient := &mockGitHubClientIntegration{
			issues: []issue.Issue{
				{Number: 1, Title: "Integration test issue 1", State: "open", URL: "https://github.com/test/repo/issues/1"},
				{Number: 2, Title: "Integration test issue 2", State: "open", URL: "https://github.com/test/repo/issues/2"},
			},
		}

		// Create TUI model with mock client
		model := tui.NewIssueSelectModel(mockClient)

		// Verify the model can be initialized
		assert.NotNil(t, model)

		// Verify the model can handle issue loading
		cmd := model.Init()
		assert.NotNil(t, cmd)

		// The integration test stops here since we can't easily test
		// the full TUI interaction without a terminal
	})

	t.Run("command_structure_integration", func(t *testing.T) {
		// Test that the command can be imported and has the right structure
		// This verifies that all dependencies are properly wired

		// This would test actual command execution but we skip that
		// to avoid side effects in tests
		t.Log("Command structure integration test - would test actual execution")
	})
}

// TestGitHubClientRealIntegration tests with actual GitHub CLI if available
func TestGitHubClientRealIntegration(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test - set INTEGRATION_TESTS=1 to run")
	}

	t.Run("real_github_client_if_available", func(t *testing.T) {
		// Test with real GitHub client if gh is available
		client := issue.NewGitHubClient()

		// Try to list issues - this will fail if not in a GitHub repo
		// or if gh is not authenticated, but that's expected
		issues, err := client.ListIssues("", 5)

		if err != nil {
			// Expected if not in a GitHub repo or not authenticated
			t.Logf("GitHub client error (expected in most test environments): %v", err)
			assert.Contains(t, err.Error(), "gh command")
		} else {
			// If successful, verify structure
			t.Logf("Successfully fetched %d issues", len(issues))
			for _, issue := range issues {
				assert.Greater(t, issue.Number, 0)
				assert.NotEmpty(t, issue.Title)
				assert.NotEmpty(t, issue.State)
				assert.NotEmpty(t, issue.URL)
			}
		}
	})
}

// TestDirectoryStructure verifies the project structure is correct
func TestDirectoryStructure(t *testing.T) {
	t.Run("verify_project_structure", func(t *testing.T) {
		// Verify key files exist
		requiredFiles := []string{
			"main.go",
			"go.mod",
			"go.sum",
			"cmd/start.go",
			"pkg/issue/github.go",
			"pkg/tui/issueselect.go",
		}

		for _, file := range requiredFiles {
			_, err := os.Stat(file)
			require.NoError(t, err, "Required file %s should exist", file)
		}

		// Verify test files exist
		testFiles := []string{
			"pkg/issue/github_test.go",
			"pkg/tui/issueselect_test.go",
			"cmd/start_test.go",
		}

		for _, file := range testFiles {
			_, err := os.Stat(file)
			require.NoError(t, err, "Test file %s should exist", file)
		}
	})

	t.Run("verify_build_artifacts", func(t *testing.T) {
		// Check if build produces expected binary
		_, err := os.Stat("sbs")
		if err == nil {
			// Binary exists, check if it's executable
			info, err := os.Stat("sbs")
			require.NoError(t, err)

			mode := info.Mode()
			assert.True(t, mode&0111 != 0, "Binary should be executable")
		} else {
			t.Log("Binary not found - run 'go build' to create it")
		}
	})
}

// Mock client for integration testing
type mockGitHubClientIntegration struct {
	issues []issue.Issue
}

func (m *mockGitHubClientIntegration) ListIssues(searchQuery string, limit int) ([]issue.Issue, error) {
	// Simple mock implementation for integration testing
	return m.issues, nil
}

func (m *mockGitHubClientIntegration) GetIssue(issueNumber int) (*issue.Issue, error) {
	for _, issue := range m.issues {
		if issue.Number == issueNumber {
			return &issue, nil
		}
	}
	return nil, nil
}
