package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"sbs/pkg/config"
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
}

func runClean(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")
	
	// Load all sessions from all repositories
	sessions, err := config.LoadAllRepositorySessions()
	if err != nil {
		return fmt.Errorf("failed to load sessions: %w", err)
	}
	
	if len(sessions) == 0 {
		fmt.Println("No sessions to clean.")
		return nil
	}
	
	// Initialize managers (no git manager needed for basic cleanup)
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
		fmt.Printf("  Issue #%d: %s\n", session.IssueNumber, session.IssueTitle)
		fmt.Printf("    Worktree: %s\n", session.WorktreePath)
		fmt.Printf("    Tmux Session: %s\n", session.TmuxSession)
		sandboxName := session.SandboxName
		if sandboxName == "" {
			sandboxName = sandboxManager.GetSandboxName(session.IssueNumber)
		}
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
		fmt.Printf("Cleaning up issue #%d...\n", session.IssueNumber)
		
		// Remove worktree (direct filesystem removal)
		if _, err := os.Stat(session.WorktreePath); err == nil {
			if err := removeWorktreeDirectory(session.WorktreePath); err != nil {
				fmt.Printf("  Warning: failed to remove worktree %s: %v\n", session.WorktreePath, err)
			} else {
				fmt.Printf("  Removed worktree: %s\n", session.WorktreePath)
			}
		} else {
			fmt.Printf("  Worktree already gone: %s\n", session.WorktreePath)
		}
		
		// Clean up sandbox
		sandboxName := session.SandboxName
		if sandboxName == "" {
			sandboxName = sandboxManager.GetSandboxName(session.IssueNumber)
		}
		
		sandboxExists, err := sandboxManager.SandboxExists(sandboxName)
		if err != nil {
			fmt.Printf("  Warning: could not check sandbox %s: %v\n", sandboxName, err)
		} else if sandboxExists {
			if err := sandboxManager.DeleteSandbox(sandboxName); err != nil {
				fmt.Printf("  Warning: failed to delete sandbox %s: %v\n", sandboxName, err)
			} else {
				fmt.Printf("  Removed sandbox: %s\n", sandboxName)
			}
		} else {
			fmt.Printf("  Sandbox already gone: %s\n", sandboxName)
		}
	}
	
	// Save updated sessions back to their respective locations
	if err := saveSessionsToTheirRepositories(activeSessions); err != nil {
		fmt.Printf("Warning: failed to save updated sessions: %v\n", err)
	}
	
	fmt.Printf("\nCleanup complete. Removed %d stale session(s).\n", len(staleSessions))
	return nil
}

// removeWorktreeDirectory safely removes a worktree directory
func removeWorktreeDirectory(worktreePath string) error {
	// Validate that this looks like a worktree path to avoid accidental deletion
	if !strings.Contains(worktreePath, "work-issue") && !strings.Contains(worktreePath, "worktree") {
		return fmt.Errorf("path doesn't appear to be a worktree: %s", worktreePath)
	}
	
	// Remove the directory
	return os.RemoveAll(worktreePath)
}

// saveSessionsToTheirRepositories saves sessions back to their appropriate locations
func saveSessionsToTheirRepositories(sessions []config.SessionMetadata) error {
	// Group sessions by repository
	sessionsByRepo := make(map[string][]config.SessionMetadata)
	var globalSessions []config.SessionMetadata
	
	for _, session := range sessions {
		if session.RepositoryRoot != "" {
			// Repository-specific session
			key := session.RepositoryRoot
			sessionsByRepo[key] = append(sessionsByRepo[key], session)
		} else {
			// Global session (backward compatibility)
			globalSessions = append(globalSessions, session)
		}
	}
	
	// Save global sessions
	if len(globalSessions) > 0 {
		if err := config.SaveSessions(globalSessions); err != nil {
			return err
		}
	}
	
	// Save repository-specific sessions
	for repoRoot, repoSessions := range sessionsByRepo {
		sessionsPath := filepath.Join(repoRoot, ".sbs", "sessions.json")
		if err := config.SaveSessionsToPath(repoSessions, sessionsPath); err != nil {
			// Don't fail completely if one repository fails
			fmt.Printf("Warning: failed to save sessions for repository %s: %v\n", repoRoot, err)
		}
	}
	
	return nil
}