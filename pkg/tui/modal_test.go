package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"sbs/pkg/config"
)

func TestModalDialogInteractions(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		initialState   bool
		expectedState  bool
		expectedAction string
	}{
		{
			name:           "show_confirmation_dialog",
			key:            "c",
			initialState:   false,
			expectedState:  true,
			expectedAction: "show_dialog",
		},
		{
			name:           "confirm_with_y_key",
			key:            "y",
			initialState:   true,
			expectedState:  false,
			expectedAction: "execute_cleanup",
		},
		{
			name:           "cancel_with_n_key",
			key:            "n",
			initialState:   true,
			expectedState:  false,
			expectedAction: "cancel_cleanup",
		},
		{
			name:           "cancel_with_escape",
			key:            "esc",
			initialState:   true,
			expectedState:  false,
			expectedAction: "cancel_cleanup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := setupTestModel()
			model.showConfirmationDialog = tt.initialState

			var keyMsg tea.KeyMsg
			switch tt.key {
			case "esc":
				keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
			default:
				keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			}

			// This will fail until we implement proper modal dialog handling
			newModel, cmd := model.Update(keyMsg)

			updatedModel := newModel.(Model)
			assert.Equal(t, tt.expectedState, updatedModel.showConfirmationDialog,
				"Modal dialog state should match expected")

			// Verify appropriate command is returned based on expectedAction
			switch tt.expectedAction {
			case "show_dialog":
				assert.Nil(t, cmd, "Show dialog should not return command")
			case "execute_cleanup":
				assert.NotNil(t, cmd, "Execute cleanup should return command")
			case "cancel_cleanup":
				assert.Nil(t, cmd, "Cancel cleanup should not return command")
			}
		})
	}
}

func TestModalDialogRendering(t *testing.T) {
	t.Run("modal_dialog_appearance", func(t *testing.T) {
		model := setupTestModel()
		model.showConfirmationDialog = true
		model.confirmationMessage = "Clean 2 stale sessions?\nIssue #123, Issue #124\n\n(y/n) Press y to confirm, n to cancel"
		model.width = 80
		model.height = 24

		// This will fail until we implement modal dialog rendering
		view := model.View()

		// Verify modal content is present
		assert.Contains(t, view, "Clean 2 stale sessions?", "Should show confirmation message")
		assert.Contains(t, view, "Issue #123", "Should show issue details")
		assert.Contains(t, view, "Issue #124", "Should show issue details")

		// Verify modal styling indicators
		assert.Contains(t, view, "y/n", "Should show confirmation options")
	})

	t.Run("modal_dialog_not_shown_when_disabled", func(t *testing.T) {
		model := setupTestModel()
		model.showConfirmationDialog = false
		model.confirmationMessage = "This should not appear"
		model.width = 80
		model.height = 24

		view := model.View()

		// Verify modal content is not present
		assert.NotContains(t, view, "This should not appear", "Should not show modal when disabled")
		assert.NotContains(t, view, "y/n", "Should not show confirmation options when disabled")
	})

	t.Run("modal_dialog_overlay_styling", func(t *testing.T) {
		model := setupTestModel()
		model.showConfirmationDialog = true
		model.confirmationMessage = "Test confirmation"
		model.width = 80
		model.height = 24

		view := model.View()

		// This will fail until we implement proper modal styling
		// Check that modal appears to be an overlay (exact styling depends on implementation)
		assert.Contains(t, view, "Test confirmation", "Should render modal content")

		// The view should still contain base content but with modal overlay
		lines := strings.Split(view, "\n")
		assert.True(t, len(lines) > 5, "View should have multiple lines with modal overlay")
	})
}

