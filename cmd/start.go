package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"sbs/pkg/config"
	"sbs/pkg/git"
	"sbs/pkg/inputsource"
	"sbs/pkg/issue"
	"sbs/pkg/repo"
	"sbs/pkg/tmux"
	"sbs/pkg/tui"
)

var startCmd = &cobra.Command{
	Use:   "start [work-item-id]",
	Short: "Start a new work environment for any work item",
	Long: `Create or resume a work environment for any work item from configured input sources.

REQUIRED: Work item ID must be in namespaced format (source:id):
  sbs start github:123       # GitHub issue
  sbs start test:quick       # Test work item
  sbs start test:hooks       # Test Claude Code hooks
  sbs start test:sandbox     # Test sandbox integration

When run without arguments, launches interactive work item selection:
  sbs start

This command will:
1. Create/switch to a work item branch (issue-{source}-{id}-{slug})
2. Create/use a worktree in ~/.work-issue-worktrees/
3. Create/attach to a tmux session (work-issue-{source}-{id})
4. Launch work-issue.sh in the session

Input sources are configured via .sbs/input-source.json in your project root.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolP("resume", "r", false, "Resume existing session without launching work-issue.sh")
	startCmd.Flags().String("command", "", "Custom command to run in tmux session")
	startCmd.Flags().Bool("no-command", false, "Start session without executing any command")
	startCmd.Flags().BoolP("verbose", "v", false, "Enable verbose debug output")
}

func runStart(cmd *cobra.Command, args []string) error {
	resume, _ := cmd.Flags().GetBool("resume")
	customCommand, _ := cmd.Flags().GetString("command")
	noCommand, _ := cmd.Flags().GetBool("no-command")
	verbose, _ := cmd.Flags().GetBool("verbose")

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

	// Create input source factory and load project-specific input source
	factory := inputsource.NewInputSourceFactory()
	inputSourceInstance, err := factory.CreateFromProject(currentRepo.Root)
	if err != nil {
		return fmt.Errorf("failed to create input source: %w", err)
	}

	// Load input source configuration for validation settings
	inputSourceConfig, err := config.LoadInputSourceConfig(currentRepo.Root)
	if err != nil {
		return fmt.Errorf("failed to load input source config: %w", err)
	}

	if verbose {
		fmt.Printf("Debug: Using input source type: %s\n", inputSourceInstance.GetType())
	}

	// Parse work item ID - either from args or interactive selection
	var workItem *inputsource.WorkItem

	if len(args) == 0 {
		// No arguments provided - launch interactive work item selection
		selectedWorkItem, err := runInteractiveWorkItemSelection(inputSourceInstance)
		if err != nil {
			return fmt.Errorf("failed to select work item: %w", err)
		}
		if selectedWorkItem == nil {
			// User quit the selection
			fmt.Println("Work item selection cancelled.")
			return nil
		}
		workItem = selectedWorkItem

		fmt.Printf("Selected work item %s: %s\n", workItem.FullID(), workItem.Title)
	} else {
		// Work item ID provided as argument
		workItemIDStr := args[0]

		// Parse the work item ID (requires namespaced format)
		parsedWorkItem, err := inputsource.ParseWorkItemID(workItemIDStr)
		if err != nil {
			return fmt.Errorf("invalid work item ID: %s (%w)", workItemIDStr, err)
		}

		// If the parsed source doesn't match the project's input source, validate it's allowed
		if parsedWorkItem.Source != inputSourceInstance.GetType() {
			if !inputSourceConfig.AllowCrossSource() {
				return fmt.Errorf("work item source '%s' doesn't match project input source '%s'. "+
					"To allow cross-source usage, set allow_cross_source: true in .sbs/input-source.json",
					parsedWorkItem.Source, inputSourceInstance.GetType())
			}
			if verbose {
				fmt.Printf("Debug: Cross-source usage allowed: using %s work item with %s project\n",
					parsedWorkItem.Source, inputSourceInstance.GetType())
			}
		}

		// Fetch the full work item details
		workItem, err = inputSourceInstance.GetWorkItem(parsedWorkItem.ID)
		if err != nil {
			return fmt.Errorf("failed to get work item %s: %w", parsedWorkItem.FullID(), err)
		}
	}

	// Initialize managers
	gitManager, err := git.NewManager(currentRepo.Root)
	if err != nil {
		return fmt.Errorf("failed to initialize git manager: %w", err)
	}

	tmuxManager := tmux.NewManager()
	// issueTracker := issue.NewTracker(repoConfig) // Not needed for new implementation

	// Load global sessions
	sessionsPath, err := config.GetGlobalSessionsPath()
	if err != nil {
		return fmt.Errorf("failed to get sessions path: %w", err)
	}

	sessions, err := config.LoadSessionsFromPath(sessionsPath)
	if err != nil {
		return fmt.Errorf("failed to load sessions: %w", err)
	}

	// Check if session already exists by namespaced ID
	existingSession := findSessionByWorkItem(sessions, workItem)
	if existingSession != nil {
		fmt.Printf("Found existing session for work item %s\n", workItem.FullID())

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

	fmt.Printf("Working on work item %s: %s\n", workItem.FullID(), workItem.Title)

	// Use namespaced branch naming
	branch := workItem.GetBranchName()
	// Create the branch using the new name
	err = createWorkItemBranch(gitManager, branch)
	if verbose {
		fmt.Printf("Debug: Using namespaced branch naming: %s\n", branch)
	}

	if err != nil {
		return fmt.Errorf("failed to create work item branch: %w", err)
	}
	fmt.Printf("Using branch: %s\n", branch)

	// Generate friendly title for sandbox environment
	friendlyTitle := generateWorkItemFriendlyTitle(currentRepo.Name, workItem)
	fmt.Printf("Friendly title: %s\n", friendlyTitle)

	// Create worktree path based on work item
	worktreePath := generateWorkItemWorktreePath(currentRepo, workItem)
	if verbose {
		fmt.Printf("Debug: Creating worktree at path: %s\n", worktreePath)
		fmt.Printf("Debug: Using branch: %s\n", branch)
		fmt.Printf("Debug: Repository root: %s\n", currentRepo.Root)
	}

	if err := gitManager.CreateWorktree(branch, worktreePath); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}
	fmt.Printf("Worktree created at: %s\n", worktreePath)

	// Create environment variables for tmux session
	tmuxEnv := tmux.CreateTmuxEnvironment(friendlyTitle)

	// Create tmux session with work item-specific name
	tmuxSessionName := generateWorkItemTmuxSessionName(currentRepo, workItem)
	session, err := createWorkItemTmuxSession(tmuxManager, workItem, worktreePath, tmuxSessionName, tmuxEnv)
	if err != nil {
		return fmt.Errorf("failed to create tmux session: %w", err)
	}
	fmt.Printf("Tmux session created: %s (SBS_TITLE=%s)\n", session.Name, friendlyTitle)

	// Get work item-specific sandbox name
	sandboxName := generateWorkItemSandboxName(currentRepo, workItem)

	// Create session metadata with input source information
	sessionMetadata := createWorkItemSessionMetadata(workItem, branch, worktreePath, session.Name,
		sandboxName, currentRepo.Name, currentRepo.Root, friendlyTitle)

	// Update sessions list
	if existingSession != nil {
		// Update existing session by namespaced ID
		for i, s := range sessions {
			if s.NamespacedID == workItem.FullID() {
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
			if err := tmuxManager.ExecuteCommand(session.Name, customCommand, nil, tmuxEnv); err != nil {
				fmt.Printf("Warning: Failed to execute custom command: %v\n", err)
			}
		} else if repoConfig.NoCommand {
			// Repository config specifies no command
			fmt.Printf("Session started without executing any command (repository config).\n")
		} else if repoConfig.TmuxCommand != "" {
			// Repository config specifies custom command
			fmt.Printf("Executing repository command in session: %s\n", repoConfig.TmuxCommand)

			// Create substitution map for parameters
			substitutions := map[string]string{
				"$1": workItem.ID,
			}

			if err := tmuxManager.ExecuteCommandWithSubstitution(session.Name, repoConfig.TmuxCommand, repoConfig.TmuxCommandArgs, substitutions, tmuxEnv); err != nil {
				fmt.Printf("Warning: Failed to execute repository command: %v\n", err)
			}
		} else {
			// Default behavior - execute work-issue.sh
			fmt.Printf("Starting work-issue.sh in session...\n")
			// Use 0 as dummy issue number since work-issue.sh will use environment variables for context
			if err := tmuxManager.StartWorkIssue(session.Name, 0, repoConfig.WorkIssueScript, tmuxEnv); err != nil {
				fmt.Printf("Warning: Failed to start work-issue.sh: %v\n", err)
			}
		}
	}

	// Show attach command
	fmt.Printf("\nWork environment ready! Use 'sbs attach %s' to connect.\n", workItem.FullID())
	return nil
}

// runInteractiveWorkItemSelection launches the TUI for work item selection
func runInteractiveWorkItemSelection(inputSource inputsource.InputSource) (*inputsource.WorkItem, error) {
	// For now, fall back to GitHub client for interactive selection
	// TODO: Implement generic work item selection TUI
	if inputSource.GetType() != "github" {
		return nil, fmt.Errorf("interactive selection not yet supported for %s input sources", inputSource.GetType())
	}

	// Use existing GitHub interactive selection as fallback
	githubClient := issue.NewGitHubClient()
	model := tui.NewIssueSelectModel(githubClient)

	program := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run work item selection interface: %w", err)
	}

	issueSelectModel, ok := finalModel.(*tui.IssueSelectModel)
	if !ok {
		return nil, fmt.Errorf("unexpected model type returned from TUI")
	}

	if issueSelectModel.IsQuit() {
		return nil, nil
	}

	selectedIssue := issueSelectModel.GetSelectedIssue()
	if selectedIssue == nil {
		return nil, fmt.Errorf("no work item was selected")
	}

	// Convert GitHub issue to WorkItem
	return &inputsource.WorkItem{
		Source: "github",
		ID:     strconv.Itoa(selectedIssue.Number),
		Title:  selectedIssue.Title,
		State:  selectedIssue.State,
		URL:    selectedIssue.URL,
	}, nil
}

// runInteractiveIssueSelection launches the TUI for issue selection (legacy compatibility)
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

// Helper functions for work item integration

// findSessionByWorkItem finds a session by work item using namespaced ID
func findSessionByWorkItem(sessions []config.SessionMetadata, workItem *inputsource.WorkItem) *config.SessionMetadata {
	// Find by namespaced ID
	for _, session := range sessions {
		if session.NamespacedID == workItem.FullID() {
			return &session
		}
	}

	return nil
}

// createWorkItemBranch creates a branch for a work item using direct git commands
func createWorkItemBranch(gitManager *git.Manager, branchName string) error {
	// Check if branch already exists
	exists, err := gitManager.BranchExists(branchName)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	if exists {
		return nil // Branch already exists
	}

	// Create the branch using the generic branch creation method
	// We use a dummy issue number (0) since the branch name is already constructed
	_, err = gitManager.CreateIssueBranch(0, "")
	if err != nil {
		return fmt.Errorf("failed to create branch %s: %w", branchName, err)
	}

	return nil
}

// generateWorkItemFriendlyTitle creates a friendly title for the work item
func generateWorkItemFriendlyTitle(repoName string, workItem *inputsource.WorkItem) string {
	// Create a consistent format for all work item sources
	title := strings.ReplaceAll(workItem.Title, " ", "-")
	title = strings.ToLower(title)

	// Limit length
	if len(title) > inputsource.MaxFriendlyTitleLength {
		title = title[:inputsource.MaxFriendlyTitleLength]
	}

	return fmt.Sprintf("%s-%s-%s", repoName, workItem.Source, workItem.ID)
}

// generateWorkItemWorktreePath creates a worktree path for the work item
func generateWorkItemWorktreePath(currentRepo *repo.Repository, workItem *inputsource.WorkItem) string {
	// Create a consistent path format for all work item sources
	baseDir := filepath.Dir(currentRepo.GetWorktreePath(1)) // Get the base worktree directory
	return filepath.Join(baseDir, fmt.Sprintf("issue-%s-%s", workItem.Source, workItem.ID))
}

// generateWorkItemTmuxSessionName creates a tmux session name for the work item
func generateWorkItemTmuxSessionName(currentRepo *repo.Repository, workItem *inputsource.WorkItem) string {
	// Create a consistent naming format for all work item sources
	return fmt.Sprintf("work-issue-%s-%s-%s",
		currentRepo.Name, workItem.Source, workItem.ID)
}

// generateWorkItemSandboxName creates a sandbox name for the work item
func generateWorkItemSandboxName(currentRepo *repo.Repository, workItem *inputsource.WorkItem) string {
	// Create a consistent naming format for all work item sources
	return fmt.Sprintf("work-issue-%s-%s-%s",
		currentRepo.Name, workItem.Source, workItem.ID)
}

// createWorkItemTmuxSession creates a tmux session for the work item
func createWorkItemTmuxSession(tmuxManager *tmux.Manager, workItem *inputsource.WorkItem,
	worktreePath, sessionName string, tmuxEnv map[string]string) (*tmux.Session, error) {

	// Use 0 as dummy issue number since all work items are handled generically now
	return tmuxManager.CreateSession(0, worktreePath, sessionName, tmuxEnv)
}

// createWorkItemSessionMetadata creates session metadata for the work item
func createWorkItemSessionMetadata(workItem *inputsource.WorkItem, branch, worktreePath,
	tmuxSession, sandboxName, repoName, repoRoot, friendlyTitle string) *config.SessionMetadata {

	return &config.SessionMetadata{
		IssueTitle:     workItem.Title,
		FriendlyTitle:  friendlyTitle,
		Branch:         branch,
		WorktreePath:   worktreePath,
		TmuxSession:    tmuxSession,
		SandboxName:    sandboxName,
		RepositoryName: repoName,
		RepositoryRoot: repoRoot,
		CreatedAt:      "", // TODO: Set timestamp
		LastActivity:   "", // TODO: Set timestamp
		Status:         "active",
		SourceType:     workItem.Source,
		NamespacedID:   workItem.FullID(),
	}
}
