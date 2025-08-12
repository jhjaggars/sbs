package tui

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"sbs/pkg/cleanup"
	"sbs/pkg/config"
	"sbs/pkg/repo"
	"sbs/pkg/sandbox"
	"sbs/pkg/status"
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
	LogView    key.Binding
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
	LogView: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "logs"),
	),
}

// ViewMode type for TUI
type ViewMode int

const (
	ViewModeRepository ViewMode = iota // Show only current repo sessions
	ViewModeGlobal                     // Show all sessions across repos
	ViewModeLog                        // Show log view for selected session (TUI-specific)
)

// LogView represents the state of the log display
type LogView struct {
	content      string
	scrollOffset int
	loading      bool
	refreshing   bool
	errorMessage string
	maxLines     int
	maxSizeBytes int // Maximum content size in bytes
}

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
	statusDetector         *status.Detector
	cleanupManager         *cleanup.CleanupManager
	config                 *config.Config
	width                  int
	height                 int
	error                  error
	showConfirmationDialog bool
	confirmationMessage    string

	// Log view state
	logView              *LogView
	previousViewMode     ViewMode
	logAutoRefreshActive bool
	logAutoRefreshMutex  sync.Mutex // Prevent multiple concurrent refreshes
	pendingCleanSessions []config.SessionMetadata
}

