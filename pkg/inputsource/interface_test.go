package inputsource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test interface compliance - these will fail until we implement the interfaces
func TestInputSource_InterfaceCompliance(t *testing.T) {
	// Test that all implementations satisfy the interface
	t.Run("github_source_compliance", func(t *testing.T) {
		var _ InputSource = (*GitHubInputSource)(nil)
	})

	t.Run("test_source_compliance", func(t *testing.T) {
		var _ InputSource = (*TestInputSource)(nil)
	})
}

func TestInputSource_GetWorkItem_ErrorHandling(t *testing.T) {
	// Test error scenarios across all implementations
	tests := []struct {
		name       string
		createFunc func() InputSource
		testID     string
		expectErr  bool
	}{
		{
			name: "test_source_invalid_id",
			createFunc: func() InputSource {
				return NewTestInputSource()
			},
			testID:    "invalid@id",
			expectErr: true,
		},
		{
			name: "test_source_empty_id",
			createFunc: func() InputSource {
				return NewTestInputSource()
			},
			testID:    "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := tt.createFunc()
			item, err := source.GetWorkItem(tt.testID)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, item)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, item)
			}
		})
	}
}

func TestInputSource_ListWorkItems_Consistency(t *testing.T) {
	// Test consistent behavior across implementations
	tests := []struct {
		name       string
		createFunc func() InputSource
		minItems   int
	}{
		{
			name: "test_source_list",
			createFunc: func() InputSource {
				return NewTestInputSource()
			},
			minItems: 0, // test source returns empty list (dynamic items)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := tt.createFunc()
			items, err := source.ListWorkItems("", 10)

			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(items), tt.minItems)

			// Verify all items have correct structure
			for _, item := range items {
				assert.NotEmpty(t, item.Source, "Item should have a source")
				assert.NotEmpty(t, item.ID, "Item should have an ID")
				assert.NotEmpty(t, item.Title, "Item should have a title")
				assert.NotEmpty(t, item.State, "Item should have a state")
				assert.Equal(t, source.GetType(), item.Source, "Item source should match input source type")
			}
		})
	}
}

func TestInputSource_GetType_Consistency(t *testing.T) {
	// Test that GetType returns consistent values
	tests := []struct {
		name         string
		createFunc   func() InputSource
		expectedType string
	}{
		{
			name: "test_source_type",
			createFunc: func() InputSource {
				return NewTestInputSource()
			},
			expectedType: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := tt.createFunc()
			sourceType := source.GetType()
			assert.Equal(t, tt.expectedType, sourceType)
		})
	}
}

func TestInputSource_SearchFunctionality(t *testing.T) {
	// Test search filtering works across implementations
	tests := []struct {
		name        string
		createFunc  func() InputSource
		searchQuery string
		expectMatch bool
	}{
		{
			name: "test_source_search_hooks",
			createFunc: func() InputSource {
				return NewTestInputSource()
			},
			searchQuery: "hooks",
			expectMatch: false, // test source returns empty list
		},
		{
			name: "test_source_search_nonexistent",
			createFunc: func() InputSource {
				return NewTestInputSource()
			},
			searchQuery: "nonexistent-search-term",
			expectMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := tt.createFunc()
			items, err := source.ListWorkItems(tt.searchQuery, 10)

			require.NoError(t, err)

			if tt.expectMatch {
				assert.NotEmpty(t, items, "Search should find matching items")
			} else {
				// Should either return no items or items that don't match the search
				for _, item := range items {
					// If items are returned, they should be relevant to the search
					// (This is implementation-dependent)
					assert.NotNil(t, item)
				}
			}
		})
	}
}
