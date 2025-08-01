package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"sbs/pkg/config"
	"sbs/pkg/git"
	"sbs/pkg/issue"
	"sbs/pkg/repo"
	"sbs/pkg/sandbox"
	"sbs/pkg/tmux"
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
	stopCmd.Flags().BoolP("delete-branch", "d", false, "Delete the associated branch when stopping the session")
}

func runStop(cmd *cobra.Command, args []string) error {
	issueNumberStr := args[0]
	issueNumber, err := strconv.Atoi(issueNumberStr)
	if err != nil {
		return fmt.Errorf("invalid issue number: %s", issueNumberStr)
	}

	// Get the delete-branch flag
	deleteBranch, _ := cmd.Flags().GetBool("delete-branch")

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

	// Handle branch deletion if requested
	if deleteBranch {
		if err := deleteBranchForSession(session); err != nil {
			fmt.Printf("Warning: failed to delete branch: %v\n", err)
		} else {
			fmt.Printf("Deleted branch: %s\n", session.Branch)
		}
	}

	fmt.Printf("Session for issue #%d stopped. Worktree preserved at: %s\n",
		issueNumber, session.WorktreePath)

	return nil
}

// deleteBranchForSession deletes the branch associated with a session
func deleteBranchForSession(session *config.SessionMetadata) error {
	if session.Branch == "" {
		return fmt.Errorf("no branch associated with session")
	}

	// Initialize repository manager to get current repo
	repoManager := repo.NewManager()
	currentRepo, err := repoManager.DetectCurrentRepository()
	if err != nil {
		return fmt.Errorf("must be run from within a git repository: %w", err)
	}

	// Initialize git manager
	gitManager, err := git.NewManager(currentRepo.Root)
	if err != nil {
		return fmt.Errorf("failed to initialize git manager: %w", err)
	}

	// Validate branch deletion is safe
	safe, warnings, err := gitManager.ValidateBranchDeletion(session.Branch)
	if err != nil {
		return fmt.Errorf("failed to validate branch deletion: %w", err)
	}

	if !safe {
		return fmt.Errorf("branch deletion not safe: %s", strings.Join(warnings, ", "))
	}

	if len(warnings) > 0 {
		for _, warning := range warnings {
			fmt.Printf("Warning: %s\n", warning)
		}
	}

	// Delete the branch
	err = gitManager.DeleteIssueBranch(session.Branch)
	if err != nil {
		return fmt.Errorf("failed to delete branch %s: %w", session.Branch, err)
	}

	return nil
}
