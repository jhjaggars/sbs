package cmd

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"sbs/pkg/config"
	"sbs/pkg/git"
	"sbs/pkg/issue"
	"sbs/pkg/repo"
	"sbs/pkg/tmux"
	"sbs/pkg/tui"
)

var startCmd = &cobra.Command{
	Use:   "start [issue-number]",
	Short: "Start a new issue work environment",
	Long: `Create or resume a work environment for a GitHub issue.

When run with an issue number:
  sbs start 123

When run without arguments, launches interactive issue selection:
  sbs start

This command will:
1. Create/switch to an issue branch (issue-<number>-<slug>)
2. Create/use a worktree in ~/.work-issue-worktrees/
3. Create/attach to a tmux session (work-issue-<number>)
4. Launch work-issue.sh in the session`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolP("resume", "r", false, "Resume existing session without launching work-issue.sh")
	startCmd.Flags().String("command", "", "Custom command to run in tmux session")
	startCmd.Flags().Bool("no-command", false, "Start session without executing any command")
}

func runStart(cmd *cobra.Command, args []string) error {
	resume, _ := cmd.Flags().GetBool("resume")
	customCommand, _ := cmd.Flags().GetString("command")
	noCommand, _ := cmd.Flags().GetBool("no-command")

	// Initialize repository context first (required for both modes)
	repoManager := repo.NewManager()
	currentRepo, err := repoManager.DetectCurrentRepository()
	if err != nil {
		return fmt.Errorf("must be run from within a git repository: %w", err)
	}

	// Load repository-aware configuration
	repoConfig, err := config.LoadConfigWithRepository(currentRepo.Root)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine issue number - either from args or interactive selection
	var issueNumber int
	if len(args) == 0 {
		// No arguments provided - launch interactive issue selection
		selectedIssue, err := runInteractiveIssueSelection()
		if err != nil {
			return fmt.Errorf("failed to select issue: %w", err)
		}
		if selectedIssue == nil {
			// User quit the selection
			fmt.Println("Issue selection cancelled.")
			return nil
		}
		issueNumber = selectedIssue.Number
		fmt.Printf("Selected issue #%d: %s\n", selectedIssue.Number, selectedIssue.Title)
	} else {
		// Issue number provided as argument
		issueNumberStr := args[0]
		var err error
		issueNumber, err = strconv.Atoi(issueNumberStr)
		if err != nil {
			return fmt.Errorf("invalid issue number: %s", issueNumberStr)
		}
	}

	// Initialize managers
	gitManager, err := git.NewManager(currentRepo.Root)
	if err != nil {
		return fmt.Errorf("failed to initialize git manager: %w", err)
	}

	tmuxManager := tmux.NewManager()
	issueTracker := issue.NewTracker(repoConfig)

	// Load global sessions
	sessions, err := config.LoadSessions()
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

	// Save updated sessions to global location
	if err := config.SaveSessions(sessions); err != nil {
		return fmt.Errorf("failed to save sessions: %w", err)
	}

	// Execute command in session unless resuming
	if !resume {
		// Determine what command to execute based on precedence:
		// 1. Command-line flags (--command, --no-command)
		// 2. Repository config
		// 3. Global config
		// 4. Default behavior (work-issue.sh)

		if noCommand {
			// Explicitly requested no command execution
			fmt.Printf("Session started without executing any command.\n")
		} else if customCommand != "" {
			// Custom command from command line
			fmt.Printf("Executing custom command in session: %s\n", customCommand)
			if err := tmuxManager.ExecuteCommand(session.Name, customCommand, nil); err != nil {
				fmt.Printf("Warning: Failed to execute custom command: %v\n", err)
			}
		} else if repoConfig.NoCommand {
			// Repository config specifies no command
			fmt.Printf("Session started without executing any command (repository config).\n")
		} else if repoConfig.TmuxCommand != "" {
			// Repository config specifies custom command
			fmt.Printf("Executing repository command in session: %s\n", repoConfig.TmuxCommand)
			if err := tmuxManager.ExecuteCommand(session.Name, repoConfig.TmuxCommand, repoConfig.TmuxCommandArgs); err != nil {
				fmt.Printf("Warning: Failed to execute repository command: %v\n", err)
			}
		} else {
			// Default behavior - execute work-issue.sh
			fmt.Printf("Starting work-issue.sh in session...\n")
			if err := tmuxManager.StartWorkIssue(session.Name, issueNumber, repoConfig.WorkIssueScript); err != nil {
				fmt.Printf("Warning: Failed to start work-issue.sh: %v\n", err)
			}
		}
	}

	fmt.Printf("\nWork environment ready! Use 'sbs attach %d' to connect.\n", issueNumber)
	return nil
}

// runInteractiveIssueSelection launches the TUI for issue selection
func runInteractiveIssueSelection() (*issue.Issue, error) {
	// Initialize GitHub client
	githubClient := issue.NewGitHubClient()

	// Create and run the issue selection TUI
	model := tui.NewIssueSelectModel(githubClient)

	program := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run issue selection interface: %w", err)
	}

	// Extract the result from the final model
	issueSelectModel, ok := finalModel.(*tui.IssueSelectModel)
	if !ok {
		return nil, fmt.Errorf("unexpected model type returned from TUI")
	}

	// Check if user quit or selected an issue
	if issueSelectModel.IsQuit() {
		return nil, nil // User quit - this is not an error
	}

	selectedIssue := issueSelectModel.GetSelectedIssue()
	if selectedIssue == nil {
		return nil, fmt.Errorf("no issue was selected")
	}

	return selectedIssue, nil
}
