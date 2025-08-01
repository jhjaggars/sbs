package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"sbs/pkg/config"
	"sbs/pkg/repo"
	"sbs/pkg/sandbox"
	"sbs/pkg/tmux"
)

type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Enter      key.Binding
	Quit       key.Binding
	Help       key.Binding
	Refresh    key.Binding
	ToggleView key.Binding
	Stop       key.Binding
	Clean      key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "attach to session"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	ToggleView: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "toggle global/repo view"),
	),
	Stop: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "stop session"),
	),
	Clean: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "clean stale"),
	),
}

type ViewMode int

const (
	ViewModeRepository ViewMode = iota // Show only current repo sessions
	ViewModeGlobal                     // Show all sessions across repos
)

type Model struct {
	sessions               []config.SessionMetadata
	tmuxSessions           []*tmux.Session
	cursor                 int
	showHelp               bool
	viewMode               ViewMode
	currentRepo            *repo.Repository
	tmuxManager            *tmux.Manager
	repoManager            *repo.Manager
	sandboxManager         *sandbox.Manager
	width                  int
	height                 int
	error                  error
	showConfirmationDialog bool
	confirmationMessage    string
	pendingCleanSessions   []config.SessionMetadata
}

func NewModel() Model {
	repoManager := repo.NewManager()
	currentRepo, _ := repoManager.DetectCurrentRepository()

	// Default to repository view if in a repo, global otherwise
	viewMode := ViewModeGlobal
	if currentRepo != nil {
		viewMode = ViewModeRepository
	}

	return Model{
		sessions:               []config.SessionMetadata{},
		cursor:                 0,
		showHelp:               false,
		viewMode:               viewMode,
		currentRepo:            currentRepo,
		tmuxManager:            tmux.NewManager(),
		repoManager:            repoManager,
		sandboxManager:         sandbox.NewManager(),
		showConfirmationDialog: false,
		confirmationMessage:    "",
		pendingCleanSessions:   []config.SessionMetadata{},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.refreshSessions(),
		tea.EnterAltScreen,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle modal dialog keys first (higher priority)
		if m.showConfirmationDialog {
			switch msg.Type {
			case tea.KeyEsc:
				m.showConfirmationDialog = false
				m.confirmationMessage = ""
				m.pendingCleanSessions = []config.SessionMetadata{}
				return m, nil
			case tea.KeyEnter:
				m.showConfirmationDialog = false
				return m, m.executeCleanup()
			case tea.KeyRunes:
				switch string(msg.Runes) {
				case "y", "Y":
					m.showConfirmationDialog = false
					return m, m.executeCleanup()
				case "n", "N":
					m.showConfirmationDialog = false
					m.confirmationMessage = ""
					m.pendingCleanSessions = []config.SessionMetadata{}
					return m, nil
				}
			}
			return m, nil
		}

		// Normal key handling when modal is not shown
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.sessions)-1 {
				m.cursor++
			}
			return m, nil

		case key.Matches(msg, keys.Enter):
			if len(m.sessions) > 0 && m.cursor < len(m.sessions) {
				sessionName := m.sessions[m.cursor].TmuxSession
				return m, m.attachToSession(sessionName)
			}
			return m, nil

		case key.Matches(msg, keys.Stop):
			if len(m.sessions) > 0 && m.cursor < len(m.sessions) {
				return m, m.stopSelectedSession()
			}
			return m, nil

		case key.Matches(msg, keys.Clean):
			return m.showCleanConfirmation(), nil

		case key.Matches(msg, keys.Help):
			m.showHelp = !m.showHelp
			return m, nil

		case key.Matches(msg, keys.Refresh):
			return m, m.refreshSessions()

		case key.Matches(msg, keys.ToggleView):
			return m.toggleViewMode(), m.refreshSessions()
		}

	case refreshMsg:
		m.sessions = msg.sessions
		m.tmuxSessions = msg.tmuxSessions
		m.error = msg.err
		return m, nil

	case attachMsg:
		if msg.err != nil {
			m.error = msg.err
		}
		return m, nil

	case stopSessionMsg:
		m.error = msg.err
		return m, m.refreshSessions()

	case cleanSessionsMsg:
		m.error = msg.err
		m.showConfirmationDialog = false
		return m, m.refreshSessions()
	}

	return m, nil
}

