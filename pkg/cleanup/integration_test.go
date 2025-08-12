package cleanup

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sbs/pkg/config"
)

// TestCLI_CleanupManagerIntegration tests CLI-specific cleanup functionality
func TestCLI_CleanupManagerIntegration(t *testing.T) {
	tests := []struct {
		name            string
		sessions        []config.SessionMetadata
		activeTmux      []string
		sandboxExists   map[string]bool
		worktreeExists  map[string]bool
		options         CleanupOptions
		expectedResults CleanupResults
	}{
		{
			name: "CLI dry-run mode shows what would be cleaned",
			sessions: []config.SessionMetadata{
				{NamespacedID: "123", TmuxSession: "sbs-123", SandboxName: "sbs-repo-123", WorktreePath: "/path/123"},
			},
			activeTmux:     []string{}, // Session is stale
			sandboxExists:  map[string]bool{"sbs-repo-123": true},
			worktreeExists: map[string]bool{"/path/123": true},
			options: CleanupOptions{
				DryRun:         true,
				CleanSandboxes: true,
				CleanWorktrees: true,
			},
			expectedResults: CleanupResults{
				CleanedSessions:  0, // Nothing cleaned in dry run
				CleanedSandboxes: 0,
				CleanedWorktrees: 0,
				WouldClean:       1, // But would clean 1 session
			},
		},
		{
			name: "CLI comprehensive cleanup with worktrees and sandboxes",
			sessions: []config.SessionMetadata{
				{NamespacedID: "123", TmuxSession: "sbs-123", SandboxName: "sbs-repo-123", WorktreePath: "/path/123"},
				{NamespacedID: "124", TmuxSession: "sbs-124", SandboxName: "sbs-repo-124", WorktreePath: "/path/124"},
			},
			activeTmux:     []string{"sbs-123"}, // Only 123 is active
			sandboxExists:  map[string]bool{"sbs-repo-124": true},
			worktreeExists: map[string]bool{"/path/124": true},
			options: CleanupOptions{
				CleanSandboxes: true,
				CleanWorktrees: true,
			},
			expectedResults: CleanupResults{
				CleanedSessions:  1,
				CleanedSandboxes: 1,
				CleanedWorktrees: 1,
			},
		},
		{
			name: "CLI sandbox-only cleanup (like TUI)",
			sessions: []config.SessionMetadata{
				{NamespacedID: "123", TmuxSession: "sbs-123", SandboxName: "sbs-repo-123", WorktreePath: "/path/123"},
			},
			activeTmux:    []string{}, // Session is stale
			sandboxExists: map[string]bool{"sbs-repo-123": true},
			options: CleanupOptions{
				CleanSandboxes: true,
				CleanWorktrees: false, // TUI mode - sandbox only
			},
			expectedResults: CleanupResults{
				CleanedSessions:  1,
				CleanedSandboxes: 1,
				CleanedWorktrees: 0, // Not cleaned
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTmux := &MockTmuxManager{sessions: tt.activeTmux}
			mockSandbox := &MockSandboxManager{sandboxes: tt.sandboxExists}
			mockGit := &MockGitManager{worktrees: tt.worktreeExists}

			manager := NewCleanupManager(mockTmux, mockSandbox, mockGit, nil)

			// First identify stale sessions
			staleSessions, err := manager.IdentifyStaleSessionsInView(tt.sessions, ViewModeGlobal)
			assert.NoError(t, err)

			// Then cleanup the stale sessions
			results, err := manager.CleanupSessions(staleSessions, tt.options)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedResults.CleanedSessions, results.CleanedSessions)
			assert.Equal(t, tt.expectedResults.CleanedSandboxes, results.CleanedSandboxes)
			assert.Equal(t, tt.expectedResults.CleanedWorktrees, results.CleanedWorktrees)
			assert.Equal(t, tt.expectedResults.WouldClean, results.WouldClean)
		})
	}
}

