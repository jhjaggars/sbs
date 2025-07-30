package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"sbs/pkg/config"
	"sbs/pkg/issue"
	"sbs/pkg/tmux"
)

var attachCmd = &cobra.Command{
	Use:   "attach <issue-number>",
	Short: "Attach to an existing work session",
	Long: `Attach to the tmux session for the specified issue number.
If the session doesn't exist, an error will be returned.`,
	Args: cobra.ExactArgs(1),
	RunE: runAttach,
}

func init() {
	rootCmd.AddCommand(attachCmd)
}

func runAttach(cmd *cobra.Command, args []string) error {
	issueNumberStr := args[0]
	issueNumber, err := strconv.Atoi(issueNumberStr)
	if err != nil {
		return fmt.Errorf("invalid issue number: %s", issueNumberStr)
	}
	
	// Load sessions
	sessions, err := config.LoadSessions()
	if err != nil {
		return fmt.Errorf("failed to load sessions: %w", err)
	}
	
	// Find session
	issueTracker := issue.NewTracker(cfg)
	session := issueTracker.FindSessionByIssue(sessions, issueNumber)
	if session == nil {
		return fmt.Errorf("no session found for issue #%d", issueNumber)
	}
	
	// Check if tmux session exists
	tmuxManager := tmux.NewManager()
	exists, err := tmuxManager.SessionExists(session.TmuxSession)
	if err != nil {
		return fmt.Errorf("failed to check tmux session: %w", err)
	}
	
	if !exists {
		return fmt.Errorf("tmux session %s does not exist", session.TmuxSession)
	}
	
	// Update last activity
	sessions = issueTracker.UpdateSessionActivity(sessions, issueNumber)
	if err := config.SaveSessions(sessions); err != nil {
		// Don't fail if we can't save - just log
		fmt.Printf("Warning: failed to update session activity: %v\n", err)
	}
	
	fmt.Printf("Attaching to session for issue #%d...\n", issueNumber)
	return tmuxManager.AttachToSession(session.TmuxSession)
}