func (m Model) View() string {
	if m.error != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit", m.error)
	}

	var b strings.Builder

	// Title with view mode indicator
	var title string
	if m.currentRepo != nil && m.viewMode == ViewModeRepository {
		title = titleStyle.Render(fmt.Sprintf("Work Issue Orchestrator (%s)", m.currentRepo.Name))
	} else {
		title = titleStyle.Render("Work Issue Orchestrator (Global)")
	}
	b.WriteString(title + "\n\n")

	// Sessions list
	if len(m.sessions) == 0 {
		b.WriteString(mutedStyle.Render("No active work sessions found.") + "\n")
		b.WriteString(mutedStyle.Render("Use 'work-orchestrator start <issue-number>' to create a new session.") + "\n")
	} else {
		// Calculate responsive column widths based on terminal width
		var widths ColumnWidths
		var headerRow string

		if m.viewMode == ViewModeGlobal {
			widths = CalculateGlobalViewWidths(m.width)
			headerRow = FormatGlobalViewHeader(widths)
		} else {
			widths = CalculateRepositoryViewWidths(m.width)
			headerRow = FormatRepositoryViewHeader(widths)
		}

		b.WriteString(tableHeaderStyle.Render(headerRow) + "\n")

		// Sessions
		for i, session := range m.sessions {
			// Determine actual status by checking tmux
			status := m.getSessionStatus(session.TmuxSession)
			lastActivity := m.formatTimeAgo(session.LastActivity)

			// Format row based on view mode using responsive widths
			var row string
			if m.viewMode == ViewModeGlobal {
				row = FormatGlobalViewRow(widths,
					session.IssueNumber,
					session.IssueTitle,
					session.RepositoryName,
					session.Branch,
					FormatStatus(status),
					lastActivity,
				)
			} else {
				row = FormatRepositoryViewRow(widths,
					session.IssueNumber,
					session.IssueTitle,
					session.Branch,
					FormatStatus(status),
					lastActivity,
				)
			}

			// Apply selection style
			if i == m.cursor {
				row = selectedRowStyle.Render(row)
			} else {
				row = tableCellStyle.Render(row)
			}

			b.WriteString(row + "\n")
		}
	}

	// Help
	if m.showHelp {
		b.WriteString("\n" + m.helpView())
	} else {
		helpText := "\nenter: attach, s: stop, c: clean, ?: help, g: toggle, r: refresh, q: quit"
		if m.currentRepo == nil && m.viewMode == ViewModeRepository {
			helpText = "\nNot in git repository - global view. enter: attach, s: stop, c: clean, ?: help, r: refresh, q: quit"
		}
		b.WriteString(helpStyle.Render(helpText))
	}

	content := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(b.String())

	// Render modal dialog overlay if shown
	if m.showConfirmationDialog {
		dialog := modalContentStyle.Render(m.confirmationMessage)

		// Center the dialog
		dialogWidth := lipgloss.Width(dialog)
		dialogHeight := lipgloss.Height(dialog)

		x := maxInt(0, (m.width-dialogWidth)/2)
		y := maxInt(0, (m.height-dialogHeight)/2)

		content = lipgloss.Place(m.width, m.height,
			lipgloss.Left, lipgloss.Top,
			lipgloss.JoinVertical(lipgloss.Left,
				strings.Repeat("\n", y),
				lipgloss.JoinHorizontal(lipgloss.Left,
					strings.Repeat(" ", x),
					dialog,
				),
			),
			lipgloss.WithWhitespaceChars(" "),
		)
	}

	return content
}

