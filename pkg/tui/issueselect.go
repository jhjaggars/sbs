package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"sbs/pkg/issue"
)

// GitHubClientInterface defines the interface for GitHub operations
type GitHubClientInterface interface {
	ListIssues(searchQuery string, limit int) ([]issue.Issue, error)
}

// IssueSelectModel represents the state of the issue selection TUI
type IssueSelectModel struct {
	// Dependencies
	githubClient GitHubClientInterface

	// State
	issues         []issue.Issue    // Current list of issues
	filteredIssues []issue.Issue    // Filtered issues based on search
	cursor         int              // Currently selected issue index
	searchInput    textinput.Model  // Search input field
	searchFocused  bool             // Whether search input is focused
	showHelp       bool             // Whether to show help text
	width          int              // Terminal width
	height         int              // Terminal height
	state          issueSelectState // Current UI state
	err            error            // Current error, if any
	selectedIssue  *issue.Issue     // Selected issue (when state is stateSelected)

	// Configuration
	issueLimit int // Maximum number of issues to fetch
}

// issueSelectState represents the different states of the UI
type issueSelectState int

const (
	stateLoading issueSelectState = iota
	stateReady
	stateSelected
	stateError
	stateQuit
)

// Message types for tea.Cmd communication
type issuesLoadedMsg struct {
	issues []issue.Issue
	err    error
}

type searchCompletedMsg struct {
	issues []issue.Issue
	query  string
	err    error
}

// NewIssueSelectModel creates a new issue selection model
func NewIssueSelectModel(githubClient GitHubClientInterface) *IssueSelectModel {
	// Initialize search input
	ti := textinput.New()
	ti.Placeholder = "Search issues..."
	ti.CharLimit = 100
	ti.Width = 50

	return &IssueSelectModel{
		githubClient:   githubClient,
		issues:         []issue.Issue{},
		filteredIssues: []issue.Issue{},
		cursor:         0,
		searchInput:    ti,
		searchFocused:  false,
		showHelp:       false,
		state:          stateLoading,
		issueLimit:     100, // Default limit
	}
}

// Init initializes the model and starts loading issues
func (m *IssueSelectModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadIssues(""), // Load all issues initially
		textinput.Blink,  // Start cursor blinking for search input
	)
}

