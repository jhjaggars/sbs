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
			{"legacy_numeric", "123", "github", "123", false},
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

	t.Run("session_migration_compatibility", func(t *testing.T) {
		// Test that existing sessions can be migrated
		legacySessions := []config.SessionMetadata{
			{
				IssueNumber: 123,
				IssueTitle:  "Legacy issue",
				Branch:      "issue-123-legacy-issue",
			},
		}

		migratedSessions, err := config.MigrateSessionMetadata(legacySessions)
		require.NoError(t, err)
		assert.Len(t, migratedSessions, 1)

		migrated := migratedSessions[0]
		assert.Equal(t, "github", migrated.SourceType)
		assert.Equal(t, "github:123", migrated.NamespacedID)
		assert.Equal(t, 123, migrated.IssueNumber) // Preserved for compatibility
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

func TestStartCommand_BackwardCompatibility(t *testing.T) {
	t.Run("legacy_numeric_argument_parsing", func(t *testing.T) {
		// Test that `sbs start 123` still works exactly as before
		// This test validates the parsing logic without running the full command

		// Simulate legacy argument parsing
		args := []string{"123"}

		// This should be parsed as a legacy GitHub issue
		workItem, err := inputsource.ParseWorkItemID(args[0])
		require.NoError(t, err)
		assert.Equal(t, "github", workItem.Source)
		assert.Equal(t, "123", workItem.ID)

		// Verify legacy branch naming is preserved
		actualLegacyBranch := workItem.GetLegacyBranchName()
		assert.Contains(t, actualLegacyBranch, "issue-123")

		// Verify full ID format
		assert.Equal(t, "github:123", workItem.FullID())
	})

	t.Run("branch_naming_compatibility", func(t *testing.T) {
		// Test branch naming compatibility for different scenarios
		tests := []struct {
			name               string
			workItem           *inputsource.WorkItem
			expectedLegacy     string
			expectedNamespaced string
		}{
			{
				name: "github_issue",
				workItem: &inputsource.WorkItem{
					Source: "github",
					ID:     "123",
					Title:  "Fix auth bug",
				},
				expectedLegacy:     "issue-123-fix-auth-bug",
				expectedNamespaced: "issue-github-123-fix-auth-bug",
			},
			{
				name: "test_item",
				workItem: &inputsource.WorkItem{
					Source: "test",
					ID:     "quick",
					Title:  "Quick development test",
				},
				expectedLegacy:     "issue-test-quick-quick-development-test",
				expectedNamespaced: "issue-test-quick-quick-development-test",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				legacyBranch := tt.workItem.GetLegacyBranchName()
				namespacedBranch := tt.workItem.GetBranchName()

				assert.Equal(t, tt.expectedLegacy, legacyBranch)
				assert.Equal(t, tt.expectedNamespaced, namespacedBranch)
			})
		}
	})
}
