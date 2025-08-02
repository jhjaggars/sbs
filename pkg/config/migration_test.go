package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionMigration_LegacyToNamespaced(t *testing.T) {
	// Test migration of existing sessions to use namespaced IDs
	legacySessions := []SessionMetadata{
		{
			IssueNumber: 123,
			IssueTitle:  "Legacy issue",
			Branch:      "issue-123-legacy-issue",
			// Missing SourceType, NamespacedID fields
		},
		{
			IssueNumber: 456,
			IssueTitle:  "Another legacy issue",
			Branch:      "issue-456-another-legacy-issue",
			Status:      "active",
			// Missing SourceType, NamespacedID fields
		},
	}

	migratedSessions, err := MigrateSessionMetadata(legacySessions)
	require.NoError(t, err)
	assert.Len(t, migratedSessions, 2)

	// Check first session migration
	migrated1 := migratedSessions[0]
	assert.Equal(t, "github", migrated1.SourceType)
	assert.Equal(t, "github:123", migrated1.NamespacedID)
	assert.Equal(t, 123, migrated1.IssueNumber) // Preserved for compatibility
	assert.Equal(t, "Legacy issue", migrated1.IssueTitle)
	assert.Equal(t, "issue-123-legacy-issue", migrated1.Branch)

	// Check second session migration
	migrated2 := migratedSessions[1]
	assert.Equal(t, "github", migrated2.SourceType)
	assert.Equal(t, "github:456", migrated2.NamespacedID)
	assert.Equal(t, 456, migrated2.IssueNumber) // Preserved for compatibility
	assert.Equal(t, "active", migrated2.Status)
}

func TestSessionMigration_AlreadyMigrated(t *testing.T) {
	// Test that already migrated sessions are not modified
	existingSessions := []SessionMetadata{
		{
			IssueNumber:  456,
			SourceType:   "test",
			NamespacedID: "test:quick",
			Branch:       "issue-test-quick-dev-test",
			IssueTitle:   "Test session",
		},
		{
			IssueNumber:  789,
			SourceType:   "github",
			NamespacedID: "github:789",
			Branch:       "issue-github-789-new-format",
			IssueTitle:   "New format session",
		},
	}

	migratedSessions, err := MigrateSessionMetadata(existingSessions)
	require.NoError(t, err)
	assert.Len(t, migratedSessions, 2)

	// Should remain unchanged
	assert.Equal(t, existingSessions[0], migratedSessions[0])
	assert.Equal(t, existingSessions[1], migratedSessions[1])
}

func TestSessionMigration_MixedSessions(t *testing.T) {
	// Test migration of mixed legacy and new sessions
	mixedSessions := []SessionMetadata{
		// Legacy session
		{
			IssueNumber: 123,
			Branch:      "issue-123-legacy",
			IssueTitle:  "Legacy issue",
		},
		// Already migrated session
		{
			IssueNumber:  456,
			SourceType:   "github",
			NamespacedID: "github:456",
			Branch:       "issue-github-456-migrated",
			IssueTitle:   "Already migrated",
		},
		// Another legacy session
		{
			IssueNumber: 789,
			Branch:      "issue-789-another-legacy",
			IssueTitle:  "Another legacy",
		},
	}

	migratedSessions, err := MigrateSessionMetadata(mixedSessions)
	require.NoError(t, err)
	assert.Len(t, migratedSessions, 3)

	// Check first session was migrated
	assert.Equal(t, "github", migratedSessions[0].SourceType)
	assert.Equal(t, "github:123", migratedSessions[0].NamespacedID)

	// Check second session unchanged
	assert.Equal(t, "github:456", migratedSessions[1].NamespacedID)
	assert.Equal(t, "github", migratedSessions[1].SourceType)

	// Check third session was migrated
	assert.Equal(t, "github", migratedSessions[2].SourceType)
	assert.Equal(t, "github:789", migratedSessions[2].NamespacedID)
}

func TestSessionMigration_NonGitHubSessions(t *testing.T) {
	// Test migration preserves non-GitHub sessions correctly
	sessions := []SessionMetadata{
		{
			SourceType:   "test",
			NamespacedID: "test:hooks",
			Branch:       "issue-test-hooks-test-hooks",
			IssueTitle:   "Test hooks session",
		},
		{
			SourceType:   "jira",
			NamespacedID: "jira:PROJ-123",
			Branch:       "issue-jira-PROJ-123-jira-issue",
			IssueTitle:   "JIRA issue session",
		},
	}

	migratedSessions, err := MigrateSessionMetadata(sessions)
	require.NoError(t, err)
	assert.Len(t, migratedSessions, 2)

	// Should remain unchanged
	assert.Equal(t, sessions, migratedSessions)
}