func (m Model) helpView() string {
	var help strings.Builder
	help.WriteString(headerStyle.Render("Help") + "\n")
	help.WriteString("↑/k    - Move up\n")
	help.WriteString("↓/j    - Move down\n")
	help.WriteString("enter  - Attach to selected session\n")
	help.WriteString("s      - Stop selected session\n")
	help.WriteString("c      - Clean stale sessions\n")
	help.WriteString("g      - Toggle global/repository view\n")
	help.WriteString("r      - Refresh session list\n")
	help.WriteString("?      - Toggle this help\n")
	help.WriteString("q      - Quit\n")
	return helpStyle.Render(help.String())
}

func (m Model) getSessionStatus(sessionName string) string {
	exists, _ := m.tmuxManager.SessionExists(sessionName)
	if exists {
		return "active"
	}
	return "stopped"
}

func (m Model) formatTimeAgo(timeStr string) string {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return "unknown"
	}

	duration := time.Since(t)
	if duration < time.Minute {
		return "now"
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm ago", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%dd ago", int(duration.Hours()/24))
	}
}

type refreshMsg struct {
	sessions     []config.SessionMetadata
	tmuxSessions []*tmux.Session
	err          error
}

type attachMsg struct {
	err error
}

type stopSessionMsg struct {
	err     error
	success bool
}

type cleanSessionsMsg struct {
	err             error
	cleanedSessions []config.SessionMetadata
}

type confirmationDialogMsg struct {
	show    bool
	message string
}

// toggleViewMode switches between repository and global view modes
func (m Model) toggleViewMode() Model {
	if m.currentRepo == nil {
		// Can't toggle to repository view if not in a repository
		return m
	}

	if m.viewMode == ViewModeRepository {
		m.viewMode = ViewModeGlobal
	} else {
		m.viewMode = ViewModeRepository
	}

	// Reset cursor when switching views
	m.cursor = 0

	return m
}

func (m Model) refreshSessions() tea.Cmd {
	return func() tea.Msg {
		// Always load from global sessions file
		allSessions, err := config.LoadAllRepositorySessions()
		if err != nil {
			return refreshMsg{err: err}
		}

		var sessions []config.SessionMetadata

		if m.viewMode == ViewModeRepository && m.currentRepo != nil {
			// Filter sessions for current repository
			for _, session := range allSessions {
				if session.RepositoryRoot == m.currentRepo.Root {
					sessions = append(sessions, session)
				}
			}
		} else {
			// Show all sessions (global view)
			sessions = allSessions
		}

		tmuxSessions, err := m.tmuxManager.ListSessions()
		if err != nil {
			return refreshMsg{err: err}
		}

		return refreshMsg{
			sessions:     sessions,
			tmuxSessions: tmuxSessions,
		}
	}
}

func (m Model) attachToSession(sessionName string) tea.Cmd {
	return func() tea.Msg {
		err := m.tmuxManager.AttachToSession(sessionName)
		return attachMsg{err: err}
	}
}

func (m Model) stopSelectedSession() tea.Cmd {
	if m.cursor < 0 || m.cursor >= len(m.sessions) {
		return func() tea.Msg {
			return stopSessionMsg{err: fmt.Errorf("no session selected"), success: false}
		}
	}

	session := m.sessions[m.cursor]
	return func() tea.Msg {
		// Check if tmux session exists
		exists, err := m.tmuxManager.SessionExists(session.TmuxSession)
		if err != nil {
			return stopSessionMsg{err: fmt.Errorf("failed to check tmux session: %w", err), success: false}
		}

		// Kill tmux session if it exists
		if exists {
			if err := m.tmuxManager.KillSession(session.TmuxSession); err != nil {
				return stopSessionMsg{err: fmt.Errorf("failed to kill tmux session: %w", err), success: false}
			}
		}

		// Stop sandbox if it exists
		sandboxName := session.SandboxName
		if sandboxName == "" {
			sandboxName = m.sandboxManager.GetSandboxName(session.IssueNumber)
		}

		sandboxExists, err := m.sandboxManager.SandboxExists(sandboxName)
		if err == nil && sandboxExists {
			if err := m.sandboxManager.DeleteSandbox(sandboxName); err != nil {
				return stopSessionMsg{err: fmt.Errorf("failed to delete sandbox: %w", err), success: false}
			}
		}

		return stopSessionMsg{err: nil, success: true}
	}
}

