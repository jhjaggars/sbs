package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"sbs/pkg/cmdlog"
)

type Manager struct {
	repoPath string
	repo     *git.Repository
}

func NewManager(repoPath string) (*Manager, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository at %s: %w", repoPath, err)
	}

	return &Manager{
		repoPath: repoPath,
		repo:     repo,
	}, nil
}

func (m *Manager) CreateIssueBranch(issueNumber int, issueTitle string) (string, error) {
	branchName := m.formatBranchName(issueNumber, issueTitle)

	// Check if branch already exists
	branches, err := m.repo.Branches()
	if err != nil {
		return "", fmt.Errorf("failed to list branches: %w", err)
	}

	var branchExists bool
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		if strings.HasSuffix(ref.Name().String(), branchName) {
			branchExists = true
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("error checking branches: %w", err)
	}

	if branchExists {
		return branchName, nil
	}

	// Get HEAD reference
	head, err := m.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Create new branch
	branchRef := plumbing.NewBranchReferenceName(branchName)
	ref := plumbing.NewHashReference(branchRef, head.Hash())

	err = m.repo.Storer.SetReference(ref)
	if err != nil {
		return "", fmt.Errorf("failed to create branch %s: %w", branchName, err)
	}

	return branchName, nil
}

// CreateBranchDirect creates a branch with the exact name provided
func (m *Manager) CreateBranchDirect(branchName string) error {
	// Check if branch already exists
	exists, err := m.BranchExists(branchName)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	if exists {
		return nil // Branch already exists
	}

	// Get HEAD reference
	head, err := m.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Create new branch
	branchRef := plumbing.NewBranchReferenceName(branchName)
	ref := plumbing.NewHashReference(branchRef, head.Hash())

	err = m.repo.Storer.SetReference(ref)
	if err != nil {
		return fmt.Errorf("failed to create branch %s: %w", branchName, err)
	}

	return nil
}

func (m *Manager) CreateWorktree(branchName string, worktreePath string) error {
	// Ensure worktree directory exists
	parentDir := filepath.Dir(worktreePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create worktree parent directory %s: %w", parentDir, err)
	}

	// Check if worktree already exists
	if _, err := os.Stat(worktreePath); err == nil {
		// Worktree exists, verify it's valid
		if m.isValidWorktree(worktreePath) {
			return nil
		}

		// Invalid worktree detected, remove it first
		if err := m.cleanupInvalidWorktree(worktreePath); err != nil {
			return fmt.Errorf("failed to cleanup invalid worktree at %s: %w", worktreePath, err)
		}
	}

	// Verify branch exists before creating worktree
	if !m.branchExists(branchName) {
		return fmt.Errorf("branch %s does not exist", branchName)
	}

	// Use git command to create worktree (go-git doesn't support worktrees well)
	cmd := exec.Command("git", "worktree", "add", worktreePath, branchName)
	cmd.Dir = m.repoPath

	// Capture both stdout and stderr for better error reporting
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create worktree at %s for branch %s: %w\nGit output: %s",
			worktreePath, branchName, err, string(output))
	}

	// Final validation that worktree was created successfully
	if !m.isValidWorktree(worktreePath) {
		return fmt.Errorf("worktree created but validation failed at %s", worktreePath)
	}

	return nil
}

func (m *Manager) RemoveWorktree(worktreePath string) error {
	// Remove worktree using git command
	cmd := exec.Command("git", "worktree", "remove", worktreePath, "--force")
	cmd.Dir = m.repoPath
	if err := cmd.Run(); err != nil {
		// If git command fails, try manual removal
		return os.RemoveAll(worktreePath)
	}
	return nil
}

func (m *Manager) ListWorktrees() ([]string, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = m.repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	var worktrees []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			if path != m.repoPath { // Skip main worktree
				worktrees = append(worktrees, path)
			}
		}
	}

	return worktrees, nil
}

