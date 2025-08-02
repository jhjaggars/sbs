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
		{"legacy_github", "123", "github", "123", false}, // backward compatibility
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

func TestWorkItem_IsLegacyFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"numeric_legacy", "123", true},
		{"numeric_with_hash", "#456", true},
		{"github_namespaced", "github:123", false},
		{"test_namespaced", "test:quick", false},
		{"non_numeric", "abc", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLegacyFormat(tt.input)
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

func TestWorkItem_GetLegacyBranchName(t *testing.T) {
	tests := []struct {
		name     string
		item     *WorkItem
		expected string
	}{
		{
			name: "github_legacy_format",
			item: &WorkItem{
				Source: "github",
				ID:     "123",
				Title:  "Fix authentication bug",
			},
			expected: "issue-123-fix-authentication-bug",
		},
		{
			name: "non_github_uses_full_format",
			item: &WorkItem{
				Source: "test",
				ID:     "quick",
				Title:  "Quick development test",
			},
			expected: "issue-test-quick-quick-development-test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.item.GetLegacyBranchName()
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

func TestWorkItem_BranchNamingCompatibility(t *testing.T) {
	t.Run("github_legacy_format", func(t *testing.T) {
		// GitHub issues should maintain backward-compatible branch names for legacy sessions
		item := &WorkItem{
			Source: "github",
			ID:     "123",
			Title:  "Fix authentication bug",
		}

		// For legacy compatibility, GitHub branches should omit source prefix
		// when created from legacy numeric input
		legacyBranch := item.GetLegacyBranchName()
		assert.Equal(t, "issue-123-fix-authentication-bug", legacyBranch)

		// New namespaced format includes source
		namespacedBranch := item.GetBranchName()
		assert.Equal(t, "issue-github-123-fix-authentication-bug", namespacedBranch)
	})

	t.Run("new_sources_use_namespaced_format", func(t *testing.T) {
		// Non-GitHub sources should always use namespaced format
		item := &WorkItem{
			Source: "test",
			ID:     "quick",
			Title:  "Quick development test",
		}

		branch := item.GetBranchName()
		assert.Equal(t, "issue-test-quick-quick-development-test", branch)

		// No legacy format for non-GitHub sources
		assert.Equal(t, branch, item.GetLegacyBranchName())
	})
}
