package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"sbs/pkg/config"
	"sbs/pkg/tmux"
)

var attachCmd = &cobra.Command{
	Use:   "attach <work-item-id>",
	Short: "Attach to an existing work session",
	Long: `Attach to the tmux session for the specified work item.
If the session doesn't exist, an error will be returned.

Work item ID formats:
  sbs attach 123         # Primary work type
  sbs attach test:my-test  # Test work type`,
	Args: cobra.ExactArgs(1),
	RunE: runAttach,
}

func init() {
	rootCmd.AddCommand(attachCmd)
}

func runAttach(cmd *cobra.Command, args []string) error {
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
	for i, s := range sessions {
		if s.NamespacedID == workItemID {
			// Update last activity timestamp
			sessions[i].LastActivity = time.Now().Format(time.RFC3339)
			break
		}
	}
	if err := config.SaveSessions(sessions); err != nil {
		// Don't fail if we can't save - just log
		fmt.Printf("Warning: failed to update session activity: %v\n", err)
	}

	// Create environment variables with friendly title if available
	var tmuxEnv map[string]string
	if session.FriendlyTitle != "" {
		tmuxEnv = tmux.CreateTmuxEnvironment(session.FriendlyTitle)
		fmt.Printf("Attaching to session for work item %s (SBS_TITLE=%s)...\n", workItemID, session.FriendlyTitle)
	} else {
		fmt.Printf("Attaching to session for work item %s...\n", workItemID)
	}

	return tmuxManager.AttachToSession(session.TmuxSession, tmuxEnv)
}
