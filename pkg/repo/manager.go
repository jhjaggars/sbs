package repo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/go-git/go-git/v5"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"sbs/pkg/cmdlog"
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
	output, err := m.runGitCommand(repoRoot, []string{"remote", "get-url", "origin"})
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
		`/([^/]+?)(?:\.git)?$`,            // Just repo name
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
	output, err := m.runGitCommand(repoRoot, []string{"remote", "get-url", "origin"})
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
func (m *Manager) SanitizeName(name string, maxLength ...int) string {
	// Handle empty input
	if name == "" {
		return ""
	}

	// First normalize Unicode characters and remove diacritics
	sanitized := m.normalizeUnicode(name)

	// Remove special characters and replace with hyphens
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	sanitized = reg.ReplaceAllString(sanitized, "-")

	// Trim hyphens from start and end
	sanitized = strings.Trim(sanitized, "-")

	// Convert to lowercase
	sanitized = strings.ToLower(sanitized)

	// Determine length limit
	lengthLimit := 30 // Default for backward compatibility
	if len(maxLength) > 0 && maxLength[0] > 0 {
		lengthLimit = maxLength[0]
	} else if len(maxLength) > 0 && maxLength[0] == 0 {
		// maxLength of 0 means use existing behavior (30 chars)
		lengthLimit = 30
	}

	// Apply length limit
	if len(sanitized) > lengthLimit {
		// Try to truncate at word boundary (hyphen)
		truncated := sanitized[:lengthLimit]

		// Find last hyphen to truncate at word boundary
		lastHyphen := strings.LastIndex(truncated, "-")
		if lastHyphen > 0 && lastHyphen < lengthLimit-1 {
			// Truncate at word boundary if there's a hyphen not at the very end
			sanitized = truncated[:lastHyphen]
		} else {
			// No good word boundary found, just truncate
			sanitized = truncated
		}

		// Ensure we don't end with a hyphen
		sanitized = strings.TrimSuffix(sanitized, "-")
	}

	return sanitized
}

// normalizeUnicode normalizes Unicode characters and removes diacritics
func (m *Manager) normalizeUnicode(input string) string {
	// Transform to remove diacritics and normalize
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, err := transform.String(t, input)
	if err != nil {
		// If transformation fails, fall back to original string
		return input
	}
	return result
}

// runGitCommand executes a git command with logging in a specific directory
func (m *Manager) runGitCommand(dir string, args []string) ([]byte, error) {
	ctx := cmdlog.LogCommandGlobal("git", args, cmdlog.GetCaller())

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	start := time.Now()
	output, err := cmd.Output()
	duration := time.Since(start)

	if err != nil {
		ctx.LogCompletion(false, getExitCode(cmd), err.Error(), duration)
		return output, err
	}

	ctx.LogCompletion(true, 0, "", duration)
	return output, nil
}

// getExitCode extracts exit code from exec.Cmd
func getExitCode(cmd *exec.Cmd) int {
	if cmd.ProcessState != nil {
		return cmd.ProcessState.ExitCode()
	}
	return -1
}
