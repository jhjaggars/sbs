package inputsource

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sbs/pkg/issue"
)

// mockGitHubClient implements the GitHub client interface for testing
type mockGitHubClient struct {
	issues          map[int]*issue.Issue
	listResult      []issue.Issue
	getIssueCalled  bool
	lastIssueNumber int
	listCalled      bool
	lastSearchQuery string
	lastLimit       int
	getIssueError   error
	listIssuesError error
}

func (m *mockGitHubClient) GetIssue(issueNumber int) (*issue.Issue, error) {
	m.getIssueCalled = true
	m.lastIssueNumber = issueNumber

	if m.getIssueError != nil {
		return nil, m.getIssueError
	}

	if issue, exists := m.issues[issueNumber]; exists {
		return issue, nil
	}

	return nil, errors.New("issue not found")
}

func (m *mockGitHubClient) ListIssues(searchQuery string, limit int) ([]issue.Issue, error) {
	m.listCalled = true
	m.lastSearchQuery = searchQuery
	m.lastLimit = limit

	if m.listIssuesError != nil {
		return nil, m.listIssuesError
	}

	return m.listResult, nil
}

func TestGitHubInputSource_GetWorkItem(t *testing.T) {
	mockClient := &mockGitHubClient{
		issues: map[int]*issue.Issue{
			123: {Number: 123, Title: "Fix auth bug", State: "open", URL: "https://github.com/test/repo/issues/123"},
		},
	}

	source := &GitHubInputSource{client: mockClient}

	t.Run("successful_get", func(t *testing.T) {
		item, err := source.GetWorkItem("123")
		require.NoError(t, err)
		assert.Equal(t, "github", item.Source)
		assert.Equal(t, "123", item.ID)
		assert.Equal(t, "Fix auth bug", item.Title)
		assert.Equal(t, "open", item.State)
		assert.Equal(t, "https://github.com/test/repo/issues/123", item.URL)
		assert.Equal(t, "github:123", item.FullID())

		// Verify mock was called correctly
		assert.True(t, mockClient.getIssueCalled)
		assert.Equal(t, 123, mockClient.lastIssueNumber)
	})

	t.Run("invalid_id_format", func(t *testing.T) {
		item, err := source.GetWorkItem("invalid")
		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Contains(t, err.Error(), "invalid GitHub issue number")
	})

	t.Run("issue_not_found", func(t *testing.T) {
		item, err := source.GetWorkItem("999")
		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Contains(t, err.Error(), "failed to get GitHub issue")
	})

	t.Run("github_client_error", func(t *testing.T) {
		mockClient.getIssueError = errors.New("GitHub API error")

		item, err := source.GetWorkItem("123")
		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Contains(t, err.Error(), "failed to get GitHub issue")

		// Reset error for other tests
		mockClient.getIssueError = nil
	})
}

func TestGitHubInputSource_BackwardCompatibility(t *testing.T) {
	// Test that existing GitHub workflows still work
	mockClient := &mockGitHubClient{
		issues: map[int]*issue.Issue{
			456: {Number: 456, Title: "Legacy issue", State: "open"},
		},
	}

	source := &GitHubInputSource{client: mockClient}
	item, err := source.GetWorkItem("456")

	require.NoError(t, err)
	assert.Equal(t, "github:456", item.FullID())
	assert.Equal(t, "issue-github-456-legacy-issue", item.GetBranchName())
	assert.Equal(t, "issue-456-legacy-issue", item.GetLegacyBranchName()) // Legacy format for GitHub
}

