package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"sbs/pkg/config"
	"sbs/pkg/git"
	"sbs/pkg/repo"
	"sbs/pkg/sandbox"
	"sbs/pkg/tmux"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up stale sessions and worktrees",
	Long: `Remove stale sessions and their associated worktrees.
A session is considered stale if its tmux session no longer exists.`,
	RunE: runClean,
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().BoolP("dry-run", "n", false, "Show what would be cleaned without actually doing it")
	cleanCmd.Flags().BoolP("force", "f", false, "Force cleanup without confirmation")

	// Enhanced cleanup modes
	cleanCmd.Flags().Bool("stale", false, "Clean only stale sessions")
	cleanCmd.Flags().Bool("orphaned", false, "Clean orphaned resources")
	cleanCmd.Flags().Bool("branches", false, "Clean orphaned branches")
	cleanCmd.Flags().Bool("all", false, "Clean all resource types")
}

// CleanupMode represents the type of cleanup to perform
type CleanupMode int

const (
	// Default cleanup mode (backwards compatible)
	CleanupModeDefault CleanupMode = iota
	// Clean only stale sessions
	CleanupModeStale
	// Clean orphaned resources
	CleanupModeOrphaned
	// Clean orphaned branches
	CleanupModeBranches
	// Clean all resource types
	CleanupModeAll
	// Combined cleanup modes
	CleanupModeStaleAndBranches
)

// determineCleanupMode determines the cleanup mode based on flags
func determineCleanupMode(stale, orphaned, branches, all bool) CleanupMode {
	// All mode overrides everything
	if all {
		return CleanupModeAll
	}

	// Check for combinations
	if stale && branches {
		return CleanupModeStaleAndBranches
	}

	// Individual modes
	if stale {
		return CleanupModeStale
	}
	if orphaned {
		return CleanupModeOrphaned
	}
	if branches {
		return CleanupModeBranches
	}

	// Default mode (backwards compatible)
	return CleanupModeDefault
}

// resolveSandboxName attempts to get the correct sandbox name for a session
func resolveSandboxName(session config.SessionMetadata, sandboxManager *sandbox.Manager) string {
	// If session has a stored sandbox name, use it
	if session.SandboxName != "" {
		return session.SandboxName
	}

	// Try to get repository-aware sandbox name if we have repository info
	if session.RepositoryName != "" {
		if session.SandboxName != "" {
			return session.SandboxName
		}
		return fmt.Sprintf("sbs-%s", session.NamespacedID)
	}

	// Use stored sandbox name if available
	if session.SandboxName != "" {
		return session.SandboxName
	}
	return fmt.Sprintf("work-issue-%s", session.NamespacedID)
}

func runClean(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")

	// Get cleanup mode flags
	staleOnly, _ := cmd.Flags().GetBool("stale")
	orphanedOnly, _ := cmd.Flags().GetBool("orphaned")
	branchesOnly, _ := cmd.Flags().GetBool("branches")
	allResources, _ := cmd.Flags().GetBool("all")

	// Determine cleanup mode
	cleanupMode := determineCleanupMode(staleOnly, orphanedOnly, branchesOnly, allResources)

	// Execute cleanup based on mode
	return executeCleanup(cleanupMode, dryRun, force)
}

// executeCleanup performs the actual cleanup based on the specified mode
func executeCleanup(mode CleanupMode, dryRun, force bool) error {
	switch mode {
	case CleanupModeDefault:
		return executeDefaultCleanup(dryRun, force)
	case CleanupModeStale:
		return executeStaleCleanup(dryRun, force)
	case CleanupModeBranches:
		return executeBranchCleanup(dryRun, force)
	case CleanupModeAll:
		return executeComprehensiveCleanup(dryRun, force)
	case CleanupModeStaleAndBranches:
		// Execute both stale and branch cleanup
		if err := executeStaleCleanup(dryRun, force); err != nil {
			return err
		}
		return executeBranchCleanup(dryRun, force)
	default:
		return executeDefaultCleanup(dryRun, force)
	}
}

