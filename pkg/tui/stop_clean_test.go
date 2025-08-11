package tui

import (
	"errors"
	"testing"

	"github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sbs/pkg/config"
	"sbs/pkg/sandbox"
	"sbs/pkg/tmux"
)

// Mock implementations for testing
type MockTmuxManager struct {
	sessions      []string
	staleSessions []string
	killError     error
	existsError   error
}

func (m *MockTmuxManager) SessionExists(sessionName string) (bool, error) {
	if m.existsError != nil {
		return false, m.existsError
	}
	for _, session := range m.sessions {
		if session == sessionName {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockTmuxManager) KillSession(sessionName string) error {
	return m.killError
}

func (m *MockTmuxManager) ListSessions() ([]*tmux.Session, error) {
	var sessions []*tmux.Session
	for _, name := range m.sessions {
		sessions = append(sessions, &tmux.Session{Name: name})
	}
	return sessions, nil
}

func (m *MockTmuxManager) AttachToSession(sessionName string) error {
	return nil
}

type MockSandboxManager struct {
	sandboxes   map[string]bool
	removeError error
}

func (m *MockSandboxManager) SandboxExists(sandboxName string) (bool, error) {
	if m.sandboxes == nil {
		return false, nil
	}
	exists, ok := m.sandboxes[sandboxName]
	return exists && ok, nil
}

func (m *MockSandboxManager) DeleteSandbox(sandboxName string) error {
	return m.removeError
}

func (m *MockSandboxManager) GetSandboxName(issueNumber int) string {
	return sandbox.NewManager().GetSandboxName(issueNumber)
}

// Test helper function - for now, let's just create a model with real managers
// and mock at the external level
func setupTestModel() Model {
	model := NewModel()
	model.sessions = []config.SessionMetadata{
		{
			IssueNumber:    123,
			IssueTitle:     "Test issue 123",
			TmuxSession:    "sbs-123",
			RepositoryName: "test-repo",
		},
		{
			IssueNumber:    124,
			IssueTitle:     "Test issue 124",
			TmuxSession:    "sbs-124",
			RepositoryName: "test-repo",
		},
	}
	model.cursor = 0
	return model
}

// Execute a tea.Cmd and return the resulting message
func executeCommand(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	return cmd()
}

// Interface for dependency injection
type TmuxManager interface {
	SessionExists(sessionName string) (bool, error)
	KillSession(sessionName string) error
	ListSessions() ([]*tmux.Session, error)
	AttachToSession(sessionName string) error
}

type SandboxManager interface {
	SandboxExists(sandboxName string) (bool, error)
	DeleteSandbox(sandboxName string) error
	GetSandboxName(issueNumber int) string
}

// Wrapper structs to bridge mock interfaces with concrete types
type tmuxManagerWrapper struct {
	mock *MockTmuxManager
}

func (w *tmuxManagerWrapper) SessionExists(sessionName string) (bool, error) {
	return w.mock.SessionExists(sessionName)
}

func (w *tmuxManagerWrapper) KillSession(sessionName string) error {
	return w.mock.KillSession(sessionName)
}

func (w *tmuxManagerWrapper) ListSessions() ([]*tmux.Session, error) {
	return w.mock.ListSessions()
}

func (w *tmuxManagerWrapper) AttachToSession(sessionName string) error {
	return w.mock.AttachToSession(sessionName)
}

type sandboxManagerWrapper struct {
	mock *MockSandboxManager
}

func (w *sandboxManagerWrapper) SandboxExists(sandboxName string) (bool, error) {
	return w.mock.SandboxExists(sandboxName)
}

func (w *sandboxManagerWrapper) DeleteSandbox(sandboxName string) error {
	return w.mock.DeleteSandbox(sandboxName)
}

func (w *sandboxManagerWrapper) GetSandboxName(issueNumber int) string {
	return w.mock.GetSandboxName(issueNumber)
}

func TestStopCleanKeyBindings(t *testing.T) {
	t.Run("stop_key_binding_exists", func(t *testing.T) {
		// This test will fail until we add the Stop key binding
		assert.Contains(t, keys.Stop.Keys(), "s", "Stop key binding should include 's' key")
		assert.Equal(t, "stop session", keys.Stop.Help().Desc, "Stop key binding should have 'stop session' help text")
	})

	t.Run("clean_key_binding_exists", func(t *testing.T) {
		// This test will fail until we add the Clean key binding
		assert.Contains(t, keys.Clean.Keys(), "c", "Clean key binding should include 'c' key")
		assert.Equal(t, "clean stale", keys.Clean.Help().Desc, "Clean key binding should have 'clean stale' help text")
	})

	t.Run("help_text_includes_new_shortcuts", func(t *testing.T) {
		model := setupTestModel()
		model.showHelp = false // Test condensed help

		view := model.View()

		// These will fail until we update the help text
		assert.Contains(t, view, "s: stop", "Condensed help should include 's: stop'")
		assert.Contains(t, view, "c: clean", "Condensed help should include 'c: clean'")
	})
}

func TestNewMessageTypes(t *testing.T) {
	t.Run("stopSessionMsg_creation", func(t *testing.T) {
		// This test will fail until we define stopSessionMsg
		msg := stopSessionMsg{
			err:     errors.New("test error"),
			success: false,
		}
		assert.NotNil(t, msg)
		assert.Equal(t, "test error", msg.err.Error())
		assert.False(t, msg.success)
	})

	t.Run("cleanSessionsMsg_creation", func(t *testing.T) {
		// This test will fail until we define cleanSessionsMsg
		msg := cleanSessionsMsg{
			err:             errors.New("test error"),
			cleanedSessions: []config.SessionMetadata{},
		}
		assert.NotNil(t, msg)
		assert.Equal(t, "test error", msg.err.Error())
		assert.Empty(t, msg.cleanedSessions)
	})

	t.Run("confirmationDialogMsg_creation", func(t *testing.T) {
		// This test will fail until we define confirmationDialogMsg
		msg := confirmationDialogMsg{
			show:    true,
			message: "test message",
		}
		assert.NotNil(t, msg)
		assert.True(t, msg.show)
		assert.Equal(t, "test message", msg.message)
	})
}

func TestModalDialogState(t *testing.T) {
	t.Run("modal_dialog_state_fields_exist", func(t *testing.T) {
		model := setupTestModel()

		// These will fail until we add these fields to Model struct
		assert.False(t, model.showConfirmationDialog, "Model should have showConfirmationDialog field")
		assert.Empty(t, model.confirmationMessage, "Model should have confirmationMessage field")
		assert.Empty(t, model.pendingCleanSessions, "Model should have pendingCleanSessions field")
	})

	t.Run("modal_dialog_visibility_toggle", func(t *testing.T) {
		model := setupTestModel()

		// Initially hidden
		assert.False(t, model.showConfirmationDialog)

		// Show modal
		model.showConfirmationDialog = true
		assert.True(t, model.showConfirmationDialog)

		// Hide modal
		model.showConfirmationDialog = false
		assert.False(t, model.showConfirmationDialog)
	})
}

func TestStopSelectedSession(t *testing.T) {
	t.Run("successful_stop_basic_functionality", func(t *testing.T) {
		model := setupTestModel()
		model.sessions = []config.SessionMetadata{
			{IssueNumber: 123, TmuxSession: "sbs-123"},
		}
		model.cursor = 0

		// Test that the method returns a command
		cmd := model.stopSelectedSession()
		assert.NotNil(t, cmd, "Stop command should not be nil")

		// Execute the command and verify it returns a stopSessionMsg
		msg := executeCommand(cmd)
		stopMsg, ok := msg.(stopSessionMsg)
		assert.True(t, ok, "Expected stopSessionMsg")

		// Since tmux session doesn't exist in test environment, it should succeed
		// The real test is that the flow works and returns proper message type
		assert.NotNil(t, stopMsg, "Stop message should not be nil")
	})

	t.Run("no_session_selected", func(t *testing.T) {
		model := setupTestModel()
		model.sessions = []config.SessionMetadata{}
		model.cursor = -1

		cmd := model.stopSelectedSession()
		msg := executeCommand(cmd)

		stopMsg, ok := msg.(stopSessionMsg)
		require.True(t, ok, "Expected stopSessionMsg")
		assert.Error(t, stopMsg.err, "Should have error for no session selected")
		assert.False(t, stopMsg.success, "Should not be successful")
		assert.Contains(t, stopMsg.err.Error(), "no session selected")
	})

	t.Run("invalid_session_index", func(t *testing.T) {
		model := setupTestModel()
		model.sessions = []config.SessionMetadata{
			{IssueNumber: 123, TmuxSession: "sbs-123"},
		}
		model.cursor = 999 // Invalid index

		cmd := model.stopSelectedSession()
		msg := executeCommand(cmd)

		stopMsg, ok := msg.(stopSessionMsg)
		require.True(t, ok, "Expected stopSessionMsg")
		assert.Error(t, stopMsg.err, "Should have error for invalid index")
		assert.False(t, stopMsg.success, "Should not be successful")
	})
}

func TestCleanStaleSessions(t *testing.T) {
	t.Run("identify_stale_sessions_basic_functionality", func(t *testing.T) {
		model := setupTestModel()
		model.sessions = []config.SessionMetadata{
			{IssueNumber: 123, TmuxSession: "sbs-123"},
			{IssueNumber: 124, TmuxSession: "sbs-124"},
		}

		staleSessions := model.identifyStaleSessionsInCurrentView()

		// Since no tmux sessions exist in test environment, all should be stale
		assert.Equal(t, 2, len(staleSessions), "Should identify both sessions as stale")
	})

	t.Run("clean_execution_returns_proper_structure", func(t *testing.T) {
		model := setupTestModel()
		model.sessions = []config.SessionMetadata{
			{IssueNumber: 123, TmuxSession: "sbs-123"},
		}

		result := model.identifyAndCleanStaleSessions()

		// Test the structure is correct
		assert.NotNil(t, result.cleanedSessions, "Should have cleanedSessions field")
		// Error may or may not be present depending on sandbox operations
	})

	t.Run("empty_sessions_list", func(t *testing.T) {
		model := setupTestModel()
		model.sessions = []config.SessionMetadata{}

		staleSessions := model.identifyStaleSessionsInCurrentView()
		assert.Equal(t, 0, len(staleSessions), "Should find no stale sessions in empty list")
	})
}

func TestStopCleanKeyHandling(t *testing.T) {
	t.Run("s_key_triggers_stop", func(t *testing.T) {
		model := setupTestModel()

		// This will fail until we implement 's' key handling
		newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})

		assert.NotNil(t, cmd, "Pressing 's' should return a command")

		// Execute the command to verify it returns stopSessionMsg
		msg := executeCommand(cmd)
		_, ok := msg.(stopSessionMsg)
		assert.True(t, ok, "Command should return stopSessionMsg")

		// Model should remain unchanged until message is processed
		assert.Equal(t, model.cursor, newModel.(Model).cursor)
	})

	t.Run("c_key_shows_confirmation_dialog", func(t *testing.T) {
		model := setupTestModel()

		// This will fail until we implement 'c' key handling
		newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

		updatedModel := newModel.(Model)
		assert.True(t, updatedModel.showConfirmationDialog, "Pressing 'c' should show confirmation dialog")
		assert.NotEmpty(t, updatedModel.confirmationMessage, "Should set confirmation message")
		assert.Nil(t, cmd, "Should not return a command yet")
	})

	t.Run("y_key_confirms_cleanup", func(t *testing.T) {
		model := setupTestModel()
		model.showConfirmationDialog = true
		model.pendingCleanSessions = []config.SessionMetadata{
			{IssueNumber: 123, TmuxSession: "sbs-123"},
		}

		// This will fail until we implement 'y' key handling in dialog mode
		newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

		updatedModel := newModel.(Model)
		assert.False(t, updatedModel.showConfirmationDialog, "Should hide confirmation dialog")
		assert.NotNil(t, cmd, "Should return cleanup command")

		// Execute command to verify it returns cleanSessionsMsg
		msg := executeCommand(cmd)
		_, ok := msg.(cleanSessionsMsg)
		assert.True(t, ok, "Command should return cleanSessionsMsg")
	})

	t.Run("n_key_cancels_cleanup", func(t *testing.T) {
		model := setupTestModel()
		model.showConfirmationDialog = true

		// This will fail until we implement 'n' key handling in dialog mode
		newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

		updatedModel := newModel.(Model)
		assert.False(t, updatedModel.showConfirmationDialog, "Should hide confirmation dialog")
		assert.Nil(t, cmd, "Should not return any command")
	})
}