func TestGitHubInputSource_ListWorkItems(t *testing.T) {
	// Test list functionality with GitHub client
	mockClient := &mockGitHubClient{
		listResult: []issue.Issue{
			{Number: 1, Title: "First issue", State: "open", URL: "https://github.com/test/repo/issues/1"},
			{Number: 2, Title: "Second issue", State: "open", URL: "https://github.com/test/repo/issues/2"},
		},
	}

	source := &GitHubInputSource{client: mockClient}

	t.Run("successful_list", func(t *testing.T) {
		items, err := source.ListWorkItems("", 10)

		require.NoError(t, err)
		assert.Len(t, items, 2)

		// Verify first item
		assert.Equal(t, "github", items[0].Source)
		assert.Equal(t, "1", items[0].ID)
		assert.Equal(t, "First issue", items[0].Title)
		assert.Equal(t, "open", items[0].State)

		// Verify second item
		assert.Equal(t, "github", items[1].Source)
		assert.Equal(t, "2", items[1].ID)
		assert.Equal(t, "Second issue", items[1].Title)

		// Verify mock was called correctly
		assert.True(t, mockClient.listCalled)
		assert.Equal(t, "", mockClient.lastSearchQuery)
		assert.Equal(t, 10, mockClient.lastLimit)
	})

	t.Run("list_with_search", func(t *testing.T) {
		items, err := source.ListWorkItems("auth", 5)

		require.NoError(t, err)
		assert.Len(t, items, 2) // Mock returns same items regardless of search

		// Verify search parameters were passed
		assert.Equal(t, "auth", mockClient.lastSearchQuery)
		assert.Equal(t, 5, mockClient.lastLimit)
	})

	t.Run("list_error", func(t *testing.T) {
		mockClient.listIssuesError = errors.New("GitHub API error")

		items, err := source.ListWorkItems("", 10)
		assert.Error(t, err)
		assert.Nil(t, items)
		assert.Contains(t, err.Error(), "failed to list GitHub issues")

		// Reset error for other tests
		mockClient.listIssuesError = nil
	})
}

func TestGitHubInputSource_GetType(t *testing.T) {
	source := NewGitHubInputSource()
	assert.Equal(t, "github", source.GetType())
}

func TestGitHubInputSource_Integration(t *testing.T) {
	// Test integration with WorkItem methods
	mockClient := &mockGitHubClient{
		issues: map[int]*issue.Issue{
			789: {Number: 789, Title: "Integration test issue", State: "closed", URL: "https://github.com/test/repo/issues/789"},
		},
	}

	source := &GitHubInputSource{client: mockClient}
	item, err := source.GetWorkItem("789")
	require.NoError(t, err)

	t.Run("full_id", func(t *testing.T) {
		assert.Equal(t, "github:789", item.FullID())
	})

	t.Run("branch_name", func(t *testing.T) {
		expected := "issue-github-789-integration-test-issue"
		assert.Equal(t, expected, item.GetBranchName())
	})

	t.Run("legacy_branch_name", func(t *testing.T) {
		// GitHub sources should use legacy format for backward compatibility
		expected := "issue-789-integration-test-issue"
		assert.Equal(t, expected, item.GetLegacyBranchName())
	})

	t.Run("state_preserved", func(t *testing.T) {
		assert.Equal(t, "closed", item.State)
	})

	t.Run("url_preserved", func(t *testing.T) {
		assert.Equal(t, "https://github.com/test/repo/issues/789", item.URL)
	})
}

func TestGitHubInputSource_EdgeCases(t *testing.T) {
	source := NewGitHubInputSource()

	t.Run("empty_id", func(t *testing.T) {
		item, err := source.GetWorkItem("")
		assert.Error(t, err)
		assert.Nil(t, item)
	})

	t.Run("negative_number", func(t *testing.T) {
		item, err := source.GetWorkItem("-1")
		assert.Error(t, err)
		assert.Nil(t, item)
	})

	t.Run("zero_number", func(t *testing.T) {
		item, err := source.GetWorkItem("0")
		assert.Error(t, err)
		assert.Nil(t, item)
	})

	t.Run("very_large_number", func(t *testing.T) {
		// This should not error on parsing, but will likely error on GitHub lookup
		item, err := source.GetWorkItem("999999999")
		assert.Error(t, err) // GitHub client will return "issue not found"
		assert.Nil(t, item)
	})
}

func TestGitHubInputSource_RealClientIntegration(t *testing.T) {
	// Test with the real client interface (without making actual calls)
	source := NewGitHubInputSource()

	// Just verify the source was created correctly
	assert.Equal(t, "github", source.GetType())
	assert.NotNil(t, source.client)

	// We can't test actual GitHub calls without network access,
	// but we can verify the structure is correct
}
