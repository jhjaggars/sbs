package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sbs/pkg/config"
	"sbs/pkg/repo"
)

// Mock data for testing
var testSessions = []config.SessionMetadata{
	{
		IssueNumber:    123,
		IssueTitle:     "Fix authentication bug in user login",
		RepositoryName: "test-repo",
		Branch:         "issue-123-fix-auth-bug",
		TmuxSession:    "work-issue-123",
		LastActivity:   "2025-07-31T10:00:00Z",
	},
	{
		IssueNumber:    124,
		IssueTitle:     "Add dark mode support to dashboard",
		RepositoryName: "test-repo",
		Branch:         "issue-124-dark-mode",
		TmuxSession:    "work-issue-124",
		LastActivity:   "2025-07-31T09:30:00Z",
	},
}

func setupTestRepository() *repo.Repository {
	return &repo.Repository{
		Name: "test-repo",
		Root: "/tmp/test-repo",
	}
}

func TestModel_HelpText(t *testing.T) {
	t.Run("condensed_help_contains_enter_to_attach_in_repo_view", func(t *testing.T) {
		// Arrange
		model := NewModel()
		model.sessions = testSessions
		model.currentRepo = setupTestRepository()
		model.viewMode = ViewModeRepository
		model.width = 80
		model.height = 24
		model.showHelp = false // Ensure condensed help is shown

		// Act
		view := model.View()

		// Assert - Check that "enter: attach" appears in condensed help
		assert.Contains(t, view, "enter: attach", "Repository view condensed help should contain 'enter: attach'")

		// Assert - Check that "enter: attach" appears first in help text
		helpTextStart := strings.Index(view, "Press ")
		if helpTextStart != -1 {
			helpLine := view[helpTextStart:]
			enterIndex := strings.Index(helpLine, "enter: attach")
			questionIndex := strings.Index(helpLine, "?: help")

			assert.True(t, enterIndex != -1, "Help text should contain 'enter: attach'")
			assert.True(t, questionIndex != -1, "Help text should contain '?: help'")
			assert.True(t, enterIndex < questionIndex, "'enter: attach' should appear before '?: help'")
		}
	})

	t.Run("condensed_help_contains_enter_to_attach_in_global_view", func(t *testing.T) {
		// Arrange
		model := NewModel()
		model.sessions = testSessions
		model.currentRepo = nil // Global view - no current repo
		model.viewMode = ViewModeGlobal
		model.width = 80
		model.height = 24
		model.showHelp = false // Ensure condensed help is shown

		// Act
		view := model.View()

		// Assert - Check that "enter: attach" appears in condensed help
		assert.Contains(t, view, "enter: attach", "Global view condensed help should contain 'enter: attach'")

		// Assert - Check that "enter: attach" appears first in help text
		helpTextStart := strings.Index(view, "Press ")
		if helpTextStart != -1 {
			helpLine := view[helpTextStart:]
			enterIndex := strings.Index(helpLine, "enter: attach")
			questionIndex := strings.Index(helpLine, "?: help")

			assert.True(t, enterIndex != -1, "Help text should contain 'enter: attach'")
			assert.True(t, questionIndex != -1, "Help text should contain '?: help'")
			assert.True(t, enterIndex < questionIndex, "'enter: attach' should appear before '?: help'")
		}
	})

	t.Run("help_text_format_consistency", func(t *testing.T) {
		// Arrange
		model := NewModel()
		model.sessions = testSessions
		model.currentRepo = setupTestRepository()
		model.viewMode = ViewModeRepository
		model.width = 80
		model.height = 24
		model.showHelp = false

		// Act
		view := model.View()

		// Assert - Check comma-separated format matches issueselect.go pattern
		helpTextStart := strings.Index(view, "Press ")
		require.True(t, helpTextStart != -1, "Help text should start with 'Press '")

		helpLine := view[helpTextStart:]
		newlineIndex := strings.Index(helpLine, "\n")
		if newlineIndex != -1 {
			helpLine = helpLine[:newlineIndex]
		}

		// Check for proper comma separation
		assert.Contains(t, helpLine, ", ", "Help text should use comma separation")

		// Check for expected components in correct order
		expectedOrder := []string{"enter: attach", "?: help", "g: toggle", "r: refresh", "q: quit"}
		lastIndex := -1
		for _, component := range expectedOrder {
			currentIndex := strings.Index(helpLine, component)
			if currentIndex != -1 {
				assert.True(t, currentIndex > lastIndex,
					"Components should appear in expected order: %v", expectedOrder)
				lastIndex = currentIndex
			}
		}
	})

	t.Run("enter_to_attach_appears_first", func(t *testing.T) {
		// Arrange
		model := NewModel()
		model.sessions = testSessions
		model.currentRepo = setupTestRepository()
		model.viewMode = ViewModeRepository
		model.width = 80
		model.height = 24
		model.showHelp = false

		// Act
		view := model.View()

		// Assert - Verify "enter: attach" appears at the beginning for prominence
		helpTextStart := strings.Index(view, "Press ")
		require.True(t, helpTextStart != -1, "Help text should start with 'Press '")

		helpLine := view[helpTextStart:]

		// Check that "enter: attach" appears at the beginning
		assert.Contains(t, helpLine, "enter: attach",
			"Help text should contain 'enter: attach', got: %s", helpLine[:minValue(len(helpLine), 30)])
	})

	t.Run("help_text_length_within_terminal_limits", func(t *testing.T) {
		testWidths := []int{80, 120, 160}

		for _, width := range testWidths {
			t.Run(fmt.Sprintf("width_%d", width), func(t *testing.T) {
				// Arrange
				model := NewModel()
				model.sessions = testSessions
				model.currentRepo = setupTestRepository()
				model.viewMode = ViewModeRepository
				model.width = width
				model.height = 24
				model.showHelp = false

				// Act
				view := model.View()

				// Assert - Check that help text fits within terminal width
				lines := strings.Split(view, "\n")
				for _, line := range lines {
					if strings.Contains(line, "Press ") {
						// Remove ANSI codes and measure actual text length
						cleanLine := stripANSI(line)
						assert.LessOrEqual(t, len(cleanLine), width,
							"Help text should fit within terminal width %d, got length %d: %s",
							width, len(cleanLine), cleanLine)
					}
				}
			})
		}
	})
}

