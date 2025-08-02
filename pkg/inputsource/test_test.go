package inputsource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestInputSource_ArbitraryItems(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		expectFound bool
		expectTitle string
	}{
		{"valid_id_quick", "quick", true, "Test work item: quick"},
		{"valid_id_my_test", "my-test", true, "Test work item: my-test"},
		{"valid_id_feature_x", "feature-x", true, "Test work item: feature-x"},
		{"valid_id_with_underscores", "test_with_underscores", true, "Test work item: test_with_underscores"},
		{"valid_id_alphanumeric", "test123", true, "Test work item: test123"},
		{"invalid_id_with_spaces", "test with spaces", false, ""},
		{"invalid_id_with_special_chars", "test@invalid", false, ""},
		{"empty_id", "", false, ""},
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
	assert.Len(t, items, 0) // Should return empty list since test items are created dynamically

	// Test with search query should also return empty
	items, err = source.ListWorkItems("test", 10)
	require.NoError(t, err)
	assert.Len(t, items, 0)
}

func TestTestInputSource_SearchFiltering(t *testing.T) {
	source := NewTestInputSource()

	// Since test items are created dynamically and ListWorkItems returns empty,
	// all search queries should return empty lists
	t.Run("search_any_term", func(t *testing.T) {
		items, err := source.ListWorkItems("hooks", 10)
		require.NoError(t, err)
		assert.Empty(t, items, "Search should return empty list for dynamically created test items")
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

	// Since test items are created dynamically, all limits should return empty
	t.Run("limit_greater_than_available", func(t *testing.T) {
		items, err := source.ListWorkItems("", 100)
		require.NoError(t, err)
		assert.Empty(t, items) // Should return empty list
	})

	t.Run("limit_less_than_available", func(t *testing.T) {
		items, err := source.ListWorkItems("", 2)
		require.NoError(t, err)
		assert.Empty(t, items) // Should return empty list
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

	t.Run("invalid_id_with_special_chars", func(t *testing.T) {
		item, err := source.GetWorkItem("test@invalid")
		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Contains(t, err.Error(), "invalid test work item ID")
	})

	t.Run("invalid_id_with_spaces", func(t *testing.T) {
		item, err := source.GetWorkItem("test with spaces")
		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Contains(t, err.Error(), "invalid test work item ID")
	})
}

func TestTestInputSource_WorkItemIntegration(t *testing.T) {
	// Test that returned WorkItems work correctly with WorkItem methods
	source := NewTestInputSource()

	item, err := source.GetWorkItem("my-test")
	require.NoError(t, err)

	t.Run("full_id", func(t *testing.T) {
		assert.Equal(t, "test:my-test", item.FullID())
	})

	t.Run("branch_name", func(t *testing.T) {
		expected := "issue-test-my-test-test-work-item-my-test"
		assert.Equal(t, expected, item.GetBranchName())
	})

}

func TestTestInputSource_ValidID(t *testing.T) {
	// Test the isValidTestID helper function
	tests := []struct {
		name     string
		id       string
		expected bool
	}{
		{"valid_alphanumeric", "test123", true},
		{"valid_with_hyphens", "test-with-hyphens", true},
		{"valid_with_underscores", "test_with_underscores", true},
		{"valid_mixed", "test-123_valid", true},
		{"invalid_with_spaces", "test with spaces", false},
		{"invalid_with_special_chars", "test@invalid", false},
		{"invalid_with_dots", "test.invalid", false},
		{"invalid_empty", "", false},
		{"valid_single_char", "a", true},
		{"valid_numbers_only", "123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidTestID(tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}