func (m *Manager) isValidWorktree(path string) bool {
	// Check if the directory exists
	if _, err := os.Stat(path); err != nil {
		return false
	}

	// Check if .git file/directory exists
	gitPath := filepath.Join(path, ".git")
	gitStat, err := os.Stat(gitPath)
	if err != nil {
		return false
	}

	// For worktrees, .git is typically a file containing the path to the git directory
	if !gitStat.IsDir() {
		// Read .git file to verify it points to a valid git directory
		content, err := os.ReadFile(gitPath)
		if err != nil {
			return false
		}

		// Content should be like "gitdir: /path/to/repo/.git/worktrees/name"
		gitdirLine := strings.TrimSpace(string(content))
		if !strings.HasPrefix(gitdirLine, "gitdir: ") {
			return false
		}

		actualGitDir := strings.TrimPrefix(gitdirLine, "gitdir: ")
		if _, err := os.Stat(actualGitDir); err != nil {
			return false
		}
	}

	// Additional check: verify this worktree is registered with git
	return m.isWorktreeRegistered(path)
}

func (m *Manager) formatBranchName(issueNumber int, issueTitle string) string {
	// Create a slug from the issue title
	slug := strings.ToLower(issueTitle)

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	// Limit length
	if len(slug) > 50 {
		slug = slug[:50]
		slug = strings.TrimSuffix(slug, "-")
	}

	return fmt.Sprintf("issue-%d-%s", issueNumber, slug)
}

func (m *Manager) GetCurrentBranch() (string, error) {
	head, err := m.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	if !head.Name().IsBranch() {
		return "", fmt.Errorf("HEAD is not pointing to a branch")
	}

	return head.Name().Short(), nil
}

// branchExists checks if a branch exists in the repository
func (m *Manager) branchExists(branchName string) bool {
	if m.repo == nil {
		return false // Handle nil repository gracefully
	}

	branches, err := m.repo.Branches()
	if err != nil {
		return false
	}

	var exists bool
	branches.ForEach(func(ref *plumbing.Reference) error {
		if strings.HasSuffix(ref.Name().String(), branchName) {
			exists = true
		}
		return nil
	})

	return exists
}

// isWorktreeRegistered verifies that the worktree is registered with git
func (m *Manager) isWorktreeRegistered(worktreePath string) bool {
	worktrees, err := m.ListWorktrees()
	if err != nil {
		return false
	}

	for _, wt := range worktrees {
		if wt == worktreePath {
			return true
		}
	}

	return false
}

// cleanupInvalidWorktree removes an invalid worktree
func (m *Manager) cleanupInvalidWorktree(worktreePath string) error {
	// First try to remove via git worktree command
	cmd := exec.Command("git", "worktree", "remove", worktreePath, "--force")
	cmd.Dir = m.repoPath

	// Capture output for debugging
	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil // Successfully removed
	}

	// If git worktree remove failed, try manual cleanup
	// First remove the directory
	if err := os.RemoveAll(worktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree directory %s: %w (git output: %s)",
			worktreePath, err, string(output))
	}

	// Then try to prune stale worktree references
	pruneCmd := exec.Command("git", "worktree", "prune")
	pruneCmd.Dir = m.repoPath
	if err := pruneCmd.Run(); err != nil {
		// Prune failure is not critical, just log it
		return fmt.Errorf("worktree directory removed but failed to prune references: %w", err)
	}

	return nil
}

// runGitCommand executes a git command with logging in the repository directory
func (m *Manager) runGitCommand(args []string) ([]byte, error) {
	ctx := cmdlog.LogCommandGlobal("git", args, cmdlog.GetCaller())

	cmd := exec.Command("git", args...)
	cmd.Dir = m.repoPath
	start := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	if err != nil {
		ctx.LogCompletion(false, getExitCode(cmd), err.Error(), duration)
		return output, err
	}

	ctx.LogCompletion(true, 0, "", duration)
	return output, nil
}

// runGitCommandRun executes a git command without capturing output, with logging
func (m *Manager) runGitCommandRun(args []string) error {
	ctx := cmdlog.LogCommandGlobal("git", args, cmdlog.GetCaller())

	cmd := exec.Command("git", args...)
	cmd.Dir = m.repoPath
	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	if err != nil {
		ctx.LogCompletion(false, getExitCode(cmd), err.Error(), duration)
		return err
	}

	ctx.LogCompletion(true, 0, "", duration)
	return nil
}