func (m Model) showCleanConfirmation() Model {
	staleSessions := m.identifyStaleSessionsInCurrentView()
	if len(staleSessions) == 0 {
		return m
	}

	m.showConfirmationDialog = true
	m.pendingCleanSessions = staleSessions

	var message strings.Builder
	if len(staleSessions) == 1 {
		message.WriteString("Clean 1 stale session?\n")
	} else {
		message.WriteString(fmt.Sprintf("Clean %d stale sessions?\n", len(staleSessions)))
	}

	for _, session := range staleSessions {
		message.WriteString(fmt.Sprintf("Issue #%d: %s\n", session.IssueNumber, session.IssueTitle))
	}
	message.WriteString("\n(y/n) Press y to confirm, n to cancel")

	m.confirmationMessage = message.String()
	return m
}

func (m Model) executeCleanup() tea.Cmd {
	sessions := m.pendingCleanSessions
	return func() tea.Msg {
		var cleanedSessions []config.SessionMetadata
		var hasErrors bool

		for _, session := range sessions {
			// Clean up sandbox
			sandboxName := session.SandboxName
			if sandboxName == "" {
				sandboxName = m.sandboxManager.GetSandboxName(session.IssueNumber)
			}

			sandboxExists, err := m.sandboxManager.SandboxExists(sandboxName)
			if err == nil && sandboxExists {
				if err := m.sandboxManager.DeleteSandbox(sandboxName); err != nil {
					hasErrors = true
					continue
				}
			}

			cleanedSessions = append(cleanedSessions, session)
		}

		var err error
		if hasErrors {
			err = fmt.Errorf("failed to clean some sessions")
		}

		return cleanSessionsMsg{
			err:             err,
			cleanedSessions: cleanedSessions,
		}
	}
}

func (m Model) identifyStaleSessionsInCurrentView() []config.SessionMetadata {
	var staleSessions []config.SessionMetadata

	for _, session := range m.sessions {
		exists, err := m.tmuxManager.SessionExists(session.TmuxSession)
		if err != nil {
			continue
		}
		if !exists {
			staleSessions = append(staleSessions, session)
		}
	}

	return staleSessions
}

func (m Model) identifyAndCleanStaleSessions() struct {
	cleanedSessions []config.SessionMetadata
	err             error
} {
	staleSessions := m.identifyStaleSessionsInCurrentView()

	var cleanedSessions []config.SessionMetadata
	var hasErrors bool

	for _, session := range staleSessions {
		// Clean up sandbox
		sandboxName := session.SandboxName
		if sandboxName == "" {
			sandboxName = m.sandboxManager.GetSandboxName(session.IssueNumber)
		}

		sandboxExists, err := m.sandboxManager.SandboxExists(sandboxName)
		if err == nil && sandboxExists {
			if err := m.sandboxManager.DeleteSandbox(sandboxName); err != nil {
				hasErrors = true
				continue
			}
		}

		cleanedSessions = append(cleanedSessions, session)
	}

	var err error
	if hasErrors {
		err = fmt.Errorf("failed to clean some sessions")
	}

	return struct {
		cleanedSessions []config.SessionMetadata
		err             error
	}{
		cleanedSessions: cleanedSessions,
		err:             err,
	}
}

// Helper function for maximum of two integers
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