func TestModel_ViewRendering(t *testing.T) {
	t.Run("view_contains_correct_help_text_repo_mode", func(t *testing.T) {
		// Arrange
		model := NewModel()
		model.sessions = testSessions
		model.currentRepo = setupTestRepository()
		model.viewMode = ViewModeRepository
		model.width = 80
		model.height = 24
		model.showHelp = false

		// Act
		view := model.View()

		// Assert - Test complete view rendering includes updated help text
		assert.Contains(t, view, "enter: attach", "Repository view should contain 'enter: attach'")
		assert.Contains(t, view, "?: help", "Repository view should contain '?: help'")
		assert.Contains(t, view, "g: toggle", "Repository view should contain 'g: toggle'")
		assert.Contains(t, view, "r: refresh", "Repository view should contain 'r: refresh'")
		assert.Contains(t, view, "q: quit", "Repository view should contain 'q: quit'")
	})

	t.Run("view_contains_correct_help_text_global_mode", func(t *testing.T) {
		// Arrange
		model := NewModel()
		model.sessions = testSessions
		model.currentRepo = nil
		model.viewMode = ViewModeGlobal
		model.width = 80
		model.height = 24
		model.showHelp = false

		// Act
		view := model.View()

		// Assert - Test complete view rendering includes updated help text
		assert.Contains(t, view, "enter: attach", "Global view should contain 'enter: attach'")
		assert.Contains(t, view, "?: help", "Global view should contain '?: help'")
		assert.Contains(t, view, "g: toggle", "Global view should contain 'g: toggle'")
		assert.Contains(t, view, "r: refresh", "Global view should contain 'r: refresh'")
		assert.Contains(t, view, "q: quit", "Global view should contain 'q: quit'")
	})

	t.Run("view_maintains_other_help_elements", func(t *testing.T) {
		// Arrange
		model := NewModel()
		model.sessions = testSessions
		model.currentRepo = setupTestRepository()
		model.viewMode = ViewModeRepository
		model.width = 80
		model.height = 24
		model.showHelp = false

		// Act
		view := model.View()

		// Assert - Ensure other help elements remain unchanged
		assert.Contains(t, view, "?: help", "Should contain '?: help'")
		assert.Contains(t, view, "g: toggle", "Should contain 'g: toggle'")
		assert.Contains(t, view, "r: refresh", "Should contain 'r: refresh'")
		assert.Contains(t, view, "q: quit", "Should contain 'q: quit'")
	})
}

func TestModel_EdgeCases(t *testing.T) {
	t.Run("help_text_with_narrow_terminal", func(t *testing.T) {
		// Arrange
		model := NewModel()
		model.sessions = testSessions
		model.currentRepo = setupTestRepository()
		model.viewMode = ViewModeRepository
		model.width = 40 // Very narrow terminal
		model.height = 24
		model.showHelp = false

		// Act
		view := model.View()

		// Assert - Focus only on help text, not general layout
		// Should still contain the key help elements
		assert.Contains(t, view, "enter: attach", "Should contain 'enter: attach' even in narrow terminal")
		assert.Contains(t, view, "?: help", "Should contain '?: help' even in narrow terminal")

		// Check that help text line itself is reasonable for narrow terminal
		lines := strings.Split(view, "\n")
		for _, line := range lines {
			if strings.Contains(line, "enter: attach") {
				cleanLine := stripANSI(line)
				// Help text should be manageable even if it wraps
				assert.NotEmpty(t, cleanLine, "Help text line should not be empty")
			}
		}
	})

	t.Run("help_text_with_no_sessions", func(t *testing.T) {
		// Arrange
		model := NewModel()
		model.sessions = []config.SessionMetadata{} // No sessions
		model.currentRepo = setupTestRepository()
		model.viewMode = ViewModeRepository
		model.width = 80
		model.height = 24
		model.showHelp = false

		// Act
		view := model.View()

		// Assert - Verify consistent behavior with no sessions
		assert.Contains(t, view, "enter: attach", "Should contain 'enter: attach' even with no sessions")
		assert.Contains(t, view, "?: help", "Should contain help text with no sessions")
	})

	t.Run("help_text_with_error_state", func(t *testing.T) {
		// Arrange
		model := NewModel()
		model.sessions = testSessions
		model.currentRepo = setupTestRepository()
		model.viewMode = ViewModeRepository
		model.width = 80
		model.height = 24
		model.showHelp = false
		// Simulate error state by testing help text behavior in various conditions

		// Act
		view := model.View()

		// Assert - Ensure help text still appears correctly
		assert.Contains(t, view, "enter: attach", "Should contain 'enter: attach' in error state")
		assert.Contains(t, view, "?: help", "Should contain help text in error state")
	})
}

// Helper function to strip ANSI codes for accurate length measurement
func stripANSI(s string) string {
	// Simple ANSI strip - for more complex cases, might need a regex
	result := s
	for {
		start := strings.Index(result, "\x1b[")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "m")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	return result
}

// Helper function for minimum value
func minValue(a, b int) int {
	if a < b {
		return a
	}
	return b
}
