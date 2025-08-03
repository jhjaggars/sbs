package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"sbs/pkg/config"
	"sbs/pkg/tui"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all active work sessions in plain text format",
	Long: `Display a plain text list of all active work sessions.
Shows session details in a formatted table for easy parsing and scripting.
Use the bare 'sbs' command to launch the interactive TUI instead.`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolP("plain", "p", false, "Show plain text output (default behavior, kept for backward compatibility)")
}

func runList(cmd *cobra.Command, args []string) error {
	plain, _ := cmd.Flags().GetBool("plain")

	// Default behavior is now plain text output
	// The --plain flag is kept for backward compatibility but is redundant
	if !plain {
		// Always show plain text output (--plain flag is now redundant but kept for compatibility)
		return runPlainList()
	}

	// Still support --plain explicitly for backward compatibility
	return runPlainList()
}

func runPlainList() error {
	// Load sessions
	sessions, err := config.LoadAllRepositorySessions()
	if err != nil {
		return fmt.Errorf("failed to load sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No active work sessions found.")
		return nil
	}

	// Determine if we should use global view (if sessions from multiple repos)
	useGlobalView := shouldUseGlobalView(sessions)

	// Print summary line
	printSummaryLine(sessions, useGlobalView)
	fmt.Println() // Empty line after summary

	// Get terminal width for column calculations
	terminalWidth := getTerminalWidth()

	// Print header and sessions using new aesthetic format
	if useGlobalView {
		printGlobalViewSessions(sessions, terminalWidth)
	} else {
		printRepositoryViewSessions(sessions, terminalWidth)
	}

	return nil
}

func printSummaryLine(sessions []config.SessionMetadata, useGlobalView bool) {
	count := len(sessions)
	sessionWord := "session"
	if count != 1 {
		sessionWord = "sessions"
	}

	if useGlobalView {
		// Multiple repositories
		repoNames := make(map[string]bool)
		for _, session := range sessions {
			repoNames[session.RepositoryName] = true
		}
		repoCount := len(repoNames)
		repoWord := "repository"
		if repoCount != 1 {
			repoWord = "repositories"
		}
		fmt.Printf("Showing %d of %d active %s across %d %s\n", count, count, sessionWord, repoCount, repoWord)
	} else {
		// Single repository
		repoName := sessions[0].RepositoryName
		fmt.Printf("Showing %d of %d active %s in %s\n", count, count, sessionWord, repoName)
	}
}

func printRepositoryViewSessions(sessions []config.SessionMetadata, terminalWidth int) {
	// Calculate column widths with new aesthetic approach
	widths := calculateAestheticRepositoryWidths(terminalWidth)

	// Create properly sized and underlined header columns
	idHeader := underlineText(padString("ID", widths.Issue))
	titleHeader := underlineText(padString("TITLE", widths.Title))
	statusHeader := underlineText(padString("STATUS", widths.Status))
	updatedHeader := underlineText(padString("UPDATED", widths.LastActivity))

	// Print header
	fmt.Printf("%s %s %s %s\n", idHeader, titleHeader, statusHeader, updatedHeader)

	// Print sessions
	for _, session := range sessions {
		lastActivity := formatRelativeTime(session.LastActivity)
		// Pad first, then colorize to avoid ANSI code alignment issues
		paddedID := fmt.Sprintf("%-*s", widths.Issue, session.NamespacedID)
		coloredID := colorizeID(paddedID)
		fmt.Printf("%s %-*s %-*s %-*s\n",
			coloredID,
			widths.Title, tui.TruncateString(session.IssueTitle, widths.Title),
			widths.Status, session.Status,
			widths.LastActivity, lastActivity)
	}
}

func printGlobalViewSessions(sessions []config.SessionMetadata, terminalWidth int) {
	// Calculate column widths with new aesthetic approach
	widths := calculateAestheticGlobalWidths(terminalWidth)

	// Create properly sized and underlined header columns
	idHeader := underlineText(padString("ID", widths.Issue))
	titleHeader := underlineText(padString("TITLE", widths.Title))
	repoHeader := underlineText(padString("REPOSITORY", widths.Repository))
	statusHeader := underlineText(padString("STATUS", widths.Status))
	updatedHeader := underlineText(padString("UPDATED", widths.LastActivity))

	// Print header
	fmt.Printf("%s %s %s %s %s\n", idHeader, titleHeader, repoHeader, statusHeader, updatedHeader)

	// Print sessions
	for _, session := range sessions {
		lastActivity := formatRelativeTime(session.LastActivity)
		// Pad first, then colorize to avoid ANSI code alignment issues
		paddedID := fmt.Sprintf("%-*s", widths.Issue, session.NamespacedID)
		coloredID := colorizeID(paddedID)
		fmt.Printf("%s %-*s %-*s %-*s %-*s\n",
			coloredID,
			widths.Title, tui.TruncateString(session.IssueTitle, widths.Title),
			widths.Repository, tui.TruncateString(session.RepositoryName, widths.Repository),
			widths.Status, session.Status,
			widths.LastActivity, lastActivity)
	}
}

