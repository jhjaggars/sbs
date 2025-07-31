package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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