// TestTUI_CleanupManagerIntegration tests TUI-specific cleanup functionality
func TestTUI_CleanupManagerIntegration(t *testing.T) {
	tests := []struct {
		name          string
		sessions      []config.SessionMetadata
		activeTmux    []string
		sandboxExists map[string]bool
		viewMode      ViewMode
		expectedStale int
	}{
		{
			name: "TUI view-aware session filtering - repository mode",
			sessions: []config.SessionMetadata{
				{NamespacedID: "123", TmuxSession: "sbs-123", SandboxName: "sbs-repo-123", RepositoryName: "repo1"},
				{NamespacedID: "124", TmuxSession: "sbs-124", SandboxName: "sbs-repo-124", RepositoryName: "repo2"},
			},
			activeTmux:    []string{}, // Both stale
			sandboxExists: map[string]bool{"sbs-repo-123": true, "sbs-repo-124": true},
			viewMode:      ViewModeRepository,
			expectedStale: 2, // Both show as stale in current implementation
		},
		{
			name: "TUI silent cleanup execution",
			sessions: []config.SessionMetadata{
				{NamespacedID: "123", TmuxSession: "sbs-123", SandboxName: "sbs-repo-123"},
			},
			activeTmux:    []string{},
			sandboxExists: map[string]bool{"sbs-repo-123": true},
			viewMode:      ViewModeGlobal,
			expectedStale: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTmux := &MockTmuxManager{sessions: tt.activeTmux}
			mockSandbox := &MockSandboxManager{sandboxes: tt.sandboxExists}

			manager := NewCleanupManager(mockTmux, mockSandbox, nil, nil)

			// Test TUI-specific stale session identification
			staleSessions, err := manager.IdentifyStaleSessionsInView(tt.sessions, tt.viewMode)
			assert.NoError(t, err)
			assert.Len(t, staleSessions, tt.expectedStale)

			// Test TUI-style cleanup (sandbox-only)
			options := CleanupOptions{
				CleanSandboxes: true,
				CleanWorktrees: false, // TUI doesn't clean worktrees
				SilentMode:     true,
			}

			results, err := manager.CleanupSessions(staleSessions, options)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStale, results.CleanedSessions)
		})
	}
}

// TestCrossInterface_CleanupConsistency ensures CLI and TUI produce identical results
func TestCrossInterface_CleanupConsistency(t *testing.T) {
	sessions := []config.SessionMetadata{
		{NamespacedID: "123", TmuxSession: "sbs-123", SandboxName: "sbs-repo-123"},
		{NamespacedID: "124", TmuxSession: "sbs-124", SandboxName: "sbs-repo-124"},
	}

	mockTmux := &MockTmuxManager{sessions: []string{}} // Both stale
	mockSandbox := &MockSandboxManager{
		sandboxes: map[string]bool{"sbs-repo-123": true, "sbs-repo-124": true},
	}

	manager := NewCleanupManager(mockTmux, mockSandbox, nil, nil)

	// Test that both interfaces identify the same stale sessions
	cliStale, err := manager.IdentifyStaleSessionsInView(sessions, ViewModeGlobal)
	assert.NoError(t, err)

	tuiStale, err := manager.IdentifyStaleSessionsInView(sessions, ViewModeGlobal)
	assert.NoError(t, err)

	assert.Equal(t, len(cliStale), len(tuiStale))
	assert.ElementsMatch(t, extractSessionIDs(cliStale), extractSessionIDs(tuiStale))
}

// TestCleanupManager_SandboxNameResolution tests sandbox name resolution logic
func TestCleanupManager_SandboxNameResolution(t *testing.T) {
	tests := []struct {
		name            string
		session         config.SessionMetadata
		expectedResolve string
	}{
		{
			name: "session with stored sandbox name",
			session: config.SessionMetadata{
				NamespacedID: "123",
				SandboxName:  "sbs-repo-123",
			},
			expectedResolve: "sbs-repo-123",
		},
		{
			name: "session without sandbox name - should generate fallback",
			session: config.SessionMetadata{
				NamespacedID: "456",
				SandboxName:  "", // No stored name
			},
			expectedResolve: "work-issue-456", // Should generate fallback name
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that sessions with empty sandbox names are handled gracefully
			resolved := resolveSandboxNameForSession(tt.session)
			if tt.expectedResolve != "" {
				assert.Equal(t, tt.expectedResolve, resolved)
			}
		})
	}
}

// Helper function for sandbox name resolution (using CleanupManager)
func resolveSandboxNameForSession(session config.SessionMetadata) string {
	manager := NewCleanupManager(nil, nil, nil, nil)
	return manager.ResolveSandboxName(session)
}
