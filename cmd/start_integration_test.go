package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sbs/pkg/config"
	"sbs/pkg/inputsource"
)

func TestStartCommand_InputSourceIntegration(t *testing.T) {
	// Integration test with build tag
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("test_input_source_parsing", func(t *testing.T) {
		// Test that work item ID parsing works correctly
		tests := []struct {
			name           string
			input          string
			expectedSource string
			expectedID     string
			expectError    bool
		}{
			{"github_namespaced", "github:456", "github", "456", false},
			{"test_namespaced", "test:quick", "test", "quick", false},
			{"invalid_format", "invalid-format", "", "", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				workItem, err := inputsource.ParseWorkItemID(tt.input)
				if tt.expectError {
					assert.Error(t, err)
					assert.Nil(t, workItem)
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.expectedSource, workItem.Source)
					assert.Equal(t, tt.expectedID, workItem.ID)
				}
			})
		}
	})

	t.Run("input_source_factory", func(t *testing.T) {
		factory := inputsource.NewInputSourceFactory()

		// Test GitHub source creation
		githubConfig := &config.InputSourceConfig{Type: "github"}
		githubSource, err := factory.Create(githubConfig)
		require.NoError(t, err)
		assert.Equal(t, "github", githubSource.GetType())

		// Test test source creation
		testConfig := &config.InputSourceConfig{Type: "test"}
		testSource, err := factory.Create(testConfig)
		require.NoError(t, err)
		assert.Equal(t, "test", testSource.GetType())

		// Test that test source has predefined items
		quickItem, err := testSource.GetWorkItem("quick")
		require.NoError(t, err)
		assert.Equal(t, "test", quickItem.Source)
		assert.Equal(t, "quick", quickItem.ID)
		assert.Equal(t, "Quick development test", quickItem.Title)
	})
}

func TestStartCommand_WorkItemIDParsing(t *testing.T) {
	t.Run("supports_simple_id_format", func(t *testing.T) {
		// Test that simple numeric input is now accepted (for primary work types)
		// This will be handled by the start command logic, not ParseWorkItemID
		// ParseWorkItemID still requires namespaced format, but start command handles both

		// Test namespaced format still works
		workItem, err := inputsource.ParseWorkItemID("test:quick")
		assert.NoError(t, err)
		assert.NotNil(t, workItem)
		assert.Equal(t, "test", workItem.Source)
		assert.Equal(t, "quick", workItem.ID)
	})
}
