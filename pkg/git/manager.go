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
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		return fmt.Errorf("failed to create worktree parent directory: %w", err)
	}

	// Check if worktree already exists
	if _, err := os.Stat(worktreePath); err == nil {
		// Worktree exists, verify it's valid
		if m.isValidWorktree(worktreePath) {
			return nil
		}
		// Remove invalid worktree
		os.RemoveAll(worktreePath)
	}

	// Use git command to create worktree (go-git doesn't support worktrees well)
	cmd := exec.Command("git", "worktree", "add", worktreePath, branchName)
	cmd.Dir = m.repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
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
	gitDir := filepath.Join(path, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
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