// getExitCode extracts exit code from exec.Cmd
func getExitCode(cmd *exec.Cmd) int {
	if cmd.ProcessState != nil {
		return cmd.ProcessState.ExitCode()
	}
	return -1
}

// Branch cleanup methods for enhanced resource management

// BranchDeletionResult represents the result of a branch deletion operation
type BranchDeletionResult struct {
	BranchName string
	Success    bool
	Message    string
	Error      error
}

// DeleteIssueBranch deletes a single issue branch safely
func (m *Manager) DeleteIssueBranch(branchName string) error {
	// Validate branch exists
	if !m.branchExists(branchName) {
		// Not an error - branch doesn't exist, which is the desired state
		return nil
	}

	// Check if it's the current branch
	currentBranch, err := m.GetCurrentBranch()
	if err == nil && currentBranch == branchName {
		return fmt.Errorf("cannot delete current branch: %s", branchName)
	}

	// Try normal deletion first
	args := []string{"branch", "-d", branchName}
	_, err = m.runGitCommand(args)
	if err != nil {
		return fmt.Errorf("failed to delete branch %s (safe deletion attempt): %w", branchName, err)
	}
	return nil
}

// DeleteIssueBranchForce forcefully deletes a branch (even if unmerged)
func (m *Manager) DeleteIssueBranchForce(branchName string) error {
	// Validate branch exists
	if !m.branchExists(branchName) {
		return nil
	}

	// Check if it's the current branch
	currentBranch, err := m.GetCurrentBranch()
	if err == nil && currentBranch == branchName {
		return fmt.Errorf("cannot delete current branch: %s", branchName)
	}

	// Force deletion
	args := []string{"branch", "-D", branchName}
	_, err = m.runGitCommand(args)
	if err != nil {
		return fmt.Errorf("failed to delete branch %s (force deletion attempt): %w", branchName, err)
	}
	return nil
}

// DeleteCurrentBranch always returns an error as a safety measure
func (m *Manager) DeleteCurrentBranch() error {
	return fmt.Errorf("cannot delete current branch - switch to another branch first")
}

// ListIssueBranches returns all branches that match the issue-* pattern
func (m *Manager) ListIssueBranches() ([]string, error) {
	args := []string{"branch", "--list", "issue-*"}
	output, err := m.runGitCommand(args)
	if err != nil {
		return nil, fmt.Errorf("failed to list issue branches: %w", err)
	}

	var branches []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Remove leading asterisk and spaces (current branch marker)
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "issue-") {
			branches = append(branches, line)
		}
	}

	return branches, nil
}

// FindOrphanedIssueBranches finds issue branches that don't have active sessions.
// It compares all existing issue branches against the provided list of active session work item IDs
// and returns branches that don't correspond to any active session.
func (m *Manager) FindOrphanedIssueBranches(activeSessionWorkItems []string) ([]string, error) {
	allIssueBranches, err := m.ListIssueBranches()
	if err != nil {
		return nil, fmt.Errorf("failed to get issue branches: %w", err)
	}

	// Create map of active work items for quick lookup
	activeWorkItems := make(map[string]bool)
	for _, workItem := range activeSessionWorkItems {
		activeWorkItems[workItem] = true
	}

	var orphanedBranches []string
	for _, branch := range allIssueBranches {
		// Extract work item ID from branch name
		workItemID := m.extractWorkItemFromBranch(branch)
		if workItemID != "" && !activeWorkItems[workItemID] {
			orphanedBranches = append(orphanedBranches, branch)
		}
	}

	return orphanedBranches, nil
}

// GetBranchAge returns the age of a branch based on its last commit.
// The age is calculated as the time elapsed since the last commit on the branch.
// Returns an error if the branch doesn't exist or if the commit information cannot be retrieved.
func (m *Manager) GetBranchAge(branchName string) (time.Duration, error) {
	if !m.branchExists(branchName) {
		return 0, fmt.Errorf("branch %s does not exist", branchName)
	}

	// Get the last commit timestamp for the branch
	args := []string{"log", "-1", "--format=%ct", branchName}
	output, err := m.runGitCommand(args)
	if err != nil {
		return 0, fmt.Errorf("failed to get branch age: %w", err)
	}

	timestampStr := strings.TrimSpace(string(output))
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	commitTime := time.Unix(timestamp, 0)
	return time.Since(commitTime), nil
}

