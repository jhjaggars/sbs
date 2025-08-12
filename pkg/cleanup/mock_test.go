package cleanup

import (
	"sbs/pkg/config"
	"sbs/pkg/git"
	"sbs/pkg/tmux"
)

// MockTmuxManager implements a mock tmux manager for testing
type MockTmuxManager struct {
	sessions []string
	error    error
}

func (m *MockTmuxManager) SessionExists(sessionName string) (bool, error) {
	if m.error != nil {
		return false, m.error
	}
	for _, session := range m.sessions {
		if session == sessionName {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockTmuxManager) KillSession(sessionName string) error {
	return m.error
}

func (m *MockTmuxManager) ListSessions() ([]*tmux.Session, error) {
	if m.error != nil {
		return nil, m.error
	}
	var sessions []*tmux.Session
	for _, name := range m.sessions {
		sessions = append(sessions, &tmux.Session{Name: name})
	}
	return sessions, nil
}

func (m *MockTmuxManager) AttachToSession(sessionName string) error {
	return m.error
}

// MockSandboxManager implements a mock sandbox manager for testing
type MockSandboxManager struct {
	sandboxes    map[string]bool
	deleteErrors map[string]error
	error        error
}

func (m *MockSandboxManager) SandboxExists(sandboxName string) (bool, error) {
	if m.error != nil {
		return false, m.error
	}
	if m.sandboxes == nil {
		return false, nil
	}
	exists, ok := m.sandboxes[sandboxName]
	return exists && ok, nil
}

func (m *MockSandboxManager) DeleteSandbox(sandboxName string) error {
	if m.error != nil {
		return m.error
	}
	if m.deleteErrors != nil {
		if err, exists := m.deleteErrors[sandboxName]; exists {
			return err
		}
	}
	return nil
}

// MockGitManager implements a mock git manager for testing
type MockGitManager struct {
	worktrees map[string]bool
	branches  []string
	error     error
}

func (m *MockGitManager) FindOrphanedIssueBranches(activeWorkItems []string) ([]string, error) {
	if m.error != nil {
		return nil, m.error
	}
	return m.branches, nil
}

func (m *MockGitManager) DeleteMultipleBranches(branches []string, force bool) ([]git.BranchDeletionResult, error) {
	if m.error != nil {
		return nil, m.error
	}
	var results []git.BranchDeletionResult
	for _, branch := range branches {
		results = append(results, git.BranchDeletionResult{
			BranchName: branch,
			Success:    true,
			Message:    "deleted",
		})
	}
	return results, nil
}

func (m *MockGitManager) WorktreeExists(path string) bool {
	if m.worktrees == nil {
		return false
	}
	exists, ok := m.worktrees[path]
	return exists && ok
}

// MockConfigManager implements a mock config manager for testing
type MockConfigManager struct {
	sessions  []config.SessionMetadata
	saveError error
}

func (m *MockConfigManager) LoadAllRepositorySessions() ([]config.SessionMetadata, error) {
	return m.sessions, nil
}

func (m *MockConfigManager) SaveSessions(sessions []config.SessionMetadata) error {
	return m.saveError
}
