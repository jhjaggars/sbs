package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"work-orchestrator/pkg/config"
	"work-orchestrator/pkg/git"
	"work-orchestrator/pkg/issue"
	"work-orchestrator/pkg/repo"
	"work-orchestrator/pkg/tmux"
)

var startCmd = &cobra.Command{
	Use:   "start <issue-number>",
	Short: "Start a new issue work environment",
	Long: `Create or resume a work environment for a GitHub issue.
This command will:
1. Create/switch to an issue branch (issue-<number>-<slug>)
2. Create/use a worktree in ~/.work-issue-worktrees/
3. Create/attach to a tmux session (work-issue-<number>)
4. Launch work-issue.sh in the session`,
	Args: cobra.ExactArgs(1),
	RunE: runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolP("resume", "r", false, "Resume existing session without launching work-issue.sh")
}

func runStart(cmd *cobra.Command, args []string) error {
	issueNumberStr := args[0]
	issueNumber, err := strconv.Atoi(issueNumberStr)
	if err != nil {
		return fmt.Errorf("invalid issue number: %s", issueNumberStr)
	}
	
	resume, _ := cmd.Flags().GetBool("resume")
	
	// Initialize repository context
	repoManager := repo.NewManager()
	currentRepo, err := repoManager.DetectCurrentRepository()
	if err != nil {
		return fmt.Errorf("must be run from within a git repository: %w", err)
	}
	
	// Initialize managers
	gitManager, err := git.NewManager(currentRepo.Root)
	if err != nil {
		return fmt.Errorf("failed to initialize git manager: %w", err)
	}
	
	tmuxManager := tmux.NewManager()
	issueTracker := issue.NewTracker(cfg)
	
	// Load repository-specific sessions
	sessionsPath := currentRepo.GetSessionsPath()
	sessions, err := config.LoadSessionsFromPath(sessionsPath)
	if err != nil {
		return fmt.Errorf("failed to load sessions: %w", err)
	}
	
	// Check if session already exists
	existingSession := issueTracker.FindSessionByIssue(sessions, issueNumber)
	if existingSession != nil {
		fmt.Printf("Found existing session for issue #%d\n", issueNumber)
		
		// Check if tmux session exists
		sessionExists, err := tmuxManager.SessionExists(existingSession.TmuxSession)
		if err != nil {
			return fmt.Errorf("failed to check tmux session: %w", err)
		}
		
		if sessionExists {
			fmt.Printf("Attaching to existing tmux session: %s\n", existingSession.TmuxSession)
			return tmuxManager.AttachToSession(existingSession.TmuxSession)
		} else {
			fmt.Printf("Tmux session not found, recreating...\n")
		}
	}
	
	// Get issue information from GitHub using gh command
	githubIssue, err := issueTracker.GetIssue(issueNumber)
	if err != nil {
		fmt.Printf("Warning: Could not fetch issue from GitHub: %v\n", err)
		githubIssue = &issue.Issue{
			Number: issueNumber,
			Title:  fmt.Sprintf("Issue #%d", issueNumber),
			State:  "unknown",
		}
	}
	
	fmt.Printf("Working on issue #%d: %s\n", githubIssue.Number, githubIssue.Title)
	
	// Create or switch to issue branch
	branch, err := gitManager.CreateIssueBranch(issueNumber, githubIssue.Title)
	if err != nil {
		return fmt.Errorf("failed to create issue branch: %w", err)
	}
	fmt.Printf("Using branch: %s\n", branch)
	
	// Create worktree using repository-aware path
	worktreePath := currentRepo.GetWorktreePath(issueNumber)
	if err := gitManager.CreateWorktree(branch, worktreePath); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}
	fmt.Printf("Worktree created at: %s\n", worktreePath)
	
	// Create tmux session with repository-scoped name
	tmuxSessionName := currentRepo.GetTmuxSessionName(issueNumber)
	session, err := tmuxManager.CreateSession(issueNumber, worktreePath, tmuxSessionName)
	if err != nil {
		return fmt.Errorf("failed to create tmux session: %w", err)
	}
	fmt.Printf("Tmux session created: %s\n", session.Name)
	
	// Get repository-scoped sandbox name
	sandboxName := currentRepo.GetSandboxName(issueNumber)
	
	// Create or update session metadata
	sessionMetadata := issueTracker.CreateSessionMetadata(
		issueNumber, githubIssue, branch, worktreePath, session.Name, sandboxName, currentRepo.Name, currentRepo.Root)
	
	// Update sessions list
	if existingSession != nil {
		// Update existing session
		for i, s := range sessions {
			if s.IssueNumber == issueNumber {
				sessions[i] = *sessionMetadata
				break
			}
		}
	} else {
		// Add new session
		sessions = append(sessions, *sessionMetadata)
	}
	
	// Save updated sessions to repository-specific location
	if err := config.SaveSessionsToPath(sessions, sessionsPath); err != nil {
		return fmt.Errorf("failed to save sessions: %w", err)
	}
	
	// Launch work-issue.sh unless resuming
	if !resume {
		fmt.Printf("Starting work-issue.sh in session...\n")
		if err := tmuxManager.StartWorkIssue(session.Name, issueNumber, cfg.WorkIssueScript); err != nil {
			fmt.Printf("Warning: Failed to start work-issue.sh: %v\n", err)
		}
	}
	
	fmt.Printf("\nWork environment ready! Use 'work-orchestrator attach %d' to connect.\n", issueNumber)
	return nil
}