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
}

type ViewMode int

const (
	ViewModeRepository ViewMode = iota // Show only current repo sessions
	ViewModeGlobal                     // Show all sessions across repos
)

type Model struct {
	sessions     []config.SessionMetadata
	tmuxSessions []*tmux.Session
	cursor       int
	showHelp     bool
	viewMode     ViewMode
	currentRepo  *repo.Repository
	tmuxManager  *tmux.Manager
	repoManager  *repo.Manager
	width        int
	height       int
	error        error
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
		sessions:    []config.SessionMetadata{},
		cursor:      0,
		showHelp:    false,
		viewMode:    viewMode,
		currentRepo: currentRepo,
		tmuxManager: tmux.NewManager(),
		repoManager: repoManager,
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
		helpText := "\nPress enter to attach, ? for help, g to toggle view, r to refresh, q to quit"
		if m.currentRepo == nil && m.viewMode == ViewModeRepository {
			helpText = "\nNot in git repository - showing global view. Press enter to attach, ? for help, r to refresh, q to quit"
		}
		b.WriteString(helpStyle.Render(helpText))
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(b.String())
}

func (m Model) helpView() string {
	var help strings.Builder
	help.WriteString(headerStyle.Render("Help") + "\n")
	help.WriteString("↑/k    - Move up\n")
	help.WriteString("↓/j    - Move down\n")
	help.WriteString("enter  - Attach to selected session\n")
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
