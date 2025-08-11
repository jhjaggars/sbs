package config

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionMetadata_ResourceTracking(t *testing.T) {
	t.Run("resource_creation_log_initialization", func(t *testing.T) {
		// Test initialization of ResourceCreationLog field
		metadata := &SessionMetadata{
			IssueNumber:         123,
			IssueTitle:          "Test Issue",
			ResourceCreationLog: []ResourceCreationEntry{},
			ResourceStatus:      "active",
			CurrentCreationStep: "completed",
		}

		// Verify empty slice is properly initialized
		assert.NotNil(t, metadata.ResourceCreationLog)
		assert.Equal(t, 0, len(metadata.ResourceCreationLog))
		assert.Equal(t, "active", metadata.ResourceStatus)
		assert.Equal(t, "completed", metadata.CurrentCreationStep)

		// Assert proper JSON serialization
		jsonData, err := json.Marshal(metadata)
		require.NoError(t, err)
		// Note: omitempty means empty slice won't appear in JSON, but that's expected behavior
		assert.Contains(t, string(jsonData), `"resource_status":"active"`)
		assert.Contains(t, string(jsonData), `"current_creation_step":"completed"`)
	})

	t.Run("resource_creation_log_operations", func(t *testing.T) {
		// Test adding resource creation entries
		metadata := &SessionMetadata{
			IssueNumber:         123,
			ResourceCreationLog: []ResourceCreationEntry{},
		}

		// Add first entry
		entry1 := ResourceCreationEntry{
			ResourceType: "branch",
			ResourceID:   "issue-123-test-branch",
			CreatedAt:    time.Now(),
			Status:       "created",
			Metadata:     map[string]interface{}{"commit_hash": "abc123"},
		}
		metadata.ResourceCreationLog = append(metadata.ResourceCreationLog, entry1)

		// Add second entry
		entry2 := ResourceCreationEntry{
			ResourceType: "worktree",
			ResourceID:   "/tmp/worktree-123",
			CreatedAt:    time.Now().Add(1 * time.Second),
			Status:       "created",
			Metadata:     map[string]interface{}{"path": "/tmp/worktree-123"},
		}
		metadata.ResourceCreationLog = append(metadata.ResourceCreationLog, entry2)

		// Verify chronological ordering and data structure integrity
		assert.Equal(t, 2, len(metadata.ResourceCreationLog))
		assert.Equal(t, "branch", metadata.ResourceCreationLog[0].ResourceType)
		assert.Equal(t, "worktree", metadata.ResourceCreationLog[1].ResourceType)
		assert.True(t, metadata.ResourceCreationLog[1].CreatedAt.After(metadata.ResourceCreationLog[0].CreatedAt))
	})

	t.Run("resource_status_tracking", func(t *testing.T) {
		// Test ResourceStatus field updates
		metadata := &SessionMetadata{
			IssueNumber:    123,
			ResourceStatus: "creating",
		}

		// Verify status transitions (creating, active, cleanup, failed)
		assert.Equal(t, "creating", metadata.ResourceStatus)

		metadata.ResourceStatus = "active"
		assert.Equal(t, "active", metadata.ResourceStatus)

		metadata.ResourceStatus = "cleanup"
		assert.Equal(t, "cleanup", metadata.ResourceStatus)

		metadata.ResourceStatus = "failed"
		assert.Equal(t, "failed", metadata.ResourceStatus)
	})

	t.Run("creation_step_tracking", func(t *testing.T) {
		// Test CurrentCreationStep field
		metadata := &SessionMetadata{
			IssueNumber:         123,
			CurrentCreationStep: "branch_creation",
		}

		// Verify step progression tracking
		assert.Equal(t, "branch_creation", metadata.CurrentCreationStep)

		metadata.CurrentCreationStep = "worktree_creation"
		assert.Equal(t, "worktree_creation", metadata.CurrentCreationStep)

		metadata.CurrentCreationStep = "tmux_creation"
		assert.Equal(t, "tmux_creation", metadata.CurrentCreationStep)

		metadata.CurrentCreationStep = "sandbox_creation"
		assert.Equal(t, "sandbox_creation", metadata.CurrentCreationStep)

		metadata.CurrentCreationStep = "completed"
		assert.Equal(t, "completed", metadata.CurrentCreationStep)
	})

	t.Run("failure_point_tracking", func(t *testing.T) {
		// Test FailurePoint field for partial failures
		metadata := &SessionMetadata{
			IssueNumber:         123,
			ResourceStatus:      "failed",
			CurrentCreationStep: "tmux_creation",
			FailurePoint:        "tmux_creation",
			FailureReason:       "Failed to create tmux session: connection refused",
		}

		// Verify failure context capture
		assert.Equal(t, "failed", metadata.ResourceStatus)
		assert.Equal(t, "tmux_creation", metadata.FailurePoint)
		assert.Equal(t, "Failed to create tmux session: connection refused", metadata.FailureReason)

		// Assert recovery information storage
		assert.Equal(t, metadata.CurrentCreationStep, metadata.FailurePoint)
	})
}

