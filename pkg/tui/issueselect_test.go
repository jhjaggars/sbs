package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sbs/pkg/issue"
)

// Mock GitHub client for TUI testing
type mockGitHubClient struct {
	issues []issue.Issue
	err    error
	calls  []string // Track method calls for verification
}

func (m *mockGitHubClient) ListIssues(searchQuery string, limit int) ([]issue.Issue, error) {
	m.calls = append(m.calls, "ListIssues")
	if m.err != nil {
		return nil, m.err
	}
	
	// Filter issues based on search query for testing
	if searchQuery == "" {
		return m.issues, nil
	}
	
	var filtered []issue.Issue
	for _, issue := range m.issues {
		if strings.Contains(strings.ToLower(issue.Title), strings.ToLower(searchQuery)) {
			filtered = append(filtered, issue)
		}
	}
	return filtered, nil
}

// Test data
var testIssues = []issue.Issue{
	{Number: 123, Title: "Fix authentication bug", State: "open", URL: "https://github.com/owner/repo/issues/123"},
	{Number: 124, Title: "Add dark mode support", State: "open", URL: "https://github.com/owner/repo/issues/124"},
	{Number: 125, Title: "Refactor database connection", State: "open", URL: "https://github.com/owner/repo/issues/125"},
}

func TestIssueSelectModel(t *testing.T) {
	t.Run("initialization", func(t *testing.T) {
		// Arrange
		mockClient := &mockGitHubClient{issues: testIssues}
		
		// Act
		model := NewIssueSelectModel(mockClient)
		
		// Assert
		assert.NotNil(t, model)
		assert.Equal(t, 0, model.cursor)
		assert.Equal(t, "", model.searchInput.Value())
		assert.False(t, model.showHelp)
		assert.Len(t, model.issues, 0) // No issues loaded initially
		assert.Equal(t, stateLoading, model.state)
	})
	
	t.Run("loading_issues", func(t *testing.T) {
		// Arrange
		mockClient := &mockGitHubClient{issues: testIssues}
		model := NewIssueSelectModel(mockClient)
		
		// Act - simulate Init() command execution
		cmd := model.Init()
		require.NotNil(t, cmd)
		
		// Simulate the load message
		loadMsg := issuesLoadedMsg{issues: testIssues, err: nil}
		newModel, _ := model.Update(loadMsg)
		issueModel := newModel.(*IssueSelectModel)
		
		// Assert
		assert.Equal(t, stateReady, issueModel.state)
		assert.Len(t, issueModel.issues, 3)
		assert.Equal(t, "Fix authentication bug", issueModel.issues[0].Title)
		// Note: mockClient.calls tracking happens when the actual command executes, not in the test
	})
	
	t.Run("keyboard_navigation", func(t *testing.T) {
		// Arrange
		mockClient := &mockGitHubClient{issues: testIssues}
		model := NewIssueSelectModel(mockClient)
		model.issues = testIssues
		model.filteredIssues = testIssues // Also set filtered issues
		model.state = stateReady
		
		// Test down arrow
		downMsg := tea.KeyMsg{Type: tea.KeyDown}
		newModel, _ := model.Update(downMsg)
		issueModel := newModel.(*IssueSelectModel)
		assert.Equal(t, 1, issueModel.cursor)
		
		// Test up arrow (should stay at 1 since we moved down once)
		upMsg := tea.KeyMsg{Type: tea.KeyUp}
		newModel, _ = issueModel.Update(upMsg)
		issueModel = newModel.(*IssueSelectModel)
		assert.Equal(t, 0, issueModel.cursor)
		
		// Test 'j' key (vim-style down)
		jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
		newModel, _ = issueModel.Update(jMsg)
		issueModel = newModel.(*IssueSelectModel)
		assert.Equal(t, 1, issueModel.cursor)
		
		// Test 'k' key (vim-style up)
		kMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
		newModel, _ = issueModel.Update(kMsg)
		issueModel = newModel.(*IssueSelectModel)
		assert.Equal(t, 0, issueModel.cursor)
		
		// Test boundary - can't go above 0
		upMsg = tea.KeyMsg{Type: tea.KeyUp}
		newModel, _ = issueModel.Update(upMsg)
		issueModel = newModel.(*IssueSelectModel)
		assert.Equal(t, 0, issueModel.cursor)
		
		// Test boundary - can't go below last item
		issueModel.cursor = len(testIssues) - 1
		downMsg = tea.KeyMsg{Type: tea.KeyDown}
		newModel, _ = issueModel.Update(downMsg)
		issueModel = newModel.(*IssueSelectModel)
		assert.Equal(t, len(testIssues)-1, issueModel.cursor)
	})
	
	t.Run("issue_selection", func(t *testing.T) {
		// Arrange
		mockClient := &mockGitHubClient{issues: testIssues}
		model := NewIssueSelectModel(mockClient)
		model.issues = testIssues
		model.filteredIssues = testIssues // Also set filtered issues
		model.state = stateReady
		model.cursor = 1 // Select second issue
		
		// Act - press Enter
		enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
		newModel, cmd := model.Update(enterMsg)
		issueModel := newModel.(*IssueSelectModel)
		
		// Assert
		assert.Equal(t, stateSelected, issueModel.state)
		assert.NotNil(t, issueModel.selectedIssue)
		assert.Equal(t, 124, issueModel.selectedIssue.Number)
		assert.NotNil(t, cmd) // Should return quit command
	})
	
	t.Run("issue_selection_with_empty_list", func(t *testing.T) {
		// Arrange
		mockClient := &mockGitHubClient{issues: []issue.Issue{}}
		model := NewIssueSelectModel(mockClient)
		model.issues = []issue.Issue{}
		model.filteredIssues = []issue.Issue{} // Also set filtered issues
		model.state = stateReady
		
		// Act - press Enter
		enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
		newModel, cmd := model.Update(enterMsg)
		issueModel := newModel.(*IssueSelectModel)
		
		// Assert
		assert.Equal(t, stateReady, issueModel.state) // Should remain in ready state
		assert.Nil(t, cmd) // Should not quit
	})
	
	t.Run("search_functionality", func(t *testing.T) {
		// Arrange
		mockClient := &mockGitHubClient{issues: testIssues}
		model := NewIssueSelectModel(mockClient)
		model.issues = testIssues
		model.state = stateReady
		
		// Focus on search input first
		tabMsg := tea.KeyMsg{Type: tea.KeyTab}
		newModel, _ := model.Update(tabMsg)
		issueModel := newModel.(*IssueSelectModel)
		assert.True(t, issueModel.searchFocused)
		
		// Type in search box
		typeMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a', 'u', 't', 'h'}}
		newModel, _ = issueModel.Update(typeMsg)
		issueModel = newModel.(*IssueSelectModel)
		
		// Verify search input received the text
		searchValue := issueModel.searchInput.Value()
		assert.Contains(t, searchValue, "auth")
	})
	
	t.Run("quit_behavior", func(t *testing.T) {
		// Arrange
		mockClient := &mockGitHubClient{issues: testIssues}
		model := NewIssueSelectModel(mockClient)
		model.state = stateReady // Set to ready state so it can handle key presses
		
		// Test 'q' key quit
		qMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
		newModel, cmd := model.Update(qMsg)
		issueModel := newModel.(*IssueSelectModel)
		assert.Equal(t, stateQuit, issueModel.state)
		assert.NotNil(t, cmd) // Should return quit command
		
		// Reset model for next test
		model = NewIssueSelectModel(mockClient)
		model.state = stateReady // Set to ready state so it can handle key presses
		
		// Test Ctrl+C quit
		ctrlCMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
		newModel, cmd = model.Update(ctrlCMsg)
		issueModel = newModel.(*IssueSelectModel)
		assert.Equal(t, stateQuit, issueModel.state)
		assert.NotNil(t, cmd) // Should return quit command
	})
	
	t.Run("error_display", func(t *testing.T) {
		// Arrange
		mockClient := &mockGitHubClient{err: assert.AnError}
		model := NewIssueSelectModel(mockClient)
		
		// Act - simulate error message
		errorMsg := issuesLoadedMsg{issues: nil, err: assert.AnError}
		newModel, _ := model.Update(errorMsg)
		issueModel := newModel.(*IssueSelectModel)
		
		// Assert
		assert.Equal(t, stateError, issueModel.state)
		assert.Equal(t, assert.AnError, issueModel.err)
	})
	
	t.Run("window_resize_handling", func(t *testing.T) {
		// Arrange
		mockClient := &mockGitHubClient{issues: testIssues}
		model := NewIssueSelectModel(mockClient)
		
		// Act
		resizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
		newModel, _ := model.Update(resizeMsg)
		issueModel := newModel.(*IssueSelectModel)
		
		// Assert
		assert.Equal(t, 120, issueModel.width)
		assert.Equal(t, 40, issueModel.height)
	})
}

