package inputsource

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkItem_ParseID(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedSource string
		expectedID     string
		expectedError  bool
	}{
		{"github_namespaced", "github:123", "github", "123", false},
		{"test_namespaced", "test:quick", "test", "quick", false},
		{"jira_namespaced", "jira:PROJ-456", "jira", "PROJ-456", false},
		{"invalid_format", "invalid-format", "", "", true},
		{"empty_source", ":123", "", "", true},
		{"empty_id", "github:", "", "", true},
		{"multiple_colons", "test:quick:extra", "", "", true},
		{"empty_string", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := ParseWorkItemID(tt.input)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, item)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedSource, item.Source)
				assert.Equal(t, tt.expectedID, item.ID)
			}
		})
	}
}

func TestWorkItem_FullID(t *testing.T) {
	tests := []struct {
		name     string
		item     *WorkItem
		expected string
	}{
		{
			name:     "github_item",
			item:     &WorkItem{Source: "github", ID: "123"},
			expected: "github:123",
		},
		{
			name:     "test_item",
			item:     &WorkItem{Source: "test", ID: "quick"},
			expected: "test:quick",
		},
		{
			name:     "jira_item",
			item:     &WorkItem{Source: "jira", ID: "PROJ-456"},
			expected: "jira:PROJ-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.item.FullID()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWorkItem_GetBranchName(t *testing.T) {
	tests := []struct {
		name     string
		item     *WorkItem
		expected string
	}{
		{
			name: "github_item",
			item: &WorkItem{
				Source: "github",
				ID:     "123",
				Title:  "Fix authentication bug",
			},
			expected: "issue-github-123-fix-authentication-bug",
		},
		{
			name: "test_item",
			item: &WorkItem{
				Source: "test",
				ID:     "quick",
				Title:  "Quick development test",
			},
			expected: "issue-test-quick-quick-development-test",
		},
		{
			name: "special_characters_in_title",
			item: &WorkItem{
				Source: "test",
				ID:     "special",
				Title:  "Fix bug with @#$% special chars!",
			},
			expected: "issue-test-special-fix-bug-with-special-chars",
		},
		{
			name: "empty_title",
			item: &WorkItem{
				Source: "test",
				ID:     "empty",
				Title:  "",
			},
			expected: "issue-test-empty",
		},
		{
			name: "whitespace_title",
			item: &WorkItem{
				Source: "test",
				ID:     "whitespace",
				Title:  "   ",
			},
			expected: "issue-test-whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.item.GetBranchName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWorkItem_EdgeCases(t *testing.T) {
	t.Run("special_characters_in_id", func(t *testing.T) {
		tests := []struct {
			name      string
			fullID    string
			expectErr bool
		}{
			{"jira_format", "jira:PROJ-123", false},
			{"underscores", "test:quick_test", false},
			{"spaces_invalid", "test:quick test", true},
			{"multiple_colons", "test:quick:extra", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				item, err := ParseWorkItemID(tt.fullID)
				if tt.expectErr {
					assert.Error(t, err)
					assert.Nil(t, item)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, item)
				}
			})
		}
	})

	t.Run("extremely_long_titles", func(t *testing.T) {
		longTitle := strings.Repeat("a", 500) // Very long title
		item := &WorkItem{
			Source: "test",
			ID:     "long",
			Title:  longTitle,
		}

		branch := item.GetBranchName()
		// Branch name should be truncated to reasonable length
		assert.Less(t, len(branch), 200) // Git has practical limits
		assert.Contains(t, branch, "issue-test-long")
	})

	t.Run("empty_and_whitespace_titles", func(t *testing.T) {
		tests := []struct {
			title          string
			expectedSuffix string
		}{
			{"", "issue-test-empty"},
			{"   ", "issue-test-empty"},
			{"   Mixed   Spaces   ", "issue-test-empty-mixed-spaces"},
		}

		for _, tt := range tests {
			item := &WorkItem{
				Source: "test",
				ID:     "empty",
				Title:  tt.title,
			}

			branch := item.GetBranchName()
			assert.Contains(t, branch, "issue-test-empty")
		}
	})
}