// Update handles messages and updates the model state
func (m *IssueSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.searchInput.Width = min(msg.Width-20, 80) // Responsive search width
		return m, nil

	case issuesLoadedMsg:
		if msg.err != nil {
			m.state = stateError
			m.err = msg.err
		} else {
			m.state = stateReady
			m.issues = msg.issues
			m.filteredIssues = msg.issues // Initially, all issues are shown
			m.cursor = 0                  // Reset cursor when issues are loaded
		}
		return m, nil

	case searchCompletedMsg:
		if msg.err != nil {
			m.state = stateError
			m.err = msg.err
		} else {
			m.filteredIssues = msg.issues
			m.cursor = 0 // Reset cursor when search results change
		}
		return m, nil

	case tea.KeyMsg:
		// Handle quit keys first (always available)
		switch msg.Type {
		case tea.KeyCtrlC:
			m.state = stateQuit
			return m, tea.Quit
		}

		// Handle other keys based on focus and state
		if m.state == stateError || m.state == stateQuit {
			switch msg.Type {
			case tea.KeyRunes:
				if len(msg.Runes) > 0 && msg.Runes[0] == 'q' {
					m.state = stateQuit
					return m, tea.Quit
				}
			}
			return m, nil
		}

		if m.state != stateReady {
			return m, nil // Don't handle keys if not ready
		}

		// Handle search input when focused
		if m.searchFocused {
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			cmds = append(cmds, cmd)

			switch msg.Type {
			case tea.KeyTab, tea.KeyShiftTab:
				// Switch focus back to issue list
				m.searchFocused = false
				m.searchInput.Blur()
			case tea.KeyEnter:
				// Perform search
				query := m.searchInput.Value()
				cmds = append(cmds, m.loadIssues(query))
			case tea.KeyEsc:
				// Cancel search and switch focus
				m.searchFocused = false
				m.searchInput.Blur()
			}
			return m, tea.Batch(cmds...)
		}

		// Handle keys when issue list is focused
		switch msg.Type {
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.filteredIssues)-1 {
				m.cursor++
			}
		case tea.KeyTab, tea.KeyShiftTab:
			// Switch focus to search input
			m.searchFocused = true
			m.searchInput.Focus()
			cmds = append(cmds, textinput.Blink)
		case tea.KeyEnter:
			// Select current issue
			if len(m.filteredIssues) > 0 && m.cursor < len(m.filteredIssues) {
				m.selectedIssue = &m.filteredIssues[m.cursor]
				m.state = stateSelected
				return m, tea.Quit
			}
		case tea.KeyRunes:
			if len(msg.Runes) > 0 {
				switch msg.Runes[0] {
				case 'q':
					m.state = stateQuit
					return m, tea.Quit
				case 'j':
					// Vim-style down
					if m.cursor < len(m.filteredIssues)-1 {
						m.cursor++
					}
				case 'k':
					// Vim-style up
					if m.cursor > 0 {
						m.cursor--
					}
				case '?':
					// Toggle help
					m.showHelp = !m.showHelp
				case 'r':
					// Refresh issues
					query := m.searchInput.Value()
					cmds = append(cmds, m.loadIssues(query))
				case '/':
					// Focus search
					m.searchFocused = true
					m.searchInput.Focus()
					cmds = append(cmds, textinput.Blink)
				default:
					// Start typing in search
					m.searchFocused = true
					m.searchInput.Focus()
					m.searchInput.SetValue(string(msg.Runes))
					cmds = append(cmds, textinput.Blink)
				}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the current state of the model
func (m *IssueSelectModel) View() string {
	switch m.state {
	case stateLoading:
		return m.loadingView()
	case stateError:
		return m.errorView()
	case stateQuit:
		return "" // Empty view when quitting
	case stateSelected:
		return fmt.Sprintf("Selected issue #%d: %s\n", m.selectedIssue.Number, m.selectedIssue.Title)
	default: // stateReady
		return m.readyView()
	}
}

// loadingView renders the loading state
func (m *IssueSelectModel) loadingView() string {
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render("Loading issues...")
}

// errorView renders the error state
func (m *IssueSelectModel) errorView() string {
	errorText := fmt.Sprintf("Error: %v\n\nPress q to quit", m.err)
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(errorText)
}

// readyView renders the main issue selection interface
func (m *IssueSelectModel) readyView() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Select an Issue")
	b.WriteString(title + "\n\n")

	// Search input
	searchLabel := "Search: "
	if m.searchFocused {
		searchLabel = "> Search: "
	}
	searchLine := searchLabel + m.searchInput.View()
	b.WriteString(searchLine + "\n\n")

	// Issues list or no issues message
	if len(m.filteredIssues) == 0 {
		if m.searchInput.Value() != "" {
			b.WriteString(mutedStyle.Render("No issues found matching your search.") + "\n")
		} else {
			b.WriteString(mutedStyle.Render("No open issues found in this repository.") + "\n")
		}
	} else {
		// Calculate responsive column widths based on terminal width
		// Use a simplified two-column layout for issue selection (Issue + Title only)
		widths := CalculateIssueSelectWidths(m.width)

		// Table header
		headerRow := FormatIssueSelectHeader(widths)
		b.WriteString(tableHeaderStyle.Render(headerRow) + "\n")

		// Issue rows
		visibleStart, visibleEnd := m.getVisibleRange()
		for i := visibleStart; i < visibleEnd; i++ {
			if i >= len(m.filteredIssues) {
				break
			}

			issue := m.filteredIssues[i]
			row := FormatIssueSelectRow(widths, issue.Number, issue.Title)

			// Apply selection style
			if i == m.cursor {
				row = selectedRowStyle.Render(row)
			} else {
				row = tableCellStyle.Render(row)
			}

			b.WriteString(row + "\n")
		}

		// Show pagination info if needed
		if len(m.filteredIssues) > m.getMaxVisibleIssues() {
			paginationInfo := fmt.Sprintf("Showing %d-%d of %d issues",
				visibleStart+1, min(visibleEnd, len(m.filteredIssues)), len(m.filteredIssues))
			b.WriteString("\n" + mutedStyle.Render(paginationInfo) + "\n")
		}
	}

	// Help text
	if m.showHelp {
		b.WriteString("\n" + m.helpView())
	} else {
		helpText := "\nPress ? for help, tab to search, enter to select, q to quit"
		b.WriteString(helpStyle.Render(helpText))
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(b.String())
}

// helpView renders the help text
func (m *IssueSelectModel) helpView() string {
	var help strings.Builder
	help.WriteString(headerStyle.Render("Help") + "\n")
	help.WriteString("↑/k    - Move up\n")
	help.WriteString("↓/j    - Move down\n")
	help.WriteString("enter  - Select issue\n")
	help.WriteString("tab    - Switch to search\n")
	help.WriteString("/      - Start search\n")
	help.WriteString("r      - Refresh issues\n")
	help.WriteString("?      - Toggle this help\n")
	help.WriteString("q      - Quit\n")
	return helpStyle.Render(help.String())
}

// getVisibleRange calculates which issues should be visible based on terminal height
func (m *IssueSelectModel) getVisibleRange() (int, int) {
	maxVisible := m.getMaxVisibleIssues()
	totalIssues := len(m.filteredIssues)

	if totalIssues <= maxVisible {
		return 0, totalIssues
	}

	// Center the cursor in the visible area when possible
	start := m.cursor - maxVisible/2
	if start < 0 {
		start = 0
	}

	end := start + maxVisible
	if end > totalIssues {
		end = totalIssues
		start = max(0, end-maxVisible)
	}

	return start, end
}

// getMaxVisibleIssues calculates how many issues can fit on screen
func (m *IssueSelectModel) getMaxVisibleIssues() int {
	// Reserve space for title, search, headers, help, padding
	reservedLines := 8
	if m.showHelp {
		reservedLines += 10 // Additional space for help text
	}

	availableLines := m.height - reservedLines
	if availableLines < 1 {
		return 1 // Always show at least one issue
	}

	return availableLines
}

// loadIssues creates a command to load issues from GitHub
func (m *IssueSelectModel) loadIssues(searchQuery string) tea.Cmd {
	return func() tea.Msg {
		issues, err := m.githubClient.ListIssues(searchQuery, m.issueLimit)
		if searchQuery != "" {
			return searchCompletedMsg{issues: issues, query: searchQuery, err: err}
		}
		return issuesLoadedMsg{issues: issues, err: err}
	}
}

// GetSelectedIssue returns the selected issue (for external use)
func (m *IssueSelectModel) GetSelectedIssue() *issue.Issue {
	return m.selectedIssue
}

// GetState returns the current state (for external use)
func (m *IssueSelectModel) GetState() issueSelectState {
	return m.state
}

// IsQuit returns true if the user chose to quit
func (m *IssueSelectModel) IsQuit() bool {
	return m.state == stateQuit
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
