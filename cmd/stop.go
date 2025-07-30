package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"work-orchestrator/pkg/config"
	"work-orchestrator/pkg/issue"
	"work-orchestrator/pkg/sandbox"
	"work-orchestrator/pkg/tmux"
)

var stopCmd = &cobra.Command{
	Use:   "stop <issue-number>",
	Short: "Stop a work session",
	Long: `Stop the tmux session for the specified issue.
The worktree and session metadata are preserved.`,
	Args: cobra.ExactArgs(1),
	RunE: runStop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop(cmd *cobra.Command, args []string) error {
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
	
	// Stop tmux session
	tmuxManager := tmux.NewManager()
	exists, err := tmuxManager.SessionExists(session.TmuxSession)
	if err != nil {
		return fmt.Errorf("failed to check tmux session: %w", err)
	}
	
	if exists {
		if err := tmuxManager.KillSession(session.TmuxSession); err != nil {
			return fmt.Errorf("failed to kill tmux session: %w", err)
		}
		fmt.Printf("Stopped tmux session: %s\n", session.TmuxSession)
	} else {
		fmt.Printf("Tmux session %s was not running\n", session.TmuxSession)
	}
	
	// Stop sandbox if it exists
	sandboxManager := sandbox.NewManager()
	sandboxName := session.SandboxName
	if sandboxName == "" {
		// For backward compatibility, generate sandbox name
		sandboxName = sandboxManager.GetSandboxName(issueNumber)
	}
	
	sandboxExists, err := sandboxManager.SandboxExists(sandboxName)
	if err != nil {
		fmt.Printf("Warning: could not check sandbox %s: %v\n", sandboxName, err)
	} else if sandboxExists {
		if err := sandboxManager.DeleteSandbox(sandboxName); err != nil {
			fmt.Printf("Warning: failed to delete sandbox %s: %v\n", sandboxName, err)
		} else {
			fmt.Printf("Deleted sandbox: %s\n", sandboxName)
		}
	} else {
		fmt.Printf("Sandbox %s was not running\n", sandboxName)
	}
	
	// Update session status
	for i, s := range sessions {
		if s.IssueNumber == issueNumber {
			sessions[i].Status = "stopped"
			break
		}
	}
	
	// Save updated sessions
	if err := config.SaveSessions(sessions); err != nil {
		return fmt.Errorf("failed to save sessions: %w", err)
	}
	
	fmt.Printf("Session for issue #%d stopped. Worktree preserved at: %s\n", 
		issueNumber, session.WorktreePath)
	
	return nil
}