// executeDefaultCleanup performs the original cleanup behavior (backwards compatible)
func executeDefaultCleanup(dryRun, force bool) error {
	// Load all sessions from all repositories
	sessions, err := config.LoadAllRepositorySessions()
	if err != nil {
		return fmt.Errorf("failed to load sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions to clean.")
		return nil
	}

	// Initialize managers
	tmuxManager := tmux.NewManager()
	sandboxManager := sandbox.NewManager()

	var staleSessions []config.SessionMetadata
	var activeSessions []config.SessionMetadata

	// Check which sessions are stale
	for _, session := range sessions {
		exists, err := tmuxManager.SessionExists(session.TmuxSession)
		if err != nil {
			fmt.Printf("Warning: could not check session %s: %v\n", session.TmuxSession, err)
			activeSessions = append(activeSessions, session)
			continue
		}

		if !exists {
			staleSessions = append(staleSessions, session)
		} else {
			activeSessions = append(activeSessions, session)
		}
	}

	if len(staleSessions) == 0 {
		fmt.Println("No stale sessions found.")
		return nil
	}

	// Show what will be cleaned
	fmt.Printf("Found %d stale session(s):\n", len(staleSessions))
	for _, session := range staleSessions {
		fmt.Printf("  Work Item %s: %s\n", session.NamespacedID, session.IssueTitle)
		fmt.Printf("    Worktree: %s\n", session.WorktreePath)
		fmt.Printf("    Tmux Session: %s\n", session.TmuxSession)
		sandboxName := resolveSandboxName(session, sandboxManager)
		fmt.Printf("    Sandbox: %s\n", sandboxName)
	}

	if dryRun {
		fmt.Println("\nDry run - no changes made.")
		return nil
	}

	// Confirm unless forced
	if !force {
		fmt.Print("\nProceed with cleanup? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Cleanup cancelled.")
			return nil
		}
	}

	// Clean up stale sessions
	fmt.Println("\nCleaning up stale sessions...")
	for _, session := range staleSessions {
		fmt.Printf("Cleaning up work item %s...\n", session.NamespacedID)

		// Remove worktree using unified git manager approach
		if err := removeWorktreeForCleanup(session.WorktreePath); err != nil {
			fmt.Printf("  Warning: failed to remove worktree %s: %v\n", session.WorktreePath, err)
		} else {
			fmt.Printf("  Removed worktree: %s\n", session.WorktreePath)
		}

		// Clean up sandbox
		sandboxName := resolveSandboxName(session, sandboxManager)

		sandboxExists, err := sandboxManager.SandboxExists(sandboxName)
		if err != nil {
			fmt.Printf("  Warning: could not check sandbox %s: %v\n", sandboxName, err)
		} else if sandboxExists {
			fmt.Printf("  Attempting to delete sandbox: %s\n", sandboxName)
			if err := sandboxManager.DeleteSandbox(sandboxName); err != nil {
				fmt.Printf("  Warning: failed to delete sandbox %s: %v\n", sandboxName, err)
			} else {
				fmt.Printf("  Removed sandbox: %s\n", sandboxName)
			}
		} else {
			fmt.Printf("  Sandbox already gone: %s\n", sandboxName)
		}
	}

	// Save updated sessions back to global location
	if err := config.SaveSessions(activeSessions); err != nil {
		fmt.Printf("Warning: failed to save updated sessions: %v\n", err)
	}

	fmt.Printf("\nCleanup complete. Removed %d stale session(s).\n", len(staleSessions))
	return nil
}

// executeStaleCleanup performs cleanup of stale sessions only
func executeStaleCleanup(dryRun, force bool) error {
	fmt.Println("Cleaning up stale sessions only...")
	return executeDefaultCleanup(dryRun, force)
}

// executeBranchCleanup performs cleanup of orphaned branches
func executeBranchCleanup(dryRun, force bool) error {
	fmt.Println("Cleaning up orphaned branches...")

	// Load sessions to determine active issues
	sessions, err := config.LoadAllRepositorySessions()
	if err != nil {
		return fmt.Errorf("failed to load sessions: %w", err)
	}

	// Get active issue numbers with robust active session detection
	tmuxManager := tmux.NewManager()
	activeWorkItems := make([]string, 0, len(sessions))
	for _, session := range sessions {
		// Only include active sessions
		if session.Status == "active" {
			// Optional: verify tmux session actually exists for more robust detection
			if exists, _ := tmuxManager.SessionExists(session.TmuxSession); exists {
				activeWorkItems = append(activeWorkItems, session.NamespacedID)
			}
		}
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

	// Find orphaned branches
	orphanedBranches, err := gitManager.FindOrphanedIssueBranches(activeWorkItems)
	if err != nil {
		return fmt.Errorf("failed to find orphaned branches: %w", err)
	}

	if len(orphanedBranches) == 0 {
		fmt.Println("No orphaned branches found.")
		return nil
	}

	// Show what will be cleaned
	fmt.Printf("Found %d orphaned branch(es):\n", len(orphanedBranches))
	for _, branch := range orphanedBranches {
		fmt.Printf("  %s\n", branch)
	}

	if dryRun {
		fmt.Println("\nDry run - no changes made.")
		return nil
	}

	// Confirm unless forced
	if !force {
		fmt.Print("\nProceed with branch cleanup? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Branch cleanup cancelled.")
			return nil
		}
	}

	// Delete orphaned branches
	results, err := gitManager.DeleteMultipleBranches(orphanedBranches, false)
	if err != nil {
		return fmt.Errorf("failed to delete branches: %w", err)
	}

	// Report results
	successCount := 0
	for _, result := range results {
		if result.Success {
			fmt.Printf("  Deleted branch: %s\n", result.BranchName)
			successCount++
		} else {
			fmt.Printf("  Failed to delete branch %s: %s\n", result.BranchName, result.Message)
		}
	}

	fmt.Printf("\nBranch cleanup complete. Removed %d branch(es).\n", successCount)
	return nil
}

// executeComprehensiveCleanup performs cleanup of all resource types
func executeComprehensiveCleanup(dryRun, force bool) error {
	fmt.Println("Performing comprehensive cleanup of all resources...")

	// Execute stale session cleanup
	if err := executeStaleCleanup(dryRun, force); err != nil {
		fmt.Printf("Warning: stale session cleanup failed: %v\n", err)
	}

	// Execute branch cleanup
	if err := executeBranchCleanup(dryRun, force); err != nil {
		fmt.Printf("Warning: branch cleanup failed: %v\n", err)
	}

	fmt.Println("Comprehensive cleanup complete.")
	return nil
}

// removeWorktreeForCleanup safely removes a worktree using git commands
func removeWorktreeForCleanup(worktreePath string) error {
	// Validate that this looks like a worktree path to avoid accidental deletion
	if !strings.Contains(worktreePath, "sbs") && !strings.Contains(worktreePath, "worktree") {
		return fmt.Errorf("path doesn't appear to be a worktree: %s", worktreePath)
	}

	// Check if worktree exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		// Worktree doesn't exist, just prune stale references
		return pruneStaleWorktreesForCleanup()
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

	// Use the enhanced worktree removal method
	return gitManager.RemoveWorktreeForSession(worktreePath)
}

// pruneStaleWorktreesForCleanup prunes stale worktrees during cleanup
func pruneStaleWorktreesForCleanup() error {
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

	// Prune stale worktrees
	return gitManager.PruneStaleWorktrees()
}
