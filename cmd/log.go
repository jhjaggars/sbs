package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"sbs/pkg/config"
	"sbs/pkg/tui"
)

var logCmd = &cobra.Command{
	Use:   "log <work-item-id>",
	Short: "Execute loghook script for a work session",
	Long: `Execute the .sbs/loghook script for the specified work item once and display the output.

Work item ID formats:
  sbs log 123         # Primary work type
  sbs log test:my-test  # Test work type

The loghook script is executed from the session's worktree directory with a 10-second timeout.`,
	Args: cobra.ExactArgs(1),
	RunE: runLog,
}

func init() {
	rootCmd.AddCommand(logCmd)
}

func runLog(cmd *cobra.Command, args []string) error {
	workItemID := args[0]

	// Load sessions
	sessions, err := config.LoadSessions()
	if err != nil {
		return fmt.Errorf("failed to load sessions: %w", err)
	}

	// Find session by namespaced ID
	var session *config.SessionMetadata
	for _, s := range sessions {
		if s.NamespacedID == workItemID {
			session = &s
			break
		}
	}
	if session == nil {
		return fmt.Errorf("no session found for work item %s", workItemID)
	}

	// Execute the loghook script
	output, err := tui.ExecuteLoghookScript(*session)
	if err != nil {
		// Print any output we got even if there was an error
		if output != "" {
			fmt.Print(output)
		}
		fmt.Fprintf(os.Stderr, "Error executing loghook script: %v\n", err)
		return err
	}

	// Print the output
	fmt.Print(output)
	return nil
}
