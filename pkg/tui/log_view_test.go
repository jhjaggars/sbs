package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"sbs/pkg/config"
)

func TestModel_LogViewKeyBinding(t *testing.T) {
	t.Run("l_key_binding_triggers_log_view", func(t *testing.T) {
		// Arrange
		model := NewModel()
		model.sessions = testSessions
		model.cursor = 0
		model.viewMode = ViewModeRepository

		// Act - simulate 'l' key press
		keyMsg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'l'},
		}
		newModel, cmd := model.Update(keyMsg)

		// Assert - should switch to log view mode
		assert.Equal(t, ViewModeLog, newModel.(Model).viewMode, "Pressing 'l' should switch to log view mode")
		assert.NotNil(t, cmd, "Pressing 'l' should return a command to execute loghook script")
	})

	t.Run("l_key_binding_help_text", func(t *testing.T) {
		// Arrange
		model := NewModel()
		model.sessions = testSessions
		model.width = 80
		model.height = 24
		model.showHelp = false

		// Act
		view := model.View()

		// Assert - help text should include 'l: logs' option
		assert.Contains(t, view, "l: logs", "Help text should include 'l: logs' option")
	})

	t.Run("l_key_binding_disabled_when_no_sessions", func(t *testing.T) {
		// Arrange
		model := NewModel()
		model.sessions = []config.SessionMetadata{} // No sessions
		model.viewMode = ViewModeRepository

		// Act - simulate 'l' key press
		keyMsg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'l'},
		}
		newModel, cmd := model.Update(keyMsg)

		// Assert - should not switch to log view when no sessions
		assert.Equal(t, ViewModeRepository, newModel.(Model).viewMode, "Should not switch to log view when no sessions")
		assert.Nil(t, cmd, "Should not return a command when no sessions")
	})

	t.Run("l_key_binding_disabled_when_no_current_session", func(t *testing.T) {
		// Arrange
		model := NewModel()
		model.sessions = testSessions
		model.cursor = -1 // No valid selection
		model.viewMode = ViewModeRepository

		// Act - simulate 'l' key press
		keyMsg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'l'},
		}
		newModel, cmd := model.Update(keyMsg)

		// Assert - should not switch to log view when no valid selection
		assert.Equal(t, ViewModeRepository, newModel.(Model).viewMode, "Should not switch to log view when no valid selection")
		assert.Nil(t, cmd, "Should not return a command when no valid selection")
	})

	t.Run("l_key_has_proper_binding_definition", func(t *testing.T) {
		// Test that the 'l' key binding is properly defined in the keyMap
		// This test checks that the LogView key binding exists and has correct help text

		// Note: This test will fail until we add the LogView key binding to the keyMap
		logBinding := keys.LogView
		assert.True(t, key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}, logBinding),
			"LogView key binding should match 'l' key")
		assert.Contains(t, logBinding.Help().Key, "l", "LogView key binding help should contain 'l'")
		assert.Contains(t, logBinding.Help().Desc, "logs", "LogView key binding help should contain 'logs'")
	})
}

func TestLogView_StateManagement(t *testing.T) {
	t.Run("log_view_creation", func(t *testing.T) {
		// Test LogView struct initialization
		model := NewModel()
		model.sessions = testSessions
		model.cursor = 0

		// Simulate pressing 'l' key to enter log view
		keyMsg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'l'},
		}
		newModel, _ := model.Update(keyMsg)
		updatedModel := newModel.(Model)

		// Assert that log view was created and initialized
		assert.Equal(t, ViewModeLog, updatedModel.viewMode, "Should be in log view mode")
		assert.NotNil(t, updatedModel.logView, "LogView should be initialized when entering log view mode")
		assert.Equal(t, 0, updatedModel.logView.scrollOffset, "LogView should start with scroll offset 0")
		assert.Empty(t, updatedModel.logView.content, "LogView should start with empty content")
		assert.True(t, updatedModel.logView.loading, "LogView should be loading initially")
	})

	t.Run("log_view_mode_toggle", func(t *testing.T) {
		// Test switching between list view and log view
		model := NewModel()
		model.sessions = testSessions
		model.cursor = 0
		originalViewMode := model.viewMode

		// Switch to log view
		model.viewMode = ViewModeLog
		assert.Equal(t, ViewModeLog, model.viewMode, "Should be in log view mode")

		// Switch back
		model.viewMode = originalViewMode
		assert.Equal(t, originalViewMode, model.viewMode, "Should return to original view mode")
	})

	t.Run("log_view_exit_on_escape", func(t *testing.T) {
		// Test ESC key exits log view
		model := NewModel()
		model.sessions = testSessions
		model.cursor = 0

		// Enter log view first
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
		updatedModel, _ := model.Update(keyMsg)
		model = updatedModel.(Model)

		// Verify we're in log view
		assert.Equal(t, ViewModeLog, model.viewMode, "Should be in log view")

		// Act - simulate ESC key press
		keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
		newModel, _ := model.Update(keyMsg)

		// Assert - should return to previous view state
		assert.Equal(t, ViewModeRepository, newModel.(Model).viewMode, "ESC should return to previous view mode")
	})

	t.Run("log_view_exit_on_q", func(t *testing.T) {
		// Test 'q' key exits log view
		model := NewModel()
		model.sessions = testSessions
		model.cursor = 0

		// Enter log view first
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
		updatedModel, _ := model.Update(keyMsg)
		model = updatedModel.(Model)

		// Verify we're in log view
		assert.Equal(t, ViewModeLog, model.viewMode, "Should be in log view")

		// Act - simulate 'q' key press
		keyMsg = tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'q'},
		}
		newModel, _ := model.Update(keyMsg)

		// Assert - should return to list view
		assert.Equal(t, ViewModeRepository, newModel.(Model).viewMode, "'q' should return to previous view mode")
	})

	t.Run("log_view_scroll_functionality", func(t *testing.T) {
		// Test up/down arrow scrolling
		model := NewModel()
		model.sessions = testSessions
		model.cursor = 0

		// Enter log view first
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
		updatedModel, _ := model.Update(keyMsg)
		model = updatedModel.(Model)

		// Set up log view with content for scrolling (more lines than can fit on screen)
		content := "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\nline11\nline12"
		model.logView.content = content
		model.logView.loading = false
		model.height = 8 // Small height to force scrolling

		// Test down scroll
		keyMsg = tea.KeyMsg{Type: tea.KeyDown}
		newModel, _ := model.Update(keyMsg)
		assert.Equal(t, 1, newModel.(Model).logView.scrollOffset, "Down arrow should increase scroll offset")

		// Test up scroll
		keyMsg = tea.KeyMsg{Type: tea.KeyUp}
		newModel, _ = newModel.Update(keyMsg)
		assert.Equal(t, 0, newModel.(Model).logView.scrollOffset, "Up arrow should decrease scroll offset")

		// Test scroll bounds - shouldn't go negative
		keyMsg = tea.KeyMsg{Type: tea.KeyUp}
		newModel, _ = newModel.Update(keyMsg)
		assert.Equal(t, 0, newModel.(Model).logView.scrollOffset, "Scroll offset should not go negative")
	})
}