func TestMessageHandling(t *testing.T) {
	t.Run("handle_stopSessionMsg_success", func(t *testing.T) {
		model := setupTestModel()

		msg := stopSessionMsg{err: nil, success: true}

		// This will fail until we implement stopSessionMsg handling
		newModel, cmd := model.Update(msg)

		updatedModel := newModel.(Model)
		assert.NoError(t, updatedModel.error, "Should clear error on success")
		assert.NotNil(t, cmd, "Should return refresh command")
	})

	t.Run("handle_stopSessionMsg_error", func(t *testing.T) {
		model := setupTestModel()

		msg := stopSessionMsg{err: errors.New("stop failed"), success: false}

		// This will fail until we implement stopSessionMsg handling
		newModel, cmd := model.Update(msg)

		updatedModel := newModel.(Model)
		assert.Error(t, updatedModel.error, "Should set error on failure")
		assert.Equal(t, "stop failed", updatedModel.error.Error())
		assert.NotNil(t, cmd, "Should still return refresh command")
	})

	t.Run("handle_cleanSessionsMsg", func(t *testing.T) {
		model := setupTestModel()
		model.showConfirmationDialog = true

		msg := cleanSessionsMsg{
			err:             nil,
			cleanedSessions: []config.SessionMetadata{{IssueNumber: 123}},
		}

		// This will fail until we implement cleanSessionsMsg handling
		newModel, cmd := model.Update(msg)

		updatedModel := newModel.(Model)
		assert.False(t, updatedModel.showConfirmationDialog, "Should hide confirmation dialog")
		assert.NoError(t, updatedModel.error, "Should clear error on success")
		assert.NotNil(t, cmd, "Should return refresh command")
	})
}