func TestSessionMetadata_ResourceTrackingBackwardCompatibility(t *testing.T) {
	t.Run("load_legacy_session_metadata", func(t *testing.T) {
		// Test loading sessions without new resource tracking fields
		legacyJSON := `{
			"issue_number": 123,
			"issue_title": "Legacy Issue",
			"branch": "issue-123-legacy-issue",
			"worktree_path": "/tmp/worktree-123",
			"tmux_session": "sbs-123",
			"sandbox_name": "sbs-123",
			"repository_name": "test-repo",
			"repository_root": "/test",
			"created_at": "2025-08-01T10:00:00Z",
			"last_activity": "2025-08-01T10:00:00Z",
			"status": "active"
		}`

		var metadata SessionMetadata
		err := json.Unmarshal([]byte(legacyJSON), &metadata)
		require.NoError(t, err)

		// Verify default values are applied appropriately
		assert.Equal(t, 123, metadata.IssueNumber)
		assert.Equal(t, "active", metadata.Status)
		assert.Equal(t, "", metadata.ResourceStatus)      // Should be empty (default)
		assert.Equal(t, "", metadata.CurrentCreationStep) // Should be empty (default)
		assert.Equal(t, "", metadata.FailurePoint)        // Should be empty (default)
		assert.Equal(t, "", metadata.FailureReason)       // Should be empty (default)
		// ResourceCreationLog will be nil initially, which is expected behavior
		assert.Equal(t, 0, len(metadata.ResourceCreationLog))

		// Assert existing functionality is preserved
		assert.Equal(t, "Legacy Issue", metadata.IssueTitle)
		assert.Equal(t, "issue-123-legacy-issue", metadata.Branch)
	})

	t.Run("migrate_existing_sessions", func(t *testing.T) {
		// Test migration of existing session metadata
		existingMetadata := SessionMetadata{
			IssueNumber:    456,
			IssueTitle:     "Existing Issue",
			Branch:         "issue-456-existing-issue",
			WorktreePath:   "/tmp/worktree-456",
			TmuxSession:    "sbs-456",
			SandboxName:    "sbs-456",
			RepositoryName: "test-repo",
			RepositoryRoot: "/test",
			CreatedAt:      "2025-08-01T10:00:00Z",
			LastActivity:   "2025-08-01T10:00:00Z",
			Status:         "active",
			// New fields intentionally omitted to simulate existing data
		}

		// Simulate loading and then saving (which would initialize new fields)
		jsonData, err := json.Marshal(existingMetadata)
		require.NoError(t, err)

		var migratedMetadata SessionMetadata
		err = json.Unmarshal(jsonData, &migratedMetadata)
		require.NoError(t, err)

		// Verify new fields are properly initialized with defaults
		assert.Equal(t, "", migratedMetadata.ResourceStatus)
		assert.Equal(t, "", migratedMetadata.CurrentCreationStep)
		assert.Equal(t, "", migratedMetadata.FailurePoint)
		assert.Equal(t, "", migratedMetadata.FailureReason)
		assert.Equal(t, 0, len(migratedMetadata.ResourceCreationLog))

		// Assert no data loss during migration
		assert.Equal(t, existingMetadata.IssueNumber, migratedMetadata.IssueNumber)
		assert.Equal(t, existingMetadata.IssueTitle, migratedMetadata.IssueTitle)
		assert.Equal(t, existingMetadata.Branch, migratedMetadata.Branch)
		assert.Equal(t, existingMetadata.Status, migratedMetadata.Status)
	})
}