// ValidateBranchDeletion checks if a branch is safe to delete.
// Returns true if the branch can be safely deleted, false if there are concerns.
// The warnings slice contains human-readable messages about potential issues.
// This method checks for: branch existence, current branch status, and unmerged changes.
func (m *Manager) ValidateBranchDeletion(branchName string) (bool, []string, error) {
	var warnings []string

	// Handle case where repository is not initialized (for testing)
	if m.repo == nil {
		return true, warnings, nil // Treat as safe in test scenarios
	}

	// Check if branch exists
	if !m.branchExists(branchName) {
		return true, warnings, nil // Already gone, safe to "delete"
	}

	// Check if it's the current branch
	currentBranch, err := m.GetCurrentBranch()
	if err == nil && currentBranch == branchName {
		return false, append(warnings, "cannot delete current branch"), nil
	}

	// Check if branch has unmerged changes
	hasUnmerged, err := m.HasUnmergedChanges(branchName)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("could not check merge status: %v", err))
	} else if hasUnmerged {
		warnings = append(warnings, "branch has unmerged changes - use force delete if intended")
		return false, warnings, nil
	}

	return true, warnings, nil
}

// HasUnmergedChanges checks if a branch has commits not merged into the main branch
func (m *Manager) HasUnmergedChanges(branchName string) (bool, error) {
	if !m.branchExists(branchName) {
		return false, nil
	}

	// Check commits in branch that are not in main/master
	mainBranches := []string{"main", "master"}
	for _, mainBranch := range mainBranches {
		if m.branchExists(mainBranch) {
			args := []string{"log", fmt.Sprintf("%s..%s", mainBranch, branchName), "--oneline"}
			output, err := m.runGitCommand(args)
			if err != nil {
				continue // Try next main branch
			}

			// If there's output, there are unmerged commits
			if strings.TrimSpace(string(output)) != "" {
				return true, nil
			}
		}
	}

	return false, nil
}

// BranchExists is a public wrapper around the private branchExists method
func (m *Manager) BranchExists(branchName string) (bool, error) {
	return m.branchExists(branchName), nil
}

// DeleteMultipleBranches deletes multiple branches, with optional dry run
func (m *Manager) DeleteMultipleBranches(branchNames []string, dryRun bool) ([]BranchDeletionResult, error) {
	results := make([]BranchDeletionResult, 0, len(branchNames))

	for _, branchName := range branchNames {
		result := BranchDeletionResult{
			BranchName: branchName,
		}

		if dryRun {
			// Validate but don't actually delete
			safe, warnings, err := m.ValidateBranchDeletion(branchName)
			if err != nil {
				result.Success = false
				result.Error = err
				result.Message = fmt.Sprintf("validation failed: %v", err)
			} else if !safe {
				result.Success = false
				result.Message = fmt.Sprintf("would NOT delete (unsafe): %s", strings.Join(warnings, ", "))
			} else {
				result.Success = true
				result.Message = "would delete (safe)"
			}
		} else {
			// Actually delete the branch
			err := m.DeleteIssueBranch(branchName)
			if err != nil {
				result.Success = false
				result.Error = err
				result.Message = fmt.Sprintf("deletion failed: %v", err)
			} else {
				result.Success = true
				result.Message = "deleted successfully"
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// extractWorkItemFromBranch extracts the work item ID from a branch name
// Expects new format "issue-source-id-some-title" only
func (m *Manager) extractWorkItemFromBranch(branchName string) string {
	if !strings.HasPrefix(branchName, "issue-") {
		return ""
	}

	parts := strings.Split(branchName, "-")
	if len(parts) < 3 {
		return ""
	}

	// New format: issue-source-id-title
	source := parts[1]
	id := parts[2]
	return source + ":" + id
}

// extractIssueNumberFromBranch extracts the issue number from a branch name like "issue-123-some-title"
// This is kept for backward compatibility but should be phased out
func (m *Manager) extractIssueNumberFromBranch(branchName string) int {
	// With namespaced work items, we don't extract numeric issue numbers anymore
	// All work items are handled generically
	return 0
}
