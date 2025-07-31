package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionMetadata_FriendlyTitleField(t *testing.T) {
	// Test that FriendlyTitle field exists and is accessible
	metadata := &SessionMetadata{
		IssueNumber:    123,
		IssueTitle:     "Fix user authentication bug",
		FriendlyTitle:  "fix-user-authentication-bug",
		Branch:         "issue-123-fix-user-authentication-bug",
		WorktreePath:   "/home/user/.work-issue-worktrees/project/issue-123",
		TmuxSession:    "work-issue-project-123",
		SandboxName:    "work-issue-project-123",
		RepositoryName: "project",
		RepositoryRoot: "/home/user/project",
		CreatedAt:      "2025-07-31T08:00:00Z",
		LastActivity:   "2025-07-31T08:00:00Z",
		Status:         "active",
	}

	assert.Equal(t, "fix-user-authentication-bug", metadata.FriendlyTitle)
	assert.Equal(t, 123, metadata.IssueNumber)
	assert.Equal(t, "Fix user authentication bug", metadata.IssueTitle)
}

func TestSessionMetadata_JSONSerialization(t *testing.T) {
	// Test that FriendlyTitle field is included in JSON output
	metadata := &SessionMetadata{
		IssueNumber:    123,
		IssueTitle:     "Fix user authentication bug",
		FriendlyTitle:  "fix-user-authentication-bug",
		Branch:         "issue-123-fix-user-authentication-bug",
		WorktreePath:   "/home/user/.work-issue-worktrees/project/issue-123",
		TmuxSession:    "work-issue-project-123",
		SandboxName:    "work-issue-project-123",
		RepositoryName: "project",
		RepositoryRoot: "/home/user/project",
		CreatedAt:      "2025-07-31T08:00:00Z",
		LastActivity:   "2025-07-31T08:00:00Z",
		Status:         "active",
	}

	jsonData, err := json.Marshal(metadata)
	require.NoError(t, err)

	// Check that the JSON contains the friendly_title field
	assert.Contains(t, string(jsonData), `"friendly_title":"fix-user-authentication-bug"`)
	assert.Contains(t, string(jsonData), `"issue_number":123`)
	assert.Contains(t, string(jsonData), `"issue_title":"Fix user authentication bug"`)
}

func TestSessionMetadata_JSONDeserialization(t *testing.T) {
	// Test that JSON with FriendlyTitle can be unmarshaled correctly
	jsonData := `{
		"issue_number": 123,
		"issue_title": "Fix user authentication bug",
		"friendly_title": "fix-user-authentication-bug",
		"branch": "issue-123-fix-user-authentication-bug",
		"worktree_path": "/home/user/.work-issue-worktrees/project/issue-123",
		"tmux_session": "work-issue-project-123",
		"sandbox_name": "work-issue-project-123",
		"repository_name": "project",
		"repository_root": "/home/user/project",
		"created_at": "2025-07-31T08:00:00Z",
		"last_activity": "2025-07-31T08:00:00Z",
		"status": "active"
	}`

	var metadata SessionMetadata
	err := json.Unmarshal([]byte(jsonData), &metadata)
	require.NoError(t, err)

	assert.Equal(t, 123, metadata.IssueNumber)
	assert.Equal(t, "Fix user authentication bug", metadata.IssueTitle)
	assert.Equal(t, "fix-user-authentication-bug", metadata.FriendlyTitle)
	assert.Equal(t, "issue-123-fix-user-authentication-bug", metadata.Branch)
	assert.Equal(t, "work-issue-project-123", metadata.TmuxSession)
	assert.Equal(t, "active", metadata.Status)
}

func TestSessionMetadata_BackwardCompatibility(t *testing.T) {
	// Test that old JSON without FriendlyTitle can be loaded without breaking
	jsonData := `{
		"issue_number": 123,
		"issue_title": "Fix user authentication bug",
		"branch": "issue-123-fix-user-authentication-bug",
		"worktree_path": "/home/user/.work-issue-worktrees/project/issue-123",
		"tmux_session": "work-issue-project-123",
		"sandbox_name": "work-issue-project-123",
		"repository_name": "project",
		"repository_root": "/home/user/project",
		"created_at": "2025-07-31T08:00:00Z",
		"last_activity": "2025-07-31T08:00:00Z",
		"status": "active"
	}`

	var metadata SessionMetadata
	err := json.Unmarshal([]byte(jsonData), &metadata)
	require.NoError(t, err)

	assert.Equal(t, 123, metadata.IssueNumber)
	assert.Equal(t, "Fix user authentication bug", metadata.IssueTitle)
	assert.Equal(t, "", metadata.FriendlyTitle) // Should be empty string (Go zero value)
	assert.Equal(t, "issue-123-fix-user-authentication-bug", metadata.Branch)
	assert.Equal(t, "work-issue-project-123", metadata.TmuxSession)
	assert.Equal(t, "active", metadata.Status)
}