func TestSessionMigration_NeedsDetection(t *testing.T) {
	// Test the helper function that detects if a session needs migration
	tests := []struct {
		name     string
		session  SessionMetadata
		expected bool
	}{
		{
			name: "legacy_session_needs_migration",
			session: SessionMetadata{
				IssueNumber: 123,
				Branch:      "issue-123-legacy",
				// Missing SourceType and NamespacedID
			},
			expected: true,
		},
		{
			name: "migrated_session_no_migration",
			session: SessionMetadata{
				IssueNumber:  456,
				SourceType:   "github",
				NamespacedID: "github:456",
			},
			expected: false,
		},
		{
			name: "test_session_no_migration",
			session: SessionMetadata{
				SourceType:   "test",
				NamespacedID: "test:quick",
			},
			expected: false,
		},
		{
			name: "empty_namespaced_id_needs_migration",
			session: SessionMetadata{
				IssueNumber:  789,
				SourceType:   "github",
				NamespacedID: "", // Empty namespaced ID
			},
			expected: true,
		},
		{
			name: "empty_source_type_needs_migration",
			session: SessionMetadata{
				IssueNumber:  789,
				SourceType:   "", // Empty source type
				NamespacedID: "github:789",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sessionNeedsMigration(&tt.session)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSessionMigration_ErrorHandling(t *testing.T) {
	t.Run("nil_sessions", func(t *testing.T) {
		migratedSessions, err := MigrateSessionMetadata(nil)
		require.NoError(t, err)
		assert.Empty(t, migratedSessions)
	})

	t.Run("empty_sessions", func(t *testing.T) {
		migratedSessions, err := MigrateSessionMetadata([]SessionMetadata{})
		require.NoError(t, err)
		assert.Empty(t, migratedSessions)
	})

	t.Run("session_with_zero_issue_number", func(t *testing.T) {
		// Test sessions are allowed to have zero issue numbers
		sessions := []SessionMetadata{
			{
				IssueNumber:  0,
				SourceType:   "test",
				NamespacedID: "test:quick",
				Branch:       "issue-test-quick-test",
			},
		}

		migratedSessions, err := MigrateSessionMetadata(sessions)
		require.NoError(t, err)
		assert.Len(t, migratedSessions, 1)
		assert.Equal(t, sessions[0], migratedSessions[0])
	})
}

func TestSessionMigration_PreservesAllFields(t *testing.T) {
	// Test that migration preserves all existing fields
	legacySession := SessionMetadata{
		IssueNumber:         123,
		IssueTitle:          "Legacy with all fields",
		FriendlyTitle:       "legacy-with-all-fields",
		Branch:              "issue-123-legacy-with-all-fields",
		WorktreePath:        "/path/to/worktree",
		TmuxSession:         "work-issue-123",
		SandboxName:         "work-issue-repo-123",
		RepositoryName:      "test-repo",
		RepositoryRoot:      "/path/to/repo",
		CreatedAt:           "2023-01-01T00:00:00Z",
		LastActivity:        "2023-01-02T00:00:00Z",
		Status:              "active",
		ResourceStatus:      "active",
		CurrentCreationStep: "complete",
		FailurePoint:        "",
		FailureReason:       "",
		ResourceCreationLog: []ResourceCreationEntry{
			{
				ResourceType: "branch",
				ResourceID:   "issue-123-legacy-with-all-fields",
				Status:       "created",
			},
		},
	}

	migratedSessions, err := MigrateSessionMetadata([]SessionMetadata{legacySession})
	require.NoError(t, err)
	assert.Len(t, migratedSessions, 1)

	migrated := migratedSessions[0]

	// Check new fields were added
	assert.Equal(t, "github", migrated.SourceType)
	assert.Equal(t, "github:123", migrated.NamespacedID)

	// Check all existing fields were preserved
	assert.Equal(t, legacySession.IssueNumber, migrated.IssueNumber)
	assert.Equal(t, legacySession.IssueTitle, migrated.IssueTitle)
	assert.Equal(t, legacySession.FriendlyTitle, migrated.FriendlyTitle)
	assert.Equal(t, legacySession.Branch, migrated.Branch)
	assert.Equal(t, legacySession.WorktreePath, migrated.WorktreePath)
	assert.Equal(t, legacySession.TmuxSession, migrated.TmuxSession)
	assert.Equal(t, legacySession.SandboxName, migrated.SandboxName)
	assert.Equal(t, legacySession.RepositoryName, migrated.RepositoryName)
	assert.Equal(t, legacySession.RepositoryRoot, migrated.RepositoryRoot)
	assert.Equal(t, legacySession.CreatedAt, migrated.CreatedAt)
	assert.Equal(t, legacySession.LastActivity, migrated.LastActivity)
	assert.Equal(t, legacySession.Status, migrated.Status)
	assert.Equal(t, legacySession.ResourceStatus, migrated.ResourceStatus)
	assert.Equal(t, legacySession.CurrentCreationStep, migrated.CurrentCreationStep)
	assert.Equal(t, legacySession.FailurePoint, migrated.FailurePoint)
	assert.Equal(t, legacySession.FailureReason, migrated.FailureReason)
	assert.Equal(t, legacySession.ResourceCreationLog, migrated.ResourceCreationLog)
}
