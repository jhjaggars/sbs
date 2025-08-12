package cleanup

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sbs/pkg/config"
)

// TestCleanupManager_Creation tests that CleanupManager struct exists and initializes correctly
func TestCleanupManager_Creation(t *testing.T) {
	// Test CleanupManager struct exists and initializes correctly
	manager := NewCleanupManager(nil, nil, nil, nil)
	assert.NotNil(t, manager)
}

// TestCleanupManager_SessionIdentification tests identification of stale sessions
func TestCleanupManager_SessionIdentification(t *testing.T) {
	tests := []struct {
		name               string
		sessions           []config.SessionMetadata
		activeTmuxSessions []string
		expectedStale      []string
		expectedError      error
	}{
		{
			name: "identify stale sessions correctly",
			sessions: []config.SessionMetadata{
				{NamespacedID: "123", TmuxSession: "sbs-123"},
				{NamespacedID: "124", TmuxSession: "sbs-124"},
			},
			activeTmuxSessions: []string{"sbs-123"},
			expectedStale:      []string{"124"},
			expectedError:      nil,
		},
		{
			name: "no stale sessions found",
			sessions: []config.SessionMetadata{
				{NamespacedID: "123", TmuxSession: "sbs-123"},
			},
			activeTmuxSessions: []string{"sbs-123"},
			expectedStale:      []string{},
			expectedError:      nil,
		},
		{
			name:               "empty sessions list",
			sessions:           []config.SessionMetadata{},
			activeTmuxSessions: []string{},
			expectedStale:      []string{},
			expectedError:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTmux := &MockTmuxManager{
				sessions: tt.activeTmuxSessions,
				error:    tt.expectedError,
			}

			manager := NewCleanupManager(mockTmux, nil, nil, nil)
			staleSessions, err := manager.IdentifyStaleSessionsInView(tt.sessions, ViewModeGlobal)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				staleIDs := extractSessionIDs(staleSessions)
				assert.ElementsMatch(t, tt.expectedStale, staleIDs)
			}
		})
	}
}

// TestCleanupManager_ResourceCleanup tests cleanup of various resources
func TestCleanupManager_ResourceCleanup(t *testing.T) {
	tests := []struct {
		name            string
		sessions        []config.SessionMetadata
		sandboxExists   map[string]bool
		worktreeExists  map[string]bool
		cleanupOptions  CleanupOptions
		expectedResults CleanupResults
		expectedError   error
	}{
		{
			name: "successful comprehensive cleanup",
			sessions: []config.SessionMetadata{
				{
					NamespacedID: "123",
					SandboxName:  "sbs-repo-123",
					WorktreePath: "/path/to/worktree-123",
					TmuxSession:  "sbs-123",
				},
			},
			sandboxExists:  map[string]bool{"sbs-repo-123": true},
			worktreeExists: map[string]bool{"/path/to/worktree-123": true},
			cleanupOptions: CleanupOptions{
				CleanSandboxes: true,
				CleanWorktrees: true,
				DryRun:         false,
			},
			expectedResults: CleanupResults{
				CleanedSessions:  1,
				CleanedSandboxes: 1,
				CleanedWorktrees: 1,
				Errors:           []error{},
			},
			expectedError: nil,
		},
		{
			name: "dry run mode",
			sessions: []config.SessionMetadata{
				{NamespacedID: "123", SandboxName: "sbs-repo-123"},
			},
			sandboxExists: map[string]bool{"sbs-repo-123": true},
			cleanupOptions: CleanupOptions{
				CleanSandboxes: true,
				DryRun:         true,
			},
			expectedResults: CleanupResults{
				CleanedSessions:  0, // Nothing actually cleaned in dry run
				CleanedSandboxes: 0,
				WouldClean:       1,
				Errors:           []error{},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSandbox := &MockSandboxManager{
				sandboxes: tt.sandboxExists,
			}

			mockGit := &MockGitManager{
				worktrees: tt.worktreeExists,
			}

			manager := NewCleanupManager(nil, mockSandbox, mockGit, nil)
			results, err := manager.CleanupSessions(tt.sessions, tt.cleanupOptions)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResults.CleanedSessions, results.CleanedSessions)
			assert.Equal(t, tt.expectedResults.CleanedSandboxes, results.CleanedSandboxes)
			assert.Equal(t, tt.expectedResults.CleanedWorktrees, results.CleanedWorktrees)
		})
	}
}

// Helper function to extract session IDs from session metadata
func extractSessionIDs(sessions []config.SessionMetadata) []string {
	ids := make([]string, len(sessions))
	for i, session := range sessions {
		ids[i] = session.NamespacedID
	}
	return ids
}