func TestSessionMetadata_DefaultFriendlyTitle(t *testing.T) {
	// Test that default value behavior works correctly
	metadata := &SessionMetadata{
		IssueNumber: 123,
		IssueTitle:  "Fix user authentication bug",
		// FriendlyTitle intentionally omitted
		Branch:         "issue-123-fix-user-authentication-bug",
		WorktreePath:   "/home/user/.work-issue-worktrees/project/issue-123",
		TmuxSession:    "work-issue-project-123",
		SandboxName:    "work-issue-project-123",
		RepositoryName: "project",
		RepositoryRoot: "/home/user/project",
		CreatedAt:      "2025-07-31T08:00:00Z",
		LastActivity:   "2025-07-31T08:00:00Z",
		Status:         "active",
	}

	// FriendlyTitle should be empty string (Go zero value for string)
	assert.Equal(t, "", metadata.FriendlyTitle)
	assert.Equal(t, 123, metadata.IssueNumber)
	assert.Equal(t, "Fix user authentication bug", metadata.IssueTitle)
}

func TestSessionMetadata_CompleteWorkflow(t *testing.T) {
	// Test a complete serialization -> deserialization workflow
	originalMetadata := &SessionMetadata{
		IssueNumber:    456,
		IssueTitle:     "Implement user profile page",
		FriendlyTitle:  "implement-user-profile-page",
		Branch:         "issue-456-implement-user-profile-page",
		WorktreePath:   "/home/user/.work-issue-worktrees/webapp/issue-456",
		TmuxSession:    "work-issue-webapp-456",
		SandboxName:    "work-issue-webapp-456",
		RepositoryName: "webapp",
		RepositoryRoot: "/home/user/webapp",
		CreatedAt:      "2025-07-31T09:00:00Z",
		LastActivity:   "2025-07-31T10:00:00Z",
		Status:         "active",
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(originalMetadata)
	require.NoError(t, err)

	// Deserialize from JSON
	var deserializedMetadata SessionMetadata
	err = json.Unmarshal(jsonData, &deserializedMetadata)
	require.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, originalMetadata.IssueNumber, deserializedMetadata.IssueNumber)
	assert.Equal(t, originalMetadata.IssueTitle, deserializedMetadata.IssueTitle)
	assert.Equal(t, originalMetadata.FriendlyTitle, deserializedMetadata.FriendlyTitle)
	assert.Equal(t, originalMetadata.Branch, deserializedMetadata.Branch)
	assert.Equal(t, originalMetadata.WorktreePath, deserializedMetadata.WorktreePath)
	assert.Equal(t, originalMetadata.TmuxSession, deserializedMetadata.TmuxSession)
	assert.Equal(t, originalMetadata.SandboxName, deserializedMetadata.SandboxName)
	assert.Equal(t, originalMetadata.RepositoryName, deserializedMetadata.RepositoryName)
	assert.Equal(t, originalMetadata.RepositoryRoot, deserializedMetadata.RepositoryRoot)
	assert.Equal(t, originalMetadata.CreatedAt, deserializedMetadata.CreatedAt)
	assert.Equal(t, originalMetadata.LastActivity, deserializedMetadata.LastActivity)
	assert.Equal(t, originalMetadata.Status, deserializedMetadata.Status)
}

func TestSessionMetadata_EmptyFriendlyTitle(t *testing.T) {
	// Test behavior when FriendlyTitle is explicitly set to empty string
	metadata := &SessionMetadata{
		IssueNumber:    789,
		IssueTitle:     "Fix critical security vulnerability",
		FriendlyTitle:  "", // Explicitly empty
		Branch:         "issue-789-fix-critical-security-vulnerability",
		WorktreePath:   "/home/user/.work-issue-worktrees/secure-app/issue-789",
		TmuxSession:    "work-issue-secure-app-789",
		SandboxName:    "work-issue-secure-app-789",
		RepositoryName: "secure-app",
		RepositoryRoot: "/home/user/secure-app",
		CreatedAt:      "2025-07-31T11:00:00Z",
		LastActivity:   "2025-07-31T12:00:00Z",
		Status:         "active",
	}

	// Serialize and deserialize
	jsonData, err := json.Marshal(metadata)
	require.NoError(t, err)

	var deserializedMetadata SessionMetadata
	err = json.Unmarshal(jsonData, &deserializedMetadata)
	require.NoError(t, err)

	// FriendlyTitle should remain empty
	assert.Equal(t, "", deserializedMetadata.FriendlyTitle)
	assert.Equal(t, 789, deserializedMetadata.IssueNumber)
	assert.Equal(t, "Fix critical security vulnerability", deserializedMetadata.IssueTitle)
}
