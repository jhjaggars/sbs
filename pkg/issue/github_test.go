package issue

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data fixtures
var sampleIssuesJSON = `[
    {
        "number": 123,
        "title": "Fix authentication bug",
        "state": "open",
        "url": "https://github.com/owner/repo/issues/123"
    },
    {
        "number": 124,
        "title": "Add dark mode support",
        "state": "open", 
        "url": "https://github.com/owner/repo/issues/124"
    }
]`

var singleIssueJSON = `[
    {
        "number": 42,
        "title": "Single issue for testing",
        "state": "open",
        "url": "https://github.com/owner/repo/issues/42"
    }
]`

var emptyIssuesJSON = `[]`

// Mock command executor for testing
type mockCommandExecutor struct {
	expectedCmd    string
	expectedArgs   []string
	mockOutput     []byte
	mockError      error
	mockStderr     []byte
	callCount      int
	actualCommands [][]string
}

func (m *mockCommandExecutor) executeCommand(name string, args ...string) ([]byte, error) {
	m.callCount++
	m.actualCommands = append(m.actualCommands, append([]string{name}, args...))
	
	if m.mockError != nil {
		if len(m.mockStderr) > 0 {
			return nil, &exec.ExitError{Stderr: m.mockStderr}
		}
		return nil, m.mockError
	}
	return m.mockOutput, nil
}

