package tui

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"sbs/pkg/config"
)

func TestLogView_AutoRefresh(t *testing.T) {
	t.Run("auto_refresh_enabled_when_log_view_active", func(t *testing.T) {
		// Test that auto-refresh starts when entering log view
		model := NewModel()
		model.sessions = testSessions
		model.cursor = 0
		model.viewMode = ViewModeLog

		// Simulate entering log view
		cmd := model.startLogAutoRefresh()

		// Verify refresh interval configuration
		assert.NotNil(t, cmd, "Should return auto-refresh command when entering log view")

		// Test that the command is a tick command (will be implemented)
		// This test will fail until we implement startLogAutoRefresh
	})

	t.Run("auto_refresh_disabled_when_log_view_inactive", func(t *testing.T) {
		// Test that auto-refresh stops when exiting log view
		model := NewModel()
		model.sessions = testSessions
		model.viewMode = ViewModeRepository // Not in log view
		model.logAutoRefreshActive = true   // Was previously active

		// Simulate exiting log view
		model.stopLogAutoRefresh()

		// Verify no background refresh activity
		assert.False(t, model.logAutoRefreshActive, "Auto-refresh should be disabled when not in log view")
	})

	t.Run("configurable_refresh_interval", func(t *testing.T) {
		// Test default 5-second interval
		model := NewModel()
		model.config = &config.Config{
			LogRefreshIntervalSecs: 0, // Use default
		}

		interval := model.getLogRefreshInterval()
		assert.Equal(t, 5*time.Second, interval, "Default log refresh interval should be 5 seconds")

		// Test custom intervals from configuration
		model.config.LogRefreshIntervalSecs = 10
		interval = model.getLogRefreshInterval()
		assert.Equal(t, 10*time.Second, interval, "Should use configured interval")

		// Test minimum/maximum interval bounds
		model.config.LogRefreshIntervalSecs = 1 // Below minimum
		interval = model.getLogRefreshInterval()
		assert.GreaterOrEqual(t, interval, 2*time.Second, "Should enforce minimum interval")

		model.config.LogRefreshIntervalSecs = 300 // Above maximum
		interval = model.getLogRefreshInterval()
		assert.LessOrEqual(t, interval, 120*time.Second, "Should enforce maximum interval")
	})

	t.Run("manual_refresh_with_r_key", func(t *testing.T) {
		// Test 'r' key triggers immediate refresh
		model := NewModel()
		model.sessions = testSessions
		model.cursor = 0

		// Enter log view first to initialize logView
		keyMsg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'l'},
		}
		updatedModel, _ := model.Update(keyMsg)
		model = updatedModel.(Model)

		// Simulate 'r' key press for manual refresh
		keyMsg = tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'r'},
		}
		newModel, cmd := model.Update(keyMsg)

		// Test refresh indication to user
		assert.NotNil(t, cmd, "'r' key should trigger refresh command")
		logView := newModel.(Model).logView
		assert.True(t, logView.refreshing, "Should indicate refresh in progress")
	})

	t.Run("refresh_error_handling", func(t *testing.T) {
		// Test behavior when refresh fails
		model := NewModel()
		model.sessions = testSessions
		model.cursor = 0
		model.viewMode = ViewModeLog
		model.logView = &LogView{}
		model.logAutoRefreshActive = true // Set auto-refresh as active

		// Simulate refresh error message
		errorMsg := logRefreshErrorMsg{
			err: assert.AnError,
		}
		newModel, _ := model.Update(errorMsg)

		// Verify error display to user
		logView := newModel.(Model).logView
		assert.NotEmpty(t, logView.errorMessage, "Should display error message")
		assert.Contains(t, logView.errorMessage, "refresh failed", "Error message should indicate refresh failure")

		// Test retry behavior - should continue trying to refresh
		assert.True(t, newModel.(Model).logAutoRefreshActive, "Should continue auto-refresh even after error")
	})

	t.Run("log_refresh_tick_message_handling", func(t *testing.T) {
		// Test handling of log refresh tick messages
		model := NewModel()
		model.sessions = testSessions
		model.cursor = 0
		model.viewMode = ViewModeLog
		model.logAutoRefreshActive = true

		// Simulate log refresh tick
		tickMsg := logRefreshTickMsg{}
		newModel, cmd := model.Update(tickMsg)

		// Should trigger refresh and schedule next tick
		assert.NotNil(t, cmd, "Log refresh tick should trigger refresh command")
		assert.Equal(t, ViewModeLog, newModel.(Model).viewMode, "Should remain in log view")
	})

	t.Run("log_refresh_result_message_handling", func(t *testing.T) {
		// Test handling of log refresh result messages
		model := NewModel()
		model.sessions = testSessions
		model.cursor = 0
		model.viewMode = ViewModeLog
		model.logView = &LogView{
			loading: true,
		}

		// Simulate successful refresh result
		resultMsg := logRefreshResultMsg{
			content: "Updated log content",
			err:     nil,
		}
		newModel, cmd := model.Update(resultMsg)

		// Should update content
		logView := newModel.(Model).logView
		assert.Equal(t, "Updated log content", logView.content, "Should update log content")
		assert.False(t, logView.loading, "Should clear loading state")
		assert.Nil(t, cmd, "Result message should not schedule next refresh - that's done by tick messages")
	})
}

// Message types are now defined in model.go

// Note: startLogAutoRefresh, stopLogAutoRefresh, and getLogRefreshInterval are now implemented in model.go
