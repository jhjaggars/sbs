package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"work-orchestrator/pkg/config"
	"work-orchestrator/pkg/tui"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all active work sessions",
	Long: `Display an interactive list of all active work sessions.
Use arrow keys or j/k to navigate, enter to attach to a session.`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolP("plain", "p", false, "Show plain text output instead of interactive TUI")
}

func runList(cmd *cobra.Command, args []string) error {
	plain, _ := cmd.Flags().GetBool("plain")
	
	if plain {
		return runPlainList()
	}
	
	// Launch interactive TUI
	model := tui.NewModel()
	program := tea.NewProgram(model, tea.WithAltScreen())
	
	_, err := program.Run()
	return err
}

func runPlainList() error {
	// Load sessions
	sessions, err := config.LoadSessions()
	if err != nil {
		return fmt.Errorf("failed to load sessions: %w", err)
	}
	
	if len(sessions) == 0 {
		fmt.Println("No active work sessions found.")
		return nil
	}
	
	// Print header
	fmt.Printf("%-6s %-50s %-20s %-10s %-15s\n",
		"Issue", "Title", "Branch", "Status", "Last Activity")
	fmt.Println("================================================================================")
	
	// Print sessions
	for _, session := range sessions {
		fmt.Printf("%-6d %-50s %-20s %-10s %-15s\n",
			session.IssueNumber,
			truncateString(session.IssueTitle, 48),
			truncateString(session.Branch, 18),
			session.Status,
			formatLastActivity(session.LastActivity),
		)
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
	return timeStr[:10] // Just show date part
}