// Helper to create test model with mock data
func createTestModel(issues []issue.Issue, err error) *IssueSelectModel {
	mockClient := &mockGitHubClient{issues: issues, err: err}
	return NewIssueSelectModel(mockClient)
}

// Test utilities for simulating key presses
func TestKeyPressSimulation(t *testing.T) {
	t.Run("simulate_various_keys", func(t *testing.T) {
		mockClient := &mockGitHubClient{issues: testIssues}
		model := NewIssueSelectModel(mockClient)
		model.issues = testIssues
		model.filteredIssues = testIssues // Set filtered issues for navigation
		model.state = stateReady
		
		// Test that our key simulation approach works
		downKey := tea.KeyMsg{Type: tea.KeyDown}
		newModel, _ := model.Update(downKey)
		issueModel := newModel.(*IssueSelectModel)
		assert.Equal(t, 1, issueModel.cursor)
		
		upKey := tea.KeyMsg{Type: tea.KeyUp}
		newModel, _ = issueModel.Update(upKey)
		issueModel = newModel.(*IssueSelectModel)
		assert.Equal(t, 0, issueModel.cursor)
	})
}

// Test view rendering (basic functionality)
func TestIssueSelectModel_View(t *testing.T) {
	t.Run("loading_state_view", func(t *testing.T) {
		// Arrange
		mockClient := &mockGitHubClient{issues: testIssues}
		model := NewIssueSelectModel(mockClient)
		model.state = stateLoading
		
		// Act
		view := model.View()
		
		// Assert
		assert.Contains(t, view, "Loading")
		assert.Contains(t, view, "issues")
	})
	
	t.Run("ready_state_view", func(t *testing.T) {
		// Arrange
		mockClient := &mockGitHubClient{issues: testIssues}
		model := NewIssueSelectModel(mockClient)
		model.issues = testIssues
		model.filteredIssues = testIssues // Set filtered issues for display
		model.state = stateReady
		model.width = 80
		model.height = 24
		
		// Act
		view := model.View()
		
		// Assert
		assert.Contains(t, view, "Select an Issue")
		assert.Contains(t, view, "Fix authentication bug")
		assert.Contains(t, view, "Add dark mode support")
		assert.Contains(t, view, "Search:")
	})
	
	t.Run("error_state_view", func(t *testing.T) {
		// Arrange
		mockClient := &mockGitHubClient{err: assert.AnError}
		model := NewIssueSelectModel(mockClient)
		model.state = stateError
		model.err = assert.AnError
		
		// Act
		view := model.View()
		
		// Assert
		assert.Contains(t, view, "Error")
		assert.Contains(t, view, "Press q to quit")
	})
	
	t.Run("no_issues_view", func(t *testing.T) {
		// Arrange
		mockClient := &mockGitHubClient{issues: []issue.Issue{}}
		model := NewIssueSelectModel(mockClient)
		model.issues = []issue.Issue{}
		model.state = stateReady
		
		// Act
		view := model.View()
		
		// Assert
		assert.Contains(t, view, "No open issues found")
	})
}