func calculateAestheticRepositoryWidths(terminalWidth int) tui.ColumnWidths {
	// Reserve space for "UPDATED" column (estimate ~20 chars for relative time like "about 2 weeks ago")
	updatedWidth := 20
	// Account for spaces between columns (3 spaces between 4 columns)
	spacingWidth := 3
	availableWidth := terminalWidth - updatedWidth - spacingWidth

	// Allocate remaining width
	const (
		minIssue  = 20
		minTitle  = 30
		minStatus = 10
	)

	widths := tui.ColumnWidths{
		Issue:        minIssue,
		Status:       minStatus,
		LastActivity: updatedWidth,
	}

	// Give remaining width to title
	remainingWidth := availableWidth - minIssue - minStatus - 2 // 2 more spaces
	if remainingWidth > minTitle {
		widths.Title = remainingWidth
	} else {
		widths.Title = minTitle
	}

	return widths
}

func calculateAestheticGlobalWidths(terminalWidth int) tui.ColumnWidths {
	// Reserve space for "UPDATED" column
	updatedWidth := 20
	// Account for spaces between columns (4 spaces between 5 columns)
	spacingWidth := 4
	availableWidth := terminalWidth - updatedWidth - spacingWidth

	// Allocate remaining width
	const (
		minIssue      = 20
		minTitle      = 25
		minRepository = 15
		minStatus     = 10
	)

	widths := tui.ColumnWidths{
		Issue:        minIssue,
		Repository:   minRepository,
		Status:       minStatus,
		LastActivity: updatedWidth,
	}

	// Give remaining width to title
	remainingWidth := availableWidth - minIssue - minRepository - minStatus - 3 // 3 more spaces
	if remainingWidth > minTitle {
		widths.Title = remainingWidth
	} else {
		widths.Title = minTitle
	}

	return widths
}

var (
	// Header style: dark gray, underlined, bold
	headerStyle = lipgloss.NewStyle().
			Underline(true).
			Bold(true).
			Foreground(lipgloss.Color("#6C7086")) // Dark gray

	// ID style: green color for work item IDs
	idStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")) // Green (matching existing accent color)
)

func underlineText(text string) string {
	return headerStyle.Render(text)
}

func colorizeID(id string) string {
	return idStyle.Render(id)
}

func padString(text string, width int) string {
	// Pad string to exact width, left-aligned
	if len(text) >= width {
		return text[:width]
	}
	return text + strings.Repeat(" ", width-len(text))
}

func formatRelativeTime(timeStr string) string {
	if timeStr == "" {
		return "unknown"
	}

	// Parse the timestamp (RFC3339 format)
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		// Fallback to just showing date part if parsing fails
		if len(timeStr) >= 10 {
			return timeStr[:10]
		}
		return timeStr
	}

	now := time.Now()
	duration := now.Sub(t)

	// Format like GitHub: "about X ago" or "X ago"
	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case duration < 7*24*time.Hour:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case duration < 30*24*time.Hour:
		weeks := int(duration.Hours() / (24 * 7))
		if weeks == 1 {
			return "about 1 week ago"
		}
		return fmt.Sprintf("about %d weeks ago", weeks)
	case duration < 365*24*time.Hour:
		months := int(duration.Hours() / (24 * 30))
		if months == 1 {
			return "about 1 month ago"
		}
		return fmt.Sprintf("about %d months ago", months)
	default:
		years := int(duration.Hours() / (24 * 365))
		if years == 1 {
			return "about 1 year ago"
		}
		return fmt.Sprintf("about %d years ago", years)
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatLastActivity(timeStr string) string {
	// This is a simplified version - in real implementation
	// you'd parse the RFC3339 timestamp and format it properly
	if len(timeStr) < 10 {
		return timeStr
	}
	return timeStr[:10] // Just show date part
}

// getTerminalWidth returns the width of the terminal, defaulting to 80 if unable to detect
func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80 // Default width if we can't detect
	}
	return width
}

// shouldUseGlobalView determines if we should show the global view based on session diversity
func shouldUseGlobalView(sessions []config.SessionMetadata) bool {
	if len(sessions) == 0 {
		return false
	}

	// Check if all sessions are from the same repository
	firstRepo := sessions[0].RepositoryName
	for _, session := range sessions[1:] {
		if session.RepositoryName != firstRepo {
			return true // Multiple repos, use global view
		}
	}

	return false // All sessions from same repo, use repository view
}

// generateSeparatorLine creates a separator line for the given width
func generateSeparatorLine(width int) string {
	if width > 100 {
		width = 100 // Cap the separator line length
	}
	separator := ""
	for i := 0; i < width; i++ {
		separator += "="
	}
	return separator
}
