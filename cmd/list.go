package cmd

import (
	"fmt"
	"os"

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

	// Get terminal width
	terminalWidth := getTerminalWidth()

	// Determine if we should use global view (if sessions from multiple repos)
	useGlobalView := shouldUseGlobalView(sessions)

	// Calculate responsive column widths
	var widths tui.ColumnWidths
	var headerRow string

	if useGlobalView {
		widths = tui.CalculateGlobalViewWidths(terminalWidth)
		headerRow = tui.FormatGlobalViewHeader(widths)
	} else {
		widths = tui.CalculateRepositoryViewWidths(terminalWidth)
		headerRow = tui.FormatRepositoryViewHeader(widths)
	}

	// Print header
	fmt.Println(headerRow)
	fmt.Println(generateSeparatorLine(terminalWidth))

	// Print sessions
	for _, session := range sessions {
		var row string
		lastActivity := formatLastActivity(session.LastActivity)

		if useGlobalView {
			row = tui.FormatGlobalViewRow(widths,
				session.NamespacedID,
				session.IssueTitle,
				session.RepositoryName,
				session.Branch,
				session.Status,
				lastActivity,
			)
		} else {
			row = tui.FormatRepositoryViewRow(widths,
				session.NamespacedID,
				session.IssueTitle,
				session.Branch,
				session.Status,
				lastActivity,
			)
		}

		fmt.Println(row)
	}

	return nil
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