func TestSessionMetadata_ResourceCreationEntry(t *testing.T) {
	t.Run("resource_entry_creation", func(t *testing.T) {
		// Test creation of ResourceCreationEntry structs
		now := time.Now()
		entry := ResourceCreationEntry{
			ResourceType: "branch",
			ResourceID:   "issue-123-test-branch",
			CreatedAt:    now,
			Status:       "created",
			Metadata: map[string]interface{}{
				"commit_hash": "abc123",
				"remote_ref":  "origin/main",
			},
		}

		// Verify all required fields are populated
		assert.Equal(t, "branch", entry.ResourceType)
		assert.Equal(t, "issue-123-test-branch", entry.ResourceID)
		assert.Equal(t, now, entry.CreatedAt)
		assert.Equal(t, "created", entry.Status)
		assert.NotNil(t, entry.Metadata)
		assert.Equal(t, "abc123", entry.Metadata["commit_hash"])
		assert.Equal(t, "origin/main", entry.Metadata["remote_ref"])
	})

	t.Run("resource_entry_serialization", func(t *testing.T) {
		// Test JSON serialization of resource entries
		now := time.Now()
		entry := ResourceCreationEntry{
			ResourceType: "worktree",
			ResourceID:   "/tmp/worktree-456",
			CreatedAt:    now,
			Status:       "created",
			Metadata: map[string]interface{}{
				"path":   "/tmp/worktree-456",
				"branch": "issue-456-test",
			},
		}

		// Verify proper field mapping
		jsonData, err := json.Marshal(entry)
		require.NoError(t, err)
		assert.Contains(t, string(jsonData), `"resource_type":"worktree"`)
		assert.Contains(t, string(jsonData), `"resource_id":"/tmp/worktree-456"`)
		assert.Contains(t, string(jsonData), `"status":"created"`)

		// Assert deserialization accuracy
		var deserializedEntry ResourceCreationEntry
		err = json.Unmarshal(jsonData, &deserializedEntry)
		require.NoError(t, err)
		assert.Equal(t, entry.ResourceType, deserializedEntry.ResourceType)
		assert.Equal(t, entry.ResourceID, deserializedEntry.ResourceID)
		assert.Equal(t, entry.Status, deserializedEntry.Status)
		assert.Equal(t, entry.Metadata["path"], deserializedEntry.Metadata["path"])
		assert.Equal(t, entry.Metadata["branch"], deserializedEntry.Metadata["branch"])
	})
}

// Test helper functions
func CreateTestSessionWithResources(issueNumber int, resources []MockResourceEntry) *SessionMetadata {
	session := &SessionMetadata{
		IssueNumber:         issueNumber,
		ResourceCreationLog: make([]ResourceCreationEntry, 0),
		ResourceStatus:      "active",
		CurrentCreationStep: "completed",
	}

	for _, resource := range resources {
		entry := ResourceCreationEntry{
			ResourceType: resource.ResourceType,
			ResourceID:   resource.ResourceID,
			CreatedAt:    resource.CreatedAt,
			Status:       resource.Status,
			Metadata:     resource.Metadata,
		}
		session.ResourceCreationLog = append(session.ResourceCreationLog, entry)
	}

	return session
}

func CreateFailedSessionAtStep(issueNumber int, failureStep string) *SessionMetadata {
	return &SessionMetadata{
		IssueNumber:         issueNumber,
		ResourceStatus:      "failed",
		CurrentCreationStep: failureStep,
		FailurePoint:        failureStep,
		FailureReason:       "Simulated failure for testing",
		ResourceCreationLog: []ResourceCreationEntry{},
	}
}

type MockResourceEntry struct {
	ResourceType string
	ResourceID   string
	CreatedAt    time.Time
	Status       string
	Metadata     map[string]interface{}
}