func NewModel() Model {
	repoManager := repo.NewManager()
	currentRepo, _ := repoManager.DetectCurrentRepository()

	// Load configuration
	cfg, _ := config.LoadConfig()
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// Default to repository view if in a repo, global otherwise
	viewMode := ViewModeGlobal
	if currentRepo != nil {
		viewMode = ViewModeRepository
	}

	tmuxManager := tmux.NewManager()
	sandboxManager := sandbox.NewManager()
	cleanupManager := cleanup.NewCleanupManager(tmuxManager, sandboxManager, nil, nil)
	return Model{
		sessions:               []config.SessionMetadata{},
		cursor:                 0,
		showHelp:               false,
		viewMode:               viewMode,
		currentRepo:            currentRepo,
		tmuxManager:            tmuxManager,
		repoManager:            repoManager,
		sandboxManager:         sandboxManager,
		statusDetector:         status.NewDetector(tmuxManager, sandboxManager),
		cleanupManager:         cleanupManager,
		config:                 cfg,
		showConfirmationDialog: false,
		confirmationMessage:    "",
		pendingCleanSessions:   []config.SessionMetadata{},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.refreshSessions(),
		tea.EnterAltScreen,
		m.tickAutoRefresh(),
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

		// Handle log view keys when in log view mode
		if m.viewMode == ViewModeLog {
			switch msg.Type {
			case tea.KeyEsc:
				// Exit log view and return to previous view
				m.viewMode = m.previousViewMode
				m.stopLogAutoRefresh()
				return m, nil
			case tea.KeyUp:
				if m.logView != nil && m.logView.scrollOffset > 0 {
					m.logView.scrollOffset--
				}
				return m, nil
			case tea.KeyDown:
				if m.logView != nil {
					lines := strings.Split(m.logView.content, "\n")
					maxScroll := len(lines) - m.height + 5 // Leave some space for UI
					if maxScroll < 0 {
						maxScroll = 0
					}
					if m.logView.scrollOffset < maxScroll {
						m.logView.scrollOffset++
					}
				}
				return m, nil
			case tea.KeyRunes:
				switch string(msg.Runes) {
				case "q":
					// Exit log view and return to previous view
					m.viewMode = m.previousViewMode
					m.stopLogAutoRefresh()
					return m, nil
				case "r":
					// Manual refresh and restart auto-refresh if it was stopped
					m.logAutoRefreshMutex.Lock()
					if m.logView != nil {
						if !m.logView.refreshing {
							m.logView.refreshing = true
							// Re-enable auto-refresh when user manually retries
							m.logAutoRefreshActive = true
							m.logAutoRefreshMutex.Unlock()
							return m, tea.Batch(
								m.refreshLogContent(),
								m.startLogAutoRefresh(),
							)
						}
					}
					m.logAutoRefreshMutex.Unlock()
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

		case key.Matches(msg, keys.LogView):
			// Enter log view mode if we have sessions and a valid selection
			if len(m.sessions) > 0 && m.cursor >= 0 && m.cursor < len(m.sessions) {
				m.previousViewMode = m.viewMode
				m.viewMode = ViewModeLog
				m.logAutoRefreshActive = true

				// Initialize log view if not already initialized
				if m.logView == nil {
					m.logView = &LogView{
						content:      "",
						scrollOffset: 0,
						loading:      true,
						refreshing:   false,
						errorMessage: "",
						maxLines:     0,
						maxSizeBytes: 1048576, // 1MB default limit
					}
				} else {
					m.logView.loading = true
				}

				// Start auto-refresh and initial content load
				return m, tea.Batch(
					m.refreshLogContent(),
					m.startLogAutoRefresh(),
				)
			}
			return m, nil
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

	case tickMsg:
		// Auto-refresh sessions and schedule next tick
		return m, tea.Batch(
			m.refreshSessions(),
			m.tickAutoRefresh(),
		)

	case logRefreshTickMsg:
		// Handle auto-refresh for log view - prevent race conditions
		if m.viewMode == ViewModeLog && m.logAutoRefreshActive {
			// Use mutex to prevent multiple concurrent refreshes
			m.logAutoRefreshMutex.Lock()
			defer m.logAutoRefreshMutex.Unlock()

			// Initialize logView if not present
			if m.logView == nil {
				m.logView = &LogView{
					content:      "",
					scrollOffset: 0,
					loading:      true,
					refreshing:   false,
					errorMessage: "",
					maxLines:     0,
					maxSizeBytes: 1048576, // 1MB default limit
				}
			}

			// Only start refresh if not already refreshing
			if !m.logView.refreshing {
				return m, tea.Batch(
					m.refreshLogContent(),
					m.startLogAutoRefresh(),
				)
			} else {
				// Still schedule next refresh even if currently refreshing
				return m, m.startLogAutoRefresh()
			}
		}
		return m, nil

	case logRefreshResultMsg:
		// Handle log refresh results
		if m.logView != nil {
			m.logView.loading = false
			m.logView.refreshing = false

			if msg.err != nil {
				m.logView.errorMessage = fmt.Sprintf("refresh failed: %v", msg.err)

				// Stop auto-refresh for persistent errors that won't be resolved by retrying
				if strings.Contains(msg.err.Error(), "script not found") ||
					strings.Contains(msg.err.Error(), "path validation failed") {
					m.logAutoRefreshActive = false
				}
			} else {
				// Apply content size limits and rotation
				content := msg.content
				if len(content) > m.logView.maxSizeBytes {
					// Truncate content to last portion that fits within limits
					lines := strings.Split(content, "\n")
					truncatedContent := ""

					// Start from the end and work backwards to preserve recent content
					for i := len(lines) - 1; i >= 0; i-- {
						newContent := lines[i] + "\n" + truncatedContent
						if len(newContent) > m.logView.maxSizeBytes {
							break
						}
						truncatedContent = newContent
					}

					if truncatedContent != "" {
						content = "[Content truncated to last " + fmt.Sprintf("%d", m.logView.maxSizeBytes/1024) + "KB]\n" + truncatedContent
					}
				}

				m.logView.content = content
				m.logView.errorMessage = ""
			}
		}
		return m, nil

	case logRefreshErrorMsg:
		// Handle log refresh errors
		if m.logView != nil {
			m.logView.loading = false
			m.logView.refreshing = false
			m.logView.errorMessage = fmt.Sprintf("refresh failed: %v", msg.err)
		}
		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	if m.error != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit", m.error)
	}

	// Handle log view rendering
	if m.viewMode == ViewModeLog {
		return m.renderLogView()
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
		b.WriteString(mutedStyle.Render("Use 'sbs start <issue-number>' to create a new session.") + "\n")
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
			// Determine actual status using status detector
			sessionStatus := m.getSessionStatus(session)

			// Format row based on view mode using responsive widths
			var row string
			if m.viewMode == ViewModeGlobal {
				row = FormatGlobalViewRow(widths,
					session.NamespacedID,
					session.IssueTitle,
					session.RepositoryName,
					session.Branch,
					FormatStatus(sessionStatus.Status),
					sessionStatus.TimeDelta,
				)
			} else {
				row = FormatRepositoryViewRow(widths,
					session.NamespacedID,
					session.IssueTitle,
					session.Branch,
					FormatStatus(sessionStatus.Status),
					sessionStatus.TimeDelta,
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
		helpText := "\nPress enter: attach, l: logs, s: stop, c: clean, ?: help, g: toggle, r: refresh, q: quit"
		if m.currentRepo == nil && m.viewMode == ViewModeRepository {
			helpText = "\nNot in git repository - global view. Press enter: attach, l: logs, s: stop, c: clean, ?: help, r: refresh, q: quit"
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
	help.WriteString("l      - View logs for selected session\n")
	help.WriteString("s      - Stop selected session\n")
	help.WriteString("c      - Clean stale sessions\n")
	help.WriteString("g      - Toggle global/repository view\n")
	help.WriteString("r      - Refresh session list\n")
	help.WriteString("?      - Toggle this help\n")
	help.WriteString("q      - Quit\n")
	return helpStyle.Render(help.String())
}

func (m Model) getSessionStatus(session config.SessionMetadata) status.SessionStatus {
	return m.statusDetector.DetectSessionStatus(session)
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

type tickMsg struct{}

// Log view message types
type logRefreshTickMsg struct{}

type logRefreshResultMsg struct {
	content string
	err     error
}

type logRefreshErrorMsg struct {
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
			return stopSessionMsg{err: fmt.Errorf("session missing sandbox name"), success: false}
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
	// Convert TUI ViewMode to cleanup ViewMode
	viewMode := cleanup.ViewModeGlobal
	if m.viewMode == ViewModeRepository {
		viewMode = cleanup.ViewModeRepository
	}

	staleSessions, err := m.cleanupManager.IdentifyStaleSessionsInView(m.sessions, viewMode)
	if err != nil || len(staleSessions) == 0 {
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
		if session.NamespacedID != "" {
			message.WriteString(fmt.Sprintf("Work Item %s: %s\n", session.NamespacedID, session.IssueTitle))
		} else {
			message.WriteString(fmt.Sprintf("Issue #%d: %s\n", session.IssueNumber, session.IssueTitle))
		}
	}
	message.WriteString("\n(y/n) Press y to confirm, n to cancel")

	m.confirmationMessage = message.String()
	return m
}

func (m Model) executeCleanup() tea.Cmd {
	sessions := m.pendingCleanSessions
	return func() tea.Msg {
		// Convert TUI ViewMode to cleanup ViewMode
		viewMode := cleanup.ViewModeGlobal
		if m.viewMode == ViewModeRepository {
			viewMode = cleanup.ViewModeRepository
		}

		options := m.cleanupManager.BuildTUICleanupOptions(viewMode, true)
		results, err := m.cleanupManager.CleanupSessions(sessions, options)

		var cleanupError error
		if err != nil {
			cleanupError = err
		} else if len(results.Errors) > 0 {
			cleanupError = fmt.Errorf("failed to clean some sessions")
		}

		return cleanSessionsMsg{
			err:             cleanupError,
			cleanedSessions: sessions[:results.CleanedSessions], // Return the sessions that were cleaned
		}
	}
}

// tickAutoRefresh creates a command that triggers auto-refresh after the configured interval
func (m Model) tickAutoRefresh() tea.Cmd {
	if !m.config.StatusTracking {
		return nil
	}

	interval := time.Duration(m.config.StatusRefreshIntervalSecs) * time.Second
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

// renderLogView renders the log view UI
func (m Model) renderLogView() string {
	var b strings.Builder

	// Title for log view
	sessionTitle := "Log View"
	if len(m.sessions) > 0 && m.cursor >= 0 && m.cursor < len(m.sessions) {
		session := m.sessions[m.cursor]
		if session.NamespacedID != "" {
			sessionTitle = fmt.Sprintf("Log View - Work Item %s: %s", session.NamespacedID, session.IssueTitle)
		} else {
			sessionTitle = fmt.Sprintf("Log View - Issue #%d: %s", session.IssueNumber, session.IssueTitle)
		}
	}
	b.WriteString(titleStyle.Render(sessionTitle) + "\n\n")

	// Log content area
	if m.logView == nil {
		b.WriteString(mutedStyle.Render("No log view initialized") + "\n")
	} else if m.logView.loading {
		b.WriteString(mutedStyle.Render("Loading log content...") + "\n")
	} else if m.logView.errorMessage != "" {
		b.WriteString(errorStyle.Render("Error: "+m.logView.errorMessage) + "\n")
		if !m.logAutoRefreshActive {
			b.WriteString(mutedStyle.Render("Auto-refresh stopped due to persistent error.") + "\n")
		}
		b.WriteString(mutedStyle.Render("Press 'r' to retry, ESC or 'q' to exit") + "\n")
	} else if m.logView.content == "" {
		b.WriteString(mutedStyle.Render("No log content available") + "\n")
	} else {
		// Display log content with scrolling
		lines := strings.Split(m.logView.content, "\n")
		displayHeight := m.height - 6 // Reserve space for title and help text

		startLine := m.logView.scrollOffset
		endLine := startLine + displayHeight

		if endLine > len(lines) {
			endLine = len(lines)
		}

		if startLine < len(lines) {
			visibleLines := lines[startLine:endLine]
			for _, line := range visibleLines {
				b.WriteString(line + "\n")
			}
		}

		// Show scroll indicators
		if len(lines) > displayHeight {
			scrollInfo := fmt.Sprintf("Lines %d-%d of %d", startLine+1, endLine, len(lines))
			b.WriteString("\n" + mutedStyle.Render(scrollInfo))
		}
	}

	// Status line
	var statusParts []string
	if m.logView != nil && m.logView.refreshing {
		statusParts = append(statusParts, "Refreshing...")
	}
	if m.logAutoRefreshActive {
		interval := m.getLogRefreshInterval()
		statusParts = append(statusParts, fmt.Sprintf("Auto-refresh: %ds", int(interval.Seconds())))
	}

	if len(statusParts) > 0 {
		b.WriteString("\n" + mutedStyle.Render(strings.Join(statusParts, " | ")))
	}

	// Help text for log view
	helpText := "\nPress ↑/↓: scroll, r: refresh, ESC/q: exit"
	b.WriteString(helpStyle.Render(helpText))

	content := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(b.String())

	return content
}

// Log view helper functions

// LogExecutionInfo contains information about script execution for audit logging
type LogExecutionInfo struct {
	ScriptPath      string    `json:"script_path"`
	WorkingDir      string    `json:"working_dir"`
	ExecutionTime   time.Time `json:"execution_time"`
	DurationMs      int64     `json:"duration_ms"`
	ExitCode        int       `json:"exit_code"`
	OutputSizeBytes int       `json:"output_size_bytes"`
	TimedOut        bool      `json:"timed_out"`
	Error           string    `json:"error,omitempty"`
}

// validateLoghookPath validates and sanitizes the loghook script path
func validateLoghookPath(worktreePath string) (string, error) {
	// Ensure it's an absolute path
	if !filepath.IsAbs(worktreePath) {
		return "", fmt.Errorf("worktree path must be absolute: %s", worktreePath)
	}

	// Clean the worktree path to resolve any .. or . components
	cleanWorktreePath := filepath.Clean(worktreePath)

	// Check for obvious path traversal attempts after cleaning
	if strings.Contains(cleanWorktreePath, "..") || strings.Contains(worktreePath, "..") {
		return "", fmt.Errorf("path traversal detected in worktree path: %s", worktreePath)
	}

	// Construct and clean the loghook path
	loghookPath := filepath.Join(cleanWorktreePath, ".sbs", "loghook")
	loghookPath = filepath.Clean(loghookPath)

	// Ensure the loghook path is still within the worktree by checking prefix
	if !strings.HasPrefix(loghookPath, cleanWorktreePath) {
		return "", fmt.Errorf("loghook path is outside worktree: %s", loghookPath)
	}

	return loghookPath, nil
}

// validateScriptSecurity performs security checks on the loghook script
func validateScriptSecurity(scriptPath string) error {
	// Get file info
	info, err := os.Stat(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to stat script at %s: %w", scriptPath, err)
	}

	// Check if it's a regular file (not a symlink, device, etc.)
	if !info.Mode().IsRegular() {
		return fmt.Errorf("script at %s is not a regular file", scriptPath)
	}

	// Check if script is executable
	if info.Mode().Perm()&0111 == 0 {
		return fmt.Errorf("permission denied: script at %s is not executable", scriptPath)
	}

	// Check file ownership (should be owned by current user for security)
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		currentUID := os.Getuid()
		if int(stat.Uid) != currentUID {
			log.Printf("Warning: loghook script at %s is not owned by current user (uid=%d, script_uid=%d)",
				scriptPath, currentUID, stat.Uid)
		}
	}

	return nil
}

// logScriptExecution logs script execution for audit trails
func logScriptExecution(info LogExecutionInfo) {
	// Log to stderr for audit trail (can be captured by system logging)
	log.Printf("AUDIT: Loghook script execution - Path: %s, Duration: %dms, ExitCode: %d, OutputSize: %d bytes, TimedOut: %t",
		info.ScriptPath, info.DurationMs, info.ExitCode, info.OutputSizeBytes, info.TimedOut)

	if info.Error != "" {
		log.Printf("AUDIT: Loghook script error - %s", info.Error)
	}
}

// executeLoghookScript executes the .sbs/loghook script with default timeout (10 seconds)
func executeLoghookScript(session config.SessionMetadata) (string, error) {
	return executeLoghookScriptWithOptions(session, 10, 1048576) // 10s timeout, 1MB limit
}

// ExecuteLoghookScript executes the loghook script for a session (exported wrapper)
func ExecuteLoghookScript(session config.SessionMetadata) (string, error) {
	return executeLoghookScript(session)
}

// executeLoghookScriptWithTimeout executes the loghook script with a custom timeout
func executeLoghookScriptWithTimeout(session config.SessionMetadata, timeoutSecs int) (string, error) {
	return executeLoghookScriptWithOptions(session, timeoutSecs, 1048576) // Custom timeout, 1MB limit
}

// executeLoghookScriptWithOptions executes the loghook script with full customization
func executeLoghookScriptWithOptions(session config.SessionMetadata, timeoutSecs int, maxOutputBytes int) (string, error) {
	startTime := time.Now()
	var execInfo LogExecutionInfo

	// Validate and sanitize the script path
	loghookPath, err := validateLoghookPath(session.WorktreePath)
	if err != nil {
		execInfo = LogExecutionInfo{
			ScriptPath:    "invalid_path",
			WorkingDir:    session.WorktreePath,
			ExecutionTime: startTime,
			Error:         err.Error(),
		}
		logScriptExecution(execInfo)
		return "", fmt.Errorf("path validation failed: %w", err)
	}

	execInfo.ScriptPath = loghookPath
	execInfo.WorkingDir = session.WorktreePath
	execInfo.ExecutionTime = startTime

	// Check if loghook script exists
	if _, err := os.Stat(loghookPath); os.IsNotExist(err) {
		execInfo.Error = "script not found"
		logScriptExecution(execInfo)
		return "No loghook script found at " + loghookPath, fmt.Errorf("loghook script not found at %s", loghookPath)
	}

	// Perform security validation
	if err := validateScriptSecurity(loghookPath); err != nil {
		execInfo.Error = err.Error()
		logScriptExecution(execInfo)
		return "", fmt.Errorf("security validation failed for %s: %w", loghookPath, err)
	}

	// Execute script with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSecs)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, loghookPath)
	cmd.Dir = session.WorktreePath // Set working directory to worktree

	// Capture output with size limits to prevent memory exhaustion
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		execInfo.Error = fmt.Sprintf("failed to create stdout pipe: %v", err)
		logScriptExecution(execInfo)
		return "", fmt.Errorf("failed to create stdout pipe for script %s: %w", loghookPath, err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		execInfo.Error = fmt.Sprintf("failed to create stderr pipe: %v", err)
		logScriptExecution(execInfo)
		return "", fmt.Errorf("failed to create stderr pipe for script %s: %w", loghookPath, err)
	}

	if err := cmd.Start(); err != nil {
		execInfo.Error = fmt.Sprintf("failed to start: %v", err)
		logScriptExecution(execInfo)
		return "", fmt.Errorf("failed to start script %s: %w", loghookPath, err)
	}

	// Read output with size limits
	var output strings.Builder
	totalBytes := 0
	buf := make([]byte, 4096)

	// Read from both stdout and stderr with size limits
	outputComplete := make(chan error, 2)

	go func() {
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				if totalBytes+n > maxOutputBytes {
					// Truncate to fit within limits
					remaining := maxOutputBytes - totalBytes
					if remaining > 0 {
						output.Write(buf[:remaining])
						totalBytes = maxOutputBytes
					}
					output.WriteString("\n[Output truncated - exceeded size limit]")
					break
				}
				output.Write(buf[:n])
				totalBytes += n
			}
			if err != nil {
				if err != io.EOF {
					outputComplete <- err
					return
				}
				break
			}
		}
		outputComplete <- nil
	}()

	go func() {
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				if totalBytes+n > maxOutputBytes {
					// Truncate to fit within limits
					remaining := maxOutputBytes - totalBytes
					if remaining > 0 {
						output.Write(buf[:remaining])
						totalBytes = maxOutputBytes
					}
					output.WriteString("\n[Error output truncated - exceeded size limit]")
					break
				}
				output.Write(buf[:n])
				totalBytes += n
			}
			if err != nil {
				if err != io.EOF {
					outputComplete <- err
					return
				}
				break
			}
		}
		outputComplete <- nil
	}()

	// Wait for output reading to complete
	for i := 0; i < 2; i++ {
		if err := <-outputComplete; err != nil {
			execInfo.Error = fmt.Sprintf("failed to read output: %v", err)
			logScriptExecution(execInfo)
			return output.String(), fmt.Errorf("failed to read output from script %s: %w", loghookPath, err)
		}
	}

	// Wait for command to complete
	err = cmd.Wait()
	execInfo.DurationMs = time.Since(startTime).Milliseconds()
	execInfo.OutputSizeBytes = totalBytes

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			execInfo.TimedOut = true
			execInfo.Error = fmt.Sprintf("timeout after %d seconds", timeoutSecs)
			logScriptExecution(execInfo)
			return output.String(), fmt.Errorf("script %s timed out after %d seconds", loghookPath, timeoutSecs)
		}

		// Get exit code if possible
		if exitError, ok := err.(*exec.ExitError); ok {
			execInfo.ExitCode = exitError.ExitCode()
		} else {
			execInfo.ExitCode = -1
		}

		execInfo.Error = fmt.Sprintf("script execution failed: %v", err)
		logScriptExecution(execInfo)
		return output.String(), fmt.Errorf("script %s execution failed: %w", loghookPath, err)
	}

	// Success case
	execInfo.ExitCode = 0
	logScriptExecution(execInfo)
	return output.String(), nil
}

// getLogRefreshInterval returns the configured log refresh interval with bounds checking
func (m Model) getLogRefreshInterval() time.Duration {
	intervalSecs := m.config.LogRefreshIntervalSecs
	if intervalSecs == 0 {
		intervalSecs = 5 // Default to 5 seconds
	}

	// Enforce bounds (2-120 seconds)
	if intervalSecs < 2 {
		intervalSecs = 2
	}
	if intervalSecs > 120 {
		intervalSecs = 120
	}

	return time.Duration(intervalSecs) * time.Second
}

// startLogAutoRefresh starts the auto-refresh mechanism for log view
func (m Model) startLogAutoRefresh() tea.Cmd {
	if m.viewMode != ViewModeLog {
		return nil
	}

	interval := m.getLogRefreshInterval()
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return logRefreshTickMsg{}
	})
}

// stopLogAutoRefresh stops the auto-refresh mechanism
func (m *Model) stopLogAutoRefresh() {
	m.logAutoRefreshActive = false
}

// refreshLogContent refreshes the log content for the current session
func (m Model) refreshLogContent() tea.Cmd {
	if m.viewMode != ViewModeLog || len(m.sessions) == 0 || m.cursor < 0 || m.cursor >= len(m.sessions) {
		return nil
	}

	session := m.sessions[m.cursor]
	return func() tea.Msg {
		// Use the refactored function with timeout from config or default
		timeoutSecs := 10 // default timeout
		if m.config != nil && m.config.StatusTimeoutSeconds > 0 {
			timeoutSecs = m.config.StatusTimeoutSeconds
		}

		content, err := executeLoghookScriptWithTimeout(session, timeoutSecs)
		return logRefreshResultMsg{
			content: content,
			err:     err,
		}
	}
}

// Helper function for maximum of two integers
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
