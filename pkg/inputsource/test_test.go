package inputsource

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestInputSource_PredefinedItems(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		expectFound bool
		expectTitle string
	}{
		{"quick_test", "quick", true, "Quick development test"},
		{"hooks_test", "hooks", true, "Test Claude Code hooks"},
		{"sandbox_test", "sandbox", true, "Test sandbox integration"},
		{"invalid_id", "nonexistent", false, ""},
	}

	source := NewTestInputSource()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := source.GetWorkItem(tt.id)
			if tt.expectFound {
				require.NoError(t, err)
				assert.Equal(t, tt.expectTitle, item.Title)
				assert.Equal(t, "test", item.Source)
				assert.Equal(t, tt.id, item.ID)
				assert.Equal(t, "open", item.State)
				assert.Equal(t, "", item.URL) // Test items don't have URLs
			} else {
				assert.Error(t, err)
				assert.Nil(t, item)
			}
		})
	}
}

func TestTestInputSource_ListWorkItems(t *testing.T) {
	source := NewTestInputSource()
	items, err := source.ListWorkItems("", 10)

	require.NoError(t, err)
	assert.Len(t, items, 3) // quick, hooks, sandbox

	// Verify all items have correct structure
	expectedItems := map[string]string{
		"quick":   "Quick development test",
		"hooks":   "Test Claude Code hooks",
		"sandbox": "Test sandbox integration",
	}

	for _, item := range items {
		assert.Equal(t, "test", item.Source)
		assert.NotEmpty(t, item.ID)
		assert.NotEmpty(t, item.Title)
		assert.Equal(t, "open", item.State)
		assert.Equal(t, "", item.URL)

		// Verify it's one of the expected items
		expectedTitle, exists := expectedItems[item.ID]
		assert.True(t, exists, "Unexpected item ID: %s", item.ID)
		assert.Equal(t, expectedTitle, item.Title)
	}
}

func TestTestInputSource_SearchFiltering(t *testing.T) {
	source := NewTestInputSource()

	t.Run("search_hooks", func(t *testing.T) {
		items, err := source.ListWorkItems("hooks", 10)
		require.NoError(t, err)

		found := false
		for _, item := range items {
			if strings.Contains(strings.ToLower(item.Title), "hooks") {
				found = true
				assert.Equal(t, "hooks", item.ID)
				break
			}
		}
		assert.True(t, found, "Search should find items containing 'hooks'")
	})

	t.Run("search_development", func(t *testing.T) {
		items, err := source.ListWorkItems("development", 10)
		require.NoError(t, err)

		found := false
		for _, item := range items {
			if strings.Contains(strings.ToLower(item.Title), "development") {
				found = true
				assert.Equal(t, "quick", item.ID)
				break
			}
		}
		assert.True(t, found, "Search should find items containing 'development'")
	})

	t.Run("search_case_insensitive", func(t *testing.T) {
		items, err := source.ListWorkItems("SANDBOX", 10)
		require.NoError(t, err)

		found := false
		for _, item := range items {
			if strings.Contains(strings.ToLower(item.Title), "sandbox") {
				found = true
				assert.Equal(t, "sandbox", item.ID)
				break
			}
		}
		assert.True(t, found, "Search should be case insensitive")
	})

	t.Run("search_no_matches", func(t *testing.T) {
		items, err := source.ListWorkItems("nonexistent-term", 10)
		require.NoError(t, err)
		assert.Empty(t, items, "Search with no matches should return empty list")
	})
}

func TestTestInputSource_GetType(t *testing.T) {
	source := NewTestInputSource()
	assert.Equal(t, "test", source.GetType())
}

func TestTestInputSource_LimitParameter(t *testing.T) {
	source := NewTestInputSource()

	t.Run("limit_greater_than_available", func(t *testing.T) {
		items, err := source.ListWorkItems("", 100)
		require.NoError(t, err)
		assert.Len(t, items, 3) // Should return all 3 items
	})

	t.Run("limit_less_than_available", func(t *testing.T) {
		items, err := source.ListWorkItems("", 2)
		require.NoError(t, err)
		assert.Len(t, items, 2) // Should return only 2 items
	})

	t.Run("limit_zero", func(t *testing.T) {
		items, err := source.ListWorkItems("", 0)
		require.NoError(t, err)
		assert.Empty(t, items) // Should return no items
	})
}

func TestTestInputSource_ErrorHandling(t *testing.T) {
	source := NewTestInputSource()

	t.Run("empty_id", func(t *testing.T) {
		item, err := source.GetWorkItem("")
		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Contains(t, err.Error(), "empty")
	})

	t.Run("whitespace_id", func(t *testing.T) {
		item, err := source.GetWorkItem("   ")
		assert.Error(t, err)
		assert.Nil(t, item)
	})

	t.Run("invalid_id", func(t *testing.T) {
		item, err := source.GetWorkItem("invalid-test-id")
		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestTestInputSource_WorkItemIntegration(t *testing.T) {
	// Test that returned WorkItems work correctly with WorkItem methods
	source := NewTestInputSource()

	item, err := source.GetWorkItem("quick")
	require.NoError(t, err)

	t.Run("full_id", func(t *testing.T) {
		assert.Equal(t, "test:quick", item.FullID())
	})

	t.Run("branch_name", func(t *testing.T) {
		expected := "issue-test-quick-quick-development-test"
		assert.Equal(t, expected, item.GetBranchName())
	})

}
