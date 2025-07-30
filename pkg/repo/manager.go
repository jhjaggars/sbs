package repo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
)

type Repository struct {
	Name   string // Short name like "myproject"
	Root   string // Full path to repository root
	Remote string // Git remote URL if available
}

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

// DetectCurrentRepository detects the current git repository context
func (m *Manager) DetectCurrentRepository() (*Repository, error) {
	// Find git repository root
	repoRoot, err := m.findGitRoot()
	if err != nil {
		return nil, fmt.Errorf("not in a git repository: %w", err)
	}
	
	// Extract repository name
	repoName := m.extractRepositoryName(repoRoot)
	
	// Get remote URL if available
	remoteURL := m.getRemoteURL(repoRoot)
	
	return &Repository{
		Name:   repoName,
		Root:   repoRoot,
		Remote: remoteURL,
	}, nil
}

// GetSessionsPath returns the path to the repository-specific sessions file
func (r *Repository) GetSessionsPath() string {
	return filepath.Join(r.Root, ".work-orchestrator", "sessions.json")
}

// GetTmuxSessionName returns the repository-scoped tmux session name
func (r *Repository) GetTmuxSessionName(issueNumber int) string {
	return fmt.Sprintf("work-issue-%s-%d", r.Name, issueNumber)
}

// GetSandboxName returns the repository-scoped sandbox name
func (r *Repository) GetSandboxName(issueNumber int) string {
	return fmt.Sprintf("work-issue-%s-%d", r.Name, issueNumber)
}

// GetWorktreePath returns the repository-scoped worktree path
func (r *Repository) GetWorktreePath(issueNumber int) string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".work-issue-worktrees", r.Name, fmt.Sprintf("issue-%d", issueNumber))
}

// findGitRoot finds the root directory of the current git repository
func (m *Manager) findGitRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	
	// Use go-git to find repository root
	repo, err := git.PlainOpenWithOptions(currentDir, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return "", err
	}
	
	workTree, err := repo.Worktree()
	if err != nil {
		return "", err
	}
	
	return workTree.Filesystem.Root(), nil
}

// extractRepositoryName extracts a short repository name from the root path
func (m *Manager) extractRepositoryName(repoRoot string) string {
	// First try to get name from git remote
	if remoteName := m.extractNameFromRemote(repoRoot); remoteName != "" {
		return remoteName
	}
	
	// Fallback to directory name
	return filepath.Base(repoRoot)
}

// extractNameFromRemote extracts repository name from git remote URL
func (m *Manager) extractNameFromRemote(repoRoot string) string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	
	remoteURL := strings.TrimSpace(string(output))
	
	// Handle different URL formats
	// SSH: git@github.com:user/repo.git
	// HTTPS: https://github.com/user/repo.git
	
	// Extract repository name from URL
	patterns := []string{
		`[:/]([^/]+)/([^/]+?)(?:\.git)?$`, // Captures user/repo
		`/([^/]+?)(?:\.git)?$`,           // Just repo name
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(remoteURL)
		if len(matches) >= 2 {
			// Return the last capture group (repository name)
			return matches[len(matches)-1]
		}
	}
	
	return ""
}

// getRemoteURL gets the git remote URL
func (m *Manager) getRemoteURL(repoRoot string) string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	
	return strings.TrimSpace(string(output))
}

// IsGitRepository checks if the current directory is in a git repository
func (m *Manager) IsGitRepository() bool {
	_, err := m.DetectCurrentRepository()
	return err == nil
}

// SanitizeName sanitizes a repository name for use in file/session names
func (m *Manager) SanitizeName(name string) string {
	// Remove special characters and replace with hyphens
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	sanitized := reg.ReplaceAllString(name, "-")
	
	// Trim hyphens from start and end
	sanitized = strings.Trim(sanitized, "-")
	
	// Convert to lowercase
	sanitized = strings.ToLower(sanitized)
	
	// Limit length
	if len(sanitized) > 30 {
		sanitized = sanitized[:30]
		sanitized = strings.TrimSuffix(sanitized, "-")
	}
	
	return sanitized
}