func TestModalDialogContent(t *testing.T) {
	t.Run("confirmation_message_shows_stale_session_count", func(t *testing.T) {
		model := setupTestModel()
		// Set up sessions that will be detected as stale (tmux sessions don't exist)
		model.sessions = []config.SessionMetadata{
			{IssueNumber: 123, IssueTitle: "Test issue 123", TmuxSession: "work-issue-123"},
			{IssueNumber: 124, IssueTitle: "Test issue 124", TmuxSession: "work-issue-124"},
		}

		// This will fail until we implement showCleanConfirmation
		updatedModel := model.showCleanConfirmation()

		assert.True(t, updatedModel.showConfirmationDialog, "Should show dialog")
		assert.Contains(t, updatedModel.confirmationMessage, "2 stale sessions", "Should show count")
		assert.Contains(t, updatedModel.confirmationMessage, "Issue #123", "Should list first issue")
		assert.Contains(t, updatedModel.confirmationMessage, "Issue #124", "Should list second issue")
	})

	t.Run("confirmation_message_shows_single_session", func(t *testing.T) {
		model := setupTestModel()
		// Set up single session that will be detected as stale
		model.sessions = []config.SessionMetadata{
			{IssueNumber: 123, IssueTitle: "Single test issue", TmuxSession: "work-issue-123"},
		}

		// This will fail until we implement showCleanConfirmation
		updatedModel := model.showCleanConfirmation()

		assert.True(t, updatedModel.showConfirmationDialog, "Should show dialog")
		assert.Contains(t, updatedModel.confirmationMessage, "1 stale session", "Should show singular count")
		assert.Contains(t, updatedModel.confirmationMessage, "Issue #123", "Should list issue")
	})

	t.Run("confirmation_message_handles_empty_list", func(t *testing.T) {
		model := setupTestModel()
		// No sessions - empty list
		model.sessions = []config.SessionMetadata{}

		// This will fail until we implement showCleanConfirmation
		updatedModel := model.showCleanConfirmation()

		// Should not show dialog for empty list
		assert.False(t, updatedModel.showConfirmationDialog, "Should not show dialog for empty list")
		assert.Empty(t, updatedModel.confirmationMessage, "Should not set message for empty list")
	})
}

func TestModalDialogKeyBindingPriority(t *testing.T) {
	t.Run("modal_dialog_keys_override_normal_keys", func(t *testing.T) {
		model := setupTestModel()
		model.showConfirmationDialog = true
		model.sessions = []config.SessionMetadata{
			{IssueNumber: 123, TmuxSession: "work-issue-123"},
		}
		model.cursor = 0

		// When modal is shown, normal navigation should be disabled
		// This will fail until we implement proper key priority handling

		// Test that 'j' (down) doesn't move cursor when modal is shown
		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		updatedModel := newModel.(Model)
		assert.Equal(t, model.cursor, updatedModel.cursor, "Cursor should not move when modal is shown")

		// Test that 'k' (up) doesn't move cursor when modal is shown
		newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		updatedModel = newModel.(Model)
		assert.Equal(t, model.cursor, updatedModel.cursor, "Cursor should not move when modal is shown")

		// Test that 'enter' confirms cleanup when modal is shown (doesn't attach)
		newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		updatedModel = newModel.(Model)
		assert.False(t, updatedModel.showConfirmationDialog, "Enter should close modal")
		assert.NotNil(t, cmd, "Enter should trigger cleanup command, not attach")
	})

	t.Run("normal_keys_work_when_modal_hidden", func(t *testing.T) {
		model := setupTestModel()
		model.showConfirmationDialog = false
		model.sessions = []config.SessionMetadata{
			{IssueNumber: 123, TmuxSession: "work-issue-123"},
			{IssueNumber: 124, TmuxSession: "work-issue-124"},
		}
		model.cursor = 0

		// Normal keys should work when modal is hidden
		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		updatedModel := newModel.(Model)
		assert.Equal(t, 1, updatedModel.cursor, "Cursor should move down when modal is hidden")
	})
}

func TestModalDialogAccessibility(t *testing.T) {
	t.Run("modal_dialog_escape_handling", func(t *testing.T) {
		model := setupTestModel()
		model.showConfirmationDialog = true

		// Test that Escape key closes modal
		newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
		updatedModel := newModel.(Model)

		assert.False(t, updatedModel.showConfirmationDialog, "Escape should close modal")
		assert.Nil(t, cmd, "Escape should not execute cleanup")
	})

	t.Run("modal_dialog_enter_handling", func(t *testing.T) {
		model := setupTestModel()
		model.showConfirmationDialog = true
		model.pendingCleanSessions = []config.SessionMetadata{
			{IssueNumber: 123, TmuxSession: "work-issue-123"},
		}

		// Test that Enter key confirms cleanup (same as 'y')
		newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		updatedModel := newModel.(Model)

		assert.False(t, updatedModel.showConfirmationDialog, "Enter should close modal")
		assert.NotNil(t, cmd, "Enter should execute cleanup")
	})
}

func TestModalDialogViewModeConsistency(t *testing.T) {
	viewModes := []ViewMode{ViewModeRepository, ViewModeGlobal}

	for _, mode := range viewModes {
		t.Run(string(rune(mode))+"_view_mode_modal_consistency", func(t *testing.T) {
			model := setupTestModel()
			model.viewMode = mode
			model.showConfirmationDialog = true
			model.confirmationMessage = "Test modal in view mode\n\n(y/n) Press y to confirm, n to cancel"

			// Modal should render consistently across view modes
			view := model.View()

			assert.Contains(t, view, "Test modal in view mode", "Modal should appear in all view modes")
			assert.Contains(t, view, "y/n", "Modal should show options in all view modes")
		})
	}
}