func TestGitHubClient_ListIssues(t *testing.T) {
	t.Run("successful_list_with_no_search", func(t *testing.T) {
		// Arrange
		mockExec := &mockCommandExecutor{
			mockOutput: []byte(sampleIssuesJSON),
		}
		client := &GitHubClient{executor: mockExec}
		
		// Act
		issues, err := client.ListIssues("", 100)
		
		// Assert
		require.NoError(t, err)
		assert.Len(t, issues, 2)
		assert.Equal(t, 123, issues[0].Number)
		assert.Equal(t, "Fix authentication bug", issues[0].Title)
		assert.Equal(t, "open", issues[0].State)
		assert.Equal(t, "https://github.com/owner/repo/issues/123", issues[0].URL)
		
		// Verify command was called correctly
		assert.Equal(t, 1, mockExec.callCount)
		expectedCmd := []string{"gh", "issue", "list", "--json", "number,title,state,url", "--state", "open", "--limit", "100"}
		assert.Equal(t, expectedCmd, mockExec.actualCommands[0])
	})
	
	t.Run("successful_list_with_search_query", func(t *testing.T) {
		// Arrange
		mockExec := &mockCommandExecutor{
			mockOutput: []byte(sampleIssuesJSON),
		}
		client := &GitHubClient{executor: mockExec}
		
		// Act
		issues, err := client.ListIssues("authentication", 100)
		
		// Assert
		require.NoError(t, err)
		assert.Len(t, issues, 2)
		
		// Verify search parameter was included
		expectedCmd := []string{"gh", "issue", "list", "--json", "number,title,state,url", "--state", "open", "--limit", "100", "--search", "authentication"}
		assert.Equal(t, expectedCmd, mockExec.actualCommands[0])
	})
	
	t.Run("successful_list_with_limit", func(t *testing.T) {
		// Arrange
		mockExec := &mockCommandExecutor{
			mockOutput: []byte(sampleIssuesJSON),
		}
		client := &GitHubClient{executor: mockExec}
		
		// Act
		issues, err := client.ListIssues("", 50)
		
		// Assert
		require.NoError(t, err)
		assert.Len(t, issues, 2)
		
		// Verify limit parameter was set correctly
		expectedCmd := []string{"gh", "issue", "list", "--json", "number,title,state,url", "--state", "open", "--limit", "50"}
		assert.Equal(t, expectedCmd, mockExec.actualCommands[0])
	})
	
	t.Run("empty_issue_list", func(t *testing.T) {
		// Arrange
		mockExec := &mockCommandExecutor{
			mockOutput: []byte(emptyIssuesJSON),
		}
		client := &GitHubClient{executor: mockExec}
		
		// Act
		issues, err := client.ListIssues("", 100)
		
		// Assert
		require.NoError(t, err)
		assert.Len(t, issues, 0)
		assert.NotNil(t, issues) // Should return empty slice, not nil
	})
	
	t.Run("single_issue_result", func(t *testing.T) {
		// Arrange
		mockExec := &mockCommandExecutor{
			mockOutput: []byte(singleIssueJSON),
		}
		client := &GitHubClient{executor: mockExec}
		
		// Act
		issues, err := client.ListIssues("", 100)
		
		// Assert
		require.NoError(t, err)
		assert.Len(t, issues, 1)
		assert.Equal(t, 42, issues[0].Number)
		assert.Equal(t, "Single issue for testing", issues[0].Title)
	})
	
	t.Run("gh_command_not_found", func(t *testing.T) {
		// Arrange
		mockExec := &mockCommandExecutor{
			mockError: &exec.Error{Name: "gh", Err: exec.ErrNotFound},
		}
		client := &GitHubClient{executor: mockExec}
		
		// Act
		issues, err := client.ListIssues("", 100)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, issues)
		assert.Contains(t, err.Error(), "gh command not found")
	})
	
	t.Run("gh_authentication_error", func(t *testing.T) {
		// Arrange
		authErrorMessage := "gh: To get started with GitHub CLI, please run: gh auth login"
		mockExec := &mockCommandExecutor{
			mockError:  &exec.ExitError{},
			mockStderr: []byte(authErrorMessage),
		}
		client := &GitHubClient{executor: mockExec}
		
		// Act
		issues, err := client.ListIssues("", 100)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, issues)
		assert.Contains(t, err.Error(), "authentication")
	})
	
	t.Run("invalid_json_response", func(t *testing.T) {
		// Arrange
		mockExec := &mockCommandExecutor{
			mockOutput: []byte(`{"invalid": json}`),
		}
		client := &GitHubClient{executor: mockExec}
		
		// Act
		issues, err := client.ListIssues("", 100)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, issues)
		assert.Contains(t, err.Error(), "failed to parse")
	})
	
	t.Run("gh_command_exit_error", func(t *testing.T) {
		// Arrange
		mockExec := &mockCommandExecutor{
			mockError:  &exec.ExitError{},
			mockStderr: []byte("Some GitHub API error"),
		}
		client := &GitHubClient{executor: mockExec}
		
		// Act
		issues, err := client.ListIssues("", 100)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, issues)
		assert.Contains(t, err.Error(), "failed to list issues")
	})
	
	t.Run("network_connectivity_error", func(t *testing.T) {
		// Arrange
		networkError := "network is unreachable"
		mockExec := &mockCommandExecutor{
			mockError:  &exec.ExitError{},
			mockStderr: []byte(networkError),
		}
		client := &GitHubClient{executor: mockExec}
		
		// Act
		issues, err := client.ListIssues("", 100)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, issues)
		assert.Contains(t, err.Error(), "failed to list issues")
	})
}

// Test the existing GetIssue method to ensure we don't break it
func TestGitHubClient_GetIssue_ExistingFunctionality(t *testing.T) {
	t.Run("get_single_issue_success", func(t *testing.T) {
		// This test ensures we don't break existing functionality
		singleIssueGetJSON := `{
			"number": 123,
			"title": "Fix authentication bug",
			"state": "open",
			"url": "https://github.com/owner/repo/issues/123"
		}`
		
		mockExec := &mockCommandExecutor{
			mockOutput: []byte(singleIssueGetJSON),
		}
		client := &GitHubClient{executor: mockExec}
		
		// Act
		issue, err := client.GetIssue(123)
		
		// Assert
		require.NoError(t, err)
		assert.Equal(t, 123, issue.Number)
		assert.Equal(t, "Fix authentication bug", issue.Title)
		
		// Verify correct command was called
		expectedCmd := []string{"gh", "issue", "view", "123", "--json", "number,title,state,url"}
		assert.Equal(t, expectedCmd, mockExec.actualCommands[0])
	})
}